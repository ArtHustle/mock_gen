// Package mock_gen generates Go mocks for interfaces declared in a source
// file. The generated mock uses unexported fields with public accessor
// methods, records every call to every method into a per-method log, and
// can optionally add a sync.Mutex (--thread-safe) and/or a dynamic return
// handler per method (--generate-return-handler).
//
// CLI entry point lives in cmd/mock_gen; this package exports the
// programmatic API (GenerateMockFileWithOptions, GenerateMockContentWithOptions)
// and the data types a custom text/template receives (MockData, etc.).
//
// See the project README for a feature overview and migration notes.
package mock_gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"slices"

	"golang.org/x/tools/go/packages"
)

//go:embed default.go.tmpl
var defaultTemplate string

// DefaultTemplate returns the built-in default template used for mock generation.
// This can be used as a starting point for creating custom templates.
func DefaultTemplate() string {
	return defaultTemplate
}

const (
	structMockSuffix = "Mock"
)

func deducePackagePath(filePath string) (string, error) {
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute file path: %w", err)
	}

	fileDir := filepath.Dir(absFilePath)

	modCmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	modCmd.Dir = fileDir
	var modOut bytes.Buffer
	modCmd.Stdout = &modOut
	err = modCmd.Run()
	if err != nil {
		return "", err
	}
	moduleDir := strings.TrimSpace(modOut.String())

	modulePathCmd := exec.Command("go", "list", "-m")
	modulePathCmd.Dir = fileDir
	var modulePathOut bytes.Buffer
	modulePathCmd.Stdout = &modulePathOut
	err = modulePathCmd.Run()
	if err != nil {
		return "", err
	}
	modulePath := strings.TrimSpace(modulePathOut.String())

	relativePath, err := filepath.Rel(moduleDir, absFilePath)
	if err != nil {
		return "", err
	}

	importPath := filepath.Join(modulePath, filepath.Dir(relativePath))
	importPath = strings.ReplaceAll(importPath, string(filepath.Separator), "/") // Ensure Unix-style slashes
	return importPath, nil
}

func (this *generator) collectInterfaceImports(
	iface *types.Interface,
	importPaths *[]*types.Package,
) {
	appendUnique := func(pkg *types.Package) {
		if slices.Contains(*importPaths, pkg) {
			return
		}
		*importPaths = append(*importPaths, pkg)
	}

	var collectImportsFromType func(t types.Type)
	collectImportsFromType = func(t types.Type) {
		switch typ := t.(type) {
		case *types.Named:
			if pkg := typ.Obj().Pkg(); pkg != nil {
				appendUnique(pkg)
			}
			if typ.TypeArgs() != nil {
				for i := range typ.TypeArgs().Len() {
					collectImportsFromType(typ.TypeArgs().At(i))
				}
			}
		case *types.Alias:
			if pkg := typ.Obj().Pkg(); pkg != nil {
				appendUnique(pkg)
			}
			if typ.TypeArgs() != nil {
				for i := range typ.TypeArgs().Len() {
					collectImportsFromType(typ.TypeArgs().At(i))
				}
			}
		case *types.Slice:
			collectImportsFromType(typ.Elem())
		case *types.Array:
			collectImportsFromType(typ.Elem())
		case *types.Map:
			collectImportsFromType(typ.Key())
			collectImportsFromType(typ.Elem())
		case *types.Pointer:
			collectImportsFromType(typ.Elem())
		case *types.Chan:
			collectImportsFromType(typ.Elem())
		case *types.Signature:
			for i := range typ.Params().Len() {
				collectImportsFromType(typ.Params().At(i).Type())
			}
			for i := range typ.Results().Len() {
				collectImportsFromType(typ.Results().At(i).Type())
			}
		case *types.Struct:
			for i := range typ.NumFields() {
				collectImportsFromType(typ.Field(i).Type())
			}
		case *types.Interface:
			for i := range typ.NumMethods() {
				collectImportsFromType(typ.Method(i).Type())
			}
		}
	}

	for i := range iface.NumMethods() {
		method := iface.Method(i)
		sig := method.Type().(*types.Signature)

		params := sig.Params()
		for j := range params.Len() {
			param := params.At(j)
			collectImportsFromType(param.Type())
		}

		results := sig.Results()
		for j := range results.Len() {
			result := results.At(j)
			collectImportsFromType(result.Type())
		}
	}
}

func toTitle(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[size:]
}

// exportedFieldName produces a Go-exported (capitalized) field name from a
// source identifier. Leading underscores are stripped first — they're a
// Go convention for "intentionally unused", which doesn't translate to the
// public surface of a recorded-call struct. If the input is all underscores
// or empty after stripping, a fallback prefix `X` is used so the result is
// always a valid exported identifier.
func exportedFieldName(s string) string {
	stripped := strings.TrimLeft(s, "_")
	if stripped == "" {
		return "X"
	}
	return toTitle(stripped)
}

