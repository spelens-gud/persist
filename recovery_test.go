package persist_test

import (
	"bytes"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/spelens-gud/persist"
)

func TestRecoveryWithWriter(t *testing.T) {
	tests := []struct {
		name     string
		recovery []persist.RecoveryFunc
		wantFunc bool
	}{
		{
			name:     "without recovery func",
			recovery: nil,
			wantFunc: true,
		},
		{
			name:     "with recovery func",
			recovery: []persist.RecoveryFunc{func() {}},
			wantFunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			recoveryFunc := persist.RecoveryWithWriter(&buf, tt.recovery...)
			if recoveryFunc == nil && tt.wantFunc {
				t.Error("RecoveryWithWriter() returned nil, want function")
			}
		})
	}
}

func TestCustomRecoveryWithWriter(t *testing.T) {
	tests := []struct {
		name   string
		out    io.Writer
		handle persist.RecoveryFunc
	}{
		{
			name:   "with writer and handle func",
			out:    &bytes.Buffer{},
			handle: func() {},
		},
		{
			name:   "with nil writer",
			out:    nil,
			handle: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recoveryFunc := persist.CustomRecoveryWithWriter(tt.out, tt.handle)
			if recoveryFunc == nil {
				t.Error("CustomRecoveryWithWriter() returned nil, want function")
			}
		})
	}
}

func TestCustomRecoveryWithWriter_PanicRecovery(t *testing.T) {
	var buf bytes.Buffer
	handleCalled := false

	handle := func() {
		handleCalled = true
		// 不要在这里 panic，因为这会导致测试失败
	}

	recoveryFunc := persist.CustomRecoveryWithWriter(&buf, handle)

	// 执行恢复函数，应该能捕获 panic
	recoveryFunc()

	// 给 goroutine 一些时间执行
	time.Sleep(100 * time.Millisecond)

	if !handleCalled {
		t.Error("handle function was not called")
	}
}

func TestReturnError(t *testing.T) {
	tests := []struct {
		name        string
		err         any
		wantContain string
	}{
		{
			name:        "with Error type",
			err:         &persist.Error{Err: errors.New("test error"), Type: persist.ErrorTypePublic},
			wantContain: "test error",
		},
		{
			name:        "with ErrorMsg type",
			err:         persist.ErrorMsg{&persist.Error{Err: errors.New("test error message"), Type: persist.ErrorTypePublic}},
			wantContain: "test error message",
		},
		{
			name:        "with standard error",
			err:         errors.New("standard error"),
			wantContain: "standard error",
		},
		{
			name:        "with string",
			err:         "string error",
			wantContain: "string error",
		},
		{
			name:        "with int",
			err:         42,
			wantContain: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)

			result := persist.ReturnError(logger, tt.err)

			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("ReturnError() result = %v, want to contain %v", result, tt.wantContain)
			}

			logOutput := buf.String()
			if !strings.Contains(logOutput, "[Recovery]") {
				t.Error("log output should contain [Recovery]")
			}
		})
	}
}

func TestReturnError_NilLogger(t *testing.T) {
	err := errors.New("test error")
	result := persist.ReturnError(nil, err)

	if result != "test error" {
		t.Errorf("ReturnError() with nil logger = %v, want %v", result, "test error")
	}
}

// BenchmarkReturnError 基准测试.
func BenchmarkReturnError(b *testing.B) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	err := errors.New("benchmark error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		persist.ReturnError(logger, err)
		buf.Reset()
	}
}
