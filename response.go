package goa

import (
	"net/http"
	"regexp"
)

// Convenience wrapper
type Headers map[string]string

// An action response
type Response struct {
	Status    int       // Response status code
	MediaType MediaType // Response media type if any
	Headers   Headers   // Response header validations, enclose values in '/' for regexp behavior
}

// WithMediaType sets the response MediaType field.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithMediaType(m MediaType) Response {
	r.MediaType = m
	return r
}

// WithLocation sets the response Location field.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithLocation(l string) Response {
	return r.WithHeader("Location", l)
}

// WithHeader initializes the response Header field if not initialized yet and
// sets the given header with the given value.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithHeader(name string, value string) Response {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[name] = value
	return r
}

// Provide helper methods for creating HTTP response from status
type f int

// HTTP Response factory
var Http f

// Factory methods

func (f) Continue() Response                   { return vanillaResponse(100) }
func (f) Ok() Response                         { return vanillaResponse(200) }
func (f) Created() Response                    { return vanillaResponse(201) }
func (f) Accepted() Response                   { return vanillaResponse(202) }
func (f) NonAuthoritative() Response           { return vanillaResponse(203) }
func (f) NoContent() Response                  { return vanillaResponse(204) }
func (f) ResetContent() Response               { return vanillaResponse(205) }
func (f) PartialContent() Response             { return vanillaResponse(206) }
func (f) MultipleChoices() Response            { return vanillaResponse(300) }
func (f) MovedPermanently() Response           { return vanillaResponse(301) }
func (f) Found() Response                      { return vanillaResponse(302) }
func (f) SeeOther() Response                   { return vanillaResponse(303) }
func (f) NotModified() Response                { return vanillaResponse(304) }
func (f) UseProxy() Response                   { return vanillaResponse(305) }
func (f) TemporaryRedirect() Response          { return vanillaResponse(307) }
func (f) BadRequest() Response                 { return vanillaResponse(400) }
func (f) Unauthorized() Response               { return vanillaResponse(401) }
func (f) PaymentRequired() Response            { return vanillaResponse(402) }
func (f) Forbidden() Response                  { return vanillaResponse(403) }
func (f) NotFound() Response                   { return vanillaResponse(404) }
func (f) MethodNotAllowed() Response           { return vanillaResponse(405) }
func (f) NotAcceptable() Response              { return vanillaResponse(406) }
func (f) ProxyAuthRequired() Response          { return vanillaResponse(407) }
func (f) RequestTimeout() Response             { return vanillaResponse(408) }
func (f) Conflict() Response                   { return vanillaResponse(409) }
func (f) Gone() Response                       { return vanillaResponse(410) }
func (f) LengthRequired() Response             { return vanillaResponse(411) }
func (f) PreconditionFailed() Response         { return vanillaResponse(412) }
func (f) RequestEntityTooLarge() Response      { return vanillaResponse(413) }
func (f) RequestUriTooLong() Response          { return vanillaResponse(414) }
func (f) UnsupportedMediaType() Response       { return vanillaResponse(415) }
func (f) RequestRangeNotSatisfiable() Response { return vanillaResponse(416) }
func (f) ExpectationFailed() Response          { return vanillaResponse(417) }
func (f) InternalError() Response              { return vanillaResponse(500) }
func (f) NotImplemented() Response             { return vanillaResponse(501) }
func (f) BadGateway() Response                 { return vanillaResponse(502) }
func (f) ServiceUnavailable() Response         { return vanillaResponse(503) }
func (f) GatewayTimeout() Response             { return vanillaResponse(504) }
func (f) HTTPVersionNotSupported() Response    { return vanillaResponse(505) }

// vanillaResponse returns a default response for the given HTTP status code
func vanillaResponse(status int) Response {
	return Response{Status: status, Description: http.StatusText(status)}
}

// matches checks whether string value matches definition, returns true if it
// does, false otherwise. If definition is a string that starts and ends with
// "/" then value is matched against a regexp built from definition otherwise
// value is compared directly with definition
func (d *Response) matches(value, match string) bool {
	ok := false
	matches := matchRegexp.FindStringSubmatch(match)
	if len(matches) > 0 {
		ok, _ = regexp.MatchString(matches[1], value)
	} else {
		ok = (value == match)
	}
	return ok
}

// Regexp used to check whether a string represents a regexp - quite meta really
var matchRegexp = regexp.MustCompile("^/(.*)/$")
