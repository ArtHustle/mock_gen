// Package mock_gen generates Go mocks for interfaces declared in a source
// file. The generated mock uses unexported fields with public accessor
// methods, records every call to every method into a per-method log, and
// can optionally add a sync.Mutex (--async) and/or a dynamic return
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

// unexportedTypeName lowercases the first rune of name. Used for the
// inspect/config namespace types (e.g. `serviceMockCalls`,
// `serviceMockSendOn`) which are unexported because users only reach them
// via method chains (mock.Calls().Send().Count() etc.) and never need to
// name them at a call site.
func unexportedTypeName(name string) string {
	if name == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(name)
	return string(unicode.ToLower(r)) + name[size:]
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
// namespaced shape: Calls() returns the per-method inspection namespace,
// On() returns the per-method configuration namespace.
type Options struct {
	// Async makes the mock safe for use from goroutines:
	//   - A sync.Mutex guards all generated state.
	//   - Methods with returns use the snapshot pattern (lock, mutate,
	//     snapshot returns, unlock, run handler / return) so user code
	//     never runs under the lock. A second lock acquire writes the
	//     recorded Return back after the call returns.
	//   - Each per-method Calls() namespace gains WaitForCount(n, timeout)
	//     and WaitForCall(predicate, timeout) for asserting on calls made
	//     by background goroutines.
	Async bool

	// AddHeaderNote prepends "// Code generated by MockGen. DO NOT EDIT."
	// to the generated file.
	AddHeaderNote bool

	// CustomTemplate, when non-empty, replaces the built-in text/template.
	// It receives a MockData. Output is run through gofmt and must be valid
	// Go. See DefaultTemplate() for the built-in as a starting point.
	CustomTemplate string

	// InspectEntry overrides the default name of the inspection entry
	// point method ("Calls"). Use when the source interface has a method
	// named Calls that would collide with the generator-emitted accessor.
	InspectEntry string

	// ConfigEntry overrides the default name of the configuration entry
	// point method ("On"). Use when the source interface has a method
	// named On that would collide with the generator-emitted accessor.
	ConfigEntry string
}

// DefaultInspectEntry is the default name of the inspection namespace
// entry point method on a generated mock.
const DefaultInspectEntry = "Calls"

// DefaultConfigEntry is the default name of the configuration namespace
// entry point method on a generated mock.
const DefaultConfigEntry = "On"

// MockData is the top-level value passed to the template.
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
	// (e.g. "ServiceMock"). Also used as a prefix for per-method type
	// names so two mocks in the same file never collide.
	MockName string
	// InspectEntryName is the method name for the inspection namespace
	// entry point — "Calls" by default, or whatever the user overrode.
	InspectEntryName string
	// ConfigEntryName is the method name for the configuration namespace
	// entry point — "On" by default, or whatever the user overrode.
	ConfigEntryName string
	// InspectType is the type name for the inspection namespace object
	// (e.g. "ServiceMockCalls"). Derived from MockName + InspectEntryName.
	InspectType string
	// ConfigType is the type name for the configuration namespace object
	// (e.g. "ServiceMockOn"). Derived from MockName + ConfigEntryName.
	ConfigType string
	// Methods is one entry per method on the interface (including methods
	// reached through embedding). Order is stable across runs.
	Methods []MethodData
}

