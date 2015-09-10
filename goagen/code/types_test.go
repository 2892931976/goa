package code_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/code"
)

var _ = Describe("code generation", func() {
	BeforeEach(func() {
		code.TempCount = 0
	})

	Describe("GoTypeDef", func() {
		Context("given an attribute definition with fields", func() {
			var att *design.AttributeDefinition
			var object design.Object
			var required *design.RequiredValidationDefinition
			var st string

			JustBeforeEach(func() {
				att = new(design.AttributeDefinition)
				att.Type = object
				if required != nil {
					att.Validations = []design.ValidationDefinition{required}
				}
				st = code.GoTypeDef(att, 0, true, false)
			})

			Context("of primitive types", func() {
				BeforeEach(func() {
					object = design.Object{
						"foo": &design.AttributeDefinition{Type: design.Integer},
						"bar": &design.AttributeDefinition{Type: design.String},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Bar string `json:\"bar,omitempty\"`\n" +
						"	Foo int `json:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

			Context("of array of primitive types", func() {
				BeforeEach(func() {
					elemType := &design.AttributeDefinition{Type: design.Integer}
					array := &design.Array{ElemType: elemType}
					object = design.Object{
						"foo": &design.AttributeDefinition{Type: array},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					Ω(st).Should(Equal("struct {\n\tFoo []int `json:\"foo,omitempty\"`\n}"))
				})
			})

			Context("of array of objects", func() {
				BeforeEach(func() {
					obj := design.Object{
						"bar": &design.AttributeDefinition{Type: design.Integer},
					}
					elemType := &design.AttributeDefinition{Type: obj}
					array := &design.Array{ElemType: elemType}
					object = design.Object{
						"foo": &design.AttributeDefinition{Type: array},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo []*struct {\n" +
						"		Bar int `json:\"bar,omitempty\"`\n" +
						"	} `json:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

			Context("that are required", func() {
				BeforeEach(func() {
					object = design.Object{
						"foo": &design.AttributeDefinition{Type: design.Integer},
					}
					required = &design.RequiredValidationDefinition{
						Names: []string{"foo"},
					}
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo int `json:\"foo\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

		})

		Context("given an array", func() {
			var elemType *design.AttributeDefinition
			var source string

			JustBeforeEach(func() {
				array := &design.Array{ElemType: elemType}
				att := &design.AttributeDefinition{Type: array}
				source = code.GoTypeDef(att, 0, true, false)
			})

			Context("of primitive type", func() {
				BeforeEach(func() {
					elemType = &design.AttributeDefinition{Type: design.Integer}
				})

				It("produces the array go code", func() {
					Ω(source).Should(Equal("[]int"))
				})

			})

			Context("of object type", func() {
				BeforeEach(func() {
					object := design.Object{
						"foo": &design.AttributeDefinition{Type: design.Integer},
						"bar": &design.AttributeDefinition{Type: design.String},
					}
					elemType = &design.AttributeDefinition{Type: object}
				})

				It("produces the array go code", func() {
					Ω(source).Should(Equal("[]*struct {\n\tBar string `json:\"bar,omitempty\"`\n\tFoo int `json:\"foo,omitempty\"`\n}"))
				})
			})
		})

	})

	Describe("Unmarshaler", func() {
		var unmarshaler string
		var context, source, target string

		BeforeEach(func() {
			context = ""
			source = "raw"
			target = "p"
		})

		Context("with a primitive type", func() {
			var p design.Primitive

			JustBeforeEach(func() {
				unmarshaler = code.PrimitiveUnmarshaler(p, context, source, target)
			})

			Context("integer", func() {
				BeforeEach(func() {
					p = design.Primitive(design.IntegerKind)
				})

				It("generates the unmarshaler code", func() {
					expected := `	if val, ok := raw.(int); ok {
		p = val
	} else {
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "int")
	}`
					Ω(unmarshaler).Should(Equal(expected))
				})
			})

			Context("string", func() {
				BeforeEach(func() {
					p = design.Primitive(design.StringKind)
				})

				It("generates the unmarshaler code", func() {
					expected := `	if val, ok := raw.(string); ok {
		p = val
	} else {
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "string")
	}`
					Ω(unmarshaler).Should(Equal(expected))
				})
			})
		})

		Context("with an array of primitive types", func() {
			var p *design.Array

			JustBeforeEach(func() {
				unmarshaler = code.ArrayUnmarshaler(p, context, source, target)
			})

			BeforeEach(func() {
				p = &design.Array{
					ElemType: &design.AttributeDefinition{
						Type: design.Primitive(design.IntegerKind),
					},
				}
			})

			It("generates the unmarshaler code", func() {
				expected := `	if val, ok := raw.([]interface{}); ok {
		p = make([]int, len(val))
		for i, v := range val {
			var tmp1 int
			if val, ok := v.(int); ok {
				tmp1 = val
			} else {
				err = goa.IncompatibleTypeError(` + "`" + `[*]` + "`" + `, v, "int")
			}
			p[i] = tmp1
		}
	} else {
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "[]interface{}")
	}`
				Ω(unmarshaler).Should(Equal(expected))
			})
		})

		Context("with a simple object", func() {
			var o design.Object

			JustBeforeEach(func() {
				unmarshaler = code.ObjectUnmarshaler(o, context, source, target)
			})

			BeforeEach(func() {
				intAtt := &design.AttributeDefinition{Type: design.Primitive(design.IntegerKind)}
				o = design.Object{"foo": intAtt}
			})

			It("generates the unmarshaler code", func() {
				expected := `	if val, ok := raw.(map[string]interface{}); ok {
		p = new(struct {
			Foo int
		})
		if v, ok := val["foo"]; ok {
			var tmp1 int
			if val, ok := v.(int); ok {
				tmp1 = val
			} else {
				err = goa.IncompatibleTypeError(` + "`" + `.Foo` + "`" + `, v, "int")
			}
			p.Foo = tmp1
		}
	} else {
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "map[string]interface{}")
	}`
				Ω(unmarshaler).Should(Equal(expected))
			})
		})

		Context("with a complex object", func() {
			var o design.Object

			JustBeforeEach(func() {
				unmarshaler = code.ObjectUnmarshaler(o, context, source, target)
			})

			BeforeEach(func() {
				ar := &design.Array{
					ElemType: &design.AttributeDefinition{
						Type: design.Primitive(design.IntegerKind),
					},
				}
				intAtt := &design.AttributeDefinition{Type: design.Primitive(design.IntegerKind)}
				arAtt := &design.AttributeDefinition{Type: ar}
				io := design.Object{"foo": intAtt, "bar": arAtt}
				ioAtt := &design.AttributeDefinition{Type: io}
				o = design.Object{"baz": ioAtt, "faz": intAtt}
			})

			It("generates the unmarshaler code", func() {
				expected := `	if val, ok := raw.(map[string]interface{}); ok {
		p = new(struct {
			Baz *struct {
				Bar []int
				Foo int
			}
			Faz int
		})
		if v, ok := val["baz"]; ok {
			var tmp1 *struct {
				Bar []int
				Foo int
			}
			if val, ok := v.(map[string]interface{}); ok {
				tmp1 = new(struct {
					Bar []int
					Foo int
				})
				if v, ok := val["bar"]; ok {
					var tmp2 []int
					if val, ok := v.([]interface{}); ok {
						tmp2 = make([]int, len(val))
						for i, v := range val {
							var tmp3 int
							if val, ok := v.(int); ok {
								tmp3 = val
							} else {
								err = goa.IncompatibleTypeError(` + "`" + `.Baz.Bar[*]` + "`" + `, v, "int")
							}
							tmp2[i] = tmp3
						}
					} else {
						err = goa.IncompatibleTypeError(` + "`" + `.Baz.Bar` + "`" + `, v, "[]interface{}")
					}
					tmp1.Bar = tmp2
				}
				if v, ok := val["foo"]; ok {
					var tmp4 int
					if val, ok := v.(int); ok {
						tmp4 = val
					} else {
						err = goa.IncompatibleTypeError(` + "`" + `.Baz.Foo` + "`" + `, v, "int")
					}
					tmp1.Foo = tmp4
				}
			} else {
				err = goa.IncompatibleTypeError(` + "`" + `.Baz` + "`" + `, v, "map[string]interface{}")
			}
			p.Baz = tmp1
		}
		if v, ok := val["faz"]; ok {
			var tmp5 int
			if val, ok := v.(int); ok {
				tmp5 = val
			} else {
				err = goa.IncompatibleTypeError(` + "`" + `.Faz` + "`" + `, v, "int")
			}
			p.Faz = tmp5
		}
	} else {
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "map[string]interface{}")
	}`
				Ω(unmarshaler).Should(Equal(expected))
			})

			Context("compiling", func() {
				var gopath, srcDir string
				var out []byte

				JustBeforeEach(func() {
					cmd := exec.Command("go", "build", "-o", "goagen")
					cmd.Env = os.Environ()
					cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s:%s", gopath, os.Getenv("GOPATH")))
					cmd.Dir = srcDir
					var err error
					out, err = cmd.CombinedOutput()
					Ω(err).ShouldNot(HaveOccurred())
				})

				BeforeEach(func() {
					var err error
					gopath, err = ioutil.TempDir("", "")
					Ω(err).ShouldNot(HaveOccurred())
					tmpl, err := template.New("main").Parse(mainTmpl)
					Ω(err).ShouldNot(HaveOccurred())
					srcDir = filepath.Join(gopath, "src", "test")
					err = os.MkdirAll(srcDir, 0755)
					Ω(err).ShouldNot(HaveOccurred())
					tmpFile, err := os.Create(filepath.Join(srcDir, "main.go"))
					Ω(err).ShouldNot(HaveOccurred())
					unmarshaler = code.ObjectUnmarshaler(o, context, source, target)
					data := map[string]interface{}{
						"raw": `interface{}(map[string]interface{}{
			"baz": map[string]interface{}{
				"foo": 345,
				"bar":[]interface{}{1,2,3},
			},
			"faz": 2,
		})`,
						"source":     unmarshaler,
						"target":     target,
						"targetType": code.GoTypeRef(o, 1),
					}
					err = tmpl.Execute(tmpFile, data)
					Ω(err).ShouldNot(HaveOccurred())
				})

				AfterEach(func() {
					os.RemoveAll(gopath)
				})

				It("compiles", func() {
					Ω(string(out)).Should(BeEmpty())

					cmd := exec.Command("./goagen")
					cmd.Env = []string{fmt.Sprintf("PATH=%s", filepath.Join(gopath, "bin"))}
					cmd.Dir = srcDir
					code, err := cmd.CombinedOutput()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(string(code)).Should(Equal(`{"Baz":{"Bar":[1,2,3],"Foo":345},"Faz":2}`))
				})

			})
		})

	})
})

const mainTmpl = `package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/raphael/goa"
)

func main() {
	var err error
	raw := {{.raw}}
	var {{.target}} {{.targetType}}
{{.source}}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	b, err := json.Marshal({{.target}})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Print(string(b))
}
`