func assignAliasesToImportPaths(importPackages []*types.Package, currentPkgPath string) map[string]*types.Package {
	aliases := make(map[string]*types.Package)

	for _, pkg := range importPackages {
		parts := strings.Split(pkg.Path(), "/")
		alias := ""

		for i := len(parts) - 1; i >= 0; i-- {
			alias = parts[i] + toTitle(alias)
			if existing, exists := aliases[alias]; !exists {
				break
			} else if strings.HasPrefix(pkg.Path(), currentPkgPath) {
				aliases[alias] = pkg
				pkg = existing
				parts = strings.Split(pkg.Path(), "/")
				i = len(parts)
				alias = ""
			}
		}

		aliases[alias] = pkg
	}

	return aliases
}

// Options configures mock generation. The zero value produces the default
// method-based shape: unexported fields, accessor methods, per-method call
// recording, no mutex.
type Options struct {
	// GenerateReturnHandler adds a per-method `<method>_return` func field
	// that, when non-nil, is invoked instead of returning the static
	// `<method>_return_X` fields. Lets tests inject dynamic return values
	// and observation hooks.
	GenerateReturnHandler bool

	// MakeThreadSafe adds a sync.Mutex to each mock and guards every
	// generated method, getter, and setter with it. Methods with return
	// values use the snapshot pattern: lock, mutate, snapshot returns into
	// locals, unlock, then return / invoke handler — so the lock is never
	// held while user code runs.
	MakeThreadSafe bool

	// AddHeaderNote prepends "// Code generated by MockGen. DO NOT EDIT."
	// to the generated file.
	AddHeaderNote bool

	// CustomTemplate, when non-empty, replaces the built-in text/template.
	// It receives a MockData. Output is run through gofmt and must be valid
	// Go. See DefaultTemplate() for the built-in as a starting point.
	CustomTemplate string
}

// MockData is the top-level value passed to the template. Custom templates
// branch on Options to render correctly under different flag combinations.
type MockData struct {
	// HeaderNote is "// Code generated by MockGen. DO NOT EDIT.\n\n" when
	// Options.AddHeaderNote is set, else "".
	HeaderNote string
	// Imports holds the package imports needed by the generated mock,
	// grouped and sorted into standard-library vs third-party.
	Imports ImportsData
	// Interfaces is one entry per generated mock, in source order.
	Interfaces []InterfaceData
	// Options is the Options the generator was invoked with.
	Options Options
}

// ImportsData groups discovered package imports.
type ImportsData struct {
	// StdImports are stdlib package import lines (e.g. `"context"`).
	StdImports []string
	// NonStdImports are third-party/local package import lines, possibly
	// with an alias prefix when needed for disambiguation.
	NonStdImports []string
}

// InterfaceData describes a single source interface and the mock that
// implements it.
type InterfaceData struct {
	// Name is the original interface name (e.g. "Service").
	Name string
	// MockName is the generated mock type name, "<Name>Mock"
	// (e.g. "ServiceMock"). Also used to prefix per-method Call struct
	// names so two mocks in the same file never collide.
	MockName string
	// Methods is one entry per method on the interface (including methods
	// reached through embedding). Order is stable across runs.
	Methods []MethodData
}

// MethodData describes one method on a generated mock. Everything the
// generator needs to render the method body, the recorded-call struct, the
// log slice, and the accessor methods lives here.
type MethodData struct {
	// Name is the original method name (e.g. "Send").
	Name string
	// CallStructName is the exported, per-method record-struct type name,
	// "<MockName>_<Name>_Call" (e.g. "ServiceMock_Send_Call"). Always set.
	CallStructName string
	// CallStructFields are the public, title-cased fields on the recorded
	// struct, one per param, in declaration order. Empty for no-arg
	// methods (the struct is then `struct{}`).
	CallStructFields []FieldData
	// CallsLogField is the unexported `[]<CallStructName>` slice field
	// name on the mock struct (e.g. "send_calls_log").
	CallsLogField string
	// CallByIndexName is the public indexer method name
	// (e.g. "Send_call"). The generated method takes one int and supports
	// negative indices (offset from the end).
	CallByIndexName string
	// ReturnHandler is non-nil when Options.GenerateReturnHandler is set
	// and the method has at least one return value.
	ReturnHandler *ReturnHandlerData
	// Returns are the unexported per-return mock fields (e.g.
	// `send_return_error`). One per method return value.
	Returns []FieldData
	// Signature is everything needed to render the method declaration and
	// its body's return / pass-through code.
	Signature SignatureData
	// Getters describes the accessor methods to generate. Always non-nil.
	Getters *GetterData
}

