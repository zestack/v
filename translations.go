package v

import (
	"fmt"
	"reflect"
)

// Translator 翻译函数签名
type Translator func(message string, params map[string]any) string

var (
	// 默认翻译函数
	defaultTranslator Translator
	// 预置的翻译信息
	translations = map[string]struct {
		message string
		trans   Translator
	}{
		"required":                {message: "{label}为必填字段"},
		"required_if":             {message: "{label}为必填字段"},
		"typeof":                  {trans: typeof},
		"is_email":                {message: "{label}不是有效的电子邮箱地址"},
		"is_e164":                 {message: "{label}不是有效的 e.164 手机号码"},
		"is_phone_number":         {message: "{label}不是有效的手机号码"},
		"is_url":                  {message: "{label}不是有效的链接"},
		"is_url_encoded":          {message: "{label}不是有效的链接"},
		"is_base64_url":           {message: "{label}不是有效的BASE64链接"},
		"is_semver":               {message: "{label}不是有效的语义化版本号"},
		"is_jwt":                  {message: "{label}不是有效的权限令牌"},
		"is_uuid":                 {message: "{label}不是有效的UUID字符串"},
		"is_uuid3":                {message: "{label}不是有效的V3版UUID字符串"},
		"is_uuid4":                {message: "{label}不是有效的V4版UUID字符串"},
		"is_uuid5":                {message: "{label}不是有效的V5版UUID字符串"},
		"is_ulid":                 {message: "{label}不是有效的ULID字符串"},
		"is_md4":                  {message: ""},
		"is_md5":                  {message: ""},
		"is_sha256":               {message: ""},
		"is_sha384":               {message: ""},
		"is_sha512":               {message: ""},
		"is_ascii":                {message: "{label}只能包含ASCII字符"},
		"is_alpha":                {message: "{label}只能包含字母"},
		"is_alphanumeric":         {message: "{label}只能包含字母和数字"},
		"is_alpha_unicode":        {message: "{label}只能包含字母和Unicode字符"},
		"is_alphanumeric_unicode": {message: "{label}只能包含字母数字和Unicode字符"},
		"is_numeric":              {message: "{label}必须是一个有效的数值"},
		"is_number":               {message: "{label}必须是一个有效的数字"},
		"is_bool":                 {message: "{label}必须是一个有效的布尔值"},
		"is_hexadecimal":          {message: "{label}必须是一个有效的十六进制"},
		"is_hexcolor":             {message: "{label}必须是一个有效的十六进制颜色"},
		"is_rgb":                  {message: "{label}必须是一个有效的RGB颜色"},
		"is_rgba":                 {message: "{label}必须是一个有效的RGBA颜色"},
		"is_hsl":                  {message: "{label}必须是一个有效的RGB颜色"},
		"is_hsla":                 {message: "{label}必须是一个有效的HSLA颜色"},
		"is_color":                {message: "{label}必须是一个有效的颜色"},
		"is_latitude":             {message: "{label}必须包含有效的纬度坐标"},
		"is_longitude":            {message: "{label}必须包含有效的经度坐标"},
		"is_json":                 {message: "{label}必须是一个JSON字符串"},
		"is_base64":               {message: "{label}必须是一个有效的Base64字符串"},
		"is_html":                 {message: "{label}必须是一个有效的网页内容"},
		"is_html_encoded":         {message: "{label}必须是一个被转义的网页内容"},
		"is_datetime":             {message: "{label}的格式必须是{layout}"},
		"is_timezone":             {message: "{label}必须是一个有效的时区"},
		"is_ipv4":                 {message: "{label}必须是一个有效的IPv4地址"},
		"is_ipv6":                 {message: "{label}必须是一个有效的IPv6地址"},
		"is_ip":                   {message: "{label}必须是一个有效的IP地址"},
		"is_mac":                  {message: "{label}必须是一个有效的MAC地址"},
		"is_file":                 {message: "{label}必须是一个有效的文件"},
		"is_dir":                  {message: "{label}必须是一个有效的目录"},
		"is_lower":                {message: "{label}必须是小写字母"},
		"is_upper":                {message: "{label}必须是大写字母"},
		"contains":                {message: "{label}必须包含文本'{substr}'"},
		"contains_any":            {message: "{label}必须包含至少一个以下字符'{chars}'"},
		"contains_rune":           {message: "{label}必须包含字符'{rune}'"},
		"excludes":                {message: "{label}不能包含文本'{substr}'"},
		"excludes_all":            {message: "{label}不能包含以下任何字符'{chars}'"},
		"excludes_rune":           {message: "{label}不能包含'{rune}'"},
		"ends_with":               {message: "{label}必须以文本'{suffix}'结尾"},
		"ends_not_with":           {message: "{label}不能以文本'{suffix}'结尾"},
		"starts_with":             {message: "{label}必须以文本'{prefix}'开头"},
		"starts_not_with":         {message: "{label}不能以文本'{prefix}'开头"},
		"one_of":                  {message: "{label}必须是[{items}]中的一个"},
		"not_empty":               {message: "{label}不能为空"},
		"length":                  {message: "{label}长度必须是{length}"},
		"min_length":              {message: "{label}最小长度为{min}"},
		"max_length":              {message: "max_length"},
		"length_between":          {message: "{label}长度必须大于或等于{min}且小于或等于{max}"},
		"greater_than":            {message: "{label}必须大于{min}"},
		"greater_equal_than":      {message: "{label}必须大于或等于{min}"},
		"equal":                   {message: "{label}必须等于{another}"},
		"not_equal":               {message: "{label}不能等于{another}"},
		"less_equal_than":         {message: "{label}必须小于或等于{max}"},
		"less_than":               {message: "{label}必须小于{max}"},
		"between":                 {message: "{label}必须大于或等于{min}且小于或等于{max}"},
		"not_between":             {message: "{label}必须小于{min}或大于{max}"},
		"some":                    {message: "{label}至少有一个子项通过验证"},
		"every":                   {message: "{label}的所有子项必须通过验证"},
		"entity_exists":           {message: "{label}不存在"},
		"entity_not_exists":       {message: "{label}已经存在"},
		"index_by":                {message: "参数不完整"},
	}

	types = map[reflect.Kind]string{
		reflect.Bool:       "布尔值",
		reflect.Int:        "整数",
		reflect.Int8:       "整数",
		reflect.Int16:      "整数",
		reflect.Int32:      "整数",
		reflect.Int64:      "整数",
		reflect.Uint:       "正整数",
		reflect.Uint8:      "正整数",
		reflect.Uint16:     "正整数",
		reflect.Uint32:     "正整数",
		reflect.Uint64:     "正整数",
		reflect.Uintptr:    "正整数",
		reflect.Float32:    "浮点数",
		reflect.Float64:    "浮点数",
		reflect.Complex64:  "复数",
		reflect.Complex128: "复数",
		reflect.Array:      "数组",
		reflect.Map:        "字典(Mapper)",
		reflect.Slice:      "切片(Slice)",
		reflect.String:     "字符串",
		reflect.Struct:     "结构体",
	}
)

func typeof(_ string, params map[string]any) string {
	kind := params["kind"].(reflect.Kind)
	if msg, ok := types[kind]; ok {
		return fmt.Sprintf("%s不是有效的%s", params["label"], msg)
	} else {
		return fmt.Sprintf("%s格式验证失败", params["label"])
	}
}

// SetDefaultTranslator 设置默认翻译函数
func SetDefaultTranslator(translator Translator) {
	defaultTranslator = translator
}

// DefaultTranslator 返回默认翻译函数
func DefaultTranslator() Translator {
	return defaultTranslator
}
