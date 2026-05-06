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

type Options struct {
	GenerateReturnHandler bool
	MakeThreadSafe        bool
	AddHeaderNote         bool
	CustomTemplate        string
}

type MockData struct {
	HeaderNote string
	Imports    ImportsData
	Interfaces []InterfaceData
	Options    Options
}

type ImportsData struct {
	StdImports    []string
	NonStdImports []string
}

type InterfaceData struct {
	Name     string
	MockName string
	Methods  []MethodData
}

type MethodData struct {
	Name          string
	CallsField    string
	Params        []FieldData
	ReturnHandler *ReturnHandlerData
	Returns       []FieldData
	Signature     SignatureData
	Getters       *GetterData
}

type FieldData struct {
	Name string
	Type string
}

type ReturnHandlerData struct {
	FieldName   string
	ParamTypes  string
	ReturnTypes string
}

type SignatureData struct {
	Params       string
	Returns      string
	ParamNames   []string
	ReturnFields []string
}

type GetterData struct {
	CallsGetter   GetterMethod
	ParamGetters  []GetterMethod
	ReturnSetters []SetterMethod
	HandlerSetter *SetterMethod
}

type GetterMethod struct {
	MethodName string
	FieldName  string
	ReturnType string
}

type SetterMethod struct {
	MethodName string
	FieldName  string
	ParamType  string
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

	mockData.Imports = this.collectImportsData(imports)

	for _, rec := range genRecs {
		ifaceData := this.collectInterfaceData(rec.name, rec.iface)
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

func GenerateMockFile(filePath string) (string, error) {
	return GenerateMockFileWithOptions(filePath, Options{})
}

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

func (this *generator) collectImportsData(importPackages []*types.Package) ImportsData {
	var stdImports, nonStdImports []string

	if this.options.MakeThreadSafe {
		stdImports = append(stdImports, strconv.Quote("sync"))
	}

	for _, pkg := range importPackages {
		impPath := pkg.Path()
		alias, hasAlias := this.importPathToAlias[impPath]

		useAlias := true
		if hasAlias {
			lastSlash := strings.LastIndex(impPath, "/")
			if alias == impPath[lastSlash+1:] {
				useAlias = false
			}
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

func (this *generator) collectInterfaceData(name string, iface *types.Interface) InterfaceData {
	ifaceData := InterfaceData{
		Name:     name,
		MockName: name + structMockSuffix,
		Methods:  []MethodData{},
	}

	for i := range iface.NumMethods() {
		method := iface.Method(i)
		methodData := this.collectMethodData(method)
		ifaceData.Methods = append(ifaceData.Methods, methodData)
	}

	return ifaceData
}

func (this *generator) collectMethodData(method *types.Func) MethodData {
	sig := method.Type().(*types.Signature)

	callsFieldName := method.Name() + "_calls"
	if this.options.MakeThreadSafe {
		callsFieldName = strings.ToLower(method.Name()[0:1]) + method.Name()[1:] + "_calls"
	}

	methodData := MethodData{
		Name:       method.Name(),
		CallsField: callsFieldName,
		Params:     []FieldData{},
		Returns:    []FieldData{},
	}

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		parType := this.getFieldType(param.Type())
		fieldName := getParamMember(method, j)
		if this.options.MakeThreadSafe {
			methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
			fieldName = methodName + "_param_" + fieldName[len(method.Name()+"_param_"):]
		}
		methodData.Params = append(methodData.Params, FieldData{
			Name: fieldName,
			Type: parType,
		})
	}

	if this.options.GenerateReturnHandler && sig.Results().Len() > 0 {
		fieldName := method.Name() + "_return"
		if this.options.MakeThreadSafe {
			fieldName = strings.ToLower(method.Name()[0:1]) + method.Name()[1:] + "_return"
		}

		paramTypes, returnTypes := this.handlerSignatureTypes(sig)
		methodData.ReturnHandler = &ReturnHandlerData{
			FieldName:   fieldName,
			ParamTypes:  paramTypes,
			ReturnTypes: returnTypes,
		}
	}

	for j := range sig.Results().Len() {
		result := sig.Results().At(j)
		fieldName := getReturnMember(method, j)
		if this.options.MakeThreadSafe {
			methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
			fieldName = methodName + "_return_" + fieldName[len(method.Name()+"_return_"):]
		}
		methodData.Returns = append(methodData.Returns, FieldData{
			Name: fieldName,
			Type: this.getFieldType(result.Type()),
		})
	}

	methodData.Signature = this.collectSignatureData(method, sig)

	if this.options.MakeThreadSafe {
		methodData.Getters = this.collectGetterData(method, sig)
	}

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
			if this.options.MakeThreadSafe {
				methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
				returnField = methodName + "_return_" + returnField[len(method.Name()+"_return_"):]
			}
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

func (this *generator) collectGetterData(method *types.Func, sig *types.Signature) *GetterData {
	getterData := &GetterData{
		ParamGetters:  []GetterMethod{},
		ReturnSetters: []SetterMethod{},
	}

	callsFieldNameLower := strings.ToLower(method.Name()[0:1]) + method.Name()[1:] + "_calls"
	getterData.CallsGetter = GetterMethod{
		MethodName: method.Name() + "_calls",
		FieldName:  callsFieldNameLower,
		ReturnType: "int",
	}

	for j := range sig.Params().Len() {
		param := sig.Params().At(j)
		fieldName := getParamMember(method, j)
		methodName := strings.ToLower(method.Name()[0:1]) + method.Name()[1:]
		fieldNameLower := methodName + "_param_" + fieldName[len(method.Name()+"_param_"):]
		parType := this.getFieldType(param.Type())

		getterData.ParamGetters = append(getterData.ParamGetters, GetterMethod{
			MethodName: fieldName,
			FieldName:  fieldNameLower,
			ReturnType: parType,
		})
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
	default:
		return ""
	}
}