// FieldData is a generic `Name Type` pair used for struct fields and
// per-param/per-return records.
type FieldData struct {
	Name string
	Type string
}

// ReturnHandlerData describes the per-method `func(...)` handler field
// emitted under Options.GenerateReturnHandler.
type ReturnHandlerData struct {
	// FieldName is the unexported handler field name (e.g. "send_return").
	FieldName string
	// ParamTypes is a comma-separated list of param types for the handler
	// signature, e.g. "context.Context, *http.Request".
	ParamTypes string
	// ReturnTypes is a comma-separated list of return types, e.g.
	// "*http.Response, error".
	ReturnTypes string
}

// SignatureData carries the parts of a method signature needed to render
// the wrapper body.
type SignatureData struct {
	// Params is the method's parameter list as it appears in the source
	// signature (e.g. "ctx context.Context, r *http.Request").
	Params string
	// Returns is the method's return clause, leading space and parens
	// included when needed (e.g. " (*http.Response, error)" or " error").
	Returns string
	// ParamNames is the param names used in the method body, one per
	// param, in declaration order. Anonymous params get generated names
	// (`param` or `param1`/`param2`/...).
	ParamNames []string
	// ReturnFields is the qualified field reference for each return,
	// e.g. ["this.send_return_response", "this.send_return_error"].
	ReturnFields []string
}

// GetterData describes the accessor methods generated for a single
// method's recorded args and configured returns.
type GetterData struct {
	// CallsGetter is `<Name>_calls() []<CallStructName>` — returns the
	// full recorded log. Defensive copy under --thread-safe.
	CallsGetter GetterMethod
	// CallByIndex is `<Name>_call(i int) <CallStructName>` — returns one
	// entry. Supports negative indices (offset from end). Panics
	// out-of-range. Always non-nil.
	CallByIndex *GetterMethod
	// CallCount is `<Name>_call_count() int` — sugared `len(_calls())`
	// that avoids the defensive slice copy. Always non-nil.
	CallCount *GetterMethod
	// ReturnSetters are setters for the static per-return mock fields,
	// one per return value (e.g. `Send_return_error(value error)`).
	ReturnSetters []SetterMethod
	// HandlerSetter is the setter for the dynamic handler. Non-nil when
	// Options.GenerateReturnHandler is set AND the method has returns.
	HandlerSetter *SetterMethod
}

// GetterMethod describes a generated read accessor.
type GetterMethod struct {
	// MethodName is the exported method name (e.g. "Send_calls").
	MethodName string
	// FieldName is the unexported underlying field the method reads.
	FieldName string
	// ReturnType is the rendered return type.
	ReturnType string
}

// SetterMethod describes a generated write accessor.
type SetterMethod struct {
	// MethodName is the exported method name (e.g. "Send_return_error").
	MethodName string
	// FieldName is the unexported underlying field the method writes.
	FieldName string
	// ParamType is the rendered parameter type.
	ParamType string
}

type generator struct {
	currentPkg *types.Package
	astFile    *ast.File
	typeDefs   map[*ast.Ident]types.Object

	importAliasToPath map[string]*types.Package
	importPathToAlias map[string]string

	options Options
}

func generateMockFromAST(
	astFile *ast.File,
	pkg *types.Package,
	typeDefs map[*ast.Ident]types.Object,
	opts Options,
) (string, error) {
	gen := generator{
		currentPkg:        pkg,
		astFile:           astFile,
		typeDefs:          typeDefs,
		importAliasToPath: make(map[string]*types.Package),
		importPathToAlias: make(map[string]string),
		options:           opts,
	}

	mockData, err := gen.collectMockData()
	if err != nil {
		return "", err
	}

	templateStr := defaultTemplate
	if opts.CustomTemplate != "" {
		templateStr = opts.CustomTemplate
	}

	tmpl, err := template.New("mock").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mockData)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	formattedBytes, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("error formatting generated code: %w", err)
	}

	return string(formattedBytes), nil
}

type genRec struct {
	name  string
	iface *types.Interface
}

func (this *generator) collectMockData() (*MockData, error) {
	genRecs := this.collectTypesToGenerate()

	imports := this.collectImports(genRecs)
	sortPackagesByPathAndGroup(imports)

	mockData := &MockData{
		Options: this.options,
	}

	if this.options.AddHeaderNote {
		mockData.HeaderNote = "// Code generated by MockGen. DO NOT EDIT.\n\n"
	}

	mockData.Imports = this.collectImportsData(imports, len(genRecs) > 0)

	for _, rec := range genRecs {
		ifaceData, err := this.collectInterfaceData(rec.name, rec.iface)
		if err != nil {
			return nil, err
		}
		mockData.Interfaces = append(mockData.Interfaces, ifaceData)
	}

	return mockData, nil
}

