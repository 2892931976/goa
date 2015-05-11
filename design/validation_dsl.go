package design

/* Validation keywords for any instance type */

// http://json-schema.org/latest/json-schema-validation.html#anchor76
func (a *AttributeDefinition) Enum(val ...interface{}) *AttributeDefinition {
	a.Validations = append(a.Validations, validateEnum(val))
	return a
}

// Set default value
func (a *AttributeDefinition) Default(def interface{}) *AttributeDefinition {
	a.DefaultValue = def
	return a
}

// Set string format
func (a *AttributeDefinition) Format(f string) *AttributeDefinition {
	a.Validations = append(a.Validations, validateFormat(f))
	return a
}

// Minimum value validation
func (a *AttributeDefinition) Minimum(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, validateIntMinimum(val))
	return a
}

// Maximum value validation
func (a *AttributeDefinition) Maximum(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, validateIntMaximum(val))
	return a
}

// Minimum length validation
func (a *AttributeDefinition) MinLength(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, validateMinLength(val))
	return a
}

// Maximum length validation
func (a *AttributeDefinition) MaxLength(val int) *AttributeDefinition {
	a.Validations = append(a.Validations, validateMaxLength(val))
	return a
}

// Maximum length validation
func (a *AttributeDefinition) Required(names ...string) *AttributeDefinition {
	if a.Type.Kind() != ObjectType {
		panic("Required validation must be applied to object types")
	}
	a.Validations = append(a.Validations, validateRequired(names))
	return a
}
