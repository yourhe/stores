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
		id, err := s.Insert(model)
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

func (s *SqlBackend) Find(q interface{}, to interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	query := reflect.ValueOf(q).Interface().(string)
	dest := reflect.ValueOf(to).Elem()                   //切片值(非指针)
	modelType := reflect.TypeOf(to).Elem().Elem().Elem() //切片成员类型(非指针)
	fnames := ExtractFieldsNames(modelType)
	rows, err := s.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	fnames = FilterStrings(fnames, func(f string) bool {
		_f := strings.ToLower(f)
		for _, c := range cols {
			if c == _f {
				return true
			}
		}
		return false
	})
	for rows.Next() {
		vptr := reflect.New(modelType)
		outs := PointersByFields(reflect.Indirect(vptr), fnames)
		err := rows.Scan(outs...)
		if err != nil {
			return err
		}
		dest.Set(reflect.Append(dest, vptr))
	}
	if err != nil {
		return err
	}
	return nil
}