func (this *generator) collectTypesToGenerate() []genRec {
	var genRecs []genRec

	for _, decl := range this.astFile.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Type == nil {
				continue
			}
			def := findTypeDef(this.typeDefs, typeSpec.Name.Name)
			if def == nil {
				continue
			}
			if iface, ok := def.Type().Underlying().(*types.Interface); ok {
				genRecs = append(genRecs, genRec{
					name:  typeSpec.Name.Name,
					iface: iface,
				})
			}
		}
	}

	return genRecs
}

func (this *generator) collectImports(genRecs []genRec) []*types.Package {
	importedPaths := []*types.Package{}

	for _, rec := range genRecs {
		this.collectInterfaceImports(rec.iface, &importedPaths)
	}

	this.importAliasToPath = assignAliasesToImportPaths(importedPaths, this.currentPkg.Path())
	for alias, pkg := range this.importAliasToPath {
		this.importPathToAlias[pkg.Path()] = alias
	}
	return importedPaths
}

func isBuiltinPackage(pkgPath string) bool {
	// Standard library packages don't have a dot in their path and are not mock_gen
	return !strings.Contains(pkgPath, ".") && pkgPath != "mock_gen"
}

func sortPackagesByPathAndGroup(packages []*types.Package) {
	// First, separate standard and non-standard packages
	var stdPkgs, nonStdPkgs []*types.Package
	for _, pkg := range packages {
		if isBuiltinPackage(pkg.Path()) {
			stdPkgs = append(stdPkgs, pkg)
		} else {
			nonStdPkgs = append(nonStdPkgs, pkg)
		}
	}

	// Sort each group
	sort.Slice(stdPkgs, func(i, j int) bool {
		return stdPkgs[i].Path() < stdPkgs[j].Path()
	})
	sort.Slice(nonStdPkgs, func(i, j int) bool {
		return nonStdPkgs[i].Path() < nonStdPkgs[j].Path()
	})

	// Combine them back with standard packages first
	copy(packages, append(stdPkgs, nonStdPkgs...))
}

// GenerateMockContentWithOptions generates a mock from Go source given as a
// string. Useful for tests and embedded use. The source is type-checked as a
// throwaway package named "mock_gen" — same-package type references will
// render with that package alias. For a real source file inside a Go module,
// use GenerateMockFileWithOptions instead.
func GenerateMockContentWithOptions(source string, opts Options) (string, error) {
	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, "", source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("error parsing source: %w", err)
	}

	myImporter := importer.ForCompiler(fileSet, "source", nil)

	conf := types.Config{
		Importer: myImporter,
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	currentPkg, err := conf.Check("mock_gen", fileSet, []*ast.File{astFile}, info)
	if err != nil {
		return "", fmt.Errorf("error type-checking source: %w", err)
	}

	return generateMockFromAST(astFile, currentPkg, info.Defs, opts)
}

// GenerateMockFile generates a mock from a Go source file using default
// Options. Shorthand for GenerateMockFileWithOptions(filePath, Options{}).
func GenerateMockFile(filePath string) (string, error) {
	return GenerateMockFileWithOptions(filePath, Options{})
}

// GenerateMockFileWithOptions generates a mock from a Go source file. The
// file must be inside a Go module (its containing directory tree must hold
// a go.mod). Types and imports are resolved relative to that module, so the
// generator can be invoked from any working directory.
func GenerateMockFileWithOptions(filePath string, opts Options) (string, error) {
	packagePath, err := deducePackagePath(filePath)
	if err != nil {
		return "", fmt.Errorf("error deducing package path: %w", err)
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute file path: %w", err)
	}
	fileDir := filepath.Dir(absFilePath)

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("error parsing file: %w", err)
	}

	pkg, err := loadPackage(packagePath, fileDir)
	if err != nil {
		return "", fmt.Errorf("error loading package: %w", err)
	}

	return generateMockFromAST(astFile, pkg.Types, pkg.TypesInfo.Defs, opts)
}

func findTypeDef(defs map[*ast.Ident]types.Object, name string) types.Object {
	for ident, def := range defs {
		if ident.Name == name {
			if _, ok := def.(*types.TypeName); ok {
				return def
			}
		}
	}
	return nil
}

func loadPackage(pattern string, dir string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for pattern %s", pattern)
	}
	pkg := pkgs[0]
	if pkg.TypesInfo == nil {
		var errMsgs []string
		for _, e := range pkg.Errors {
			errMsgs = append(errMsgs, e.Msg)
		}
		return nil, fmt.Errorf("package %s loaded without type info (errors: %v)", pattern, errMsgs)
	}
	return pkg, nil
}

