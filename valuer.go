package v

import (
	"fmt"
	"reflect"
	"strings"

	"zestack.dev/is"
)

// Ruler 规则验证函数签名
type Ruler func(any) error

// Valuer 基本值验证器
type Valuer struct {
	field    string    // 字段名称，如：username
	label    string    // 数据标签，对应字段名，如：用户名
	value    any       // 参与验证的值
	requires []Checker // 空值验证器列表
	rules    []Ruler   // 参与验证的规则列表
}

// Value 创建一条验证器
func Value(value any, field, label string) *Valuer {
	return &Valuer{
		field:    field,
		label:    label,
		value:    value,
		requires: []Checker{},
		rules:    []Ruler{},
	}
}

// Validate 实现验证器接口
func (v *Valuer) Validate() error {
	if is.Empty(v.value) {
		for _, require := range v.requires {
			if err := require(); err != nil {
				return err
			}
		}
		return nil
	}

	// simple nil
	value := v.value
	if value == nil {
		return nil
	} else if reflect.TypeOf(value).Kind() == reflect.Ptr {
		rv := reflect.ValueOf(value)
		if rv.IsNil() {
			return nil
		} else {
			value = rv.Elem().Interface()
		}
	}

	// call rules
	for _, rule := range v.rules {
		if err := rule(value); err != nil {
			return err
		}
	}

	return nil
}

func (v *Valuer) mistake(err error, options ...ErrorOption) *Error {
	if m, ok := err.(*Error); ok {
		return m
	}
	e := &Error{error: err}
	e.field = v.field
	e.label = v.label
	e.value = v.value
	for _, option := range options {
		option(e)
	}
	return e
}

func (v *Valuer) newError(code string, options []ErrorOption) *Error {
	e := NewError(code, options...)
	e.label = v.label
	e.value = v.value
	e.field = v.field
	return e
}

func (v *Valuer) addRule(rule Ruler) *Valuer {
	v.rules = append(v.rules, rule)
	return v
}

func (v *Valuer) simple(code string, check func(any) bool, options []ErrorOption) *Valuer {
	return v.addRule(func(val any) error {
		if check(val) {
			return nil
		}
		return v.newError(code, options)
	})
}

func (v *Valuer) string(code string, check func(string) bool, options []ErrorOption) *Valuer {
	return v.simple(code, func(a any) bool { return check(toString(a)) }, options)
}

func (v *Valuer) Custom(code string, check func(val any) any, options ...ErrorOption) *Valuer {
	v.rules = append(v.rules, func(val any) error {
		if res := check(val); res == false {
			return v.newError(code, options) // 验证失败
		} else if res == true || res == nil {
			return nil // 验证成功
		} else if err, ok := res.(error); ok {
			if isBuiltinError(err) {
				return err
			}
			m := v.newError(code, options)
			m.error = err
			//if str := err.Error(); m.format != "" && str != "" {
			//	m.format = fmt.Sprintf("%s(%s)", m.format, str)
			//}
			return m
		} else {
			panic("must return a bool, a nil or a error pointer")
		}
	})
	return v
}

// Required 值是否必须（值不为空）
func (v *Valuer) Required(options ...ErrorOption) *Valuer {
	v.requires = append(v.requires, func() error {
		return v.newError("required", options)
	})
	return v
}

// RequiredIf 满足条件必须
func (v *Valuer) RequiredIf(condition bool, options ...ErrorOption) *Valuer {
	v.requires = append(v.requires, func() error {
		if condition {
			return v.newError("required_if", options)
		}
		return nil
	})

	return v
}

// RequiredWith 依赖其它值判断是否必须
func (v *Valuer) RequiredWith(values []any, options ...ErrorOption) *Valuer {
	v.requires = append(v.requires, func() error {
		for _, value := range values {
			if !is.Empty(value) {
				return v.newError("required_with", options)
			}
		}
		return nil
	})

	return v
}

func (v *Valuer) When(condition bool, then func(*Valuer)) *Valuer {
	if condition && then != nil {
		v.addRule(func(a any) error {
			x := Value(v.value, v.field, v.label)
			then(x)
			return x.Validate()
		})
	}
	return v
}

