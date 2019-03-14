package sql2

// "github.com/jmoiron/sqlx"

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func (s *SqlBackend) Write(key string, model interface{}) error {
	if model == nil && key != "" {
		//删除
		id, _ := strconv.ParseInt(key, 10, 64)
		return s.Delete(id)
	}
	if key == "" && model != nil {
		id, err := s.Create(model)
		if err != nil {
			return err
		}
		v := reflect.Indirect(reflect.ValueOf(model))
		idv := v.FieldByNameFunc(func(f string) bool {
			return strings.ToLower(f) == s.pk
		})
		idv.SetInt(id)
		return nil
	}
	if key != "" && model != nil {
		//修改
		id, _ := strconv.ParseInt(key, 10, 64)
		return s.Update(id, model)
	}
	return errors.New("arguments wrong")
}

func (s *SqlBackend) Read(key string, model interface{}) error {
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return err
	}
	return s.Get(id, model)
}

func (s *SqlBackend) ReadAll(dest interface{}) error {
	return s.List(dest)
}

func (s *SqlBackend) Keys(strs ...string) []string {
	panic("not implemented")
}

func (s *SqlBackend) Perfix() string {
	panic("not implemented")
}

func (s *SqlBackend) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.DB.Exec(query, args...)
}

func (s *SqlBackend) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.DB.Query(query, args...)
}
