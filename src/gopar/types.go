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

var BuiltinTypes map[string]Type

func init() {
	fmt.Println("Init: types.go")
	BuiltinTypes = make(map[string]Type, 0)

	builtin := []string{
		"uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64",
		"float32", "float64", "complex64", "complex128", "uint", "int", "uintptr",
		"rune", "byte", "bool", "nil",
	}
	for _, ident := range builtin {
		BuiltinTypes[ident] = newFinalBaseType(&ast.Ident{Name: ident})
	}

	BuiltinTypes["true"] = BuiltinTypes["bool"]
	BuiltinTypes["false"] = BuiltinTypes["bool"]
	BuiltinTypes["iota"] = BuiltinTypes["int"]
	BuiltinTypes["string"] = newCustomIndexedType(BuiltinTypes["byte"], BuiltinTypes["int"], true)
	errorTyp := newStructType(nil)
	errorTyp.AddMethod("Error", newCustomFuncType(func(args []Type) Type {
		return BuiltinTypes["string"]
	}).SetParameterAccess(true))
	BuiltinTypes["error"] = errorTyp

	// builtin functions
	BuiltinTypes["len"] = newCustomFuncType(func(args []Type) Type {
		return BuiltinTypes["int"]
	}).SetParameterAccess(true)

	BuiltinTypes["cap"] = BuiltinTypes["len"]

	BuiltinTypes["new"] = newCustomFuncType(func(args []Type) Type {
		return newPointerTypeFromT(args[0])
	}).SetParameterAccess(true)
	BuiltinTypes["make"] = newCustomFuncType(func(args []Type) Type {
		return args[0]
	}).SetParameterAccess(true)
	BuiltinTypes["append"] = newCustomFuncType(func(args []Type) Type {
		return args[0]
	}).SetParameterAccess(false)

	BuiltinTypes["copy"] = newCustomFuncType(func(args []Type) Type {
		return BuiltinTypes["int"]
	}).SetParameterAccess(false, true)

	BuiltinTypes["close"] = newCustomFuncType(func(args []Type) Type {
		return nil
	}).SetParameterAccess(false)

	BuiltinTypes["delete"] = newCustomFuncType(func(args []Type) Type {
		return nil
	}).SetParameterAccess(false, true)

	BuiltinTypes["complex"] = newCustomFuncType(func(args []Type) Type {
		width := args[0].Definition().(*ast.Ident).Name[len("float"):]
		return BuiltinTypes["complex"+width]
	}).SetParameterAccess(true)
	BuiltinTypes["real"] = newCustomFuncType(func(args []Type) Type {
		width := args[0].Definition().(*ast.Ident).Name[len("complex"):]
		return BuiltinTypes["float"+width]
	}).SetParameterAccess(true)
	BuiltinTypes["imag"] = BuiltinTypes["real"]

	BuiltinTypes["panic"] = newCustomFuncType(func(args []Type) Type {
		return nil
	}).SetParameterAccess(true)

	BuiltinTypes["recover"] = newCustomFuncType(func(args []Type) Type {
		return nil // TODO: this should return interface{}
	}).SetParameterAccess(true)

	BuiltinTypes["print"] = newCustomFuncType(func(args []Type) Type {
		return nil
	}).SetParameterAccess(true)

	// builtin packages

	// unsafe
	//BuiltinTypes["unsafe"] =

	// reflect
	//BuiltinTypes["reflect"] = 
}

type Type interface {
	Complete(Resolver)
	Definition() ast.Node
	Fields() []string            // return an ordered list of all fields in this type
	Field(string) Type           // return .field's type
	AddMethod(string, *FuncType) // add a new method to this type
	Method(string) *FuncType     // return .method() type
	Dereference() Type           // the return type of a *dereference operation
	IndexKey() Type              // type for [key]
	IndexValue() Type            // the return type of an [index] or <-chan operation
	Call([]Type) Type            // the return types of calling this type
	Math(Type, token.Token) Type // outcome of any math operation with another type
	PassByValue() bool           // is this type passed by value or reference?
	String() string              // a representation of this type
	CType() string               // type signature (int* for slice)
	CDecl() string               // type definition (typedef struct {} something)
}