func (v *Valuer) Match(handle func(m *Matcher)) *Valuer {
	return v.addRule(func(a any) error {
		m := Match(v.value, v.field, v.label)
		handle(m)
		return m.Validate()
	})
}

func (v *Valuer) Typeof(kind reflect.Kind, options ...ErrorOption) *Valuer {
	return v.addRule(func(val any) error {
		if reflect.TypeOf(val).Kind() != kind {
			options = merge(options, ErrorParam("kind", kind))
			return v.newError("typeof", options)
		}
		return nil
	})
}

func (v *Valuer) IsString(options ...ErrorOption) *Valuer {
	options = append(options, ErrorCode("is_string"))
	return v.Typeof(reflect.String, options...)
}

func (v *Valuer) IsEmail(options ...ErrorOption) *Valuer {
	return v.string("is_email", is.Email, options)
}

func (v *Valuer) IsE164(options ...ErrorOption) *Valuer {
	return v.string("is_e164", is.E164, options)
}

func (v *Valuer) IsPhoneNumber(options ...ErrorOption) *Valuer {
	return v.string("is_phone_number", is.PhoneNumber, options)
}

func (v *Valuer) IsURL(options ...ErrorOption) *Valuer {
	return v.string("is_url", is.URL, options)
}

func (v *Valuer) IsURLEncoded(options ...ErrorOption) *Valuer {
	return v.string("is_url_encoded", is.URLEncoded, options)
}

func (v *Valuer) IsBase64URL(options ...ErrorOption) *Valuer {
	return v.string("is_base64_url", is.Base64URL, options)
}

func (v *Valuer) IsSemver(options ...ErrorOption) *Valuer {
	return v.string("is_semver", is.Semver, options)
}

func (v *Valuer) IsJwt(options ...ErrorOption) *Valuer {
	return v.string("is_jwt", is.JWT, options)
}

func (v *Valuer) IsUUID(options ...ErrorOption) *Valuer {
	return v.string("is_uuid", is.UUID, options)
}

func (v *Valuer) IsUUID5(options ...ErrorOption) *Valuer {
	return v.string("is_uuid5", is.UUID5, options)
}

func (v *Valuer) IsUUID4(options ...ErrorOption) *Valuer {
	return v.string("is_uuid4", is.UUID4, options)
}

func (v *Valuer) IsUUID3(options ...ErrorOption) *Valuer {
	return v.string("is_uuid3", is.UUID3, options)
}

func (v *Valuer) IsULID(options ...ErrorOption) *Valuer {
	return v.string("is_ulid", is.ULID, options)
}

func (v *Valuer) IsMD4(options ...ErrorOption) *Valuer {
	return v.string("is_md4", is.MD4, options)
}

func (v *Valuer) IsMD5(options ...ErrorOption) *Valuer {
	return v.string("is_md5", is.MD5, options)
}

func (v *Valuer) IsSHA256(options ...ErrorOption) *Valuer {
	return v.string("is_sha256", is.SHA256, options)
}

func (v *Valuer) IsSHA384(options ...ErrorOption) *Valuer {
	return v.string("is_sha384", is.SHA384, options)
}

func (v *Valuer) IsSHA512(options ...ErrorOption) *Valuer {
	return v.string("is_sha512", is.SHA512, options)
}

func (v *Valuer) IsAscii(options ...ErrorOption) *Valuer {
	return v.string("is_ascii", is.ASCII, options)
}

func (v *Valuer) IsAlpha(options ...ErrorOption) *Valuer {
	return v.string("is_alpha", is.Alpha, options)
}

func (v *Valuer) IsAlphanumeric(options ...ErrorOption) *Valuer {
	return v.string("is_alphanumeric", is.Alphanumeric, options)
}

func (v *Valuer) IsAlphaUnicode(options ...ErrorOption) *Valuer {
	return v.string("is_alpha_unicode", is.AlphaUnicode, options)
}

func (v *Valuer) IsAlphanumericUnicode(options ...ErrorOption) *Valuer {
	return v.string("is_alphanumeric_unicode", is.AlphanumericUnicode, options)
}

func (v *Valuer) IsNumeric(options ...ErrorOption) *Valuer {
	return v.simple("is_numeric", is.Numeric[any], options)
}

