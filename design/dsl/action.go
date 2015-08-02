package dsl

import . "github.com/raphael/goa/design"

// Action defines an action definition DSL.
//
// Action("Update", func() {
//     Description("Update account")
//     Routing(
//         PUT("/:id"),
//         PUT("/organizations/:org/accounts/:id"),
//     )
//     Headers(func() {
//         Header("Authorization", String)
//         Header("X-Account", Integer)
//         Required("Authorization", "X-Account")
//     })
//     Params(func() {
//         Param("id", Integer, "Account ID")
//         Required("id")
//     })
//     Payload(func() {
//         Member("name")
//         Member("year")
//     })
//     Responses(
//         NoContent(),
//         NotFound(),
//     )
// })
func Action(name string, dsl func()) {
	if r, ok := resourceDefinition(); ok {
		action := &design.ActionDefinition{Name: name}
		if !executeDSL(dsl, action) {
			return
		}
		r.Actions[name] = action
	}
}

// Routing adds one or more routes to the action
func Routing(routes ...*RouteDefinition) {
	if a, ok := actionDefinition(); ok {
		a.Routes = append(a.Routes, routes...)
	}
}

// GET creates a route using the GET HTTP method
func GET(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "GET", Path: path}
}

// HEAD creates a route using the HEAD HTTP method
func HEAD(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "HEAD", Path: path}
}

// POST creates a route using the POST HTTP method
func POST(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "POST", Path: path}
}

// PUT creates a route using the PUT HTTP method
func PUT(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "PUT", Path: path}
}

// DELETE creates a route using the DELETE HTTP method
func DELETE(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "DELETE", Path: path}
}

// TRACE creates a route using the TRACE HTTP method
func TRACE(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "TRACE", Path: path}
}

// CONNECT creates a route using the GET HTTP method
func CONNECT(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "CONNECT", Path: path}
}

// PATCH creates a route using the PATCH HTTP method
func PATCH(path string) *RouteDefinition {
	return &RouteDefinition{Verb: "PATCH", Path: path}
}

// Headers computes the action headers from the given DSL.
func Headers(dsl func()) {
	if a, ok := actionDefinition(); ok {
		headers := new(AttributeDefinition)
		if executeDSL(dsl, headers) {
			a.Headers = headers
		}
	}
}

// Params computes the action parameters from the given DSL.
func Params(dsl func()) {
	if a, ok := actionDefinition(); ok {
		params := new(AttributeDefinition)
		if executeDSL(dsl, params) {
			a.Params = params
		}
	}
}

// Payload sets the action payload attributes.
func Payload(dsl func()) {
	if a, ok := actionDefinition(); ok {
		payload := new(AttributeDefinition)
		if executeDSL(dsl, payload) {
			a.Payload = payload
		}
	}
}

// Response records a possible action response.
func Response(resp *ResponseDefinition) {
	if a, ok := actionDefinition(); ok {
		for _, r := range a.Responses {
			if r.Status == resp.Status {
				fail
			}
		}
		a.Responses = append(a.Responses, resp)
	}
}
