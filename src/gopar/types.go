// Types
//
// Supports operations on types, such as figuring out the type of a struct
// field or array access. Also figures out the result of a binary expression
// between two types, or a dereference (*) or address-of (&) operation.
//
// Type definitions are fully-defined (Type.final = true), all other types are
// references to them.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

type Type interface {
	Complete(Resolver)
	Definition() ast.Node
	Fields() []string  // return an ordered list of all fields in this type
	Field(string) Type // return .field's type
	Dereference() Type // the return type of a *dereference operation
	Reference() Type   // return type type of a &reference operation
	IndexKey() Type    // the return type of an [index] or <-chan operation
	IndexValue() Type
	Call() []Type   // the return types of calling this type
	Math(Type) Type // outcome of any math operation with another type
	String() string // a representation of this type
}

type BaseType struct {
	ast.Node // definition node
}

func newBaseType(node ast.Node) *BaseType {
	return &BaseType{Node: node}
}

func (t *BaseType) Complete(resolver Resolver) {
	return
}

func (t *BaseType) Definition() ast.Node {
	return t.Node
}

func (t *BaseType) Fields() []string {
	return nil
}

func (t *BaseType) Field(name string) Type {
	return nil
}

func (t *BaseType) Dereference() Type {
	return nil
}

func (t *BaseType) Reference() Type {
	return nil
}

func (t *BaseType) IndexKey() Type {
	return nil
}

func (t *BaseType) IndexValue() Type {
	return nil
}

func (t *BaseType) Call() []Type {
	return nil
}

func (t *BaseType) Math(other Type) Type {
	return nil
}

func (t *BaseType) String() string {
	return fmt.Sprintf("Type<%T %v>", t.Node, t.Node)
}

type StructType struct {
	BaseType
	fieldOrder []string
	fields     map[string]Type
}

func newStructType(node ast.Node) *StructType {
	return &StructType{
		BaseType:   newBaseType(node),
		fieldOrder: make([]string, 0),
		fields:     make(map[string]Type),
	}
}

// fill in all struct fields
func (t *StructType) Complete(resolver Resolver) {
	switch e := t.Node.(type) {
	case *ast.StructType:
		for _, field := range e.Fields.List {
			fieldTyp := TypeOf(field.Type, resolver)
			// Embedded fields
			// *Struct1
			// Struct2
			// abc.Struct3
			if len(field.Names) == 0 {
				var ig IdentifierGroup
				AccessIdentBuild(&ig, field.Type, nil)
				name := ig.group[len(ig.group)-1].id
				t.addField(name, fieldTyp)
			} else {
				for _, name := range field.Names {
					t.addField(name.Name, fieldTyp)
				}
			}
		}
	}
}

func (t *StructType) addField(name string, typ *Type) {
	t.fields[name] = typ
	t.fieldOrder = append(t.fieldOrder, name)
	return
}

// an array, list, map or chan type
type IndexedType struct {
	BaseType
	key   Type
	value Type
}

func newIndexedType(node ast.Node) *IndexedType {
	return &IndexedType{
		BaseType: newBaseType(node),
	}
}

// fill in key and value sections
func (t *IndexedType) Complete(resolver Resolver) {
	return
}

func (t *IndexedType) IndexKey() Type {
	return t.key
}

func (t *IndexedType) IndexValue() Type {
	return t.value
}

// a pointer type
type PointerType struct {
	BaseType
	inner Type
}

func newPointerType(node ast.Node) *PointerType {
	return &PointerType{
		BaseType: newBaseType(node),
	}
}

// Resolve the inner type
func (t *PointerType) Complete(resolver Resolver) {
	return
}

// a function
type FuncType struct {
	BaseType
	args    []Type
	results []Type
}

func newFuncType(node ast.Node) *FuncType {
	return &FuncType{
		BaseType: newBaseType(node),
		args:     make([]Type, 0),
		results:  make([]Type, 0),
	}
}

// Create a new type from a declaration Node
func TypeDecl(expr ast.Node) Type {
	switch n := t.Node.(type) {
	case *ast.ChanType, *ast.ArrayType, *ast.MapType:
		return newIndexedType(n)
	case *ast.StructType, *ast.InterfaceType:
		return newStructType(n)
	case *ast.StarExpr:
		return newPointerType(n)
	case *ast.FuncType:
		return newFuncType(n)
	}
	return nil
}

