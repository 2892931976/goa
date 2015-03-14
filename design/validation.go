package design

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Required value validation
func (p *Property) Required() *Property {
	p.Validations = append(p.Validations, validateRequired(p.Name))
	return p
}

// Set default value
func (p *Property) Default(def interface{}) *Property {
	p.DefaultValue = def
	return p
}

// Set string format
func (p *Property) Format(f string) *Property {
	p.Validations = append(p.Validations, validateFormat(p.Name, f))
	return p
}

// Minimum value validation
func (p *Property) Minimum(val int) *Property {
	p.Validations = append(p.Validations, validateIntMinimum(p.Name, val))
	return p
}

// Maximum value validation
func (p *Property) Maximum(val int) *Property {
	p.Validations = append(p.Validations, validateIntMaximum(p.Name, val))
	return p
}

// Minimum length validation
func (p *Property) MinLength(val int) *Property {
	p.Validations = append(p.Validations, validateMinLength(p.Name, val))
	return p
}

// Maximum length validation
func (p *Property) MaxLength(val int) *Property {
	p.Validations = append(p.Validations, validateMaxLength(p.Name, val))
	return p
}

// Enum validation
func (p *Property) Enum(val ...interface{}) *Property {
	p.Validations = append(p.Validations, validateEnum(p.Name, val))
	return p
}

// validateRequired returns a validation function that checks whether given value is nil
func validateRequired(name string) Validation {
	return func(val interface{}) error {
		if val == nil {
			return fmt.Errorf("Property %s is required.", name)
		}
		return nil
	}
}

// Regular expression used to validate RFC1035 hostnames
var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

// Simple regular expression for IPv4 values, more rigorous checking is done via net.ParseIP
var ipv4Regex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)

// validateFormat returns a validation function that validates the format of the given string
// The format specification follows the json schema draft 4 validation extension.
// see http://json-schema.org/latest/json-schema-validation.html#anchor105
// Supported formats are:
// - "date-time": RFC3339 date time value
// - "email": RFC5322 email address
// - "hostname": RFC1035 Internet host name
// - "ipv4" and "ipv6": RFC2673 and RFC2373 IP address values
// - "uri": RFC3986 URI value
// - "mac": IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address value
// - "cidr": RFC4632 and RFC4291 CIDR notation IP address value
func validateFormat(name, f string) Validation {
	return func(val interface{}) error {
		if val == nil {
			return nil
		}
		if sval, ok := val.(string); !ok {
			return fmt.Errorf("%s has an invalid type, got '%v', need string",
				name, val)
		} else {
			var err error
			switch strings.ToLower(f) {
			case "date-time":
				_, err = time.Parse(sval, time.RFC3339)
			case "email":
				_, err = mail.ParseAddress(sval)
			case "hostname":
				if !hostnameRegex.MatchString(sval) {
					err = fmt.Errorf("invalid hostname value '%s', does not match %s",
						sval, hostnameRegex.String())
				}
			case "ipv4", "ipv6":
				ip := net.ParseIP(sval)
				if ip == nil {
					err = fmt.Errorf("invalid %s value '%s'", f, sval)
				}
				if f == "ipv4" {
					if !ipv4Regex.MatchString(sval) {
						err = fmt.Errorf("invalid IPv4 value '%s'", sval)
					}
				}
			case "uri":
				_, err = url.ParseRequestURI(sval)
			case "mac":
				_, err = net.ParseMAC(sval)
			case "cidr":
				_, _, err = net.ParseCIDR(sval)
			default:
				err = fmt.Errorf("unknown validation format '%s'", f)
			}
			return fmt.Errorf("%s has an invalid value: %s", name, err)
		}
	}
}

// validateIntMaxValue returns a validation function that checks whether given value is a int that
// is lesser than max.
func validateIntMaximum(name string, max int) Validation {
	return func(val interface{}) error {
		if val == nil {
			return nil
		}
		if ival, ok := val.(int); !ok {
			return fmt.Errorf("%s has an invalid type, got '%v', need integer",
				name, val)
		} else if ival > max {
			return fmt.Errorf("%s has an invalid value: maximum allowed is %v, got %v",
				name, max, ival)
		}
		return nil
	}
}

// validateIntMinValue returns a validation function that checks whether given value is a int that
// is greater than min.
func validateIntMinimum(name string, min int) Validation {
	return func(val interface{}) error {
		if val == nil {
			return nil
		}
		if ival, ok := val.(int); !ok {
			return fmt.Errorf("%s has an invalid type, got '%v', need integer",
				name, val)
		} else if ival < min {
			return fmt.Errorf("%s has an invalid value: minimum allowed is %v, got %v",
				name, min, ival)
		}
		return nil
	}
}

// validateMinLength returns a validation function that checks whether given string or array has
// at least the number of given characters or elements.
func validateMinLength(name string, min int) Validation {
	return func(val interface{}) error {
		if val == nil {
			return nil
		}
		if sval, ok := val.(string); ok {
			if len(sval) < min {
				return fmt.Errorf("%s has an invalid value: minimum length is %v, got '%s' (%d characters)",
					name, min, sval, len(sval))
			}
		} else {
			k := reflect.TypeOf(val).Kind()
			if k == reflect.Slice || k == reflect.Array {
				v := reflect.ValueOf(val)
				if v.Len() < min {
					return fmt.Errorf("%s has an invalid value: minimum length is %v but actual length is %d, got %v",
						name, min, v.Len(), v)
				}
			} else {
				return fmt.Errorf("%s has an invalid type, got '%v', need string or array",
					name, val)
			}
		}
		return nil
	}
}

// validateMaxLength returns a validation function that checks whether given string or array has
// at most the number of given characters or elements.
func validateMaxLength(name string, max int) Validation {
	return func(val interface{}) error {
		if val == nil {
			return nil
		}
		if sval, ok := val.(string); ok {
			if len(sval) > max {
				return fmt.Errorf("%s has an invalid value: maximum length is %v, got '%s' (%d characters)",
					name, max, sval, len(sval))
			}
		} else {
			k := reflect.TypeOf(val).Kind()
			if k == reflect.Slice || k == reflect.Array {
				v := reflect.ValueOf(val)
				if v.Len() > max {
					return fmt.Errorf("%s has an invalid value: maximum length is %v but actual length is %d, got %v",
						name, max, v.Len(), v)
				}
			} else {
				return fmt.Errorf("%s has an invalid type, got '%v', need string or array",
					name, val)
			}
		}
		return nil
	}
}

// validateEnum returns a validation function that checks whether given value is one of the
// valid values.
func validateEnum(name string, valid []interface{}) Validation {
	return func(val interface{}) error {
		ok := false
		for _, v := range valid {
			if v == val {
				ok = true
				break
			}
		}
		if !ok {
			sValid := make([]string, len(valid))
			for i, v := range valid {
				sValid[i] = fmt.Sprintf("%v", v)
			}
			return fmt.Errorf("%s has an invalid value: allowed values are %s, got %v",
				name, strings.Join(sValid, ", "), val)
		}
		return nil
	}
}
