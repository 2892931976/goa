package main

var _ = goa.registerHandlers(
	goa.handler{"bottles", "list", listBottlesHandler},
	goa.handler{"bottles", "show", showBottlesHandler},
	goa.handler{"bottles", "create", createBottlesHandler},
	goa.handler{"bottles", "update", updateBottlesHandler},
	goa.handler{"bottles", "delete", deleteBottlesHandler},
)

func listBottlesHandler(c *goa.Context) *goa.Response {
	ctx := ListBottleContext{Context: c}
	return c.Action(&ctx)
}

func showBottlesHandler(c *goa.Context) *goa.Response {
	ctx := ShowBottleContext{Context: c}
	return c.Action(&ctx)
}

func createBottlesHandler(c *goa.Context) *goa.Response {
	ctx := CreateBottleContext{Context: c}
	return c.Action(&ctx)
}

func updateBottlesHandler(c *goa.Context) *goa.Response {
	ctx := UpdateBottleContext{Context: c}
	return c.Action(&ctx)
}

func deleteBottlesHandler(c *goa.Context) *goa.Response {
	ctx := DeleteBottleContext{Context: c}
	return c.Action(&ctx)
}
