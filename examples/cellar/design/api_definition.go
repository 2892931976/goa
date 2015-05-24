package main

import . "github.com/raphael/goa/design"

var _ = API("cellar", func() {

	Title("The virtual wine cellar")
	Description("A basic example of a CRUD API implemented with goa")
	BasePath("/:accountID")

	BaseParams(
		Param("accountID", Integer,
			"API request account. All actions operate on resources belonging to the account."),
	)

	ResponseTemplate("NotFound", func() {
		Description("Resource not found")
		Status(404)
		MediaType("application/json")
	})

	ResponseTemplate("Ok", func(mt string) {
		Description("Resource listing")
		Status(200)
		MediaType(mt)
	})

	ResponseTemplate("Created", func() {
		Description("Resource created")
		Status(201)
	})

	Trait("Authenticated", func() {
		Headers(
			Header("Auth-Token", Required()),
		)
	})
})