func (v *Valuer) IsNumber(options ...ErrorOption) *Valuer {
	return v.simple("is_number", is.Number[any], options)
}

func (v *Valuer) IsBool(options ...ErrorOption) *Valuer {
	return v.simple("is_bool", is.Boolean[any], options)
}

func (v *Valuer) IsHexadecimal(options ...ErrorOption) *Valuer {
	return v.string("is_hexadecimal", is.Hexadecimal, options)
}

func (v *Valuer) IsHexColor(options ...ErrorOption) *Valuer {
	return v.string("is_hexcolor", is.HEXColor, options)
}

func (v *Valuer) IsRgb(options ...ErrorOption) *Valuer {
	return v.string("is_rgb", is.RGB, options)
}

func (v *Valuer) IsRgba(options ...ErrorOption) *Valuer {
	return v.string("is_rgba", is.RGBA, options)
}

func (v *Valuer) IsHsl(options ...ErrorOption) *Valuer {
	return v.string("is_hsl", is.HSL, options)
}

func (v *Valuer) IsHsla(options ...ErrorOption) *Valuer {
	return v.string("is_hsla", is.HSLA, options)
}

func (v *Valuer) IsColor(options ...ErrorOption) *Valuer {
	return v.string("is_color", is.Color, options)
}

func (v *Valuer) IsLatitude(options ...ErrorOption) *Valuer {
	return v.simple("is_latitude", is.Latitude[any], options)
}

func (v *Valuer) IsLongitude(options ...ErrorOption) *Valuer {
	return v.simple("is_longitude", is.Longitude[any], options)
}

func (v *Valuer) IsJson(options ...ErrorOption) *Valuer {
	return v.simple("is_json", is.JSON[any], options)
}

func (v *Valuer) IsBase64(options ...ErrorOption) *Valuer {
	return v.string("is_base64", is.Base64, options)
}

func (v *Valuer) IsHTML(options ...ErrorOption) *Valuer {
	return v.string("is_html", is.HTML, options)
}

func (v *Valuer) IsHTMLEncoded(options ...ErrorOption) *Valuer {
	return v.string("is_html_encoded", is.HTMLEncoded, options)
}

func (v *Valuer) IsDatetime(layout string, options ...ErrorOption) *Valuer {
	return v.simple(
		"is_datetime",
		func(a any) bool { return is.Datetime(toString(a), layout) },
		append([]ErrorOption{ErrorParam("layout", layout)}, options...),
	)
}

func (v *Valuer) IsTimezone(options ...ErrorOption) *Valuer {
	return v.string("is_timezone", is.Timezone, options)
}

func (v *Valuer) IsIPv4(options ...ErrorOption) *Valuer {
	return v.string("is_ipv4", is.IPv4, options)
}

func (v *Valuer) IsIPv6(options ...ErrorOption) *Valuer {
	return v.string("is_ipv6", is.IPv6, options)
}

func (v *Valuer) IsIP(options ...ErrorOption) *Valuer {
	return v.string("is_ip", is.IP, options)
}

func (v *Valuer) IsMAC(options ...ErrorOption) *Valuer {
	return v.string("is_mac", is.MAC, options)
}

func (v *Valuer) IsFile(options ...ErrorOption) *Valuer {
	return v.simple("is_file", is.File, options)
}

func (v *Valuer) IsDir(options ...ErrorOption) *Valuer {
	return v.simple("is_file", is.Dir, options)
}

func (v *Valuer) IsLower(options ...ErrorOption) *Valuer {
	return v.string("is_lower", is.Lowercase, options)
}

func (v *Valuer) IsUpper(options ...ErrorOption) *Valuer {
	return v.string("is_upper", is.Uppercase, options)
}

func (v *Valuer) Contains(substr string, options ...ErrorOption) *Valuer {
	return v.simple(
		"contains",
		func(a any) bool { return strings.Contains(toString(a), substr) },
		merge(options, ErrorParam("substr", substr)),
	)
}

func (v *Valuer) ContainsAny(chars string, options ...ErrorOption) *Valuer {
	return v.simple(
		"contains_any",
		func(a any) bool { return strings.ContainsAny(toString(a), chars) },
		merge(options, ErrorParam("chars", chars)),
	)
}

