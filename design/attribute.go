package design

import "github.com/goadesign/goa/eval"

type (
	// AttributeExpr defines a object field with optional description, default value and
	// validations.
	AttributeExpr struct {
		// DSLFunc contains the DSL used to initialize the expression.
		eval.DSLFunc
		// Attribute type
		Type DataType
		// Attribute reference type if any
		Reference DataType
		// Optional description
		Description string
		// Optional validations
		Validation *ValidationExpr
		// Metadata is a list of key/value pairs
		Metadata MetadataExpr
		// Optional member default value
		DefaultValue interface{}
	}

	// CompositeExpr defines a generic composite expression that contains an attribute.
	// This makes it possible for plugins to use attributes in their own data structures.
	CompositeExpr interface {
		// Attribute returns the composite expression embedded attribute.
		Attribute() *AttributeExpr
	}

	// ValidationExpr contains validation rules for an attribute.
	ValidationExpr struct {
		// Values represents an enum validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor76.
		Values []interface{}
		// Format represents a format validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor104.
		Format string
		// PatternValidationExpr represents a pattern validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor33
		Pattern string
		// Minimum represents an minimum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor21.
		Minimum *float64
		// Maximum represents a maximum value validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor17.
		Maximum *float64
		// MinLength represents an minimum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor29.
		MinLength *int
		// MaxLength represents an maximum length validation as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor26.
		MaxLength *int
		// Required list the required fields of object attributes as described at
		// http://json-schema.org/latest/json-schema-validation.html#anchor61.
		Required []string
	}
)

// EvalName returns the name used by the DSL evaluation.
func (a *AttributeExpr) EvalName() string {
	return "attribute"
}

// Walk traverses the data structure recursively and calls the given function once
// on each field starting with the field returned by Expr.
func (a *AttributeExpr) Walk(walker func(*AttributeExpr) error) error {
	return walk(a, walker, make(map[string]bool))
}

// Recursive implementation of the Walk methods. Takes care of avoiding infinite recursions by
// keeping track of types that have already been walked.
func walk(at *AttributeExpr, walker func(*AttributeExpr) error, seen map[string]bool) error {
	if err := walker(at); err != nil {
		return err
	}
	walkUt := func(ut UserType) error {
		if _, ok := seen[ut.Name()]; ok {
			return nil
		}
		seen[ut.Name()] = true
		return walk(ut.Attribute(), walker, seen)
	}
	switch actual := at.Type.(type) {
	case Primitive:
		return nil
	case *Array:
		return walk(actual.ElemType, walker, seen)
	case *Map:
		if err := walk(actual.KeyType, walker, seen); err != nil {
			return err
		}
		return walk(actual.ElemType, walker, seen)
	case Object:
		for _, cat := range actual {
			if err := walk(cat, walker, seen); err != nil {
				return err
			}
		}
	case UserType:
		return walkUt(actual)
	default:
		panic("unknown field type") // bug
	}
	return nil
}

// Context returns the generic definition name used in error messages.
func (v *ValidationExpr) Context() string {
	return "validation"
}

// Merge merges other into v.
func (v *ValidationExpr) Merge(other *ValidationExpr) {
	if v.Values == nil {
		v.Values = other.Values
	}
	if v.Format == "" {
		v.Format = other.Format
	}
	if v.Pattern == "" {
		v.Pattern = other.Pattern
	}
	if v.Minimum == nil || (other.Minimum != nil && *v.Minimum > *other.Minimum) {
		v.Minimum = other.Minimum
	}
	if v.Maximum == nil || (other.Maximum != nil && *v.Maximum < *other.Maximum) {
		v.Maximum = other.Maximum
	}
	if v.MinLength == nil || (other.MinLength != nil && *v.MinLength > *other.MinLength) {
		v.MinLength = other.MinLength
	}
	if v.MaxLength == nil || (other.MaxLength != nil && *v.MaxLength < *other.MaxLength) {
		v.MaxLength = other.MaxLength
	}
	v.AddRequired(other.Required)
}

// AddRequired merges the required fields from other into v
func (v *ValidationExpr) AddRequired(required []string) {
	for _, r := range required {
		found := false
		for _, rr := range v.Required {
			if r == rr {
				found = true
				break
			}
		}
		if !found {
			v.Required = append(v.Required, r)
		}
	}
}

// HasRequiredOnly returns true if the validation only has the Required field with a non-zero value.
func (v *ValidationExpr) HasRequiredOnly() bool {
	if len(v.Values) > 0 {
		return false
	}
	if v.Format != "" || v.Pattern != "" {
		return false
	}
	if (v.Minimum != nil) || (v.Maximum != nil) || (v.MaxLength != nil) {
		return false
	}
	return true
}

// Dup makes a shallow dup of the validation.
func (v *ValidationExpr) Dup() *ValidationExpr {
	return &ValidationExpr{
		Values:    v.Values,
		Format:    v.Format,
		Pattern:   v.Pattern,
		Minimum:   v.Minimum,
		Maximum:   v.Maximum,
		MinLength: v.MinLength,
		MaxLength: v.MaxLength,
		Required:  v.Required,
	}
}
