package persist_test

import (
	"errors"
	"testing"

	"encoding/json"

	"github.com/spelens-gud/persist"
)

// TestError_SetType 测试设置错误类型.
func TestError_SetType(t *testing.T) {
	err := &persist.Error{
		Err:  errors.New("test error"),
		Type: persist.ErrorTypePrivate,
	}

	result := err.SetType(persist.ErrorTypePublic)

	if result.Type != persist.ErrorTypePublic {
		t.Errorf("Expected type %d, got %d", persist.ErrorTypePublic, result.Type)
	}

	if !errors.Is(result, err) {
		t.Error("SetType should return the same instance")
	}
}

// TestError_SetMeta 测试设置错误元数据.
func TestError_SetMeta(t *testing.T) {
	err := &persist.Error{
		Err: errors.New("test error"),
	}

	meta := map[string]string{"key": "value"}
	result := err.SetMeta(meta)

	if result.Meta == nil {
		t.Error("Meta was not set")
	}

	if !errors.Is(result, err) {
		t.Error("SetMeta should return the same instance")
	}
}

// TestError_Error 测试 Error 方法返回正确的错误消息.
func TestError_Error(t *testing.T) {
	originalErr := errors.New("original error message")
	err := &persist.Error{
		Err: originalErr,
	}

	if err.Error() != "original error message" {
		t.Errorf("Expected 'original error message', got '%s'", err.Error())
	}
}

// TestError_IsType 测试 IsType 方法的各种场景.
func TestError_IsType(t *testing.T) {
	tests := []struct {
		name      string
		errType   persist.ErrorType
		checkType persist.ErrorType
		expected  bool
	}{
		{
			name:      "exact match",
			errType:   persist.ErrorTypePublic,
			checkType: persist.ErrorTypePublic,
			expected:  true,
		},
		{
			name:      "no match",
			errType:   persist.ErrorTypePublic,
			checkType: persist.ErrorTypePrivate,
			expected:  false,
		},
		{
			name:      "combined flags match",
			errType:   persist.ErrorTypePublic | persist.ErrorTypeOp,
			checkType: persist.ErrorTypePublic,
			expected:  true,
		},
		{
			name:      "any type",
			errType:   persist.ErrorTypePublic,
			checkType: persist.ErrorTypeAny,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &persist.Error{
				Err:  errors.New("test"),
				Type: tt.errType,
			}

			result := err.IsType(tt.checkType)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestError_Unwrap 测试 Unwrap 方法返回原始错误.
func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("wrapped error")
	err := &persist.Error{
		Err: originalErr,
	}

	unwrapped := err.Unwrap()
	if !errors.Is(unwrapped, originalErr) {
		t.Error("Unwrap should return the original error")
	}
}

// TestError_JSON 测试 JSON 方法的各种元数据类型.
func TestError_JSON(t *testing.T) {
	tests := []struct {
		name     string
		err      *persist.Error
		expected map[string]interface{}
	}{
		{
			name: "with struct meta",
			err: &persist.Error{
				Err: errors.New("test error"),
				Meta: struct {
					Field string `json:"field"`
				}{Field: "value"},
			},
			expected: map[string]interface{}{
				"field": "value",
			},
		},
		{
			name: "with map meta",
			err: &persist.Error{
				Err: errors.New("test error"),
				Meta: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expected: map[string]interface{}{
				"key1":  "value1",
				"key2":  "value2",
				"error": "test error",
			},
		},
		{
			name: "with other meta type",
			err: &persist.Error{
				Err:  errors.New("test error"),
				Meta: "string meta",
			},
			expected: map[string]interface{}{
				"meta":  "string meta",
				"error": "test error",
			},
		},
		{
			name: "no meta",
			err: &persist.Error{
				Err: errors.New("test error"),
			},
			expected: map[string]interface{}{
				"error": "test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.JSON()

			if tt.name == "with struct meta" {
				if _, ok := result.(struct {
					Field string `json:"field"`
				}); !ok {
					t.Error("Expected struct meta to be returned as-is")
				}

				return
			}

			resultMap, ok := result.(persist.H)
			if !ok {
				t.Fatalf("Expected result to be H type, got %T", result)
			}

			for key, expectedValue := range tt.expected {
				if resultMap[key] != expectedValue {
					t.Errorf("Expected %s to be %v, got %v", key, expectedValue, resultMap[key])
				}
			}
		})
	}
}

// TestError_MarshalJSON 测试 MarshalJSON 方法生成正确的 JSON.
func TestError_MarshalJSON(t *testing.T) {
	err := &persist.Error{
		Err:  errors.New("test error"),
		Meta: map[string]string{"key": "value"},
	}

	data, jsonErr := err.MarshalJSON()
	if jsonErr != nil {
		t.Fatalf("MarshalJSON failed: %v", jsonErr)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", unmarshalErr)
	}

	if result["error"] != "test error" {
		t.Errorf("Expected error to be 'test error', got %v", result["error"])
	}

	if result["key"] != "value" {
		t.Errorf("Expected key to be 'value', got %v", result["key"])
	}
}
func TestErrorMsg_ByType(t *testing.T) {
	errs := persist.ErrorMsg{
		&persist.Error{Err: errors.New("error1"), Type: persist.ErrorTypePublic},
		&persist.Error{Err: errors.New("error2"), Type: persist.ErrorTypePrivate},
		&persist.Error{Err: errors.New("error3"), Type: persist.ErrorTypePublic | persist.ErrorTypeOp},
		&persist.Error{Err: errors.New("error4"), Type: persist.ErrorTypeLoad},
	}

	tests := []struct {
		name     string
		typ      persist.ErrorType
		expected int
	}{
		{
			name:     "public errors",
			typ:      persist.ErrorTypePublic,
			expected: 2,
		},
		{
			name:     "private errors",
			typ:      persist.ErrorTypePrivate,
			expected: 1,
		},
		{
			name:     "bind errors",
			typ:      persist.ErrorTypeOp,
			expected: 1,
		},
		{
			name:     "any errors",
			typ:      persist.ErrorTypeAny,
			expected: 4,
		},
		{
			name:     "render errors",
			typ:      persist.ErrorTypeLoad,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errs.ByType(tt.typ)
			if len(result) != tt.expected {
				t.Errorf("Expected %d errors, got %d", tt.expected, len(result))
			}
		})
	}

	emptyErrors := persist.ErrorMsg{}
	result := emptyErrors.ByType(persist.ErrorTypePublic)
	if result != nil {
		t.Error("Expected nil for empty ErrorMsg")
	}
}

// TestErrorMsg_Last 测试 Last 方法返回最后一个错误.
func TestErrorMsg_Last(t *testing.T) {
	lastErr := &persist.Error{Err: errors.New("last error")}
	errs := persist.ErrorMsg{
		&persist.Error{Err: errors.New("first error")},
		&persist.Error{Err: errors.New("second error")},
		lastErr,
	}

	result := errs.Last()
	if !errors.Is(result, lastErr) {
		t.Error("Last() should return the last error")
	}

	emptyErrors := persist.ErrorMsg{}
	result = emptyErrors.Last()
	if result != nil {
		t.Error("Last() should return nil for empty ErrorMsg")
	}
}

// TestErrorMsg_Errors 测试 Errors 方法返回正确的错误消息切片.
func TestErrorMsg_Errors(t *testing.T) {
	errs := persist.ErrorMsg{
		&persist.Error{Err: errors.New("error1")},
		&persist.Error{Err: errors.New("error2")},
		&persist.Error{Err: errors.New("error3")},
	}

	result := errs.Errors()
	expected := []string{"error1", "error2", "error3"}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d errors, got %d", len(expected), len(result))
	}

	for i, err := range result {
		if err != expected[i] {
			t.Errorf("Expected error %d to be '%s', got '%s'", i, expected[i], err)
		}
	}

	// Test empty ErrorMsg
	emptyErrors := persist.ErrorMsg{}
	result = emptyErrors.Errors()
	if result != nil {
		t.Error("Errors() should return nil for empty ErrorMsg")
	}
}

// TestErrorMsg_JSON 测试 JSON 方法返回正确的 JSON 结构.
func TestErrorMsg_JSON(t *testing.T) {
	tests := []struct {
		name     string
		errors   persist.ErrorMsg
		expected interface{}
	}{
		{
			name:     "empty errors",
			errors:   persist.ErrorMsg{},
			expected: nil,
		},
		{
			name: "single error",
			errors: persist.ErrorMsg{
				&persist.Error{Err: errors.New("single error")},
			},
			expected: map[string]interface{}{
				"error": "single error",
			},
		},
		{
			name: "multiple errors",
			errors: persist.ErrorMsg{
				&persist.Error{Err: errors.New("error1")},
				&persist.Error{Err: errors.New("error2")},
			},
			expected: []interface{}{
				map[string]interface{}{"error": "error1"},
				map[string]interface{}{"error": "error2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.JSON()

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}

				return
			}

			// For single error case
			if tt.name == "single error" {
				resultMap, ok := result.(persist.H)
				if !ok {
					t.Fatalf("Expected H type, got %T", result)
				}
				expectedMap := tt.expected.(map[string]interface{})
				if resultMap["error"] != expectedMap["error"] {
					t.Errorf("Expected error to be %v, got %v", expectedMap["error"], resultMap["error"])
				}

				return
			}

			if tt.name == "multiple errors" {
				resultSlice, ok := result.([]interface{})
				if !ok {
					t.Fatalf("Expected []interface{} type, got %T", result)
				}
				if len(resultSlice) != 2 {
					t.Errorf("Expected 2 errors, got %d", len(resultSlice))
				}
			}
		})
	}
}