func (this *generator) collectImportsData(importPackages []*types.Package, hasInterfaces bool) ImportsData {
	var stdImports, nonStdImports []string

	if this.options.MakeThreadSafe {
		stdImports = append(stdImports, strconv.Quote("sync"))
	}
	// callIndex (emitted per mock by the default template) uses strconv.Itoa
	// for panic messages. Custom templates manage their own imports.
	if hasInterfaces && this.options.CustomTemplate == "" {
		stdImports = append(stdImports, strconv.Quote("strconv"))
	}

	for _, pkg := range importPackages {
		impPath := pkg.Path()
		alias, hasAlias := this.importPathToAlias[impPath]

		// Emit an explicit `alias "path"` whenever the alias used in type
		// references differs from the package's actual name (which is what
		// Go would default to without an alias). This catches:
		//   - collision-resolved aliases (e.g. bModel "path/b/model")
		//   - same-package types where the module path's last segment
		//     differs from the package name (issue #20)
		useAlias := true
		if hasAlias && alias == pkg.Name() {
			useAlias = false
		}

		impStr := strconv.Quote(impPath)
		if useAlias {
			if alias != "" {
				impStr = alias + " " + impStr
			} else if pkg.Name() != "" {
				impStr = pkg.Name() + " " + impStr
			}
		}

		if isBuiltinPackage(impPath) {
			stdImports = append(stdImports, impStr)
		} else {
			nonStdImports = append(nonStdImports, impStr)
		}
	}

	sort.Strings(stdImports)
	sort.Strings(nonStdImports)

	return ImportsData{
		StdImports:    stdImports,
		NonStdImports: nonStdImports,
	}
}

func (this *generator) getFullName(obj types.Object, typeArgs *types.TypeList) string {
	typeSB := strings.Builder{}
	pkg := obj.Pkg()
	var pkgAlias string
	if pkg != nil {
		pkgAlias = this.importPathToAlias[pkg.Path()]
	}
	if pkgAlias != "" {
		typeSB.WriteString(pkgAlias + "." + obj.Name())
	} else {
		typeSB.WriteString(obj.Name())
	}

	if typeArgs != nil && typeArgs.Len() > 0 {
		typeSB.WriteByte('[')
		for i := range typeArgs.Len() {
			if i > 0 {
				typeSB.WriteString(", ")
			}
			argType := typeArgs.At(i)
			typeSB.WriteString(this.getFieldType(argType))
		}
		typeSB.WriteByte(']')
	}

	return typeSB.String()
}

func (this *generator) getTypeFullName(namedType *types.Named) string {
	return this.getFullName(namedType.Obj(), namedType.TypeArgs())
}

func (this *generator) getAliasFullName(namedType *types.Alias) string {
	return this.getFullName(namedType.Obj(), namedType.TypeArgs())
}

func (this *generator) collectInterfaceData(name string, iface *types.Interface) (InterfaceData, error) {
	ifaceData := InterfaceData{
		Name:     name,
		MockName: name + structMockSuffix,
		Methods:  []MethodData{},
	}

	for i := range iface.NumMethods() {
		method := iface.Method(i)
		methodData := this.collectMethodData(method, ifaceData.MockName)
		ifaceData.Methods = append(ifaceData.Methods, methodData)
	}

	if err := validateNoNameCollisions(name, ifaceData.Methods); err != nil {
		return InterfaceData{}, err
	}

	return ifaceData, nil
}

// validateNoNameCollisions checks that no method on the interface has a name
// that would collide with a generator-emitted method on the mock. Generated
// names per source method `X`: X_calls, X_call, X_call_count, X_return,
// X_return_<each return>. Generated mock-level names: Reset_history.
//
// If a collision is detected, returns an error pointing at the offending
// user-defined method so the user can rename it.
func validateNoNameCollisions(interfaceName string, methods []MethodData) error {
	// Build the set of names the generator will emit for OTHER methods.
	generated := map[string]string{} // generated name -> owning source method
	generated["Reset_history"] = "<mock-level reset>"
	for _, m := range methods {
		generated[m.Name+"_calls"] = m.Name
		generated[m.Name+"_call"] = m.Name
		generated[m.Name+"_call_count"] = m.Name
		if m.ReturnHandler != nil {
			generated[m.Name+"_return"] = m.Name
		}
		for _, r := range m.Getters.ReturnSetters {
			generated[r.MethodName] = m.Name
		}
	}
	// Each user method must not match one of the generated names except its
	// own (e.g. method Foo doesn't conflict with its own Foo_calls generator;
	// only with the generator names of *other* methods).
	for _, m := range methods {
		owner, conflicts := generated[m.Name]
		if !conflicts {
			continue
		}
		if owner == m.Name {
			// Shouldn't happen — m.Name itself isn't one of the suffixed names.
			continue
		}
		return fmt.Errorf(
			"interface %s declares method %q which collides with a generator-emitted method "+
				"(would clash with the accessor for source method %q). "+
				"Rename %q to avoid the conflict.",
			interfaceName, m.Name, owner, m.Name,
		)
	}
	return nil
}

