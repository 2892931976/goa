/*
/ Echo goa app which implements a single "echo" action.
/ Send requests with curl via:
/
/     curl http://localhost:8080/api/echo?value=foo
*/
package main

import (
	. "github.com/raphael/goa"
	"log"
	"net/http"
	"os"
)

// Echo resource - only one action: "echo"
var resource = Resource{
	RoutePrefix: "/echo",
	Actions: Actions{
		"echo": Action{
			Route: GET("?value={value}"), // Capture param in "value"
			Params: Params{
				"value": Attribute{Type: String, Required: true},
			},
			Responses: Responses{
				"ok": Response{Status: 200}, // Only one response code possible
			},
		},
	},
}

// Controller struct
type EchoController struct {
	Controller
}

// Action implementation
func (c *EchoController) Echo(request Request, value string) {
	request.Respond(value) // Send 200 response, use "value" param as body
}

// Entry point
func main() {
	app := New("/api")                      // Create application
	app.Mount(&EchoController{}, &resource) // Mount resource and corresponding controller
	l := log.New(os.Stdout, "[echo] ", 0)
	addr := "localhost:8080"
	l.Printf("listening on %s", addr)
	l.Printf("Routes:")
	app.PrintRoutes()
	l.Printf("  try with `curl http://%s/api/echo?value=foo`", addr)
	l.Fatal(http.ListenAndServe(addr, app)) // app implements http.Handlefunc
}
