package aliyunoss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/yourhe/stores"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	Endpoint        = "oss-cn-xxxx.aliyuncs.com"
	AccessKeyID     = "xxx"
	AccessKeySecret = "xxx"
	BucketName      = "xx"
	Prefix          = "xx/"
)

var _ stores.Store = (*Store)(nil)
var SplitKeys = stores.SplitKeys

type Option func(*Store)

func Suffix(suffix string) Option {
	return func(s *Store) {
		s.suffix = suffix
	}
}

func OssPrefix(prefix string) Option {
	return func(s *Store) {
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		s.prefix = prefix
	}
}

type Store struct {
	cfg    oss.Config
	client *oss.Client
	db     *oss.Bucket
	suffix string
	prefix string
}

func NewStore(opts ...Option) *Store {
	cfg := oss.Config{
		Endpoint:        Endpoint,
		AccessKeyID:     AccessKeyID,
		AccessKeySecret: AccessKeySecret,
	}
	client, _ := oss.New(
		cfg.Endpoint,
		cfg.AccessKeyID,
		cfg.AccessKeySecret,
	)
	db, _ := client.Bucket(BucketName)
	s := &Store{
		cfg:    cfg,
		client: client,
		db:     db,
	}
	s.suffix = ".json"
	s.prefix = "users/"
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Store) Write(key string, in interface{}) error {
	key = s.MakeKey(key)
	// remove
	if in == nil {
		return s.db.DeleteObject(key)
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	return s.db.PutObject(key, r)
	// panic("not implemented")
}

func (s *Store) Read(key string, to interface{}) error {
	v := stores.IPtrToValue(to)
	if v.Kind() == reflect.Slice {
		return s.readAll(key, to)
	}
	return s.read(key, to)
}
func (s *Store) read(key string, to interface{}) error {
	key = s.MakeKey(key)

	w, err := s.db.GetObject(key)
	if err != nil {
		return err
	}
	defer w.Close()
	return json.NewDecoder(w).Decode(to)
}

func (s *Store) readAll(key string, to interface{}) error {
	keys := s.Keys(key)
	v := reflect.ValueOf(to)
	nv, err := stores.MakeSlice(v, len(keys))
	if err != nil {
		return err
	}

	v = *nv
	for i, subkey := range keys {
		dst := v.Index(i)
		dst.Set(reflect.New(dst.Type().Elem()))
		if err := s.Read(key+subkey, dst.Interface()); err != nil {
			return err
		}
	}
	return err

}
func (s *Store) ReadAll(to interface{}) error {
	// s.db.ListObjects
	return s.readAll("", to)
}

func (s *Store) Keys(dirs ...string) []string {
	key := ""
	if len(dirs) > 0 {
		key = dirs[0]
	}
	key = s.MakeKey(key)
	path, _ := SplitKeys(key)
	prefix := oss.Prefix(path + "/")
	delimiter := oss.Delimiter("/")
	results, err := s.db.ListObjects(prefix, delimiter)
	if err != nil {
		return nil
	}
	var keys []string //= make([]string, len(results.Objects), len(results.Objects))
	var i = 0
	for _, obj := range results.Objects {
		if stores.IsPath(obj.Key) {
			continue
		}
		key := strings.Replace(obj.Key, s.suffix, "", -1)
		if strings.HasPrefix(key, s.prefix) {
			key = key[len(s.prefix):]
		}
		_, key = SplitKeys(key)

		keys = append(keys, key)
		i++
	}
	return keys
}

func (s *Store) MakeKey(key string) string {
	path, key := SplitKeys(key)
	path = s.prefix + path
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}
	if key != "" {
		key = fmt.Sprintf("%s%s", key, s.suffix)
	}
	return fmt.Sprintf("%s%s", path, key)
}

func (s *Store) Perfix() string {
	return s.prefix
}
