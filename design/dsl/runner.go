package dsl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	. "github.com/raphael/goa/design"
)

var (
	// Errors contains the DSL execution errors if any.
	Errors MultiError

	// Global DSL evaluation stack
	ctxStack contextStack
)

type (
	// MultiError collects all DSL errors. It implements error.
	MultiError []*Error

	// Error represents an error that occurred while running the API DSL.
	// It contains the name of the file and line number of where the error
	// occurred as well as the original Go error.
	Error struct {
		GoError error
		File    string
		Line    int
	}

	// DSL evaluation contexts stack
	contextStack []DSLDefinition
)

// RunDSL runs all the registered top level DSLs and returns any error.
// This function is called by the client package init.
// goagen creates that function during code generation.
func RunDSL() error {
	if Design == nil {
		return nil
	}
	Errors = nil
	// First run the top level API DSL to initialize responses and
	// response templates needed by resources.
	executeDSL(Design.DSL, Design)
	// Then run the user type DSLs
	for _, t := range Design.Types {
		executeDSL(t.DSL, t.AttributeDefinition)
	}
	// Then the media type DSLs
	for _, mt := range Design.MediaTypes {
		executeDSL(mt.DSL, mt)
	}
	// And now that we have everything the resources.
	for _, r := range Design.Resources {
		executeDSL(r.DSL, r)
	}

	// Validate DSL
	if err := Design.Validate(); err != nil {
		return err
	}
	if Errors != nil {
		return Errors
	}

	// Second pass post-validation does final merges with defaults and base types.
	for _, t := range Design.Types {
		finalizeType(t)
	}
	for _, mt := range Design.MediaTypes {
		finalizeMediaType(mt)
	}
	for _, r := range Design.Resources {
		finalizeResource(r)
	}

	return nil
}