// TestErrorMsg_MarshalJSON 测试 MarshalJSON 方法生成正确的 JSON.
func TestErrorMsg_MarshalJSON(t *testing.T) {
	errs := persist.ErrorMsg{
		&persist.Error{Err: errors.New("error1")},
		&persist.Error{Err: errors.New("error2")},
	}

	data, err := errs.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result []map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", unmarshalErr)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result))
	}

	if result[0]["error"] != "error1" {
		t.Errorf("Expected first error to be 'error1', got %v", result[0]["error"])
	}

	if result[1]["error"] != "error2" {
		t.Errorf("Expected second error to be 'error2', got %v", result[1]["error"])
	}
}

// TestErrorMsg_String 测试 String 方法返回正确的格式化字符串.
func TestErrorMsg_String(t *testing.T) {
	tests := []struct {
		name     string
		errors   persist.ErrorMsg
		expected string
	}{
		{
			name:     "empty errors",
			errors:   persist.ErrorMsg{},
			expected: "",
		},
		{
			name: "single error without meta",
			errors: persist.ErrorMsg{
				&persist.Error{Err: errors.New("test error")},
			},
			expected: "Error #01: test error\n",
		},
		{
			name: "single error with meta",
			errors: persist.ErrorMsg{
				&persist.Error{
					Err:  errors.New("test error"),
					Meta: "some meta",
				},
			},
			expected: "Error #01: test error\n     Meta: some meta\n",
		},
		{
			name: "multiple errors",
			errors: persist.ErrorMsg{
				&persist.Error{Err: errors.New("error1")},
				&persist.Error{
					Err:  errors.New("error2"),
					Meta: "meta2",
				},
			},
			expected: "Error #01: error1\nError #02: error2\n     Meta: meta2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.String()
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
