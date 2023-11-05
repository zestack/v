package v

import (
	"fmt"
	"strings"
)

var emptyErrors []*Error

func init() {
	emptyErrors = make([]*Error, 0)
}

// Error 验证错误结构体
type Error struct {
	error  error
	code   string
	format string
	params map[string]any
	field  string
	label  string
	value  any
}

// ErrorOption 错误配置函数签名
type ErrorOption func(*Error)

func merge(options []ErrorOption, presets ...ErrorOption) []ErrorOption {
	return append(presets, options...)
}

// NewError 创建错误实例
func NewError(code string, options ...ErrorOption) *Error {
	e := Error{code: code}
	for _, option := range options {
		option(&e)
	}
	return &e
}

// ErrorFormat 设置错误格式化字符串
func ErrorFormat(format string) ErrorOption {
	return func(e *Error) {
		e.format = format
	}
}

// ErrorParam 设置验证参数
func ErrorParam(key string, value any) ErrorOption {
	return func(e *Error) {
		if e.params == nil {
			e.params = make(map[string]any)
		}
		e.params[key] = value
	}
}

// ErrorCode 设置错误代码
func ErrorCode(code string) ErrorOption {
	return func(e *Error) {
		e.code = code
	}
}

// Code 返回错误代码
func (e *Error) Code() string {
	return e.code
}

// Format 返回错误格式化模板
func (e *Error) Format() string {
	return e.format
}

// Params 返回错误格式化蚕食
func (e *Error) Params() map[string]any {
	p := map[string]any{}
	if e.params != nil {
		for k, v := range e.params {
			p[k] = v
		}
	}
	return p
}

// Field 返回字段名
func (e *Error) Field() string {
	return e.field
}

// Label 返回错误标签
func (e *Error) Label() string {
	return e.label
}

// Value 返回用于验证的值
func (e *Error) Value() any {
	return e.value
}

// String 实现 fmt.Stringer 接口，返回格式化后的字符串
func (e *Error) String() string {
	message := e.format
	params := e.Params()
	params["label"] = e.label
	params["value"] = e.value
	//params["field"] = e.field
	// 定义了消息或翻译函数
	if t, found := translations[e.code]; found {
		if message == "" {
			message = t.message
		}
		if t.trans != nil {
			return t.trans(message, params)
		}
	}
	// 设置了默认翻译函数
	if defaultTranslator != nil {
		return defaultTranslator(message, params)
	}
	for key, value := range params {
		message = strings.ReplaceAll(message, "{"+key+"}", fmt.Sprintf("%v", value))
	}
	return message
}

// Error 实现内置错误接口（优先使用内部错误）
func (e *Error) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return e.String()
}

// Errors 错误集
type Errors struct {
	errors []*Error
}

// IsEmpty 是否存在错误
func (e *Errors) IsEmpty() bool {
	if e == nil || e.errors == nil {
		return true
	}
	return len(e.errors) == 0
}

// Add 添加一个错误
func (e *Errors) Add(err error) {
	if err == nil {
		return
	}
	if e.errors == nil {
		e.errors = make([]*Error, 0)
	}
	if ex, ok := err.(*Errors); ok {
		if !ex.IsEmpty() {
			e.errors = append(e.errors, ex.errors...)
		}
	} else if ex, ok := err.(*Error); ok && ex != nil {
		e.errors = append(e.errors, ex)
	} else {
		e.errors = append(e.errors, &Error{error: err})
	}
}

// First 返回第一个错误实例，如果不存在则返回 nil
func (e *Errors) First() *Error {
	if e.IsEmpty() {
		return nil
	}
	return e.errors[0]
}

// Get 获取指定标签的错误列表，如果不存在将返回 nil
func (e *Errors) Get(field string) []*Error {
	if e.IsEmpty() {
		return nil
	}
	errs := make([]*Error, 0)
	for _, err := range e.errors {
		if err.field == field {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (e *Errors) All() []*Error {
	if e == nil {
		return emptyErrors
	}
	return e.errors
}

// ToMap 根据错误标签分组
func (e *Errors) ToMap() map[string][]*Error {
	if e == nil || e.IsEmpty() {
		return nil
	}
	errs := map[string][]*Error{}
	for _, err := range e.errors {
		if _, ok := errs[err.field]; !ok {
			errs[err.field] = []*Error{}
		}
		errs[err.field] = append(errs[err.field], err)
	}
	return errs
}

func (e *Errors) String() string {
	errsMap := e.ToMap()
	if errsMap == nil {
		return ""
	}

	var errors []string
	for _, errs := range errsMap {
		buf := strings.Builder{}
		for i, err := range errs {
			if i > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(err.String())
		}
		errors = append(errors, buf.String())
	}
	return strings.Join(errors, "\n")
}

func (e *Errors) Error() string {
	return e.String()
}

func isBuiltinError(err error) bool {
	if _, ok := err.(*Error); ok {
		return true
	}
	if _, ok := err.(*Errors); ok {
		return true
	}
	return false
}
