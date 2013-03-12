// Types
//
// Supports operations on types, such as figuring out the type of a struct
// field or array access. Also figures out the result of a binary expression
// between two types, or a dereference (*) or address-of (&) operation.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
)

var builtinIdents map[string]bool = map[string]bool{
	"bool":      true,
	"byte":      true,
	"complex64": true,
}

type TypeCategory uint

const (
	UnknownType TypeCategory = iota
	BasicType
	ChanType
	StructType
	ArrayType
	SliceType
	MapType
	InterfaceType
	PointerType
	FuncType
	ImportType
)

type Type struct {
	ast.Node   // pointer
	fieldOrder []string
	fields     map[string]*Type
	methods    map[string]*Type
}

// Resolve all first-level fields, but not embedded fields
func NewType(expr ast.Node) *Type {
	typ := &Type{Node: expr}
	if typ.HasFields() {
		typ.fields = make(map[string]*Type)
		typ.methods = make(map[string]*Type)
	}
	return typ
}

// Finish filling in this type's embedded structures
func (t *Type) Complete(resolver Resolver) {
	return
}

// Add a method to this type
func (t *Type) AddMethod(name string, method *ast.FuncDecl) {
	t.methods[name] = NewType(method.Type)
	return
}

func (t *Type) Category() TypeCategory {
	switch n := t.Node.(type) {
	case *ast.ChanType:
		return ChanType
	case *ast.StructType:
		return StructType
	case *ast.ArrayType:
		if n.Len != nil {
			return ArrayType
		} else {
			return SliceType
		}
	case *ast.MapType:
		return MapType
	case *ast.InterfaceType:
		return InterfaceType
	case *ast.StarExpr:
		return PointerType
	case *ast.FuncType:
		return FuncType
	}
	return UnknownType
}

func (t *Type) HasFields() bool {
	return (t.Category() == StructType || t.Category() == InterfaceType)
}

// Return all fields of this struct or interface
func (t *Type) Fields() []string {
	if t.HasFields() {
		return t.fieldOrder
	}
	return nil
}

// Return the type of a struct or interface
func (t *Type) Field(name string) *Type {
	if t.HasFields() {
		return t.fields[name]
	}
	return NewType(nil)
}

func (t *Type) String() string {
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
			for _, field := range t.Fields.List {
				if len(field.Names) == 0 {
					Format(field.Type)
					buffer.WriteString(", ")
				}
				for _, name := range field.Names {
					buffer.WriteString(name.Name)
					buffer.WriteString(" ")
					Format(field.Type)
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
			buffer.WriteString(fmt.Sprintf("<%T %+v>", cur, cur))
		}
		return ""
	}
	Format(t.Node)
	return buffer.String()
}

// Takes an identifier, returns the node that defines it. This should search all
// scopes up to the package level.
type Resolver func(ident string) *Type

func TypeOf(expr ast.Node, resolver Resolver) *Type {
	switch t := expr.(type) {
	case *ast.CallExpr:
		var fnType *Type
		switch f := t.Fun.(type) {
		case *ast.Ident:
			if f.Name == "make" {
				fnType = NewType(&ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{&ast.Field{Type: t.Args[0]}}}})
			} else if f.Name == "len" {
				fnType = NewType(&ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{&ast.Field{Type: &ast.Ident{Name: "int"}}}}})
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