func (v *Valuer) ContainsRune(rune rune, options ...ErrorOption) *Valuer {
	return v.simple(
		"contains_rune",
		func(a any) bool { return strings.ContainsRune(toString(a), rune) },
		merge(options, ErrorParam("rune", rune)),
	)
}

func (v *Valuer) Excludes(substr string, options ...ErrorOption) *Valuer {
	return v.simple(
		"excludes",
		func(val any) bool { return !strings.Contains(toString(val), substr) },
		merge(options, ErrorParam("substr", substr)),
	)
}

func (v *Valuer) ExcludesAll(chars string, options ...ErrorOption) *Valuer {
	return v.simple(
		"excludes_all",
		func(val any) bool { return !strings.ContainsAny(toString(val), chars) },
		merge(options, ErrorParam("chars", chars)),
	)
}

func (v *Valuer) ExcludesRune(rune rune, options ...ErrorOption) *Valuer {
	return v.simple(
		"excludes_rune",
		func(val any) bool { return !strings.ContainsRune(toString(val), rune) },
		merge(options, ErrorParam("rune", rune)),
	)
}

func (v *Valuer) EndsWith(suffix string, options ...ErrorOption) *Valuer {
	return v.simple(
		"ends_with",
		func(a any) bool { return strings.HasSuffix(toString(a), suffix) },
		merge(options, ErrorParam("suffix", suffix)),
	)
}

func (v *Valuer) EndsNotWith(suffix string, options ...ErrorOption) *Valuer {
	return v.simple(
		"ends_not_with",
		func(val any) bool { return !strings.HasSuffix(toString(val), suffix) },
		merge(options, ErrorParam("suffix", suffix)),
	)
}

func (v *Valuer) StartsWith(prefix string, options ...ErrorOption) *Valuer {
	return v.simple(
		"starts_with",
		func(a any) bool { return strings.HasPrefix(toString(a), prefix) },
		merge(options, ErrorParam("prefix", prefix)),
	)
}

func (v *Valuer) StartsNotWith(prefix string, options ...ErrorOption) *Valuer {
	return v.simple(
		"starts_not_with",
		func(val any) bool { return !strings.HasPrefix(toString(val), prefix) },
		merge(options, ErrorParam("prefix", prefix)),
	)
}

func (v *Valuer) OneOf(items []any, options ...ErrorOption) *Valuer {
	return v.simple(
		"one_of",
		func(value any) bool { return is.OneOf(value, items) },
		merge(options, ErrorParam("items", items)),
	)
}

func (v *Valuer) NotEmpty(options ...ErrorOption) *Valuer {
	return v.simple("not_empty", is.NotEmpty[any], options)
}

func (v *Valuer) Length(n int, options ...ErrorOption) *Valuer {
	return v.simple(
		"length",
		func(a any) bool { return is.Length(a, n, "=") },
		merge(options, ErrorParam("length", n)),
	)
}

func (v *Valuer) MinLength(min int, options ...ErrorOption) *Valuer {
	return v.simple(
		"min_length",
		func(a any) bool { return is.Length(a, min, ">=") },
		merge(options, ErrorParam("min", min)),
	)
}

func (v *Valuer) MaxLength(max int, options ...ErrorOption) *Valuer {
	return v.simple(
		"max_length",
		func(a any) bool { return is.Length(a, max, "<=") },
		merge(options, ErrorParam("max", max)),
	)
}

func (v *Valuer) LengthBetween(min, max int, options ...ErrorOption) *Valuer {
	return v.simple(
		"length_between",
		func(a any) bool { return is.LengthBetween(a, min, max) },
		merge(options, ErrorParam("min", min), ErrorParam("max", max)),
	)
}

func (v *Valuer) GreaterThan(min any, options ...ErrorOption) *Valuer {
	return v.simple(
		"greater_than",
		func(a any) bool { return is.GreaterThan(a, min) },
		merge(options, ErrorParam("min", min)),
	)
}