// MethodData describes one method on a generated mock. Every field needed
// to render the interface method body, the per-call record, the per-method
// inspect/config types, and the configuration storage lives here.
type MethodData struct {
	// Name is the original method name (e.g. "Send").
	Name string
	// Signature carries the parts of the source method signature.
	Signature SignatureData

	// CallType is the exported per-call record type name
	// (e.g. "ServiceMockSendCall").
	CallType string
	// ParamType is the exported sub-struct type name holding the recorded
	// param values (e.g. "ServiceMockSendParam"). Empty struct when the
	// method has no params.
	ParamType string
	// ReturnType is the exported sub-struct type name holding the recorded
	// return values (e.g. "ServiceMockSendReturn"). Empty struct when the
	// method has no return values.
	ReturnType string

	// ParamFields are public, title-cased fields on ParamType, one per
	// method param, in declaration order.
	ParamFields []FieldData
	// ReturnFields are public, title-cased fields on ReturnType, one per
	// method return, in declaration order.
	ReturnFields []FieldData

	// LogField is the unexported `[]<CallType>` slice field name on the
	// mock struct (e.g. "sendLog").
	LogField string
	// CfgField is the unexported per-method config struct field on the
	// mock struct (e.g. "sendCfg").
	CfgField string
	// CfgType is the unexported per-method config struct type name
	// (e.g. "serviceMockSendCfg"). Holds armed returns + handler.
	CfgType string
	// CfgReturnFields are the unexported fields on CfgType, one per
	// method return — holding the armed static value.
	CfgReturnFields []FieldData
	// HandlerFieldName is the name of the unexported handler field on
	// CfgType (e.g. "handler"). Empty when the method has no returns
	// (a handler with no returns adds no value over a pre-call hook).
	HandlerFieldName string
	// HandlerType is the rendered handler function type
	// (e.g. "func(context.Context, string) (string, error)"). Empty when
	// HandlerFieldName is empty.
	HandlerType string

	// InspectMethodType is the per-method inspection type name
	// (e.g. "ServiceMockSendCalls"). Hangs off the InterfaceData's
	// InspectType.
	InspectMethodType string
	// ConfigMethodType is the per-method configuration type name
	// (e.g. "ServiceMockSendOn"). Hangs off the InterfaceData's
	// ConfigType.
	ConfigMethodType string
	// PerReturnSetterType is the unexported type name for the per-return
	// setter namespace returned by `Return()` on the per-method config
	// object (e.g. "serviceMockSendReturn"). Empty when the method has
	// no returns.
	PerReturnSetterType string

	// ReturnsSetterName is the setter that takes all returns positionally
	// (e.g. "ReturnAll"). Empty when the method has no returns.
	ReturnsSetterName string
	// CustomReturnSetterName is the setter that installs a dynamic
	// function returning the values (e.g. "ReturnCustom"). Empty when
	// the method has no returns.
	CustomReturnSetterName string
	// IndividualReturnSetters are per-return setters living under the
	// `Return()` namespace — MethodName is the bare return name
	// (e.g. "Id", "Err"). One per return value.
	IndividualReturnSetters []SetterMethod
}

// FieldData is a generic `Name Type` pair used for struct fields.
type FieldData struct {
	Name string
	Type string
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
	// ReturnTypes is the comma-separated list of return type strings,
	// e.g. "string, error". Used to render handler types.
	ReturnTypes string
	// ParamTypes is the comma-separated list of param type strings,
	// e.g. "context.Context, string". Used to render handler types.
	ParamTypes string
}

