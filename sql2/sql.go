package sql2

import (
	"reflect"

	"database/sql"
	"errors"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/jmoiron/sqlx"
)

type Backend interface {
	Create(interface{}) (int64, error)
	Get(int64, interface{}) error
	Delete(int64) error
	List(interface{}) error
	Update(int64, interface{}) error
}

type DB struct {
	*sql.DB
}

func NewDB(sqlType string, conn string) (*DB, error) {
	db, err := sql.Open(sqlType, conn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

type SqlBackend struct {
	DB    *DB
	table string
	pk    string
}

//func NewSqlBackend(db *sqlx.DB) error{
func NewSqlBackend(db *DB) (*SqlBackend, error) {
	sql := &SqlBackend{
		DB: db,
	}
	return sql, nil
}

func (s *SqlBackend) SetTable(name string) {
	if s.table == "" {
		s.table = name
	}
}

func (s *SqlBackend) SetPKField(name string) {
	if s.pk == "" {
		s.pk = name
	}
}

func (s *SqlBackend) IsReady() error {
	//TODO:检查连接是否正常，表名，主键值是否设置
	if s.table == "" {
		return errors.New("SQL ERROR: Table not set")
	}
	if s.pk == "" {
		return errors.New("SQL ERROR: PKField not set")
	}
	return nil
}

func (s *SqlBackend) Create(model interface{}) (int64, error) {
	if err := s.IsReady(); err != nil {
		return 0, err
	}
	//TODO:类型判断函数
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := TypeFields(t)
	ff = StringFilter(ff, func(f string) bool {
		return strings.ToLower(f) != s.pk
	})
	vv := ValuesByFields(v, ff)
	query := InsertSQL(s.table, StringMap(ff, strings.ToLower))
	res, err := s.DB.Exec(query, vv...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SqlBackend) Update(id int64, model interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := TypeFields(t)
	ff = StringFilter(ff, func(f string) bool {
		return strings.ToLower(f) != s.pk
	})
	vv := ValuesByFields(v, ff)
	query := UpdateSQL(s.table, StringMap(ff, strings.ToLower)) + " where " + s.pk + "=" + strconv.FormatInt(id, 10)
	_, err := s.DB.Exec(query, vv...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqlBackend) Get(key int64, dest interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	t := reflect.TypeOf(dest).Elem()
	v := reflect.Indirect(reflect.ValueOf(dest))
	ff := TypeFields(t)
	pp := PointersByFields(v, ff)
	query := SelectSQL(s.table, StringMap(ff, strings.ToLower)) + " where " + s.pk + " = ?"
	err := s.DB.QueryRow(query, key).Scan(pp...)
	return err
}

func (s *SqlBackend) Delete(key int64) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	_, err := s.DB.Exec("DELETE FROM "+s.table+" where id=?", key)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqlBackend) List(list interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	dest := reflect.ValueOf(list).Elem()
	t := reflect.TypeOf(list).Elem().Elem().Elem()
	ff := TypeFields(t)
	query := SelectSQL(s.table, StringMap(ff, strings.ToLower))
	rows, err := s.DB.Query(query)
	defer rows.Close()
	for rows.Next() {
		vptr := reflect.New(t)
		pp := PointersByFields(reflect.Indirect(vptr), ff)
		err := rows.Scan(pp...)
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

func InsertSQL(table string, ff []string) string {
	stubs := StringMap(ff, func(f string) string {
		return "?"
	})
	fs := "(" + strings.Join(ff, ",") + ")"
	vs := "(" + strings.Join(stubs, ",") + ")"
	return "INSERT INTO " + table + " " + fs + " VALUES " + vs
}

func UpdateSQL(table string, ff []string) string {
	stubs := StringMap(ff, func(f string) string {
		return f + "=?"
	})
	return "UPDATE " + table + " SET " + strings.Join(stubs, ",")
}

func SelectSQL(table string, ff []string) string {
	fs := strings.Join(ff, ",")
	return "select " + fs + " from " + table
}

func TypeFields(t reflect.Type) []string {
	ff := []string{}
	for i := 0; i < t.NumField(); i++ {
		ff = append(ff, t.Field(i).Name)
	}
	return ff
}

func StringFilter(ss []string, match func(string) bool) []string {
	nss := []string{}
	for _, s := range ss {
		if match(s) {
			nss = append(nss, s)
		}
	}
	return nss
}

func StringMap(ss []string, mapper func(string) string) []string {
	nss := []string{}
	for _, s := range ss {
		nss = append(nss, mapper(s))
	}
	return nss
}

func ValuesByFields(v reflect.Value, ff []string) []interface{} {
	vv := []interface{}{}
	for _, f := range ff {
		vv = append(vv, v.FieldByName(f).Interface())
	}
	return vv
}

func PointersByFields(v reflect.Value, ff []string) []interface{} {
	vv := []interface{}{}
	for _, f := range ff {
		vv = append(vv, v.FieldByName(f).Addr().Interface())
	}
	return vv
}
