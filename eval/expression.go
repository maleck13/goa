package eval

type (
	// Expr built by the engine through the DSL functions.
	Expr interface {
		// EvalName is the qualified name of the DSL expression e.g. "resource bottle".
		EvalName() string
	}

	// A Root expression represents an entry point to the executed DSL: upon execution the
	// DSL engine iterates over all root expressions and calls their IterateSets methods to
	// iterate over the sub-expressions.
	Root interface {
		// DSLName is displayed by the runner upon executing the DSL.
		// Registered DSL roots must have unique DSL names.
		DSLName() string
		// DependsOn returns the list of other DSL roots this root depends on.
		// The engine uses this function to order the execution of the DSL roots.
		DependsOn() []Root
		// IterateSets implements the visitor pattern: is is called by the engine so the
		// DSL can control the order of execution. IterateSets calls back the engine via
		// the given iterator as many times as needed providing the expression sets on each
		// callback.
		IterateSets(SetIterator)
	}

	// A Source expression embeds DSL to be executed after the process is loaded.
	Source interface {
		// DSL returns the DSL used to initialize the expression in a second pass.
		DSL() func()
	}

	// A Validator expression can be validated.
	Validator interface {
		// Validate returns nil if the expression contains no validation error.
		// The Validate implementation may take advantage of ValidationErrors to report
		// more than one errors at a time.
		Validate() error
	}

	// A Finalizer expression requires an additional pass after the DSL has executed and has
	// been validated (e.g. to merge generated expressions or initialize default values).
	Finalizer interface {
		// Finalize is run by the engine as the last step. Finalize cannot fail, any
		// potential failure should be returned by implementing Validator instead.
		Finalize()
	}

	// DSLFunc is a type that DSL expressions may embed to store DSL. It implements Source.
	DSLFunc func()

	// TopExpr is the type of Top.
	TopExpr string

	// ExprSet is a sequence of expressions processed in order. Each DSL implementation
	// provides an arbitrary number of expression sets to the engine via iterators (see the
	// Root interface IterateSets method).
	//
	// The items in the set may implement the Source, Validator and/or Finalizer interfaces to
	// enable the corresponding behaviors during DSL execution. The engine first runs the
	// expression DSLs (for the ones that implement Source) then validates them (for the ones
	// that implement Validator) and finalizes them (for the ones that implement Finalizer).
	ExprSet []Expr

	// SetIterator is the function signature used to iterate over expression sets with
	// IterateSets.
	SetIterator func(s ExprSet) error
)

// Top is the expression returned by Current when the execution stack is empty.
const Top TopExpr = "top-level"

// DSL returns the DSL function.
func (f DSLFunc) DSL() func() {
	return f
}

// Name of top expression is "top-level" to help with error messages.
func (t TopExpr) EvalName() string { return string(t) }