func (this *generator) collectMethodData(method *types.Func, mockName string) MethodData {
	sig := method.Type().(*types.Signature)
	methodLower := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]

	methodData := MethodData{
		Name:             method.Name(),
		CallStructName:   mockName + "_" + method.Name() + "_Call",
		CallsLogField:    methodLower + "_calls_log",
		CallByIndexName:  method.Name() + "_call",
		CallStructFields: []FieldData{},
		Returns:          []FieldData{},
	}

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		rawFieldName := getParamMember(method, j)
		suffix := rawFieldName[len(method.Name()+"_param_"):]
		methodData.CallStructFields = append(methodData.CallStructFields, FieldData{
			Name: exportedFieldName(suffix),
			Type: this.getFieldType(param.Type()),
		})
	}

	if this.options.GenerateReturnHandler && sig.Results().Len() > 0 {
		paramTypes, returnTypes := this.handlerSignatureTypes(sig)
		methodData.ReturnHandler = &ReturnHandlerData{
			FieldName:   methodLower + "_return",
			ParamTypes:  paramTypes,
			ReturnTypes: returnTypes,
		}
	}

	for j := range sig.Results().Len() {
		result := sig.Results().At(j)
		rawFieldName := getReturnMember(method, j)
		methodData.Returns = append(methodData.Returns, FieldData{
			Name: methodLower + "_return_" + rawFieldName[len(method.Name()+"_return_"):],
			Type: this.getFieldType(result.Type()),
		})
	}

	methodData.Signature = this.collectSignatureData(method, sig)
	methodData.Getters = this.collectGetterData(method, sig, methodData.CallStructName, methodData.CallsLogField)

	return methodData
}

func (this *generator) collectSignatureData(method *types.Func, sig *types.Signature) SignatureData {
	var params, returns strings.Builder
	paramNames := []string{}
	returnFields := []string{}

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		if j > 0 {
			params.WriteString(", ")
		}

		paramName := getParamName(sig, j)
		paramNames = append(paramNames, paramName)

		var parType string
		if sig.Variadic() && j == sig.Params().Len()-1 {
			parType = " ..." + this.getFieldType(param.Type().(*types.Slice).Elem())
		} else if j >= sig.Params().Len()-1 || sig.Params().At(j+1).Type().String() != param.Type().String() {
			parType = " " + this.getFieldType(param.Type())
		}
		params.WriteString(paramName + parType)
	}

	if sig.Results().Len() > 0 {
		returns.WriteByte(' ')
		useParens := sig.Results().Len() > 1 || sig.Results().At(0).Name() != ""
		if useParens {
			returns.WriteByte('(')
		}
		for j := range sig.Results().Len() {
			result := sig.Results().At(j)
			if j > 0 {
				returns.WriteString(", ")
			}
			pref := result.Name()
			if pref != "" {
				pref += " "
			}
			returns.WriteString(pref + this.getFieldType(result.Type()))

			returnField := getReturnMember(method, j)
			methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
			returnField = methodName + "_return_" + returnField[len(method.Name()+"_return_"):]
			returnFields = append(returnFields, "this."+returnField)
		}
		if useParens {
			returns.WriteString(")")
		}
	}

	return SignatureData{
		Params:       params.String(),
		Returns:      returns.String(),
		ParamNames:   paramNames,
		ReturnFields: returnFields,
	}
}