type BaseType struct {
	ast.Node                        // definition node
	methods    map[string]*FuncType // every type can have a method set
	underlying Type                 // if this type is a direct shadow of an existing type
	complete   bool                 // ensure each type is only completed once
}

func newBaseType(node ast.Node) *BaseType {
	return &BaseType{Node: node, methods: make(map[string]*FuncType)}
}

func newShadowType(typ Type) *BaseType {
	t := newBaseType(nil)
	t.underlying = typ
	return t
}

func (t *BaseType) completed() bool {
	if !t.complete {
		t.complete = true
		return false
	}
	return true
}

func (t *BaseType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	t.underlying = TypeOfDecl(t.Node, resolver)
	return
}

// All methods to go to the underlying type
func (t *BaseType) Definition() ast.Node {
	if t.underlying == nil {
		return t.Node
	}
	return t.underlying.Definition()
}

func (t *BaseType) Fields() []string {
	if t.underlying != nil {
		return t.underlying.Fields()
	}
	return nil
}

func (t *BaseType) Field(name string) Type {
	if t.underlying != nil {
		return t.underlying.Field(name)
	}
	return nil
}

// Methods get added to the current type
func (t *BaseType) AddMethod(name string, f *FuncType) {
	t.methods[name] = f
}

func (t *BaseType) Method(name string) *FuncType {
	if method, ok := t.methods[name]; ok {
		fmt.Println(name, "=", method)
		return method
	}
	if t.underlying != nil {
		return t.underlying.Method(name)
	}
	fmt.Println("METHOD NOT FOUND", name, t.methods)
	return nil
}

func (t *BaseType) Dereference() Type {
	if t.underlying == nil {
		return nil
	}
	return t.underlying.Dereference()
}

func (t *BaseType) IndexKey() Type {
	if t.underlying != nil {
		return t.underlying.IndexKey()
	}
	return nil
}

func (t *BaseType) IndexValue() Type {
	if t.underlying != nil {
		return t.underlying.IndexValue()
	}
	return nil
}

func (t *BaseType) Call(args []Type) Type {
	return t.underlying.Call(args)
}

// The declaration of the type, such as "typedef Pixel struct {}"
func (t *BaseType) CDecl() string {
	return "UNKNOWN DECL"
}

// The name of the type, such as "*Pixel" or "int32[]"
func (t *BaseType) CType() string {
	return "UNKNOWN TYPE"
}

func (t *BaseType) Math(other Type, op token.Token) Type {
	if t.underlying != nil {
		return BinaryOp(t.underlying, op, other)
	}
	return BinaryOp(t, op, other)
}

func (t *BaseType) PassByValue() bool {
	if t.underlying == nil {
		return true
	}
	return t.underlying.PassByValue()
}

func (typ *BaseType) String() string {
	var buffer bytes.Buffer
	if typ.underlying != nil {
		buffer.WriteString("Shadow{")
		buffer.WriteString(typ.underlying.String())
		buffer.WriteString("}")
	} else {
		buffer.WriteString("BaseType{")
		buffer.WriteString(typ.Node.(*ast.Ident).Name)
		buffer.WriteString("} ")
	}
	buffer.WriteString(typ.MethodSet())
	return buffer.String()
}

func (t *BaseType) MethodSet() string {
	var buffer bytes.Buffer
	if len(t.methods) > 0 {
		buffer.WriteString(" methods{")
		for k, method := range t.methods {
			if method.pointerMethod {
				buffer.WriteString("*")
			}
			buffer.WriteString(k + ",")
		}
		buffer.WriteString("}")
	}
	return buffer.String()
}

type FinalBaseType struct {
	*BaseType
}

func newFinalBaseType(node ast.Node) *FinalBaseType {
	t := &FinalBaseType{
		BaseType: newBaseType(node),
	}
	t.completed()
	return t
}

func (t *FinalBaseType) Complete(resolver Resolver) {
	// do nothing
	return
}

type ConstType struct {
	*BaseType
	value string
}

func newConstType(node ast.Node) *ConstType {
	return &ConstType{
		BaseType: newBaseType(node),
	}
}