// SetterMethod describes a generated write accessor.
type SetterMethod struct {
	// MethodName is the exported method name (e.g. "ReturnsID").
	MethodName string
	// FieldName is the unexported underlying field the method writes
	// (on the per-method Cfg struct).
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

	if this.options.Async {
		stdImports = append(stdImports, strconv.Quote("sync"))
		// WaitForCount / WaitForCall (emitted only under --async by the
		// default template) use time.Now / time.Sleep / errors.New.
		// Custom templates manage their own imports.
		if this.options.CustomTemplate == "" {
			stdImports = append(stdImports, strconv.Quote("errors"))
			stdImports = append(stdImports, strconv.Quote("time"))
		}
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
	mockName := name + structMockSuffix
	inspectEntry := this.options.InspectEntry
	if inspectEntry == "" {
		inspectEntry = DefaultInspectEntry
	}
	configEntry := this.options.ConfigEntry
	if configEntry == "" {
		configEntry = DefaultConfigEntry
	}

	ifaceData := InterfaceData{
		Name:             name,
		MockName:         mockName,
		InspectEntryName: inspectEntry,
		ConfigEntryName:  configEntry,
		InspectType:      unexportedTypeName(mockName + inspectEntry),
		ConfigType:       unexportedTypeName(mockName + configEntry),
		Methods:          []MethodData{},
	}

	for i := range iface.NumMethods() {
		method := iface.Method(i)
		methodData := this.collectMethodData(method, mockName)
		ifaceData.Methods = append(ifaceData.Methods, methodData)
	}

	if err := validateNoEntryCollisions(name, ifaceData.Methods, inspectEntry, configEntry); err != nil {
		return InterfaceData{}, err
	}

	return ifaceData, nil
}

// validateNoEntryCollisions checks that no source-interface method name
// collides with the inspection or configuration entry-point methods that
// the generator emits on the mock. Also checks Reset (mock-level).
//
// On collision, returns an error directing the user to override via
// Options.InspectEntry or Options.ConfigEntry (CLI: --inspect-entry,
// --config-entry).
func validateNoEntryCollisions(interfaceName string, methods []MethodData, inspectEntry, configEntry string) error {
	// Reserved mock-level method names. Constructor (New<MockName>) lives at
	// package scope, not method scope, so it can't collide with interface
	// methods — no check needed for it.
	for _, m := range methods {
		switch m.Name {
		case inspectEntry:
			return fmt.Errorf(
				"interface %s declares method %q which collides with the inspection entry-point. "+
					"Override the entry name with Options.InspectEntry (CLI: --inspect-entry NAME).",
				interfaceName, m.Name,
			)
		case configEntry:
			return fmt.Errorf(
				"interface %s declares method %q which collides with the configuration entry-point. "+
					"Override the entry name with Options.ConfigEntry (CLI: --config-entry NAME).",
				interfaceName, m.Name,
			)
		case "Reset":
			return fmt.Errorf(
				"interface %s declares method %q which collides with the mock-level Reset method. "+
					"Rename the source method to avoid the conflict.",
				interfaceName, m.Name,
			)
		}
	}
	return nil
}

func (this *generator) collectMethodData(method *types.Func, mockName string) MethodData {
	sig := method.Type().(*types.Signature)
	methodLower := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
	prefix := mockName + method.Name()

	methodData := MethodData{
		Name:              method.Name(),
		CallType:          prefix + "Call",
		ParamType:         prefix + "Param",
		ReturnType:        prefix + "Return",
		LogField:          methodLower + "Log",
		CfgField:          methodLower + "Cfg",
		CfgType:           strings.ToLower(mockName[0:1]) + mockName[1:] + method.Name() + "Cfg",
		InspectMethodType: unexportedTypeName(mockName + method.Name() + "Calls"),
		ConfigMethodType:  unexportedTypeName(mockName + method.Name() + "On"),
		ParamFields:       []FieldData{},
		ReturnFields:      []FieldData{},
		CfgReturnFields:   []FieldData{},
	}

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		rawFieldName := getParamMember(method, j)
		suffix := rawFieldName[len(method.Name()+"_param_"):]
		methodData.ParamFields = append(methodData.ParamFields, FieldData{
			Name: exportedFieldName(suffix),
			Type: this.getFieldType(param.Type()),
		})
	}

	for j := range sig.Results().Len() {
		result := sig.Results().At(j)
		rawFieldName := getReturnMember(method, j)
		suffix := rawFieldName[len(method.Name()+"_return_"):]
		exported := exportedFieldName(suffix)
		methodData.ReturnFields = append(methodData.ReturnFields, FieldData{
			Name: exported,
			Type: this.getFieldType(result.Type()),
		})
		// Cfg storage uses lowercase first letter to keep the field
		// unexported (Cfg type is unexported anyway).
		cfgName := strings.ToLower(exported[0:1]) + exported[1:]
		methodData.CfgReturnFields = append(methodData.CfgReturnFields, FieldData{
			Name: cfgName,
			Type: this.getFieldType(result.Type()),
		})
	}

	methodData.Signature = this.collectSignatureData(method, sig)

	if sig.Results().Len() > 0 {
		methodData.HandlerFieldName = "handler"
		methodData.HandlerType = this.handlerFuncType(sig)
		methodData.ReturnsSetterName = "ReturnAll"
		methodData.CustomReturnSetterName = "ReturnCustom"
		methodData.PerReturnSetterType = unexportedTypeName(mockName + method.Name() + "Return")
		for i, rf := range methodData.ReturnFields {
			methodData.IndividualReturnSetters = append(
				methodData.IndividualReturnSetters,
				SetterMethod{
					MethodName: rf.Name,
					FieldName:  methodData.CfgReturnFields[i].Name,
					ParamType:  rf.Type,
				},
			)
		}
	}

	return methodData
}

func (this *generator) collectSignatureData(method *types.Func, sig *types.Signature) SignatureData {
	var params, returns strings.Builder
	paramNames := []string{}
	var paramTypesList, returnTypesList []string

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		if j > 0 {
			params.WriteString(", ")
		}

		paramName := getParamName(sig, j)
		paramNames = append(paramNames, paramName)

		var renderedType string
		if sig.Variadic() && j == sig.Params().Len()-1 {
			renderedType = "..." + this.getFieldType(param.Type().(*types.Slice).Elem())
		} else {
			renderedType = this.getFieldType(param.Type())
		}
		paramTypesList = append(paramTypesList, renderedType)

		var parType string
		if sig.Variadic() && j == sig.Params().Len()-1 {
			parType = " " + renderedType
		} else if j >= sig.Params().Len()-1 || sig.Params().At(j+1).Type().String() != param.Type().String() {
			parType = " " + renderedType
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
			rt := this.getFieldType(result.Type())
			returns.WriteString(pref + rt)
			returnTypesList = append(returnTypesList, rt)
		}
		if useParens {
			returns.WriteString(")")
		}
	}

	return SignatureData{
		Params:      params.String(),
		Returns:     returns.String(),
		ParamNames:  paramNames,
		ParamTypes:  strings.Join(paramTypesList, ", "),
		ReturnTypes: strings.Join(returnTypesList, ", "),
	}
}

// handlerFuncType returns the rendered function type for a method's handler,
// e.g. "func(context.Context, string) (string, error)" or "func() error".
func (this *generator) handlerFuncType(sig *types.Signature) string {
	paramTypes, returnTypes := this.handlerSignatureTypes(sig)
	funcType := "func(" + paramTypes + ") "
	if sig.Results().Len() > 1 {
		funcType += "(" + returnTypes + ")"
	} else {
		funcType += returnTypes
	}
	return funcType
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