func (this *generator) collectGetterData(method *types.Func, sig *types.Signature, callStructName, callsLogField string) *GetterData {
	getterData := &GetterData{
		ReturnSetters: []SetterMethod{},
	}

	getterData.CallsGetter = GetterMethod{
		MethodName: method.Name() + "_calls",
		FieldName:  callsLogField,
		ReturnType: "[]" + callStructName,
	}
	getterData.CallByIndex = &GetterMethod{
		MethodName: method.Name() + "_call",
		FieldName:  callsLogField,
		ReturnType: callStructName,
	}
	getterData.CallCount = &GetterMethod{
		MethodName: method.Name() + "_call_count",
		FieldName:  callsLogField,
		ReturnType: "int",
	}

	for j := range sig.Results().Len() {
		result := sig.Results().At(j)
		fieldName := getReturnMember(method, j)
		methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
		fieldNameLower := methodName + "_return_" + fieldName[len(method.Name()+"_return_"):]
		resType := this.getFieldType(result.Type())

		getterData.ReturnSetters = append(getterData.ReturnSetters, SetterMethod{
			MethodName: fieldName,
			FieldName:  fieldNameLower,
			ParamType:  resType,
		})
	}

	if this.options.GenerateReturnHandler && sig.Results().Len() > 0 {
		methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
		paramTypes, returnTypes := this.handlerSignatureTypes(sig)
		funcType := "func(" + paramTypes + ") "
		if sig.Results().Len() > 1 {
			funcType += "(" + returnTypes + ")"
		} else {
			funcType += returnTypes
		}
		getterData.HandlerSetter = &SetterMethod{
			MethodName: method.Name() + "_return",
			FieldName:  methodName + "_return",
			ParamType:  funcType,
		}
	}

	return getterData
}

func (this *generator) handlerSignatureTypes(sig *types.Signature) (string, string) {
	var paramTypes, returnTypes strings.Builder
	for j := range sig.Params().Len() {
		if j > 0 {
			paramTypes.WriteString(", ")
		}
		paramTypes.WriteString(this.getFieldType(sig.Params().At(j).Type()))
	}
	for j := range sig.Results().Len() {
		if j > 0 {
			returnTypes.WriteString(", ")
		}
		returnTypes.WriteString(this.getFieldType(sig.Results().At(j).Type()))
	}
	return paramTypes.String(), returnTypes.String()
}

func getReturnMember(funcObj *types.Func, returnIndex int) string {
	sig := funcObj.Type().(*types.Signature)
	results := sig.Results()
	if results.Len() <= returnIndex {
		return ""
	}

	result := results.At(returnIndex)
	name := result.Name()
	if name == "" {
		name = getNameOfType(result.Type())
		memInd := getMemberIndex(sig.Results(), returnIndex)
		if memInd != 0 {
			name += strconv.Itoa(memInd)
		}
	}

	return funcObj.Name() + "_return_" + name
}

