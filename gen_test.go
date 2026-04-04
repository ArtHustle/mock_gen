package mock_gen

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"
)

// TestFindTypeDef_PrefersTypeName verifies that findTypeDef returns the type declaration
// (types.TypeName) rather than a struct field that happens to share the same identifier name.
//
// This is a regression test for the bug where an interface named "Provider" was not found
// when the same package had a struct with a field also named "Provider". Because
// pkg.TypesInfo.Defs maps *ast.Ident keys and map iteration is non-deterministic, findTypeDef
// could return the struct field instead of the interface declaration, producing an empty mock.
func TestFindTypeDef_PrefersTypeName(t *testing.T) {
	fset := token.NewFileSet()

	// Ident for the interface type declaration: "type Provider interface{...}"
	typeNameIdent := &ast.Ident{Name: "Provider", NamePos: token.Pos(fset.Base())}

	// Ident for a struct field: "Provider string" in Config
	fieldIdent := &ast.Ident{Name: "Provider", NamePos: token.Pos(fset.Base() + 1)}

	pkg := types.NewPackage("example.com/service", "service")

	// types.TypeName represents a type declaration
	typeName := types.NewTypeName(token.NoPos, pkg, "Provider", types.NewInterfaceType(nil, nil))

	// types.Var with IsField=true represents a struct field
	field := types.NewField(token.NoPos, pkg, "Provider", types.Typ[types.String], false)

	// Build a Defs map with both entries — order matters for the bug:
	// we put the field first so a naive first-match would return it.
	defs := map[*ast.Ident]types.Object{
		fieldIdent:    field,
		typeNameIdent: typeName,
	}

	result := findTypeDef(defs, "Provider")

	if result == nil {
		t.Fatal("findTypeDef returned nil, expected the TypeName")
	}
	if _, ok := result.(*types.TypeName); !ok {
		t.Fatalf("findTypeDef returned %T (%v), expected *types.TypeName", result, result)
	}
}
