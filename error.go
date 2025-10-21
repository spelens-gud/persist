package persist

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// ErrorType 是一个无符号 64 位的错误代码.
type ErrorType uint64

const (
	// ErrorTypeState 状态错误.
	ErrorTypeState ErrorType = 1 << 63
	// ErrorTypeLoad 加载错误.
	ErrorTypeLoad ErrorType = 1 << 62
	// ErrorTypeOp 操作错误.
	ErrorTypeOp ErrorType = 1 << 61
	// ErrorTypePrivate 私有错误.
	ErrorTypePrivate ErrorType = 1 << 0
	// ErrorTypePublic 公共错误.
	ErrorTypePublic ErrorType = 1 << 1
	// ErrorTypeAny 任何其他错误.
	ErrorTypeAny ErrorType = 1<<64 - 1
)

// Error 表示错误的规范.
type Error struct {
	Err  error     // Err 是底层的错误.
	Type ErrorType // Type 是错误的类型.
	Meta any       // Meta 是与错误相关的任何元数据.
}

// ErrorMsg 是错误的切片.
type ErrorMsg []*Error

// H 是一个方便的 map[string]any 别名.
type H map[string]any

// 确保 Error 实现了 error 接口.
var _ error = (*Error)(nil)

// SetType 设置错误的类型.
func (msg *Error) SetType(flags ErrorType) *Error {
	msg.Type = flags

	return msg
}

// SetMeta 设置错误的元数据.
func (msg *Error) SetMeta(data any) *Error {
	msg.Meta = data

	return msg
}

// JSON 创建一个适当格式化的 JSON.
func (msg *Error) JSON() any {
	jsonData := H{}
	if msg.Meta != nil {
		value := reflect.ValueOf(msg.Meta)
		switch value.Kind() {
		case reflect.Struct:
			return msg.Meta
		case reflect.Map:
			for _, key := range value.MapKeys() {
				jsonData[key.String()] = value.MapIndex(key).Interface()
			}
		default:
			jsonData["meta"] = msg.Meta
		}
	}
	if _, ok := jsonData["error"]; !ok {
		jsonData["error"] = msg.Error()
	}

	return jsonData
}

// MarshalJSON 返回错误的 JSON 编码.
func (msg *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.JSON())
}

// Error 实现 error 接口.
func (msg *Error) Error() string {
	return msg.Err.Error()
}

// IsType 检查错误类型是否与提供的标志匹配.
func (msg *Error) IsType(flags ErrorType) bool {
	return (msg.Type & flags) > 0
}

// Unwrap 返回底层错误.
func (msg *Error) Unwrap() error {
	return msg.Err
}

// Println 根据错误类型打印带颜色的错误信息.
func (msg *Error) Println(out io.Writer) {
	if out == nil {
		out = DefaultErrorWriter
	}

	// 创建日志格式化参数
	params := LogFormatterParams{
		TimeStamp: time.Now(),
		Error:     msg,
		isTerm:    true,
		Keys:      make(map[any]any),
	}

	// 获取错误类型对应的颜色
	colorCode := params.ErrorTypeColor()
	resetCode := params.ResetColor()

	// 格式化并打印错误信息
	errorMsg := fmt.Sprintf("%s[%s] %s - %s%s\n",
		colorCode,
		getErrorTypeName(msg.Type),
		timeFormat(params.TimeStamp),
		msg.Error(),
		resetCode)

	fmt.Fprint(out, errorMsg)
}

// ByType 根据错误类型过滤错误信息.
func (a ErrorMsg) ByType(typ ErrorType) ErrorMsg {
	if len(a) == 0 {
		return nil
	}
	if typ == ErrorTypeAny {
		return a
	}
	var result ErrorMsg
	for _, msg := range a {
		if msg.IsType(typ) {
			result = append(result, msg)
		}
	}

	return result
}

// Last 返回最后一个错误信息，如果没有错误则返回 nil.
func (a ErrorMsg) Last() *Error {
	if length := len(a); length > 0 {
		return a[length-1]
	}

	return nil
}

// Errors 返回错误信息的字符串切片表示形式.
func (a ErrorMsg) Errors() []string {
	if len(a) == 0 {
		return nil
	}
	errorStrings := make([]string, len(a))
	for i, err := range a {
		errorStrings[i] = err.Error()
	}

	return errorStrings
}

// JSON 对返回错误信息的 JSON 表示形式.
func (a ErrorMsg) JSON() any {
	switch length := len(a); length {
	case 0:
		return nil
	case 1:
		return a.Last().JSON()
	default:
		jsonData := make([]any, length)
		for i, err := range a {
			jsonData[i] = err.JSON()
		}

		return jsonData
	}
}

// MarshalJSON 返回错误信息的 JSON 编码.
func (a ErrorMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.JSON())
}

// String 返回错误信息的字符串表示形式.
func (a ErrorMsg) String() string {
	if len(a) == 0 {
		return ""
	}
	var buffer strings.Builder
	for i, msg := range a {
		fmt.Fprintf(&buffer, "Error #%02d: %s\n", i+1, msg.Err)
		if msg.Meta != nil {
			fmt.Fprintf(&buffer, "     Meta: %v\n", msg.Meta)
		}
	}

	return buffer.String()
}

// Println 打印错误消息切片，每个错误根据类型显示不同颜色.
func (a ErrorMsg) Println(out io.Writer) {
	if out == nil {
		out = DefaultErrorWriter
	}

	for _, err := range a {
		err.Println(out)
	}
}
