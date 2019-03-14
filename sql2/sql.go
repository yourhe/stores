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

// var NotStructPointerErr = errors.New("输入对象必须是结构体指针")
// var NotStructPointerSliceErr = errors.New("输入对象必须是结构体指针数组指针")

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

func (s *SqlBackend) IsPK(key string) bool {
	return strings.ToLower(key) == strings.ToLower(s.pk)
}

// Insert 插入对象，model参数必须是结构体指针，如 &Sth{}
func (s *SqlBackend) Insert(model interface{}) (int64, error) {
	if err := s.IsReady(); err != nil {
		return 0, err
	}
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := ExtractFieldsNames(t)
	ff = FilterStrings(ff, func(f string) bool {
		return !s.IsPK(f)
	})
	args := ValuesByFields(v, ff)
	query := BuildInsertSQL(s.table, MapStrings(ff, strings.ToLower))
	res, err := s.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update 修改对象，model参数必须是结构体指针，如 &Sth{}
func (s *SqlBackend) Update(id int64, model interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := ExtractFieldsNames(t)
	ff = FilterStrings(ff, func(f string) bool {
		return !s.IsPK(f)
	})
	args := ValuesByFields(v, ff)
	query := BuildUpdateSQL(s.table, MapStrings(ff, strings.ToLower)) + " where " + s.pk + "=" + strconv.FormatInt(id, 10)
	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}

// Get 获取对象，model参数是结构体指针，如 &Sth{}
func (s *SqlBackend) Get(key int64, dest interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	t := reflect.TypeOf(dest).Elem()
	v := reflect.Indirect(reflect.ValueOf(dest))
	ff := ExtractFieldsNames(t)
	outs := PointersByFields(v, ff)
	query := BuildSelectSQL(s.table, MapStrings(ff, strings.ToLower)) + " where " + s.pk + " = ?"
	err := s.DB.QueryRow(query, key).Scan(outs...)
	return err
}

// Delete 删除对象
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

// List 获取对象集合，list参数必须是结构体指针切片指针，如 *[]*Sth{}
func (s *SqlBackend) List(list interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	dest := reflect.ValueOf(list).Elem()
	t := reflect.TypeOf(list).Elem().Elem().Elem()
	ff := ExtractFieldsNames(t)
	query := BuildSelectSQL(s.table, MapStrings(ff, strings.ToLower))
	rows, err := s.DB.Query(query)
	defer rows.Close()
	for rows.Next() {
		vptr := reflect.New(t)
		outs := PointersByFields(reflect.Indirect(vptr), ff)
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

func BuildInsertSQL(table string, fields []string) string {
	stubs := MapStrings(fields, func(string) string {
		//生成占位符
		return "?"
	})
	fs := "(" + strings.Join(fields, ",") + ")"
	vs := "(" + strings.Join(stubs, ",") + ")"
	return "INSERT INTO " + table + " " + fs + " VALUES " + vs
}

func BuildUpdateSQL(table string, fields []string) string {
	stubs := MapStrings(fields, func(f string) string {
		return f + "=?"
	})
	return "UPDATE " + table + " SET " + strings.Join(stubs, ",")
}

func BuildSelectSQL(table string, fields []string) string {
	fs := strings.Join(fields, ",")
	return "select " + fs + " from " + table
}

func ExtractFieldsNames(t reflect.Type) []string {
	ff := []string{}
	for i := 0; i < t.NumField(); i++ {
		ff = append(ff, t.Field(i).Name)
	}
	return ff
}

func FilterStrings(ss []string, match func(string) bool) []string {
	nss := []string{}
	for _, s := range ss {
		if match(s) {
			nss = append(nss, s)
		}
	}
	return nss
}

func MapStrings(ss []string, mapper func(string) string) []string {
	nss := []string{}
	for _, s := range ss {
		nss = append(nss, mapper(s))
	}
	return nss
}

func ValuesByFields(v reflect.Value, fields []string) []interface{} {
	vv := []interface{}{}
	for _, f := range fields {
		vv = append(vv, v.FieldByName(f).Interface())
	}
	return vv
}

func PointersByFields(v reflect.Value, fields []string) []interface{} {
	vv := []interface{}{}
	for _, f := range fields {
		vv = append(vv, v.FieldByName(f).Addr().Interface())
	}
	return vv
}
