package dsl_test

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
)

var _ = Describe("Action", func() {
	var name string
	var dsl func()
	var action *ActionDefinition

	BeforeEach(func() {
		Design = nil
		DSLErrors = nil
		name = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		Resource("res", func() {
			Action(name, dsl)
		})
		RunDSL()
		if r, ok := Design.Resources["res"]; ok {
			action = r.Actions[name]
		}
	})

	Context("with only a name", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an invalid action", func() {
			Ω(DSLErrors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).Should(HaveOccurred())
		})
	})

	Context("with a name and DSL defining a route", func() {
		var route = GET("/:id")

		BeforeEach(func() {
			name = "foo"
			dsl = func() { Routing(route) }
		})

		It("produces a valid action definition with the route and default status of 200 set", func() {
			Ω(DSLErrors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Name).Should(Equal(name))
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Routes).ShouldNot(BeNil())
			Ω(action.Routes).Should(HaveLen(1))
			Ω(action.Routes[0]).Should(Equal(route))
		})
	})

	Context("with a name and DSL defining a description, route, headers, payload and responses", func() {
		const typeName = "typeName"
		const description = "description"
		const headerName = "Foo"

		BeforeEach(func() {
			Type(typeName, func() {
				Attribute("name")
			})
			name = "foo"
			dsl = func() {
				Description(description)
				Routing(GET("/:id"))
				Headers(func() { Header(headerName) })
				Payload(typeName)
				Response(NoContent)
			}
		})

		It("produces a valid action with the given properties", func() {
			Ω(DSLErrors).ShouldNot(HaveOccurred())
			Ω(action).ShouldNot(BeNil())
			Ω(action.Validate()).ShouldNot(HaveOccurred())
			Ω(action.Name).Should(Equal(name))
			Ω(action.Description).Should(Equal(description))
			Ω(action.Routes).Should(HaveLen(1))
			Ω(action.Responses).ShouldNot(BeNil())
			Ω(action.Responses).Should(HaveLen(1))
			Ω(action.Responses).Should(HaveKey("NoContent"))
			Ω(action.Headers.Type.(Object)).Should(HaveKey(headerName))
			Ω(action.Headers).ShouldNot(BeNil())
			Ω(action.Headers.Type).Should(BeAssignableToTypeOf(Object{}))
			Ω(action.Headers.Type.(Object)).Should(HaveLen(1))
			Ω(action.Headers.Type.(Object)).Should(HaveKey(headerName))
		})
	})

	Context("using a response template", func() {
		const tmplName = "tmpl"
		const respMediaType = "media"
		const respStatus = 200
		const respName = "respName"

		BeforeEach(func() {
			name = "foo"
			API("test", func() {
				ResponseTemplate(tmplName, func(status, name string) {
					st, err := strconv.Atoi(status)
					if err != nil {
						ReportError(err.Error())
						return
					}
					Status(st)
					Name(name)
				})
			})
		})

		Context("called correctly", func() {
			BeforeEach(func() {
				dsl = func() {
					Routing(GET("/:id"))
					Response(tmplName, strconv.Itoa(respStatus), respName, func() {
						MediaType(respMediaType)
					})
				}
			})

			It("defines the response definition using the template", func() {
				Ω(DSLErrors).ShouldNot(HaveOccurred())
				Ω(action).ShouldNot(BeNil())
				Ω(action.Responses).ShouldNot(BeNil())
				Ω(action.Responses).Should(HaveLen(1))
				Ω(action.Responses).Should(HaveKey(tmplName))
				resp := action.Responses[tmplName]
				Ω(resp.Name).Should(Equal(respName))
				Ω(resp.Status).Should(Equal(respStatus))
				Ω(resp.MediaType).Should(Equal(respMediaType))
			})
		})

		Context("called incorrectly", func() {
			BeforeEach(func() {
				dsl = func() {
					Routing(GET("/id"))
					Response(tmplName, "not an integer", respName, func() {
						MediaType(respMediaType)
					})
				}
			})

			It("fails", func() {
				Ω(DSLErrors).Should(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Payload", func() {
	Context("with a payload definition", func() {
		BeforeEach(func() {
			Design = nil
			Resource("foo", func() {
				Action("bar", func() {
					Payload(func() {
						Member("name")
						Required("name")
					})
				})
			})
		})

		JustBeforeEach(func() {
			RunDSL()
		})

		It("generates the payload type", func() {
			Ω(DSLErrors).ShouldNot(HaveOccurred())
			Ω(Design).ShouldNot(BeNil())
			Ω(Design.Resources).Should(HaveKey("foo"))
			Ω(Design.Resources["foo"].Actions).Should(HaveKey("bar"))
			Ω(Design.Resources["foo"].Actions["bar"].Payload).ShouldNot(BeNil())
			Ω(Design.Resources["foo"].Actions["bar"].Payload).ShouldNot(BeNil())
		})
	})
})