// Current evaluation context, i.e. object being currently built by DSL
func (s contextStack) current() DSLDefinition {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// Error returns the error message.
func (m MultiError) Error() string {
	msgs := make([]string, len(m))
	for i, de := range m {
		msgs[i] = de.Error()
	}
	return strings.Join(msgs, "\n")
}

// Error returns the underlying error message.
func (de *Error) Error() (res string) {
	if err := de.GoError; err != nil {
		res = fmt.Sprintf("[%s:%d] %s", de.File, de.Line, err.Error())
	}
	return
}

// executeDSL runs DSL in given evaluation context and returns true if successful.
// It appends to Errors in case of failure (and returns false).
func executeDSL(dsl func(), ctx DSLDefinition) bool {
	if dsl == nil {
		return true
	}
	initCount := len(Errors)
	ctxStack = append(ctxStack, ctx)
	dsl()
	ctxStack = ctxStack[:len(ctxStack)-1]
	return len(Errors) <= initCount
}

// finalizeMediaType merges any base type attribute into the media type attributes
func finalizeMediaType(mt *MediaTypeDefinition) {
	if mt.BaseType != nil {
		if bat := mt.AttributeDefinition; bat != nil {
			mt.AttributeDefinition.Inherit(bat)
		}
	}
}

// finalizeType merges any base type attribute into the type attributes
func finalizeType(ut *UserTypeDefinition) {
	if ut.BaseType != nil {
		if bat := ut.AttributeDefinition; bat != nil {
			ut.AttributeDefinition.Inherit(bat)
		}
	}
}

// finalizeResource makes the final pass at the resource DSL. This is needed so that the order
// of DSL function calls is irrelevant. For example a resource response may be defined after an
// action refers to it.
func finalizeResource(r *ResourceDefinition) {
	r.IterateActions(func(a *ActionDefinition) error {
		// 1. Merge response definitions
		for name, resp := range a.Responses {
			if pr, ok := a.Parent.Responses[name]; ok {
				resp.Merge(pr)
			}
			if ar, ok := Design.Responses[name]; ok {
				resp.Merge(ar)
			}
			if dr, ok := Design.DefaultResponses[name]; ok {
				resp.Merge(dr)
			}
		}
		// 2. Create implicit action parameters for path wildcards that dont' have one
		for _, r := range a.Routes {
			wcs := ExtractWildcards(r.FullPath())
			for _, wc := range wcs {
				found := false
				var o Object
				if a.Params != nil {
					o = a.Params.Type.ToObject()
				} else {
					o = Object{}
					a.Params = &AttributeDefinition{Type: o}
				}
				for n := range o {
					if n == wc {
						found = true
						break
					}
				}
				if !found {
					o[wc] = &AttributeDefinition{Type: String}
				}
			}
		}
		return nil
	})
}

// incompatibleDSL should be called by DSL functions when they are
// invoked in an incorrect context (e.g. "Params" in "Resource").
func incompatibleDSL(dslFunc string) {
	elems := strings.Split(dslFunc, ".")
	ReportError("invalid use of %s", elems[len(elems)-1])
}

// invalidArgError records an invalid argument error.
// It is used by DSL functions that take dynamic arguments.
func invalidArgError(expected string, actual interface{}) {
	ReportError("cannot use %#v (type %s) as type %s",
		actual, reflect.TypeOf(actual), expected)
}

// ReportError records a DSL error for reporting post DSL execution.
func ReportError(fm string, vals ...interface{}) {
	var suffix string
	if cur := ctxStack.current(); cur != nil {
		suffix = fmt.Sprintf(" in %s", cur.Context())
	} else {
		suffix = " (top level)"
	}
	err := fmt.Errorf(fm+suffix, vals...)
	file, line := computeErrorLocation()
	Errors = append(Errors, &Error{
		GoError: err,
		File:    file,
		Line:    line,
	})
}

// computeErrorLocation implements a heuristic to find the location in the user
// code where the error occurred. It walks back the callstack until the file
// doesn't match "/goa/design/*.go".
// When successful it returns the file name and line number, empty string and
// 0 otherwise.
func computeErrorLocation() (file string, line int) {
	depth := 2
	_, file, line, _ = runtime.Caller(depth)
	ok := strings.HasSuffix(file, "_test.go") // Be nice with tests
	if !ok {
		nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
		ok = !nok
	}
	for !ok {
		depth++
		_, file, line, _ = runtime.Caller(depth)
		ok = strings.HasSuffix(file, "_test.go")
		if !ok {
			nok, _ := regexp.MatchString(`/goa/design/.+\.go$`, file)
			ok = !nok
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		return
	}
	f, err := filepath.Rel(wd, file)
	if err != nil {
		return
	}
	file = f
	return
}

// topLevelDefinition returns true if the currently evaluated DSL is a root
// DSL (i.e. is not being run in the context of another definition).
func topLevelDefinition(failItNotTopLevel bool) bool {
	top := ctxStack.current() == nil
	if failItNotTopLevel && !top {
		incompatibleDSL(caller())
	}
	return top
}

// actionDefinition returns true and current context if it is an ActionDefinition,
// nil and false otherwise.
func actionDefinition(failIfNotAction bool) (*ActionDefinition, bool) {
	a, ok := ctxStack.current().(*ActionDefinition)
	if !ok && failIfNotAction {
		incompatibleDSL(caller())
	}
	return a, ok
}

// apiDefinition returns true and current context if it is an APIDefinition,
// nil and false otherwise.
func apiDefinition(failIfNotAPI bool) (*APIDefinition, bool) {
	a, ok := ctxStack.current().(*APIDefinition)
	if !ok && failIfNotAPI {
		incompatibleDSL(caller())
	}
	return a, ok
}

// mediaTypeDefinition returns true and current context if it is a MediaTypeDefinition,
// nil and false otherwise.
func mediaTypeDefinition(failIfNotMT bool) (*MediaTypeDefinition, bool) {
	m, ok := ctxStack.current().(*MediaTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDSL(caller())
	}
	return m, ok
}

// typeDefinition returns true and current context if it is a UserTypeDefinition,
// nil and false otherwise.
func typeDefinition(failIfNotMT bool) (*UserTypeDefinition, bool) {
	m, ok := ctxStack.current().(*UserTypeDefinition)
	if !ok && failIfNotMT {
		incompatibleDSL(caller())
	}
	return m, ok
}

// attribute returns true and current context if it is an Attribute,
// nil and false otherwise.
func attributeDefinition(failIfNotAttribute bool) (*AttributeDefinition, bool) {
	a, ok := ctxStack.current().(*AttributeDefinition)
	if !ok && failIfNotAttribute {
		incompatibleDSL(caller())
	}
	return a, ok
}

// resourceDefinition returns true and current context if it is a ResourceDefinition,
// nil and false otherwise.
func resourceDefinition(failIfNotResource bool) (*ResourceDefinition, bool) {
	r, ok := ctxStack.current().(*ResourceDefinition)
	if !ok && failIfNotResource {
		incompatibleDSL(caller())
	}
	return r, ok
}

// responseDefinition returns true and current context if it is a ResponseDefinition,
// nil and false otherwise.
func responseDefinition(failIfNotResponse bool) (*ResponseDefinition, bool) {
	r, ok := ctxStack.current().(*ResponseDefinition)
	if !ok && failIfNotResponse {
		incompatibleDSL(caller())
	}
	return r, ok
}

// Name of calling function.
func caller() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "<unknown>"
	}
	return runtime.FuncForPC(pc).Name()
}
