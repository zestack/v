package v

import (
	"errors"
	"strings"

	"zestack.dev/is"
)

// Validatable 验证功能接口
type Validatable interface {
	Validate() error
}

// Checker 功能验证函数签名
type Checker func() error

// Validate 实现 Validatable 接口
func (c Checker) Validate() error {
	return c()
}

// Wrap 将值和值的验证函数包装成验证器
func Wrap(value any, check func(any) error) Checker {
	return func() error {
		return check(value)
	}
}

// Every 每一项都要验证通过
func Every(validators ...Validatable) Checker {
	return func() error {
		for _, validator := range validators {
			if err := validator.Validate(); err != nil {
				return err
			}
		}
		return nil
	}
}

// Some 任意一项验证通过即可
func Some(validators ...Validatable) Checker {
	return func() error {
		errs := &Errors{}
		var hasOk bool
		for _, validator := range validators {
			err := validator.Validate()
			if err != nil {
				errs.Add(err)
			} else {
				hasOk = true
			}
		}
		if hasOk || errs.IsEmpty() {
			return nil
		}
		buf := strings.Builder{}
		buf.WriteString("一下错误至少满足一项：\n")
		for _, line := range strings.Split(errs.Error(), "\n") {
			buf.WriteString("  " + line + "\n")
		}
		return errors.New(strings.TrimSpace(buf.String()))
	}
}

// Validate 执行多个验证器
func Validate(validations ...Validatable) error {
	var errs Errors
	for _, validation := range validations {
		if validation == nil {
			continue
		}
		if err := validation.Validate(); err != nil {
			errs.Add(err)
		}
	}
	if errs.IsEmpty() {
		return nil
	}
	return &errs
}

// Check 逐条执行验证器，一旦验证未通过，立即返回
func Check(validations ...Validatable) error {
	for _, validation := range validations {
		if validation == nil {
			continue
		}
		if err := validation.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// IndexBy 分组验证，只要其中一组验证通过就返回
func IndexBy(index *int, values [][]any, options ...ErrorOption) Checker {
	return func() error {
		for i, items := range values {
			count := 0
			for _, item := range items {
				if is.Empty(item) {
					break
				}
				count++
			}
			if count > 0 && count == len(items) {
				*index = i
				return nil
			}
		}
		return NewError("index_by", options...)
	}
}

// Map 通过 map 构建值验证器
func Map(data map[string]any) func(name, label string) *Valuer {
	return func(name, label string) *Valuer {
		if val, ok := data[name]; ok {
			return Value(val, name, label)
		} else {
			return Value(nil, name, label)
		}
	}
}
