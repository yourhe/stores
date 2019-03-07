package memory

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/yourhe/stores"
)

//NewMemoryStore ...
func NewMemoryStore() *Memory {
	return &Memory{
		list: make(map[string]map[string]interface{}),
		// isSync: make(map[string]bool),
		mux: sync.RWMutex{},
	}
}

//Memory ...
type Memory struct {
	list    map[string]map[string]interface{}
	keys    []string
	mux     sync.RWMutex
	mapSync sync.Map
}

// CreateUser 新建用户
// func (m *Memory) CreateUser(si *users.SystemInfo) error {
// 	if si.Id == "" {
// 		si.Id = m.GenID()
// 	}
// 	return m.Write(si.Id, si)
// }

// Write 写入内存
// Write("key",User{})
func (m *Memory) Write(key string, value interface{}) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	path, subkey := SplitKeys(key)
	// "key/key1"
	// list[key][key] = value

	// value != nil save
	if value != nil {
		if m.list[path] == nil {
			m.list[path] = make(map[string]interface{})
		}
		m.list[path][subkey] = value
		return nil
	}
	//value is nil, delete that
	if _, ok := m.list[path][subkey]; ok {
		delete(m.list[path], key)
	}

	return nil
}

//Read2 请使用Read
func (m *Memory) Read2(key string, value *interface{}) error {
	panic("Read2 该方法已删除")

	var ok bool
	m.mux.RLock()
	defer m.mux.RUnlock()
	path, subkey := SplitKeys(key)
	*value, ok = m.list[path][subkey]
	if !ok {
		return fmt.Errorf("not found")
	}
	return nil
}

// ErrPtrNeeded ,,,
var ErrPtrNeeded = errors.New("provided target must be a pointer to a valid variable")

var SplitKeys = stores.SplitKeys

// Read 读取
// Read("key",&User{})
func (m *Memory) Read(key string, to interface{}) error {
	path, subkey := SplitKeys(key)
	dstt, dstv := getModelDefinition(to)
	// fmt.Println(dstt.Kind(), key)
	if dstt.Kind() == reflect.Slice {
		return m.readAll(key, to)
	}
	m.mux.RLock()
	defer m.mux.RUnlock()
	var ok bool
	var val interface{}
	val, ok = m.list[path][subkey]
	if !ok {
		return stores.NotFoundError
	}
	_, sstv := getModelDefinition(val)
	dstv.Set(sstv)
	return nil
}

// ReadAll read all
// var users = []*users{}
// m.ReadAll(&users)
func (m *Memory) ReadAll(to interface{}) (err error) {
	return m.readAll("", to)
}

func (m *Memory) readAll(path string, to interface{}) (err error) {
	// m.mux.RLock()
	// defer m.mux.RUnlock()

	keys := m.Keys(path)
	if len(keys) < 1 {
		return stores.NotFoundError
	}

	v := reflect.ValueOf(to)
	// dv := reflect.Indirect(reflect.ValueOf(to))
	// if dv.Kind() != reflect.Slice || dv.Type().Elem().Kind() != reflect.Ptr {
	// 	return stores.NotSliceError
	// }
	// dv.Set(reflect.MakeSlice(dv.Type(), len(m.list[path]), len(m.list[path])))

	nv, err := stores.MakeSlice(v, len(keys))
	if err != nil {
		return err
	}
	v = *nv
	// var i = 0
	// keys := m.Keys(path)
	// for key := range m.list[path] {
	for i, key := range keys {

		if v.Index(i).IsNil() {
			v.Index(i).Set(reflect.New(v.Index(i).Type().Elem()))
		}
		m.Read(key, v.Index(i).Interface())
		// i++
	}
	return

}

func (m *Memory) ReadAll2(values *[]interface{}) error {
	panic("ReadAll2 该方法已删除")
	m.mux.RLock()
	defer m.mux.RUnlock()
	for key := range m.list {
		*values = append(*values, m.list[key])
	}
	return nil
}

func getModelDefinition(s interface{}) (reflect.Type, reflect.Value) {
	v := reflect.ValueOf(s)
	v = reflect.Indirect(v)
	t := v.Type()
	return t, v
}

func (m *Memory) Keys(path ...string) []string {
	m.mux.RLock()
	defer m.mux.RUnlock()
	p := strings.Join(path, "/")
	keys := make([]string, len(m.list[p]))
	var i = 0
	for key := range m.list[p] {
		keys[i] = key
		i++
	}
	return keys
}

func (m *Memory) Perfix() string {
	panic("not implemented")
}
