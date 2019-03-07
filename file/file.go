package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/yourhe/stores"

	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
)

type FileStore struct {
	dir    string
	suffix string
	uuid   ksuid.KSUID
	mux    sync.RWMutex
	log    *log.Logger
}

type Option func(*FileStore)

func Dir(dir string) Option {
	return func(fs *FileStore) {
		fs.dir = dir
	}
}

func Suffix(suffix string) Option {
	return func(fs *FileStore) {
		fs.suffix = suffix
	}
}

func NewFileStore(opts ...Option) *FileStore {
	fs := &FileStore{
		dir:    "users",
		suffix: ".json",
		log:    log.StandardLogger(),
	}
	for _, opt := range opts {
		opt(fs)
	}
	if err := os.MkdirAll(fs.dir, 0755); err != nil {
		panic(err)
	}
	return fs
}

func (f *FileStore) getFileName(key string) string {
	suffix := f.suffix
	if strings.HasSuffix(key, suffix) {
		return filepath.Join(f.dir, key)
	}
	return fmt.Sprintf("%s%s", filepath.Join(f.dir, key), suffix)
}
func (f *FileStore) Write(key string, value interface{}) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	fname := f.getFileName(key)
	fname, err := filepath.Abs(fname)
	if err != nil {
		return err
	}
	if value == nil {
		return os.RemoveAll(fname)
	}

	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = os.Stat(fname)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(fname), os.ModePerm); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(fname, bs, os.ModePerm)
}

func (f *FileStore) Read(key string, to interface{}) error {
	return f.read(key, to, false)
}
func (f *FileStore) read(key string, to interface{}, isfile bool) error {
	if reflect.Indirect(reflect.ValueOf(to)).Kind() == reflect.Slice {
		return f.readAll(filepath.Join(f.dir, key), to)
	}

	f.mux.RLock()
	defer f.mux.RUnlock()
	fname := key
	if !isfile {
		fname = f.getFileName(key)
	}
	fname, _ = filepath.Abs(fname)
	fi, err := os.Open(fname)
	if fi != nil {
		defer fi.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(fi).Decode(to)
}

func (f *FileStore) ReadAll(to interface{}) (err error) {
	return f.readAll(f.dir, to)
}
func (f *FileStore) readAll(dir string, to interface{}) (err error) {
	// f.mux.RLock()
	// defer f.mux.RUnlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	ndir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	pattern := filepath.Join(ndir, "*"+f.suffix)
	matchs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	// 反射得到struct后，重新New对象，再赋值
	v := reflect.ValueOf(to)
	// if v.Kind() != reflect.Ptr {
	// 	return fmt.Errorf("错误: have: %s, want: %s", v.Kind(), reflect.Ptr)
	// }
	// dv := reflect.Indirect(v)
	// if dv.Kind() != reflect.Slice {
	// 	return fmt.Errorf("错误: have: %s, want: %s", dv.Kind(), reflect.Slice)
	// }

	nv, err := stores.MakeSlice(v, len(matchs))
	if err != nil {
		return err
	}
	v = *nv
	// dv.Set(reflect.MakeSlice(dv.Type(), len(matchs), len(matchs)))
	var i = 0
	for _, fname := range matchs {
		di := v.Index(i)
		i++
		pk := reflect.New(di.Type().Elem())
		di.Set(pk)
		err = f.read(fname, pk.Interface(), true)
		if err != nil {
			f.log.Warn(err.Error(), fname)
			continue
		}
	}

	return nil
}

func (f *FileStore) Keys(path ...string) []string {
	f.mux.RLock()
	defer f.mux.RUnlock()

	ndir, err := filepath.Abs(f.dir)
	if err != nil {
		return nil
	}
	for _, p := range path {
		ndir = ndir + "/" + p
	}
	pattern := filepath.Join(ndir, "*"+f.suffix)
	matchs, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	for i, _ := range matchs {
		name := filepath.Base(matchs[i])
		ext := filepath.Ext(matchs[i])

		matchs[i] = name[:len(name)-len(ext)]
	}

	return matchs
}

func (f *FileStore) Perfix() string {
	return f.dir
}