func (t *ConstType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	switch t.Node.(*ast.BasicLit).Kind {
	case token.FLOAT:
		t.BaseType.underlying = resolver("float64")
	case token.INT:
		t.BaseType.underlying = resolver("int")
	case token.STRING:
		t.BaseType.underlying = resolver("string")
	}
	t.value = t.Node.(*ast.BasicLit).Value
	return
}

func (t *ConstType) String() string {
	return fmt.Sprintf("%s=%s", t.BaseType.String(), t.value)
}

// Structs AND Interfaces
type StructType struct {
	*BaseType
	fieldOrder []string
	fields     map[string]Type
	embedded   []Type // could be *StructType or *PointerType
	iface      bool
}

func newStructType(node ast.Node) *StructType {
	return &StructType{
		BaseType:   newBaseType(node),
		fieldOrder: make([]string, 0),
		fields:     make(map[string]Type),
	}
}

// fill in all struct fields, but not those carried from embedded structs
func (t *StructType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	switch e := t.Node.(type) {
	case *ast.StructType:
		for _, field := range e.Fields.List {
			fieldTyp := TypeOfDecl(field.Type, resolver)
			// Embedded fields
			// *Struct1
			// Struct2
			// abc.Struct3
			if len(field.Names) == 0 {
				var ig IdentifierGroup
				AccessIdentBuild(&ig, field.Type, nil)
				name := ig.group[len(ig.group)-1].id
				t.addField(name, fieldTyp)
				t.embedded = append(t.embedded, fieldTyp)
			} else {
				for _, name := range field.Names {
					t.addField(name.Name, fieldTyp)
				}
			}
		}
	case *ast.InterfaceType:
		t.iface = true
		for _, method := range e.Methods.List {
			fmt.Printf("Interface method: %+v\n", method)
			switch m := method.Type.(type) {
			case *ast.FuncType:
				methodTyp := newMethodType(m, nil, t)
				methodTyp.Complete(resolver)
				methodTyp.pointerMethod = true // all interface methods take pointers
				for _, name := range method.Names {
					t.AddMethod(name.Name, methodTyp)
				}
			case *ast.Ident:
				// embedded
				t.embedded = append(t.embedded, resolver(m.Name).(*StructType))
			}
		}
	}
}

func (t *StructType) addField(name string, typ Type) {
	t.fields[name] = typ
	t.fieldOrder = append(t.fieldOrder, name)
	return
}

func (t *StructType) Fields() []string {
	return t.fieldOrder
}

func (t *StructType) Field(name string) Type {
	if builtin, ok := t.fields[name]; ok {
		return builtin
	}
	// else search the embedded fields
	for _, embedded := range t.embedded {
		if field := embedded.Field(name); field != nil {
			return field
		}
	}
	return nil
}

func (t *StructType) Method(name string) *FuncType {
	if builtin := t.BaseType.Method(name); builtin != nil {
		return builtin
	}
	// else search the embedded methods
	for _, embedded := range t.embedded {
		if method := embedded.Method(name); method != nil {
			return method
		}
	}
	return nil
}

