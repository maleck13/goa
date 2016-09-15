package dsl

import "github.com/goadesign/goa/design"

// ArrayOf creates an array type from its element type.
//
// ArrayOf may be used wherever types can.
// ArrayOf takes one argument: the type of the array elements either by name or by reference.
//
// Example:
//
//    var Bottle = Type("Bottle", func() {
//        Attribute("name")
//    })
//
//    var Account = Type("Account", func() {
//        Attribute("bottles", ArrayOf(Bottle), "Account bottles", func() {
//            MinLength(1)
//        })
//    })
//
func ArrayOf(t design.DataType) *design.Array {
	at := design.AttributeExpr{Type: t}
	return &design.Array{ElemType: &at}
}

// MapOf creates a map from its key and element types.
//
// MapOf may be used wherever types can.
// MapOf takes two arguments: the key and value types either by name of by reference.
//
// Example:
//
//    var Bottle = Type("Bottle", func() {
//        Attribute("name")
//    })
//
//    var Review = Type("Review", func() {
//        Attribute("ratings", MapOf(Bottle, Int32), "Bottle ratings")
//    })
//
func MapOf(k, v design.DataType) *design.Map {
	kat := design.AttributeExpr{Type: k}
	vat := design.AttributeExpr{Type: v}
	return &design.Map{KeyType: &kat, ElemType: &vat}
}