func (typ *BaseType) String() string {
	var buffer bytes.Buffer
	var Format func(cur ast.Node) string
	Format = func(cur ast.Node) string {
		switch t := cur.(type) {
		case *ast.Ident:
			buffer.WriteString(t.Name)
		case *ast.BasicLit:
			buffer.WriteString(t.Value)
		case *ast.ArrayType:
			buffer.WriteString("[")
			if t.Len != nil {
				Format(t.Len)
			}
			buffer.WriteString("]")
			Format(t.Elt)
		case *ast.StarExpr:
			buffer.WriteString("*")
			Format(t.X)
		case *ast.StructType:
			buffer.WriteString("struct {")
			for i, field := range typ.Fields() {
				buffer.WriteString(fmt.Sprintf("%s %v", field, typ.Field(field)))
				if i != len(typ.Fields())-1 {
					buffer.WriteString(", ")
				}
			}
			buffer.WriteString("}")
		case *ast.FuncType:
			buffer.WriteString("func (")
			if t.Params != nil {
				for j, arg := range t.Params.List {
					for i, name := range arg.Names {
						buffer.WriteString(name.Name)
						if i != len(arg.Names)-1 {
							buffer.WriteString(", ")
						}
					}
					buffer.WriteString(" ")
					Format(arg.Type)
					if j != len(t.Params.List)-1 {
						buffer.WriteString(", ")
					}
				}
			}
			buffer.WriteString(")")
			if t.Results != nil {
				buffer.WriteString(" (")
				for j, arg := range t.Results.List {
					for i, name := range arg.Names {
						buffer.WriteString(name.Name)
						if i != len(arg.Names)-1 {
							buffer.WriteString(", ")
						}
					}
					buffer.WriteString(" ")
					Format(arg.Type)
					if j != len(t.Params.List)-1 {
						buffer.WriteString(", ")
					}
				}
				buffer.WriteString(")")
			}
		case *ast.MapType:
			buffer.WriteString("map[")
			Format(t.Key)
			buffer.WriteString("]")
			Format(t.Value)
		case *ast.ChanType:
			if t.Dir == ast.SEND {
				buffer.WriteString("->")
			} else if t.Dir == ast.RECV {
				buffer.WriteString("<-")
			}
			buffer.WriteString("chan ")
			Format(t.Value)
		default:
			buffer.WriteString(fmt.Sprintf("Type<%T %+v>", cur, cur))
		}
		return ""
	}
	Format(typ.Node)
	return buffer.String()
}

// Takes an identifier, returns the node that defines it. This should search all
// scopes up to the package level.
type Resolver func(ident string) *Type

func TypeOf(expr ast.Node, resolver Resolver) Type {
	switch t := expr.(type) {
	case *ast.CallExpr:
		var fnType *Type
		switch f := t.Fun.(type) {
		case *ast.Ident:
			if f.Name == "make" {
				fnType = NewType(&ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{&ast.Field{Type: t.Args[0]}}}})
			} else {
				fnType = resolver(f.Name)
			}
		case *ast.FuncLit:
			fnType = NewType(f.Type)
		}
		return fnType
	case *ast.Ident:
		identType := resolver(t.Name)
		return identType
	case *ast.BasicLit:
		switch t.Kind {
		case token.FLOAT:
			expr = &ast.Ident{Name: "float64"}
		case token.INT:
			expr = &ast.Ident{Name: "int"}
		case token.STRING:
			expr = &ast.Ident{Name: "string"}
		}
	case *ast.IndexExpr:
		return TypeOf(t.X, resolver)
	case *ast.UnaryExpr:
		// &something
		if t.Op == token.AND {
			innerTyp := TypeOf(t.X, resolver)
			expr = &ast.StarExpr{X: innerTyp.Node.(ast.Expr)}
		} else if t.Op == token.ARROW {
			chanTyp := TypeOf(t.X, resolver).Node.(*ast.ChanType)
			expr = chanTyp.Value
		}
	case *ast.StarExpr:
		ptrType := TypeOf(t.X, resolver)
		expr = &ast.StarExpr{X: ptrType.Node.(ast.Expr)}
	case *ast.CompositeLit:
		// Something{}
		return TypeOf(t.Type, resolver)
	case *ast.ArrayType, *ast.ChanType, *ast.MapType:
		// none
	default:
		fmt.Printf("Unhandled TypeOf(%T %+v)\n", expr, expr)
	}
	return NewType(expr)
}

// Helper functions for constructing C/OpenCL structures:
// http://golang.org/ref/spec#Size_and_alignment_guarantees
// http://www.khronos.org/registry/cl/sdk/1.1/docs/man/xhtml/attributes-types.html
// https://code.google.com/p/go/source/browse/src/pkg/go/types/sizes.go
func SizeOf(typ Type) int64 {
	return 1
}

func AlignOf(typ Type) int64 {
	return 1
}