func (t *StructType) String() string {
	var buffer bytes.Buffer
	if !t.iface {
		buffer.WriteString("struct {")
	} else {
		buffer.WriteString("interface {")
	}
	for i, field := range t.Fields() {
		buffer.WriteString(fmt.Sprintf("%s", field))
		if i != len(t.Fields())-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("}")
	if len(t.embedded) > 0 {
		buffer.WriteString("embedded {")
		for i, e := range t.embedded {
			buffer.WriteString(e.String())
			if i != len(t.embedded)-1 {
				buffer.WriteString(", ")
			}
		}
		buffer.WriteString("}")
	}
	buffer.WriteString(t.MethodSet())

	return buffer.String()
}

// an array, list, map or chan type
type IndexedType struct {
	*BaseType
	key     Type
	value   Type
	byValue bool // is this type passed by value (array) or reference (map, slice)
}

func newIndexedType(node ast.Node) *IndexedType {
	return &IndexedType{
		BaseType: newBaseType(node),
	}
}

func newCustomIndexedType(value Type, key Type, byValue bool) *IndexedType {
	t := newIndexedType(nil)
	t.key = key
	t.value = value
	t.byValue = byValue
	return t
}

// fill in key and value sections
func (typ *IndexedType) Complete(resolver Resolver) {
	if typ.completed() {
		return
	}
	switch t := typ.Node.(type) {
	case *ast.ArrayType:
		if t.Len != nil {
			typ.byValue = true // array
		} else {
			typ.byValue = false // slice
		}
		typ.key = BuiltinTypes["int"]
		typ.value = TypeOfDecl(t.Elt, resolver)
	case *ast.MapType:
		typ.byValue = false
		typ.key = TypeOfDecl(t.Key, resolver)
		typ.value = TypeOfDecl(t.Value, resolver)
	case *ast.ChanType:
		typ.byValue = false
		typ.value = TypeOfDecl(t.Value, resolver)
	}
	return
}

func (t *IndexedType) IndexKey() Type {
	return t.key
}

// IndexValue will return a MultiType for the val, ok := map[key] expressions
func (t *IndexedType) IndexValue() Type {
	return newMultiType(t.value, BuiltinTypes["bool"])
}

func (t *IndexedType) PassByValue() bool {
	return t.byValue
}

func (typ *IndexedType) String() string {
	var buffer bytes.Buffer
	switch t := typ.Node.(type) {
	case *ast.ArrayType:
		buffer.WriteString("[")
		if t.Len != nil {
			buffer.WriteString(typ.IndexKey().String())
		}
		buffer.WriteString("]")
	case *ast.MapType:
		buffer.WriteString("map[")
		buffer.WriteString(typ.IndexKey().String())
		buffer.WriteString("]")
	case *ast.ChanType:
		if t.Dir == ast.SEND {
			buffer.WriteString("->")
		} else if t.Dir == ast.RECV {
			buffer.WriteString("<-")
		}
		buffer.WriteString("chan ")
	default:
		buffer.WriteString("custom[")
		buffer.WriteString(typ.IndexKey().String())
		buffer.WriteString("]")
	}
	buffer.WriteString(typ.value.String())
	buffer.WriteString(typ.MethodSet())

	return buffer.String()
}

// a pointer type
type PointerType struct {
	*BaseType
	inner Type
}

func newPointerType(node ast.Node) *PointerType {
	return &PointerType{
		BaseType: newBaseType(node),
	}
}

func newPointerTypeFromT(typ Type) *PointerType {
	return &PointerType{
		BaseType: newBaseType(nil),
		inner:    typ,
	}
}

func (t *PointerType) PassByValue() bool {
	return false
}

// Resolve the inner type
func (t *PointerType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	expr := t.Node.(*ast.StarExpr).X
	t.inner = TypeOfDecl(expr, resolver)
	return
}

func (t *PointerType) Dereference() Type {
	return t.inner
}

// Methods all go to the inner type
func (t *PointerType) AddMethod(name string, f *FuncType) {
	t.inner.AddMethod(name, f)
}

func (t *PointerType) Method(name string) *FuncType {
	// check both T* and T
	return t.inner.Method(name)
}

func (t *PointerType) Fields() []string {
	return t.inner.Fields()
}

func (t *PointerType) Field(name string) Type {
	return t.inner.Field(name)
}

func (t *PointerType) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("*")
	buffer.WriteString(t.inner.String())
	return buffer.String()
}

// a function
type FuncType struct {
	*BaseType
	params                []Type
	results               []Type
	ellipsis              bool // is the last argument in this function an ellipsis?
	method, pointerMethod bool
	receiver              Type           // if method == true
	name                  string         // if we know the name of this function
	body                  *ast.BlockStmt // the actual declaration, if possible; not there for builtin/C functions
	typ                   *ast.FuncType
	customCall            func([]Type) Type // a custom function for builtin magic (make, new, unsafe.*, etc)
	noWriteMask           []bool            // an optional mask for functions without bodies, true means the arg wasn't written to
}

// decl [optional], but used as the defining node in a BasicBlock, so it must
// be returned from BaseType.Definition()
func newFuncType(typ *ast.FuncType, decl *ast.FuncDecl) *FuncType {
	t := &FuncType{
		BaseType: newBaseType(decl),
		typ:      typ,
	}
	if decl != nil {
		t.name = decl.Name.Name
		t.body = decl.Body
	}
	return t
}

