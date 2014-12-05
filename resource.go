package goa

import "fmt"

// Resource definitions describe REST resources exposed by the application API.
// They can be versioned so that multiple versions can be exposed (usually for backwards compatibility). Clients
// specify the version they want to use through the X-API-VERSION request header. If an api version is specified in the
// resource definition then clients must specify the version header or get back a response with status code 404.
//
// The definition also includes a description, a route prefix, a media type and action definitions.
// The route prefix is the common path under which all resource actions are located. The complete URL for an action
// is application path + controller prefix + resource prefix + action path.
// The media type describes the fields of the resource (see media_type.go). The resource actions may define a
// different media type for their responses, typically the "index" and "show" actions re-use the resource media type.
// Action definitions list all the actions supported by the resource (both CRUD and other actions), see the Action
// struct.
type Resource struct {
	Description string
	ApiVersion  string
	RoutePrefix string
	MediaType   MediaType
	Actions     map[string]Action

	controller Controller
	pActions   map[string]*Action // Avoid copying action objects once resource is mounted
}

// Action definitions define a route which consists of one ore more pairs of HTTP verb and path. They also optionally
// define the action parameters (variables defined in the route path) and payload (request body content). Parameters
// and payload are described using attributes which may include validations. goa takes care of validating and coercing
// the parameters and payload fields (and returns a response with status code 400 and a description of the validation
// error in the body in case of failure).
//
// The Multipart field specifies whether the request body must(RequiresMultipart) or can (SupportsMultipart) use a
// multipart content type. Multipart requests can be used to implement bulk actions - for example bulk updates. Each
// part contains the payload for a single resource, the same payload that would be used to apply the action to that
// resource in a standard (non-multipart) request.
//
// Action definitions may also specify a list of supported filters - for example an index action may support filtering
// the list of results given resource field values. Filters are defined using attributes, they are specified by the
// client using the special "filters" URL query string, the syntax is:
//
//   "?filters[]=some_field==some_value&&filters[]=other_field==other_value"
//
// Filters are readily available to the action implementation after they have been validated and coerced by goa. The
// exact semantic is up to the action implementation.
//
// Action definitions also specify the set of views supported by the action. Different views may render the media type
// differently (ommitting certain attributes or links, see media_type.go). As with filters the client specifies the
// view in the special "view" URL query string:
//
//  "?view=tiny"
//
// Finally, action definitions describe the set of potential responses they may return and for each response the status
// code, compulsory headers and a media type (if different from the resource media type). These response definitions
// are named so that the action implementation can create a response from its definition name.
type Action struct {
	Name        string
	Description string
	Route       Route
	Params      Attributes
	Payload     Attributes
	Filters     Attributes
	Views       []string
	Responses   Responses
	Multipart   int

	// Internal fields

	pResponses map[string]*Response // Avoid copying response objects once resource is mounted
	resource   *Resource            // Parent resource definition, initialized by goa
}

// ValidateResponse checks that the response content matches one of the action response definitions if any
func (a *Action) ValidateResponse(data ResponseData) error {
	if len(a.pResponses) == 0 {
		return nil
	}
	// We cheat a little here, if the response is a standardResponse and its definition field is initialized then use
	// that - otherwise try all candidate definitions
	if std, ok := data.(*standardResponse); ok {
		if r := std.definition; r != nil {
			return r.Validate(data)
		}
	}
	for _, r := range a.pResponses {
		if err := r.Validate(data); err == nil {
			return nil
		}
	}
	return fmt.Errorf("Response %v does not match any of action '%s' response definitions", data, a.Name)
}

// Interface implemented by action route
type Route interface {
	GetRawRoutes() [][]string // Retrieve pair of HTTP verb and action path
}

// Possible values for the Action struct "Multipart" field
const (
	SupportsMultipart = iota // Action request body may use multipart content type
	RequiresMultipart        // Action request body must use multipart content type
)

// Map of action definitions keyed by action name
type Actions map[string]Action

// Map of response definitions keyed by response name
type Responses map[string]Response

// HTTP verbs enum type
type httpVerb string

//  Route struct
type SingleRoute struct {
	Verb httpVerb // Route HTTP verb
	Path string   // Route path
}

// HTTP Verbs enum
const (
	options httpVerb = "OPTIONS"
	get     httpVerb = "GET"
	head    httpVerb = "HEAD"
	post    httpVerb = "POST"
	put     httpVerb = "PUT"
	delete_ httpVerb = "DELETE"
	trace   httpVerb = "TRACE"
	connect httpVerb = "CONNECT"
	patch   httpVerb = "PATCH"
)

// OPTIONS creates a route with OPTIONS verb and given path
func OPTIONS(path string) Route {
	return SingleRoute{options, path}
}

// GET creates a route with OPTIONS verb and given path
func GET(path string) Route {
	return SingleRoute{get, path}
}

// HEAD creates a route with OPTIONS verb and given path
func HEAD(path string) Route {
	return SingleRoute{head, path}
}

// POST creates a route with OPTIONS verb and given path
func POST(path string) Route {
	return SingleRoute{post, path}
}

// PUT creates a route with OPTIONS verb and given path
func PUT(path string) Route {
	return SingleRoute{put, path}
}

// DELETE creates a route with OPTIONS verb and given path
func DELETE(path string) Route {
	return SingleRoute{delete_, path}
}

// TRACE creates a route with OPTIONS verb and given path
func TRACE(path string) Route {
	return SingleRoute{trace, path}
}

// CONNECT creates a route with OPTIONS verb and given path
func CONNECT(path string) Route {
	return SingleRoute{connect, path}
}

// PATCH creates a route with OPTIONS verb and given path
func PATCH(path string) Route {
	return SingleRoute{patch, path}
}

// A multi-route is an array of routes
type MultiRoute []SingleRoute

// Multi creates a multi-route from the given list of routes
func Multi(routes ...SingleRoute) MultiRoute {
	return MultiRoute(routes)
}

// GetRawRoutes returns the pair of HTTP verb and path for the route
func (r SingleRoute) GetRawRoutes() [][]string {
	return [][]string{{string(r.Verb), r.Path}}
}

// GetRawRoutes returns the list of pairs of HTTP verb and path for the multi-route
func (m MultiRoute) GetRawRoutes() [][]string {
	routes := make([][]string, len(m))
	for _, r := range m {
		routes = append(routes, []string{string(r.Verb), r.Path})
	}
	return routes
}
