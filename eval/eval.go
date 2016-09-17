package eval

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// RunDSL iterates through the root expressions and calls IterateSets on each to retrieve the
// expression sets. It iterates over the expression sets multiple times to first execute the DSL,
// then validate the resulting expressions and lastly to finalize them. The executed DSL may
// register additional roots during initial execution via Register to have them be executed (last)
// in the same run.
func RunDSL() error {
	roots, err := Context.SortRoots()
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return nil
	}
	executed := 0
	recursed := 0
	for executed < len(roots) {
		recursed++
		start := executed
		executed = len(roots)
		for _, root := range roots[start:] {
			root.IterateSets(runSet)
		}
		if recursed > 100 {
			// Let's cross that bridge once we get there
			return fmt.Errorf("too many generated roots, infinite loop?")
		}
	}
	if Context.Errors != nil {
		return Context.Errors
	}
	for _, root := range roots {
		root.IterateSets(validateSet)
	}
	if Context.Errors != nil {
		return Context.Errors
	}
	for _, root := range roots {
		root.IterateSets(finalizeSet)
	}

	return nil
}

// runSet executes the DSL for all expressions in the given set. The expression DSLs may append to
// the set as they execute.
func runSet(set ExprSet) error {
	executed := 0
	recursed := 0
	for executed < len(set) {
		recursed++
		for _, def := range set[executed:] {
			executed++
			if source, ok := def.(Source); ok {
				Execute(source.DSL(), def)
			}
		}
		if recursed > 100 {
			return fmt.Errorf("too many generated expressions, infinite loop?")
		}
	}
	return nil
}

// validateSet runs the validation on all the set expressions that define one.
func validateSet(set ExprSet) error {
	errors := &ValidationErrors{}
	for _, def := range set {
		if validate, ok := def.(Validator); ok {
			if err := validate.Validate(); err != nil {
				errors.AddError(def, err)
			}
		}
	}
	if len(errors.Errors) > 0 {
		Context.Record(&Error{GoError: errors})
	}
	return Context.Errors
}

// finalizeSet runs the validation on all the set expressions that define one.
func finalizeSet(set ExprSet) error {
	for _, def := range set {
		if f, ok := def.(Finalizer); ok {
			f.Finalize()
		}
	}
	return nil
}

// Execute runs the given DSL to initialize the given expression. It returns true on success.
// It returns false and appends to i.Errors on failure.
// Note that Run takes care of calling Execute on all expressions that implement Source.
// This function is intended for use by expressions that run the DSL at declaration time rather than
// store the DSL for execution by the dsl engine (usually simple independent expressions).
// The DSL should use ReportError to record DSL execution errors.
func Execute(dsl func(), def Expr) bool {
	if dsl == nil {
		return true
	}
	initCount := len(Context.Errors.(MultiError))
	Context.Stack = append(Context.Stack, def)
	dsl()
	Context.Stack = Context.Stack[:len(Context.Stack)-1]
	return len(Context.Errors.(MultiError)) <= initCount
}

// Current returns the expression whose DSL is currently being executed.
// As a special case Current returns Top when the execution stack is empty.
func Current() Expr {
	current := Context.Stack.Current()
	if current == nil {
		return Top
	}
	return current
}

// ReportError records a DSL error for reporting post DSL execution.  It accepts a format and values
// a la fmt.Printf.
func ReportError(fm string, vals ...interface{}) {
	var suffix string
	if cur := Context.Stack.Current(); cur != nil {
		if name := cur.EvalName(); name != "" {
			suffix = fmt.Sprintf(" in %s", name)
		}
	} else {
		suffix = " (top level)"
	}
	err := fmt.Errorf(fm+suffix, vals...)
	file, line := computeErrorLocation()
	Context.Record(&Error{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// IncompatibleDSL should be called by DSL functions when they are invoked in an incorrect context
// (e.g. "Params" in "Resource").
func IncompatibleDSL() {
	elems := strings.Split(caller(), ".")
	ReportError("invalid use of %s", elems[len(elems)-1])
}

// InvalidArgError records an invalid argument error.  It is used by DSL functions that take dynamic
// arguments.
func InvalidArgError(expected string, actual interface{}) {
	ReportError("cannot use %#v (type %s) as type %s", actual, reflect.TypeOf(actual), expected)
}

// ValidationErrors records the errors encountered when running Validate.
type ValidationErrors struct {
	Errors []error
	Exprs  []Expr
}

// Error implements the error interface.
func (verr *ValidationErrors) Error() string {
	msg := make([]string, len(verr.Errors))
	for i, err := range verr.Errors {
		msg[i] = fmt.Sprintf("%s: %s", verr.Exprs[i].EvalName(), err)
	}
	return strings.Join(msg, "\n")
}

// Merge merges validation errors into the target.
func (verr *ValidationErrors) Merge(err *ValidationErrors) {
	if err == nil {
		return
	}
	verr.Errors = append(verr.Errors, err.Errors...)
	verr.Exprs = append(verr.Exprs, err.Exprs...)
}

// Add adds a validation error to the target.
func (verr *ValidationErrors) Add(def Expr, format string, vals ...interface{}) {
	verr.AddError(def, fmt.Errorf(format, vals...))
}

// AddError adds a validation error to the target.
// AddError "flattens" validation errors so that the recorded errors are never ValidationErrors
// themselves.
func (verr *ValidationErrors) AddError(def Expr, err error) {
	if v, ok := err.(*ValidationErrors); ok {
		verr.Errors = append(verr.Errors, v.Errors...)
		verr.Exprs = append(verr.Exprs, v.Exprs...)
		return
	}
	verr.Errors = append(verr.Errors, err)
	verr.Exprs = append(verr.Exprs, def)
}

// caller returns the name of calling function.
func caller() string {
	pc, file, _, ok := runtime.Caller(2)
	if ok && filepath.Base(file) == "current.go" {
		pc, _, _, ok = runtime.Caller(3)
	}
	if !ok {
		return "<unknown>"
	}

	return runtime.FuncForPC(pc).Name()
}
