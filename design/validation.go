package design

import (
	"fmt"
	"mime"
	"strings"
)

// Validate tests whether the API definition is consistent: all resource parent names resolve to
// an actual resource.
func (a *APIDefinition) Validate() error {
	for _, r := range a.Resources {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("Resource %s: %s", r.Name, err)
		}
		if r.ParentName == "" {
			continue
		}
		if _, ok := Design.Resources[r.ParentName]; !ok {
			return fmt.Errorf("Resource %s: Unknown parent resource %s", r.Name, r.ParentName)
		}
	}
	return nil
}

// Validate tests whether the resource definition is consistent: action names are valid and each action is
// valid.
func (r *ResourceDefinition) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("Resource name cannot be empty")
	}
	found := false
	for _, a := range r.Actions {
		if a.Name == r.CanonicalAction {
			found = true
		}
		if err := a.Validate(); err != nil {
			return fmt.Errorf("Action %s: %s", a.Name, err)
		}
	}
	if r.CanonicalAction != "" && !found {
		return fmt.Errorf("Unknown canonical action '%s'", r.CanonicalAction)
	}
	if r.BaseParams != nil {
		baseParams, ok := r.BaseParams.Type.(Object)
		if !ok {
			return fmt.Errorf("Invalid type for BaseParams, must be an Object")
		}
		vs := ParamsRegex.FindAllStringSubmatch(r.BasePath, -1)
		if len(vs) > 1 {
			vars := vs[1]
			if len(vars) != len(baseParams) {
				return fmt.Errorf("BasePath defines parameters %s but BaseParams has %d elements",
					strings.Join([]string{
						strings.Join(vars[:len(vars)-1], ", "),
						vars[len(vars)-1],
					}, " and "),
					len(baseParams),
				)
			}
			for _, v := range vars {
				found := false
				for n := range baseParams {
					if v == n {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("Variable %s from base path %s does not match any parameter from BaseParams",
						v, r.BasePath)
				}
			}
		} else {
			if len(baseParams) > 0 {
				return fmt.Errorf("BasePath does not use variables defines in BaseParams")
			}
		}
	}
	if r.ParentName != "" {
		if _, ok := Design.Resources[r.ParentName]; !ok {
			return fmt.Errorf("Parent resource named %#v not found", r.ParentName)
		}
	}
	if r.MediaType != "" {
		if mt, ok := Design.MediaTypes[r.MediaType]; ok {
			if err := mt.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate tests whether the action definition is consistent: parameters have unique names and it has at least
// one response.
func (a *ActionDefinition) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("Action name cannot be empty")
	}
	if len(a.Routes) == 0 {
		return fmt.Errorf("No route defined for action %s", a.Name)
	}
	for i, r := range a.Responses {
		for j, r2 := range a.Responses {
			if i != j && r.Status == r2.Status {
				return fmt.Errorf("Multiple response definitions with status code %d", r.Status)
			}
		}
		if err := r.Validate(); err != nil {
			if r.Status == 0 {
				return fmt.Errorf("invalid response definition: %s", err)
			}
			return fmt.Errorf("invalid %d response definition: %s", r.Status, err)
		}
	}
	if err := a.ValidateParams(); err != nil {
		return err
	}
	if a.Payload != nil {
		if err := a.Payload.Validate(); err != nil {
			return fmt.Errorf(`invalid payload definition for action "%s": %s`,
				a.Name, err)
		}
	}
	return nil
}

// ValidateParams checks the action parameters (make sure they have names, members and types).
func (a *ActionDefinition) ValidateParams() error {
	if a.Params == nil {
		return nil
	}
	params, ok := a.Params.Type.(Object)
	if !ok {
		return fmt.Errorf(`"Params" field of action "%s" is not an object`, a.Name)
	}
	for n, p := range params {
		if n == "" {
			return fmt.Errorf("%s has parameter with no name", a.Name)
		} else if p == nil {
			return fmt.Errorf("definition of parameter %s of action %s cannot be nil",
				n, a.Name)
		} else if p.Type == nil {
			return fmt.Errorf("type of parameter %s of action %s cannot be nil",
				n, a.Name)
		}
		if p.Type.Kind() == ObjectKind {
			return fmt.Errorf(`parameter %s of action %s cannot be an object, only action payloads may be of type object`,
				n, a.Name)
		}
		if err := p.Validate(); err != nil {
			return fmt.Errorf(`invalid definition for parameter %s of action %s: %s`,
				n, a.Name, err)
		}
	}
	return nil
}

// Validate tests whether the attribute definition is consistent: required fields exist.
func (a *AttributeDefinition) Validate() error {
	o, isObject := a.Type.(Object)
	for _, v := range a.Validations {
		if r, ok := v.(*RequiredValidationDefinition); ok {
			if !isObject {
				return fmt.Errorf(`only objects may define a "Required" validation`)
			}
			for _, n := range r.Names {
				var found bool
				for an := range o {
					if n == an {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf(`required field "%s" does not exist`, n)
				}
			}
		}
	}
	if isObject {
		for _, att := range o {
			if err := att.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Validate checks that the response definition is consistent: its status is set and the media
// type definition if any is valid.
func (r *ResponseDefinition) Validate() error {
	if r.Status == 0 {
		return fmt.Errorf("response status not defined")
	}
	if r.MediaType != "" {
		if mt, ok := Design.MediaTypes[r.MediaType]; ok {
			if err := mt.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate checks that the user type definition is consistent: it has a name.
func (u *UserTypeDefinition) Validate() error {
	if u.TypeName == "" {
		return fmt.Errorf("User type must have a name")
	}
	return nil
}

// Validate checks that the media type definition is consistent: its identifier is a valid media
// type identifier.
func (m *MediaTypeDefinition) Validate() error {
	if m.TypeName == "" {
		return fmt.Errorf("Media type must have a name")
	}
	if m.Identifier != "" {
		_, _, err := mime.ParseMediaType(m.Identifier)
		if err != nil {
			return fmt.Errorf("invalid media type identifier: %s", err)
		}
	} else {
		m.Identifier = "plain/text"
	}
	if m.Type == nil { // TBD move this to somewhere else than validation code
		m.Type = String
	}
	if o := m.Type.ToObject(); o != nil {
		for n, att := range o {
			if att.View != "" {
				cmt, ok := att.Type.(*MediaTypeDefinition)
				if !ok {
					return fmt.Errorf("attribute %s of media type %s defines a view for rendering but its type is not MediaTypeDefinition",
						n, m.Identifier)
				}
				if _, ok := cmt.Views[att.View]; !ok {
					return fmt.Errorf("attribute %s of media type %s uses unknown view %#v",
						n, m.Identifier, att.View)
				}
			}
		}
	}
	return nil
}

// Validate checks that the link definition is consistent: it has a media type or the name of an
// attribute part of the parent media type.
func (l *LinkDefinition) Validate() error {
	mediaType := l.MediaType
	if mediaType == nil {
		if l.Name == "" {
			return fmt.Errorf("Links must have a name")
		}
		if l.Parent == nil {
			return fmt.Errorf("Link %#v must have a parent media type", l.Name)
		}
		if l.Parent.ToObject() == nil {
			return fmt.Errorf("Link %#v parent media type must be an Object", l.Name)
		}
		att, ok := l.Parent.ToObject()[l.Name]
		if !ok {
			return fmt.Errorf("Link %#v name must match one of the parent media type attribute names", l.Name)
		}
		if mediaType, ok = att.Type.(*MediaTypeDefinition); !ok {
			return fmt.Errorf("Link %#v attribute type must be a media type", l.Name)
		}
	}
	viewFound := false
	view := l.View
	for v := range mediaType.Views {
		if v == view {
			viewFound = true
			break
		}
	}
	if !viewFound {
		return fmt.Errorf("Link %#v of media type %s uses view %#v which does not exist on target media type %#v",
			l.Name, l.Parent.Identifier, view, mediaType.Identifier)
	}
	return nil
}