func newFuncLit(decl *ast.FuncLit) *FuncType {
	t := newFuncType(decl.Type, nil)
	t.body = decl.Body
	return t
}

func newMethodType(typ *ast.FuncType, decl *ast.FuncDecl, recv Type) *FuncType {
	funcTyp := newFuncType(typ, decl)
	funcTyp.method = true
	funcTyp.pointerMethod = recv.Dereference() != nil
	funcTyp.receiver = recv
	return funcTyp
}

func newCustomFuncType(f func([]Type) Type) *FuncType {
	t := newFuncType(nil, nil)
	t.customCall = f
	return t
}

// fill in params and results
func (t *FuncType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	expr := t.typ
	if expr.Params != nil {
		for _, arg := range expr.Params.List {
			argType := TypeOfDecl(arg.Type, resolver)
			// no name args
			i := len(arg.Names)
			if i == 0 {
				i = 1
			}
			for j := 0; j < i; j++ {
				t.params = append(t.params, argType)
			}
			if _, ellipsis := arg.Type.(*ast.Ellipsis); ellipsis {
				t.ellipsis = true
			}
		}
	}
	if expr.Results != nil {
		for _, result := range expr.Results.List {
			resultType := TypeOfDecl(result.Type, resolver)
			i := len(result.Names)
			if i == 0 {
				i = 1
			}
			for j := 0; j < i; j++ {
				t.results = append(t.results, resultType)
			}
		}
	}
	return
}

func (t *FuncType) Call(args []Type) Type {
	if t.customCall != nil {
		return t.customCall(args)
	} else {
		fmt.Println("CALL", t.results)
		if len(t.results) > 1 {
			return newMultiType(t.results...)
		} else {
			return t.results[0]
		}
	}
	return nil
}

func (t *FuncType) SetParameterAccess(mask ...bool) *FuncType {
	t.noWriteMask = mask
	return t // allow method chaining
}

// Return true if this parameter is pass-by-value, false otherwise
func (t *FuncType) GetParameterAccess(index int) bool {
	if t.noWriteMask == nil {
		if len(t.params) <= index {
			return t.params[len(t.params)-1].PassByValue()
		}
		return t.params[index].PassByValue()
	}
	if len(t.noWriteMask) <= index {
		// return the value of the last argument for "..."" functions
		return t.noWriteMask[len(t.noWriteMask)-1]
	}
	return t.noWriteMask[index]
}

func (t *FuncType) PassByValue() bool {
	return false
}

func (t *FuncType) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("func ")
	if t.method {
		buffer.WriteString("(")
		buffer.WriteString(t.receiver.String())
		buffer.WriteString(") ")
	}
	if len(t.name) > 0 {
		buffer.WriteString(t.name)
	}
	buffer.WriteString("(")
	for i, param := range t.params {
		//fmt.Printf("%+v\n", t.params)
		buffer.WriteString(param.String())
		if i < len(t.params)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")
	if len(t.results) > 0 {
		buffer.WriteString(" (")
		for i, result := range t.results {
			buffer.WriteString(result.String())
			if i < len(t.results)-1 {
				buffer.WriteString(", ")
			}
			if t.ellipsis && i == len(t.results)-1 {
				buffer.WriteString("...")
			}
		}
		buffer.WriteString(")")
	}
	if t.body == nil {
		buffer.WriteString(" <builtin>")
	}
	return buffer.String()
}

// Represent multiple return values in a single Type
type MultiType struct {
	*BaseType
	values []Type
}

func newMultiType(values ...Type) *MultiType {
	t := &MultiType{
		BaseType: newBaseType(values[0].Definition()),
		values:   values,
	}
	t.BaseType.underlying = values[0]
	return t
}

func (t *MultiType) Expand() []Type {
	return t.values
}

// A type that will be resolved in the future through .Complete()
type FutureType struct {
	*BaseType
}

func newFutureType() *FutureType {
	return &FutureType{
		BaseType: newBaseType(nil),
	}
}

func (t *FutureType) Finish(base Type) {
	if base == nil {
		panic("BaseType = nil")
	}
	t.BaseType.underlying = base
}

type PackageType struct {
	*BaseType
	resolver Resolver
	path     string
}

func newPackageType(node ast.Node) Type {
	return &PackageType{
		BaseType: newBaseType(node),
	}
}

