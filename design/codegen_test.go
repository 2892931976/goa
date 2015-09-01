package design_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
)

var _ = Describe("GoTypeDef", func() {
	Context("given an attribute definition with fields", func() {
		var att *design.AttributeDefinition
		var object design.Object
		var st string

		JustBeforeEach(func() {
			if att == nil {
				att = new(design.AttributeDefinition)
			}
			att.Type = object
			st = design.GoTypeDef(att, 0, true, false)
		})

		Context("of primitive types", func() {
			BeforeEach(func() {
				object = design.Object{
					"foo": &design.AttributeDefinition{Type: design.Integer},
					"bar": &design.AttributeDefinition{Type: design.String},
				}
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
				required := &design.RequiredValidationDefinition{
					Names: []string{"foo"},
				}
				att = &design.AttributeDefinition{
					Validations: []design.ValidationDefinition{required},
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
			source = design.GoTypeDef(att, 0, true, false)
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

var _ = Describe("Unmarshaler", func() {
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
			unmarshaler = design.PrimitiveUnmarshaler(p, context, source, target)
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
			unmarshaler = design.ArrayUnmarshaler(p, context, source, target)
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
		err = goa.IncompatibleTypeError(` + "``" + `, raw, "[]int")
	}`
			Ω(unmarshaler).Should(Equal(expected))
		})
	})

	/* Context("with a simple object", func() {*/
	//var o design.Object

	//JustBeforeEach(func() {
	//unmarshaler = design.ObjectUnmarshaler(o, context, source, target)
	//})

	//BeforeEach(func() {
	//intAtt := &design.AttributeDefinition{Type: design.Primitive(design.IntegerKind)}
	//o = design.Object{"foo": intAtt}
	//})

	//It("generates the unmarshaler code", func() {
	//expected := `	if val, ok := raw.(map[string]interface{}); ok {
	//p = make(map[string]interface{})
	//if v, ok := val["foo"]; ok {
	//if val, ok := v.(int); ok {
	//p["foo"] = val
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["foo"]` + "`" + `, v, "int")
	//}
	//}
	//} else {
	//err = goa.IncompatibleTypeError(` + "``" + `, raw, ` + "`" + `struct {
	//Foo int
	//}` + "`" + `)
	//}`
	//Ω(unmarshaler).Should(Equal(expected))
	//})
	//})

	//Context("with a complex object", func() {
	//var o design.Object

	//JustBeforeEach(func() {
	//unmarshaler = design.ObjectUnmarshaler(o, context, source, target)
	//})

	//BeforeEach(func() {
	//ar := &design.Array{
	//ElemType: &design.AttributeDefinition{
	//Type: design.Primitive(design.IntegerKind),
	//},
	//}
	//intAtt := &design.AttributeDefinition{Type: design.Primitive(design.IntegerKind)}
	//arAtt := &design.AttributeDefinition{Type: ar}
	//io := design.Object{"foo": intAtt, "bar": arAtt}
	//ioAtt := &design.AttributeDefinition{Type: io}
	//o = design.Object{"baz": ioAtt, "faz": intAtt}
	//})

	//It("generates the unmarshaler code", func() {
	//expected := `	if val, ok := raw.(map[string]interface{}); ok {
	//p = make(map[string]interface{})
	//if v, ok := val["baz"]; ok {
	//if val, ok := v.(map[string]interface{}); ok {
	//p["baz"] = make(map[string]interface{})
	//if v, ok := val["bar"]; ok {
	//if val, ok := v.([]interface{}); ok {
	//p["baz"]["bar"] = make([]int, len(val))
	//for i, v := range val {
	//if val, ok := v.(int); ok {
	//p["baz"]["bar"][i] = val
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["baz"]["bar"][*]` + "`" + `, v, "int")
	//}
	//}
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["baz"]["bar"]` + "`" + `, v, "[]int")
	//}
	//}
	//if v, ok := val["foo"]; ok {
	//if val, ok := v.(int); ok {
	//p["baz"]["foo"] = val
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["baz"]["foo"]` + "`" + `, v, "int")
	//}
	//}
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["baz"]` + "`" + `, v, ` + "`" + `struct {
	//Bar []int
	//Foo int
	//}` + "`" + `)
	//}
	//}
	//if v, ok := val["faz"]; ok {
	//if val, ok := v.(int); ok {
	//p["faz"] = val
	//} else {
	//err = goa.IncompatibleTypeError(` + "`" + `["faz"]` + "`" + `, v, "int")
	//}
	//}
	//} else {
	//err = goa.IncompatibleTypeError(` + "``" + `, raw, ` + "`" + `struct {
	//Baz struct {
	//Bar []int
	//Foo int
	//}
	//Faz int
	//}` + "`" + `)
	//}`
	//Ω(unmarshaler).Should(Equal(expected))
	//})

	//Context("compiling", func() {
	//var gopath, srcDir string
	//var out []byte

	//JustBeforeEach(func() {
	//cmd := exec.Command("go", "build", "-o", "codegen")
	//cmd.Env = []string{fmt.Sprintf("PATH=%s", os.Getenv("PATH"))}
	//cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s:%s", gopath, os.Getenv("GOPATH")))
	//cmd.Dir = srcDir
	//out, _ = cmd.CombinedOutput()
	//})

	//BeforeEach(func() {
	//var err error
	//gopath, err = ioutil.TempDir("", "")
	//Ω(err).ShouldNot(HaveOccurred())
	//tmpl, err := template.New("main").Parse(mainTmpl)
	//Ω(err).ShouldNot(HaveOccurred())
	//srcDir = filepath.Join(gopath, "src", "test")
	//err = os.MkdirAll(srcDir, 0755)
	//Ω(err).ShouldNot(HaveOccurred())
	//tmpFile, err := os.Create(filepath.Join(srcDir, "main.go"))
	//Ω(err).ShouldNot(HaveOccurred())
	//unmarshaler = design.ObjectUnmarshaler(o, context, source, target)
	//data := map[string]interface{}{
	//"raw": `interface{}(map[string]interface{}{
	//"baz": map[string]interface{}{
	//"foo": 345,
	//"bar":[]int{1,2,3},
	//},
	//"faz": 2,
	//})`,
	//"source":     unmarshaler,
	//"target":     target,
	//"targetType": design.GoTypeRef(o),
	//}
	//err = tmpl.Execute(tmpFile, data)
	//Ω(err).ShouldNot(HaveOccurred())
	//})

	//AfterEach(func() {
	////	os.RemoveAll(gopath)
	//})

	//It("compiles", func() {
	//Ω(string(out)).Should(BeEmpty())

	//cmd := exec.Command("codegen")
	//cmd.Env = []string{fmt.Sprintf("PATH=%s", filepath.Join(gopath, "bin"))}
	//code, err := cmd.CombinedOutput()
	//Ω(err).ShouldNot(HaveOccurred())
	//Ω(code).Should(Equal("ok"))
	//})

	/*})*/
	//})

})

const mainTmpl = `package main

import (
	"fmt"
	"os"

	"github.com/raphael/goa"
)

func main() {
	var err error
	raw := {{.raw}}
	var {{.target}} {{.targetType}}
{{.source}}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
	}
	fmt.Printf("%v", p)
}
`
