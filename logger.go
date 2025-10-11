package persist

import (
	"fmt"
	"io"
	"time"
)

// consoleColorModeValue 打印日志时的颜色模式.
type consoleColorModeValue int

const (
	// autoColor 根据终端是否支持颜色自动决定是否打印颜色.
	autoColor consoleColorModeValue = iota
	// disableColor 禁用颜色.
	disableColor
	// forceColor 强制启用颜色.
	forceColor
)

const (
	green   = "\033[30;42m" // 绿色背景，白色文字 - 成功状态
	white   = "\033[30;47m" // 白色背景，黑色文字 - 默认/未知
	yellow  = "\033[30;43m" // 黄色背景，黑色文字 - 操作警告
	red     = "\033[30;41m" // 红色背景，白色文字 - 公共错误
	blue    = "\033[30;44m" // 蓝色背景，白色文字 - 私有错误
	magenta = "\033[30;45m" // 紫色背景，白色文字 - 加载错误
	cyan    = "\033[30;46m" // 青色背景，黑色文字 - 状态错误
	reset   = "\033[0m"     // 重置颜色
)

// consoleColorMode 当前的颜色模式，默认是自动.
var consoleColorMode = autoColor

// LoggerConfig 定义了 Logger 中间件的配置.
type LoggerConfig struct {
	Formatter LogFormatter
	Output    io.Writer
	SkipPaths []string
	Skip      Skipper
}

// Skipper 是一个函数，用于根据提供的 error 决定是否跳过日志记录.
type Skipper func(c error)

// LogFormatter 是传递给日志格式化函数的签名.
type LogFormatter func(params LogFormatterParams) string

// LogFormatterParams 是传递给日志格式化函数的参数结构体.
type LogFormatterParams struct {
	TimeStamp time.Time     // 记录日志的时间戳
	Latency   time.Duration // 处理所花费的时间
	Error     *Error        // 错误信息
	isTerm    bool          // 是否输出到终端
	Keys      map[any]any   // 上下文中的键值对
}

// ErrorTypeColor 根据错误类别对应的颜色代码.
func (p *LogFormatterParams) ErrorTypeColor() string {
	errorType := p.Error.Type

	switch errorType {
	case ErrorTypeState:
		return cyan // 青色背景 - 状态相关错误，相对温和
	case ErrorTypeLoad:
		return magenta // 紫色背景 - 加载错误，中等严重性
	case ErrorTypeOp:
		return yellow // 黄色背景 - 操作错误，警告级别
	case ErrorTypePrivate:
		return blue // 蓝色背景 - 内部错误，需要开发者关注
	case ErrorTypePublic:
		return red // 红色背景 - 用户可见错误，最高优先级
	case ErrorTypeAny:
		return green // 绿色背景 - 通用/成功状态
	default:
		return white // 白色背景 - 未知类型
	}
}

// ResetColor 返回重置颜色的代码.
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// PrintErrorWithColor 根据错误类型打印带颜色的错误信息.
func PrintErrorWithColor(err *Error, out io.Writer) {
	if out == nil {
		out = DefaultErrorWriter
	}

	// 创建日志格式化参数
	params := LogFormatterParams{
		TimeStamp: time.Now(),
		Error:     err,
		isTerm:    true,
		Keys:      make(map[any]any),
	}

	// 获取错误类型对应的颜色
	colorCode := params.ErrorTypeColor()
	resetCode := params.ResetColor()

	// 格式化并打印错误信息
	errorMsg := fmt.Sprintf("%s[%s] %s - %s%s\n",
		colorCode,
		getErrorTypeName(err.Type),
		timeFormat(params.TimeStamp),
		err.Error(),
		resetCode)

	fmt.Fprint(out, errorMsg)
}

// PrintErrorMsgWithColor 打印错误消息切片，每个错误根据类型显示不同颜色.
func PrintErrorMsgWithColor(errors ErrorMsg, out io.Writer) {
	if out == nil {
		out = DefaultErrorWriter
	}

	for i, err := range errors {
		fmt.Fprintf(out, "Error #%02d: ", i+1)
		PrintErrorWithColor(err, out)
		if err.Meta != nil {
			fmt.Fprintf(out, "     Meta: %v\n", err.Meta)
		}
	}
}

// getErrorTypeName 根据错误类型返回对应的名称.
func getErrorTypeName(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeState:
		return "STATE"
	case ErrorTypeLoad:
		return "LOAD"
	case ErrorTypeOp:
		return "OPERATION"
	case ErrorTypePrivate:
		return "PRIVATE"
	case ErrorTypePublic:
		return "PUBLIC"
	case ErrorTypeAny:
		return "ANY"
	default:
		return "UNKNOWN"
	}
}