func (v *Valuer) GreaterEqualThan(n any, options ...ErrorOption) *Valuer {
	return v.simple(
		"greater_equal_than",
		func(a any) bool { return is.GreaterEqualThan(a, n) },
		merge(options, ErrorParam("min", n)),
	)
}

func (v *Valuer) Equal(another any, options ...ErrorOption) *Valuer {
	return v.simple(
		"equal",
		func(a any) bool { return is.Equal(a, another) },
		merge(options, ErrorParam("another", another)),
	)
}

func (v *Valuer) NotEqual(another any, options ...ErrorOption) *Valuer {
	return v.simple(
		"not_equal",
		func(a any) bool { return is.NotEqual(a, another) },
		merge(options, ErrorParam("another", another)),
	)
}

func (v *Valuer) LessEqualThan(max any, options ...ErrorOption) *Valuer {
	return v.simple(
		"less_equal_than",
		func(a any) bool { return is.LessEqualThan(a, max) },
		merge(options, ErrorParam("max", max)),
	)
}

func (v *Valuer) LessThan(max any, options ...ErrorOption) *Valuer {
	return v.simple(
		"less_than",
		func(a any) bool { return is.LessThan(a, max) },
		merge(options, ErrorParam("max", max)),
	)
}

func (v *Valuer) Between(min, max any, options ...ErrorOption) *Valuer {
	return v.simple(
		"between",
		func(a any) bool { return is.Between(a, min, max) },
		merge(options, ErrorParam("min", min), ErrorParam("max", max)),
	)
}

func (v *Valuer) NotBetween(min, max any, options ...ErrorOption) *Valuer {
	return v.simple(
		"not_between",
		func(a any) bool { return is.NotBetween(a, min, max) },
		merge(options, ErrorParam("min", min), ErrorParam("max", max)),
	)
}

type Item struct {
	Key   any
	Index int
	Value any
}

func (v *Valuer) itemize(handle func(item *Item) any, every bool, options []ErrorOption) *Valuer {
	check := func(item *Item) (bool, error) {
		res := handle(item)
		if x, ok := res.(Validatable); ok {
			if err := x.Validate(); res != nil {
				return false, err
			} else {
				return !every, nil
			}
		}
		if res == true || res == nil {
			if !every {
				// some 成功
				return true, nil
			}
			return false, nil
		}
		if res == false && every {
			if every {
				// every 失败
				return false, v.newError("every", options)
			}
			return false, nil
		}
		if err, ok := res.(error); ok {
			// 自定义错误
			return false, v.mistake(err, options...)
		}
		panic(fmt.Errorf("expect a bool, a Validatable or a error, got %+v", res))
	}

	v.addRule(func(a any) error {
		rv := reflect.Indirect(reflect.ValueOf(a))
		rt := rv.Type()
		switch k := rt.Kind(); k {
		case reflect.Array, reflect.Slice:
			for i := 0; i < rv.Len(); i++ {
				skip, err := check(&Item{
					Key:   nil,
					Index: i,
					Value: rv.Index(i).Interface(),
				})
				if err != nil {
					return v.mistake(err)
				}
				if skip {
					break
				}
			}
		case reflect.Map:
			iter := rv.MapRange()
			for iter.Next() {
				skip, err := check(&Item{
					Key:   iter.Key().String(),
					Value: iter.Value().Interface(),
				})
				if err != nil {
					return v.mistake(err)
				}
				if skip {
					break
				}
			}
		case reflect.Struct:
			for i := 0; i < rt.NumField(); i++ {
				field := rt.Field(i)
				skip, err := check(&Item{
					Key:   field.Name,
					Value: rv.FieldByName(field.Name).Interface(),
				})
				if err != nil {
					return v.mistake(err)
				}
				if skip {
					break
				}
			}
		default:
			panic("invalid value")
		}

		if every {
			return nil
		}

		return v.newError("some", options)
	})

	return v
}

func (v *Valuer) Every(handle func(item *Item) any, options ...ErrorOption) *Valuer {
	return v.itemize(handle, true, options)
}

func (v *Valuer) Some(handle func(item *Item) any, options ...ErrorOption) *Valuer {
	return v.itemize(handle, false, options)
}

func toString(val any) string {
	if str, ok := val.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", val)
}
