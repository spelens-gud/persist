package persist

import (
	"os"
	"reflect"
)

// GetFieldNames 获取结构体字段名
func GetFieldNames(obj any) []string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}

// GetFieldNum 获取结构体字段数量
func GetFieldNum(obj any) int {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.NumField()
}

// GetFieldValueByTag 获取结构体字段值通过tag
func GetFieldValueByTag(obj any, tagName, tagValue string) (value any, found bool) {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get(tagName) == tagValue {
			return v.Field(i).Interface(), true
		}
	}
	return nil, false
}

// DirExists 判断路径是否存在
func DirExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
