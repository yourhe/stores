package stores // import "github.com/yourhe/stores"

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Store 存储接口
type Store interface {
	Write(string, interface{}) error
	// Read2(string, *interface{}) error
	Read(string, interface{}) error // key, out set value interface{} like &user is point interface{}
	// ReadAll2(*[]interface{}) error
	ReadAll(interface{}) error //out set value interface{} like &[]users is array struct
	Keys(...string) []string   //out set value interface{} like &[]users is array struct
	Perfix() string
}

var NotFoundError = errors.New("key not found")
var NotPrtError = errors.New("输出对象必须是指针结构")
var NotSliceError = errors.New("输出对象必须是Slice结构")

func IPtrToValue(iptr interface{}) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(iptr))
}

func SplitKeys(key string) (path string, subkey string) {
	idx := strings.LastIndex(key, "/")
	if idx > 0 {
		path = key[:idx]
		subkey = key[idx+1:]
	} else {
		subkey = key
	}
	return path, subkey
}

func ValidPtr(v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("错误: have: %s, want: %s", v.Kind(), reflect.Ptr)
	}
	return nil
}

func VaildListStruct(v reflect.Value) error {
	if err := ValidPtr(v); err != nil {
		return err
	}
	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("错误: have: %s, want: %s", v.Kind(), reflect.Slice)
	}
	return nil
}

func MakeSlice(v reflect.Value, count int) (*reflect.Value, error) {
	if err := VaildListStruct(v); err != nil {
		return nil, err
	}
	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	v.Set(reflect.MakeSlice(v.Type(), count, count))
	return &v, nil
}

func IsPath(key string) bool {
	return strings.HasSuffix(key, "/") || key == ""
}
