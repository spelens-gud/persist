package persist

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const dunno = "???"

var dunnoBytes = []byte(dunno)

// DefaultWriter 默认的 io.Writer 的调试输出和中间件输出，如 Logger() 或 Recovery().
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter 用于调试错误的默认 io.Writer.
var DefaultErrorWriter io.Writer = os.Stderr

// RecoveryFunc 定义了可传递给 CustomRecovery 的函数.
type RecoveryFunc func()

// RecoveryWithWriter 返回一个中间件, 用于给定的 writer, 从任何 panic 中恢复, 如果有 panic 则写入 500.
func RecoveryWithWriter(out io.Writer, recovery ...RecoveryFunc) func() {
	if len(recovery) > 0 {
		return CustomRecoveryWithWriter(out, recovery[0])
	}

	return CustomRecoveryWithWriter(out, defaultHandleRecovery)
}

// CustomRecoveryWithWriter 返回一个中间件, 用于给定的 writer, 从任何 panic 中恢复, 并调用提供的 handle 函数来处理它.
func CustomRecoveryWithWriter(out io.Writer, handle RecoveryFunc) func() {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "\n\n\x1b[31m", log.LstdFlags)
		logger.SetOutput(out)
	}

	return func() {
		defer func() {
			var err any
			if err = recover(); err == nil {
				return
			}

			ReturnError(logger, err)
		}()

		go handle()
	}
}

func ReturnError(logger *log.Logger, err any) (errorOutput string) {
	switch e := err.(type) {
	case *Error:
		if jsonBytes, jsonErr := e.MarshalJSON(); jsonErr == nil {
			errorOutput = string(jsonBytes)
		} else {
			errorOutput = e.Error()
		}
	case ErrorMsg:
		if jsonBytes, jsonErr := e.MarshalJSON(); jsonErr == nil {
			errorOutput = string(jsonBytes)
		} else {
			errorOutput = e.String()
		}
	case error:
		errorOutput = e.Error()
	default:
		errorOutput = fmt.Sprintf("%v", err)
	}

	if logger != nil {
		const stackSkip = 3 // 跳过调用栈的前 3 帧
		logger.Printf("[Recovery] %s panic recovered:\n%s\n%s%s",
			timeFormat(time.Now()), errorOutput, stack(stackSkip), "\033[0m")
	}

	return
}

// stack 返回调用栈的格式化表示, 跳过前 skip 个帧.
func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}

	return buf.Bytes()
}

// function 返回函数名, 如果无法获取则返回 "???".
func function(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := fn.Name()

	if lastSlash := strings.LastIndexByte(name, '/'); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := strings.IndexByte(name, '.'); period >= 0 {
		name = name[period+1:]
	}
	name = strings.ReplaceAll(name, "·", ".")

	return name
}

// source 返回第 n 行的源代码.
func source(lines [][]byte, n int) []byte {
	n--
	if n < 0 || n >= len(lines) {
		return dunnoBytes
	}

	return bytes.TrimSpace(lines[n])
}

// defaultHandleRecovery 恢复的默认处理程序.
func defaultHandleRecovery() {
	// 默认恢复处理程序, 可根据需要自定义.
}

// timeFormat 格式化时间为 "YYYY/MM/DD - HH:MM:SS".
func timeFormat(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}
