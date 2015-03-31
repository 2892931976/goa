package writers

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/alecthomas/kingpin"
)

// Middleware writer.
type middlewareGenWriter struct {
	genTmpl        *template.Template
	middlewareTmpl string
	routerTmpl     string
}

// Create middleware writer.
func NewMiddlewareGenWriter() (Writer, error) {
	genTmpl, err := template.New("middleware-gen").Parse(middlewareGenTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to create middleware-gen template, %s", err)
	}
	return &middlewareGenWriter{genTmpl: genTmpl, middlewareTmpl: middlewareTmpl, routerTmpl: routerTmpl}, nil
}

func (w *middlewareGenWriter) Source() string {
	var buf bytes.Buffer
	kingpin.FatalIfError(w.genTmpl.Execute(&buf, w), "middleware-gen template")
	return buf.String()
}

func (w *middlewareGenWriter) FunctionName() string {
	return "genMiddleware"
}

const middlewareGenTmpl = `
var resMiddlewareTmpl *template.Template

func {{.FunctionName}}(resource *design.Resource) error {
	if resRouterTmpl == nil {
		resRouterTmpl, err := template.New("router").Parse(routerTmpl)
		if err != nil {
			return fmt.Errorf("failed to create router template, %s", err)
		}
	}
	if resMiddlewareTmpl == nil {
		funcMap := template.FuncMap{"joinNames": joinNames, "literal": literal}
		resMiddlewareTmpl, err := template.New("middleware").Funcs(funcMap).Parse(middlewareTmpl)
		if err != nil {
			return fmt.Errorf("failed to create middleware template, %s", err)
		}
	}
	if err := resRouterTmpl.Execute(resource, w); err != nil {
		return fmt.Errorf("failed to generate %s router: %s", name, err)
	}
	if err := resMiddlewareTmpl.Execute(resource, w); err != nil {
		return fmt.Errorf("failed to generate %s middleware: %s", name, err)
	}
}

const routerTmpl = ` + "`" + `
{{.routerTmpl}}
` + "`" + `

const middlewareTmpl = ` + "`" + `
{{.middlewareTmpl}}
` + "`" + `
`
const routerTmpl = `
func {{.Name}}Router() { {{$resource = .}}
	router := httpRouter.New(){{$actionName, $action := range .Actions}}
	router.{{$action.HttpMethod}}(path.Join("{{$resource.BasePath}}", "{{$action.Path}}"), {{$action.Name}}{{$resource.Name}})
	{{end}} return router
}
`

const middlewareTmpl = `{{$resource = .}}{{range $actionName, $action := range .Actions}}
func {{$actionName}}{{$resource.Name}}(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	h := goa.New{{$resource.Name}}Handler(w, r){{range $name, $param := $action.PathParams}}
	{{$name}}, err := {{$param.Member.Type.Name}}.Load(params.ByName("{{$name}}"))
	if err != nil {
		goa.RespondBadRequest(w, "Invalid param '{{$name}}': %s", err)
		return
	}{{end}}{{if $action.QueryParams}}
	query := r.URL.Query()
	{{range $name, $param := $action.QueryParams}}{{$name}}, err := {{$param.Member.Type.Name}}.Load(query["{{$name}}"]{{if not (eq $param.Member.Type.Name "array")}}[0]{{end}})
	if err != nil {
		goa.RespondBadRequest(w, "Invalid param '{{$name}}': %s", err)
		return
	}
	{{end}}{{end}}{{if .action.Payload}}
	b, err := h.LoadRequestBody(r)
	if err != nil {
		goa.RespondBadRequest(w, err)
		return
	}
	raw, err := res.Actions["{{$actionName}}"].Payload.Load("payload", b)
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	var payload {{$actionName}}Payload
	err = goa.InitStruct(&payload, raw.(map[string]interface{}))
	if err != nil {
		goa.RespondBadRequest(w, err.Error())
		return
	}
	resp := h.{{$actionName}}({{if $action.Payload}}&payload{{end}}{{if $action.PathParams}}, {{joinNames $action.PathParams}}{{end}}{{if $action.QueryParams}}{{joinNames $action.QueryParams}}{{end}})
	if resp == nil {
		// Response already written by handler
		return
	}
	{{if .Responses}}ok := resp.Status == 400 || resp.Status == 500
	if !ok {
		{{range $action.Responses}}if resp.Status == {{.Status}} {
			ok = true{{if .MediaType}}
			resp.Header.Set("Content-Type", "{{.MediaType.Identifier}}+json")
		}{{end}}
	}{{$name, $value := range .HeaderPatterns}}
	h := resp.Header.Get("{{$name}}")
	if !regexp.MatchString("{{$value}}", h) {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced invalid ${{name}} header value.", h))
		return
	}{{end}}	
	{{end}} }
	if !ok {
		goa.RespondInternalError(w, fmt.Printf("API bug, code produced unknown status code %d", resp.Status))
		return
	}
	{{end}}{{/* if .Responses */}}
	resp.Write(w)
}
`
