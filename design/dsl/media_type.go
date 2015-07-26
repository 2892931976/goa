package design

import (
	"fmt"
	"strings"

	"github.com/raphael/goa/design"
)

// NewMediaType creates new media type from its identifier, description and type.
// Initializes a default view that returns all the media type members.
func NewMediaType(id, desc string, o Object) *design.MediaTypeDefinition {
	mt := MediaTypeDefinition{Object: o, Identifier: id, Description: desc, Links: make(map[string]*Link)}
	mt.Views = map[string]*View{"default": &View{Name: "default", Object: o}}
	return &mt
}

// View adds a new view to the media type.
// It returns the view so it can be modified further.
// This method ignore passed-in property names that do not exist in media type.
func (m *design.MediaTypeDefinition) View(name string, members ...string) *View {
	o := make(Object, len(members))
	i := 0
	for n, p := range m.Object {
		found := false
		for _, m := range members {
			if m == n {
				found = true
				break
			}
		}
		if found {
			o[n] = p
			i += 1
		}
	}
	view := View{Name: name, Object: o, MediaType: m}
	m.Views[name] = &view
	return &view
}

// As sets the list of member names rendered by view.
// If a member is a media type then the view used to render it defaults to the view with same name.
// The view used to renber media types members can be explicitely set using the syntax
// "<member name>:<view name>". For example:
//     m.View("expanded").As("id", "expensive_attribute:default")
func (v *View) With(members ...string) *View {
	o := Object{}
	for _, m := range members {
		elems := strings.SplitN(m, ":", 2)
		mm, ok := v.MediaType.Object[elems[0]]
		if !ok {
			panic(fmt.Sprintf("Invalid view member '%s', no such media type member.", m))
		}
		if len(elems) > 1 {
			if mm.Type.Kind() != ObjectType {
				panic(fmt.Sprintf("Cannot use view '%s' to render media type member '%s': not a media type", elems[1], elems[0]))
			}
		}
		o[m] = mm
	}
	v.Object = o
	return v
}

// Links specifies the list of links rendered with this media type.
func (v *View) Link(links ...string) *View {
	for _, l := range links {
		if _, ok := v.MediaType.Links[l]; !ok {
			panic(fmt.Sprintf("Invalid view link '%s', no such media type link.", l))
		}
	}
	v.Links = append(v.Links, links...)
	return v
}

// Link adds a new link to the given media type member.
// It returns the link so it can be modified further.
func (m *design.MediaTypeDefinition) Link(name string) *Link {
	member, ok := m.Object[name]
	if !ok {
		panic(fmt.Sprintf("Invalid  link '%s', no such media type member.", name))
	}
	link := Link{Name: name, Member: member, MediaType: m}
	m.Links[name] = &link
	return &link
}

// As overrides the link name.
// It returns the link so it can be modified further.
func (l *Link) As(name string) *Link {
	delete(l.MediaType.Links, l.Name)
	l.Name = name
	l.MediaType.Links[name] = l
	return l
}

// CollectionOf creates a collection media type from its element media type.
// A collection media type represents the content of responses that return a
// collection of resources such as "index" actions.
func CollectionOf(m *design.MediaTypeDefinition) *MediaTypeDefinition {
	col := *m
	col.isCollection = true
	return &col
}

// TBD: the functions below are runtime functions - not design function
// Equivalent code should be generated to avoid runtime dependencies on
// design code

// Render accepts either a struct or a map indexed by keys.
// If given a struct Render picks the struct fields whose names match the view property names.
// If the fields are tagged with json tags then Render uses the tag names to do the comparison with
// view property names.
// If given a map indexed by strings then Renders picks the keys with the same name as the view
// property names.
// If given an array then checks that media type is a collection then apply algorithm recursively
// on each element of the array.
// Once the resulting map has been built the values are validated using the view property
// validations.
/*design.func (m *MediaTypeDefinition) Render(value interface{}, viewName string) (interface{}, error) {*/
//if value == nil {
//return make(map[string]interface{}), nil
//}
//if _, ok := m.Views[viewName]; !ok {
//return nil, fmt.Errorf("View '%s' not found", viewName)
//}
//var rendered map[string]interface{}
//switch reflect.TypeOf(value).Kind() {
//case reflect.Slice:
//s := reflect.ValueOf(value)
//res := make([]interface{}, s.Len())
//for i := 0; i < s.Len(); i++ {
//var err error
//if res[i], err = m.Render(s.Index(i).Interface(), viewName); err != nil {
//return nil, err
//}
//}
//return res, nil
//case reflect.Struct:
//var err error
//if rendered, err = m.renderStruct(value, viewName); err != nil {
//return nil, err
//}
//case reflect.Map:
//var err error
//if rendered, err = m.renderMap(value.(map[string]interface{}), viewName); err != nil {
//return nil, err
//}
//case reflect.Ptr:
//return m.Render(reflect.ValueOf(value).Elem().Interface(), viewName)
//default:
//return nil, fmt.Errorf("Rendered value must be a map, a struct, a slice of maps or structs, or a pointer to one of these. Given value was a %v.",
//reflect.TypeOf(value))
//}
//if err := m.validate(rendered); err != nil {
//return nil, err
//}
//return rendered, nil
/*}*/

// Render given struct
// Builds map with values corresponding to fields with media type property names then validates it
// View name must be valid
/*design.func (m *MediaTypeDefinition) renderStruct(value interface{}, viewName string) (map[string]interface{}, error) {*/
//t := reflect.TypeOf(value)
//v := reflect.ValueOf(value)
//numField := t.NumField()
//rendered := make(map[string]interface{})
//view := m.Views[viewName]
//for i := 0; i < numField; i++ {
//field := t.Field(i)
//name := field.Name
//if member, ok := view.Object[name]; ok {
//raw := v.FieldByName(field.Name).Interface()
//val, err := member.Type.Load(raw)
//if err != nil {
//return nil, err
//}
//rendered[name] = val
//} else {
//member := m.Object[name]
//if member == nil {
//return nil, fmt.Errorf("Cannot render unknown member '%s'", name)
//}
//if member.DefaultValue != nil {
//rendered[name] = member.DefaultValue
//}
//}
//}
//return rendered, nil
/*}*/

// Render given map
// Builds map with values corresponding to media type property names then validates it
// View name must be valid
/*design.func (m *MediaTypeDefinition) renderMap(value map[string]interface{}, viewName string) (map[string]interface{}, error) {*/
//rendered := make(map[string]interface{})
//view := m.Views[viewName]
//for n, v := range value {
//if _, ok := view.Object[n]; ok {
//rendered[n] = v
//}
//}
//return rendered, nil
//}

// First make sure that any property with default value has its value initialized in the map, then
// run property validation functions on value associated with property name.
/*design.func (m *MediaTypeDefinition) validate(rendered map[string]interface{}) error {*/
//for n, p := range m.Object {
//if _, ok := rendered[n]; !ok {
//if p.DefaultValue != nil {
//rendered[n] = p.DefaultValue
//}
//}
//}
//for n, v := range rendered {
//p := m.Object[n]
//for _, validate := range p.Validations {
//if err := validate(n, v); err != nil {
//return err
//}
//}
//}
//return nil
/*}*/