func (t *PackageType) Complete(resolver Resolver) {
	if t.completed() {
		return
	}
	t.resolver = resolver
	t.path = t.Node.(*ast.ImportSpec).Path.Value
}

func (t *PackageType) Field(name string) Type {
	// TODO: enforce exported fields only?
	return t.resolver(name)
}

func (t *PackageType) Method(name string) *FuncType {
	if funcTyp, ok := t.resolver(name).(*FuncType); ok {
		return funcTyp
	}
	return nil
}

func (t *PackageType) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("package ")
	buffer.WriteString(t.path)
	return buffer.String()
}

// The outcome of a binary operation
func BinaryOp(X Type, op token.Token, Y Type) Type {
	switch op {
	case token.LAND, token.LOR, token.NEQ, token.LEQ, token.GEQ, token.EQL,
		token.LSS, token.GTR:
		return BuiltinTypes["bool"]
	default:
		// Binary operations are always between two of the same types, unless
		// shifting. Untyped constants are converted to the type of the other
		// operand
		_, xConst := X.(*ConstType)
		_, yConst := Y.(*ConstType)
		switch {
		case !xConst && !yConst:
			return X
		case xConst && !yConst:
			return X
		case yConst && !xConst:
			return Y
		default:
			// TODO: see http://golang.org/ref/spec#Constant_expressions
			return X
		}
	}
	return nil
}

// Takes an identifier, returns the node that defines it. This should search all
// scopes up to the package level.
type Resolver func(ident string) Type

// Create a resolver for types
func MakeResolver(block *BasicBlock, p *Package, c *Compiler) Resolver {
	return func(ident string) Type {
		for child := block; child != nil; child = child.parent {
			defineData := child.Get(AccessPassType).(*AccessPassData)
			if result, ok := defineData.defines[ident]; ok {
				return result
			}
		}
		packageScope := c.GetPassResult(DefinedTypesPassType, p).(*DefinedTypesData)
		if identType, ok := packageScope.defined[ident]; ok {
			return identType
		}
		for _, embedded := range packageScope.embedded {
			if identType := embedded.Field(ident); identType != nil {
				return identType
			}
		}
		if identType, ok := BuiltinTypes[ident]; ok {
			return identType
		}
		return nil
	}
}

func TypeOf(expr ast.Node, resolver Resolver) Type {
	fmt.Printf("TypeOf (%T %+v)\n", expr, expr)
	t := typeOf(expr, resolver, false, true)
	fmt.Printf("==> %s\n", t)
	return t
}

// Used to find the types of arguments or definitions, they vary by how they
// handle *pointers
func TypeOfDecl(expr ast.Node, resolver Resolver) Type {
	fmt.Printf("TypeOfDecl (%T %+v)\n", expr, expr)
	t := typeOf(expr, resolver, true, true)
	fmt.Printf("==> %s\n", t)
	return t
}

// Create the new Type for this declaration, but don't Complete() it to avoid
// recursive loops.
func TypeDecl(expr ast.Node, resolver Resolver) Type {
	fmt.Printf("TypeDecl (%T %+v)\n", expr, expr)
	t := typeOf(expr, resolver, true, false)
	fmt.Printf("==> %T\n", t)
	return t
}

