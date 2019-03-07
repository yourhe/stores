package fastcache

import (
	"reflect"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/yourhe/stores"
	"github.com/yourhe/stores/memory"
)

var _ stores.Store = (*FastCache)(nil) // Verify that FastCache implements stores.Store.
const Name = "FastCache"

//FastCache ...
type FastCache struct {
	mem    stores.Store
	file   stores.Store
	isSync map[string]bool
	mux    sync.RWMutex
}

func NewFastCacheStore(store stores.Store, tp interface{}) *FastCache {
	fc := &FastCache{
		mem:    memory.NewMemoryStore(),
		file:   store,
		isSync: make(map[string]bool),
		mux:    sync.RWMutex{},
	}
	// keys := fc.file.Keys()
	// itp := reflect.ValueOf(tp).Type()
	// for _, key := range keys {
	// 	it := reflect.New(itp)

	// 	// reflect.MakeSlice(tp,len(keys),len(keys))
	// 	// var tmp = new(interface{})
	// 	fc.file.Read(key, it.Interface())
	// 	fc.mem.Write(key, it.Interface())
	// }
	// fmt.Println(keys, fc.mem.Keys())
	return fc
}

func (fc *FastCache) Write(key string, in interface{}) (err error) {
	err = fc.file.Write(key, in)
	if err != nil {
		return
	}
	err = fc.mem.Write(key, in)
	if err != nil {
		return
	}
	return

}

func (fc *FastCache) Read(key string, out interface{}) (err error) {
	dst := reflect.ValueOf(out)
	if dst.Kind() != reflect.Ptr {
		return stores.NotPrtError
	}
	fc.mux.RLock()
	ok := fc.isSync[key]
	fc.mux.RUnlock()
	if !ok && IsPath(key) && reflect.Indirect(dst).Kind() == reflect.Slice {
		goto SYNC
	}
	err = fc.mem.Read(key, out)
	if err == nil || err != stores.NotFoundError {
		// fmt.Println("out mem read")
		return
	}
SYNC:
	log.Debug(key, "out mem read", IsPath(key), reflect.Indirect(dst).Kind())
	if IsPath(key) && reflect.Indirect(dst).Kind() == reflect.Slice {

		keys := fc.file.Keys(key)
		t := reflect.ValueOf(out).Elem()
		t.Set(reflect.MakeSlice(t.Type(), len(keys), len(keys)))
		for i, subkey := range keys {
			di := t.Index(i)
			pk := reflect.New(di.Type().Elem())
			di.Set(pk)
			err = fc.Read(key+subkey, di.Interface())
		}
		fc.mux.Lock()
		fc.isSync[key] = true
		fc.mux.Unlock()

	} else {
		err = fc.file.Read(key, out)
		if err == nil && out != nil {
			err = fc.mem.Write(key, out)

		}
	}

	return
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

func IsPath(key string) bool {
	return strings.HasSuffix(key, "/") || key == ""
}

func (fc *FastCache) ReadAll(out interface{}) (err error) {
	return fc.readAll(out)
}

func (fc *FastCache) readAll(out interface{}) (err error) {
	return fc.Read("", out)
	err = fc.mem.ReadAll(out)
	if err == nil && err != stores.NotFoundError {
		// fmt.Println("out mem ReadAll")
		return
	}
	keys := fc.file.Keys()
	t := reflect.ValueOf(out)
	nt, err := stores.MakeSlice(t, len(keys))
	if err != nil {
		return err
	}
	t = *nt
	for i, key := range keys {
		di := t.Index(i)
		pk := reflect.New(di.Type().Elem())
		di.Set(pk)
		fc.Read(key, di.Interface())
	}

	return
}

func (fc *FastCache) Keys(path ...string) []string {
	return fc.mem.Keys()
}

func (fc *FastCache) Perfix() string {
	panic("not implemented")
}
