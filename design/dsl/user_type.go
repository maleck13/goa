package dsl

// Type describes a user type.
//
// Type is a top level definition.
// Type takes two arguments: the type name and the defining DSL.
//
// Example:
//
//     var SumPayload = Type("SumPayload", func() {
//         Description("Type sent to add endpoint")
//
//         Attribute("a", String)                 // Define string field "a"
//         Attribute("b", Int32, "operand")       // Define int32 field "b" with description
//         Attribute("operands", ArrayOf(Int32))  // Define int32 array field
//         Attribute("ops", MapOf(String, Int32)) // Define map[string]int32 field
//         Attribute("c", SumMod)                 // Define field using user type
//
//         Required("a")                      // Required fields must be present
//         Required("b", "c")                 // in serialized data.
//     })
//
func Type(name string, dsl func()) {
}