func typeOf(expr ast.Node, resolver Resolver, definition bool, complete bool) Type {
	switch t := expr.(type) {
	case *ast.CallExpr:
		callType := TypeOf(t.Fun, resolver)
		// is this a type conversion or a function call?
		switch callType.(type) {
		case *FuncType:
			fmt.Println("CALLING")
			// gather arguments to pass to Call()
			var args []Type
			for _, argExpr := range t.Args {
				args = append(args, TypeOf(argExpr, resolver))
			}
			return callType.Call(args)
		default:
			fmt.Println("TYPE CONV")
			// int32(X)
			// deal with (*int32) parentheses manually
			var retTyp Type
			switch inner := t.Fun.(type) {
			case *ast.ParenExpr:
				retTyp = TypeOfDecl(inner.X, resolver)
			default:
				retTyp = TypeOfDecl(t.Fun, resolver)
			}
			return retTyp
		}
	case *ast.StructType, *ast.InterfaceType:
		structTyp := newStructType(t)
		if complete {
			structTyp.Complete(resolver)
		}
		return structTyp
	case *ast.FuncDecl:
		funcTyp := newFuncType(t.Type, t)
		if complete {
			funcTyp.Complete(resolver)
		}
		return funcTyp
	case *ast.FuncLit:
		funcTyp := newFuncLit(t)
		if complete {
			funcTyp.Complete(resolver)
		}
		return funcTyp
	case *ast.FuncType:
		// for arguments/local variables that are functions
		funcTyp := newFuncType(t, nil)
		if complete {
			funcTyp.Complete(resolver)
		}
		return funcTyp
	case *ast.Ident:
		return resolver(t.Name)
	case *ast.BasicLit:
		constTyp := newConstType(t)
		if complete {
			constTyp.Complete(resolver)
		}
		return constTyp
	case *ast.IndexExpr:
		indexer := TypeOf(t.X, resolver)
		fmt.Println(indexer)
		return indexer.IndexValue()
	case *ast.UnaryExpr:
		// &something
		switch t.Op {
		case token.AND:
			refTyp := newPointerType(&ast.StarExpr{X: t.X})
			if complete {
				refTyp.Complete(resolver)
			}
			return refTyp
		// <-chan
		case token.ARROW:
			chanTyp := TypeOf(t.X, resolver)
			return chanTyp.IndexValue()
		case token.NOT:
			return BuiltinTypes["bool"]
		}
		// token = -,+
		return TypeOf(t.X, resolver)
	case *ast.StarExpr:
		if definition {
			ptrTyp := newPointerType(t)
			if complete {
				ptrTyp.Complete(resolver)
			}
			return ptrTyp
		}
		ptrType := TypeOf(t.X, resolver)
		fmt.Println("Dereferencing", ptrType.String())
		return ptrType.Dereference()
	case *ast.CompositeLit:
		// Something{}
		return TypeOf(t.Type, resolver)
	case *ast.BinaryExpr:
		xTyp := TypeOf(t.X, resolver)
		yTyp := TypeOf(t.Y, resolver)
		fmt.Println(xTyp, t.X, t.Op, yTyp, t.Y)
		result := xTyp.Math(yTyp, t.Op)
		fmt.Println("=>", result)
		return result
	case *ast.ArrayType, *ast.ChanType, *ast.MapType:
		indexTyp := newIndexedType(t)
		if complete {
			indexTyp.Complete(resolver)
		}
		return indexTyp
	case *ast.SelectorExpr:
		innerTyp := TypeOf(t.X, resolver)
		fmt.Printf("%T %+v\n", innerTyp, innerTyp)
		if fieldTyp := innerTyp.Field(t.Sel.Name); fieldTyp != nil {
			return fieldTyp
		}

		// TODO: enforce *T vs T method sets here?
		methodTyp := innerTyp.Method(t.Sel.Name)
		return methodTyp
	case *ast.TypeAssertExpr:
		if t.Type == nil {
			// x.(type) switch statement
			return TypeOf(t.X, resolver)
		}
		assertedTyp := TypeOfDecl(t.Type, resolver)
		// possibly 2 return values, but by default MultiType will act like the
		// first type
		return newMultiType(assertedTyp, BuiltinTypes["bool"])
	case *ast.ParenExpr:
		return TypeOf(t.X, resolver)
	case *ast.Ellipsis:
		return TypeOf(t.Elt, resolver)
	case *ast.SliceExpr:
		sliceTyp := TypeOf(t.X, resolver)
		return newCustomIndexedType(sliceTyp.IndexValue(), sliceTyp.IndexKey(), false)
	default:
		fmt.Printf("Unhandled TypeOf(%T %+v)\n", expr, expr)
	}
	return nil
}

// Helper functions for constructing C/OpenCL structures:
// http://golang.org/ref/spec#Size_and_alignment_guarantees
// http://www.khronos.org/registry/cl/sdk/1.1/docs/man/xhtml/attributes-types.html
// https://code.google.com/p/go/source/browse/go/types/sizes.go?repo=exp
func SizeOf(typ Type) int64 {
	return 1
}

func AlignOf(typ Type) int64 {
	return 1
}
