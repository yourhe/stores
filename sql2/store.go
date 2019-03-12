package sql2

// "github.com/jmoiron/sqlx"

import (
	"errors"
	"strconv"
)

func (s *SqlBackend) Write(key string, model interface{}) error {
	if model == nil && key != "" {
		//删除
		id, _ := strconv.ParseInt(key, 10, 64)
		return s.Delete(id)
	}
	if key == "" && model != nil {
		_, err := s.Create(model) //QUETIONS:id自增情况下，不应忽略id
		return err
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