func isValidRune(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

func camelCaseComponents(str string) string {
	var sb strings.Builder
	isSplit := false
	for _, ch := range str {
		if isValidRune(ch) {
			if isSplit {
				sb.WriteRune(unicode.ToUpper(ch))
				isSplit = false
			} else {
				sb.WriteRune(ch)
			}
		} else {
			isSplit = true
		}
	}
	return sb.String()
}

func getNameOfType(t types.Type) string {
	typeStr := t.String()
	isArr := strings.HasPrefix(typeStr, "[]") || strings.HasPrefix(typeStr, "...")
	typeStr = removePackages(typeStr)
	typeStr = camelCaseComponents(typeStr)
	name := strings.ToLower(string(typeStr[0])) + typeStr[1:]
	if isArr {
		name += "Arr"
	}
	return name
}

// turns string "op.Op[pkg.Gen[pkg.T]]" into "Op[Gen[T]]"
func removePackages(str string) string {
	var sb strings.Builder
	startPos := 0
	for {
		dotInd := strings.Index(str[startPos:], ".")
		if dotInd == -1 {
			sb.WriteString(str[startPos:])
			break
		}
		dotInd += startPos
		i := dotInd - 1
		for ; i >= 0 && (isValidRune(rune(str[i])) || str[i] == '/'); i-- {
		}
		sb.WriteString(str[startPos : i+1])
		startPos = dotInd + 1
	}
	return sb.String()
}

func getParamMember(funcObj *types.Func, paramIndex int) string {
	sig := funcObj.Type().(*types.Signature)
	params := sig.Params()
	if params.Len() <= paramIndex {
		return ""
	}

	param := params.At(paramIndex)
	name := param.Name()
	if name == "" {
		name = getNameOfType(param.Type())
	}

	memInd := getMemberIndex(sig.Params(), paramIndex)
	if memInd != 0 {
		name += strconv.Itoa(memInd)
	}

	return funcObj.Name() + "_param_" + name
}

func getParamName(sig *types.Signature, paramIndex int) string {
	params := sig.Params()
	if params.Len() <= paramIndex {
		return ""
	}

	param := params.At(paramIndex)
	if param.Name() != "" {
		return param.Name()
	} else if params.Len() == 1 {
		return "param"
	} else {
		return "param" + strconv.Itoa(paramIndex+1)
	}
}

func getMemberIndex(tuple *types.Tuple, returnIndex int) int {
	if tuple == nil || tuple.Len() < 2 {
		return 0
	}

	for i := range tuple.Len() {
		if tuple.At(i).Name() != "" {
			return 0
		}
	}

	t := tuple.At(returnIndex).Type()
	typeName := t.String()
	if namedType, isNamed := t.(*types.Named); isNamed {
		typeName = namedType.Obj().Name()
	}

	currentIndex := 1
	totalSeen := 1
	for i := range tuple.Len() {
		if i == returnIndex {
			continue
		}
		curType := tuple.At(i).Type()
		curTypeName := curType.String()
		if namedType, isNamed := curType.(*types.Named); isNamed {
			curTypeName = namedType.Obj().Name()
		}
		if typeName == curTypeName {
			if i < returnIndex {
				currentIndex++
			}
			totalSeen++
		}
	}

	if totalSeen > 1 {
		return currentIndex
	}

	return 0
}

func (this *generator) getFieldType(t types.Type) string {
	switch typ := t.(type) {
	case *types.Basic:
		return typ.Name()
	case *types.Array:
		return "[" + strconv.Itoa(int(typ.Len())) + "]" + this.getFieldType(typ.Elem())
	case *types.Slice:
		return "[]" + this.getFieldType(typ.Elem())
	case *types.Pointer:
		return "*" + this.getFieldType(typ.Elem())
	case *types.Map:
		return "map[" + this.getFieldType(typ.Key()) + "]" + this.getFieldType(typ.Elem())
	case *types.Chan:
		switch typ.Dir() {
		case types.SendOnly:
			return "chan<- " + this.getFieldType(typ.Elem())
		case types.RecvOnly:
			return "<-chan " + this.getFieldType(typ.Elem())
		default:
			return "chan " + this.getFieldType(typ.Elem())
		}
	case *types.Alias:
		return this.getAliasFullName(typ)
	case *types.Named:
		return this.getTypeFullName(typ)
	case *types.Signature:
		return this.renderSignature(typ)
	case *types.Struct:
		return this.renderStruct(typ)
	case *types.Interface:
		return this.renderInterface(typ)
	default:
		return ""
	}
}

// renderStruct renders a *types.Struct as Go source: "struct{}" when empty,
// otherwise "struct{ Field1 T1; Field2 T2; ... }". Embedded fields and tags
// are not specially handled — they render as their canonical names and types.
func (this *generator) renderStruct(s *types.Struct) string {
	if s.NumFields() == 0 {
		return "struct{}"
	}
	var sb strings.Builder
	sb.WriteString("struct{ ")
	for i := range s.NumFields() {
		if i > 0 {
			sb.WriteString("; ")
		}
		f := s.Field(i)
		sb.WriteString(f.Name())
		sb.WriteByte(' ')
		sb.WriteString(this.getFieldType(f.Type()))
	}
	sb.WriteString(" }")
	return sb.String()
}

// renderInterface renders a *types.Interface as Go source: "any" when the
// interface is empty (no methods, no embedded constraints), otherwise
// "interface{ Method1(...) ...; ... }".
func (this *generator) renderInterface(iface *types.Interface) string {
	if iface.Empty() {
		return "any"
	}
	var sb strings.Builder
	sb.WriteString("interface{ ")
	for i := range iface.NumMethods() {
		if i > 0 {
			sb.WriteString("; ")
		}
		m := iface.Method(i)
		sb.WriteString(m.Name())
		// m.Type() is a *types.Signature; renderSignature emits "func(...)..."
		// but for interface methods we want "Name(...) ..." (no "func").
		// Drop the leading "func" prefix.
		sig := this.renderSignature(m.Type().(*types.Signature))
		sb.WriteString(sig[len("func"):])
	}
	sb.WriteString(" }")
	return sb.String()
}

// renderSignature renders a *types.Signature as Go source: e.g.
// "func(int) error", "func() (int, error)". Variadic last params render
// with "..." prefix. Result names are intentionally dropped — they are
// not part of the type identity.
func (this *generator) renderSignature(sig *types.Signature) string {
	var sb strings.Builder
	sb.WriteString("func(")
	for i := range sig.Params().Len() {
		if i > 0 {
			sb.WriteString(", ")
		}
		p := sig.Params().At(i)
		if sig.Variadic() && i == sig.Params().Len()-1 {
			sb.WriteString("..." + this.getFieldType(p.Type().(*types.Slice).Elem()))
		} else {
			sb.WriteString(this.getFieldType(p.Type()))
		}
	}
	sb.WriteByte(')')
	switch sig.Results().Len() {
	case 0:
	case 1:
		sb.WriteByte(' ')
		sb.WriteString(this.getFieldType(sig.Results().At(0).Type()))
	default:
		sb.WriteString(" (")
		for i := range sig.Results().Len() {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(this.getFieldType(sig.Results().At(i).Type()))
		}
		sb.WriteByte(')')
	}
	return sb.String()
}
