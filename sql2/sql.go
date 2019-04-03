package sql2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/jmoiron/sqlx"
)

type Backend interface {
	Insert(interface{}) (int64, error)
	Get(int64, interface{}) error
	Delete(int64, interface{}) error
	List(interface{}) error
	Update(int64, interface{}) error
}

type Builder interface {
	Migrate(interface{}) error
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

type StructFieldFilter func(reflect.StructField) bool
type StructFieldMapper func(reflect.StructField) string

var (
	DefaultFieldFilter = func(reflect.StructField) bool { return true }
	DefaultFieldMapper = func(f reflect.StructField) string { return strings.ToLower(f.Name) }
	ProtoFieldFilter   = func(f reflect.StructField) bool {
		if f.Tag.Get("json") != "-" {
			return true
		}
		return false
	}
)

type SqlBackend struct {
	DB          *DB
	table       string
	pk          string
	fieldFilter StructFieldFilter
	fieldMapper StructFieldMapper
}

func NewSqlBackend(db *DB) (*SqlBackend, error) {
	sql := &SqlBackend{
		DB: db,
	}
	sql.SetFieldMapper(DefaultFieldMapper)
	sql.SetFieldFilter(DefaultFieldFilter)
	return sql, nil
}

func (s *SqlBackend) SetFieldFilter(fn StructFieldFilter) {
	s.fieldFilter = fn
}

func (s *SqlBackend) SetFieldMapper(fn StructFieldMapper) {
	s.fieldMapper = fn
}

func (s *SqlBackend) SetPKField(name string) {
	if s.pk == "" {
		s.pk = name
	}
}

func (s *SqlBackend) SetTable(name string) {
	if s.table == "" {
		s.table = name
	}
}

func (s *SqlBackend) Migrate(model interface{}) error {
	//TODO:
	//若表不存在，则创建表
	//若表存在修改，则修改表(预计只能增加字段)
	return nil
}

func (s *SqlBackend) Insert(model interface{}) (int64, error) {
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := StructFields(t, s.Filter("Insert"))
	fmt.Println(ff)
	vv := FieldsValues(v, ff)
	query := s.Sql("Insert", s.TableName(t.Name()), ff)
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

func (s *SqlBackend) Get(key int64, model interface{}) error {
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := StructFields(t, s.Filter("Get"))
	ptrs := FieldsPointers(v, ff)
	query := s.Sql("Get", s.TableName(t.Name()), ff) + " where " + s.IdCond(key)
	err := s.DB.QueryRow(query).Scan(ptrs...)
	return err
}

func (s *SqlBackend) Update(key int64, model interface{}) error {
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	ff := StructFields(t, s.Filter("Update"))
	vv := FieldsValues(v, ff)
	query := s.Sql("Update", s.TableName(t.Name()), ff) + " where " + s.IdCond(key)
	_, err := s.DB.Exec(query, vv...)
	return err
}

func (s *SqlBackend) IdCond(id int64) string {
	return s.pk + "=" + strconv.FormatInt(id, 10)
}

func (s *SqlBackend) List(list interface{}) error {
	lv := reflect.ValueOf(list).Elem()             //切片值(非指针)
	t := reflect.TypeOf(list).Elem().Elem().Elem() //切片成员类型(非指针)
	ff := StructFields(t, s.Filter("List"))
	query := s.Sql("List", s.TableName(t.Name()), ff)
	rows, err := s.DB.Query(query)
	defer rows.Close()
	for rows.Next() {
		vptr := reflect.New(t)
		ptrs := FieldsPointers(reflect.Indirect(vptr), ff)
		err := rows.Scan(ptrs...)
		if err != nil {
			return err
		}
		lv.Set(reflect.Append(lv, vptr))
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *SqlBackend) Delete(key int64, model interface{}) error {
	t := reflect.TypeOf(model).Elem()
	query := s.Sql("Delete", s.TableName(t.Name()), nil) + " where " + s.IdCond(key)
	_, err := s.DB.Exec(query)
	return err
}

func (s *SqlBackend) TableFieldsNames(fields []reflect.StructField) []string {
	fnames := make([]string, len(fields))
	for i, f := range fields {
		fnames[i] = s.fieldMapper(f)
	}
	return fnames
}

func (s *SqlBackend) TableName(modelName string) string {
	if s.table == "" {
		return modelName
	}
	return s.table
}

func (s *SqlBackend) Sql(t string, table string, fields []reflect.StructField) string {
	switch t {
	case "List", "Get":
		fnames := s.TableFieldsNames(fields)
		return "SELECT " + strings.Join(fnames, ",") + " FROM " + table

	case "Insert":
		fnames := s.TableFieldsNames(fields)
		fs := "(" + strings.Join(fnames, ",") + ")"
		vs := "(" + strings.Join(FieldsStubs(fnames), ",") + ")"
		return "INSERT INTO " + table + " " + fs + " VALUES " + vs

	case "Update":
		fnames := s.TableFieldsNames(fields)
		stubs := MapStrings(fnames, func(f string) string {
			return f + "=?"
		})
		return "UPDATE " + table + " SET " + strings.Join(stubs, ",")

	case "Delete":
		return "DELETE FROM " + table
	}
	return ""
}

func (s *SqlBackend) Filter(t string) StructFieldFilter {
	//t : List Get Insert Update
	switch t {
	case "List", "Get":
		return s.fieldFilter
	case "Insert", "Update":
		return func(f reflect.StructField) bool {
			isPk := false
			if f.Name == s.pk {
				isPk = true
			}
			return s.fieldFilter(f) && !isPk
		}
	}
	return DefaultFieldFilter
}

func StructFields(t reflect.Type, need StructFieldFilter) []reflect.StructField {
	ff := []reflect.StructField{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if need(f) {
			ff = append(ff, f)
		}
	}
	return ff
}

func FieldsValues(v reflect.Value, fnames []reflect.StructField) []interface{} {
	vv := []interface{}{}
	for _, f := range fnames {
		vv = append(vv, v.FieldByName(f.Name).Interface())
	}
	return vv
}

func FieldsPointers(v reflect.Value, fnames []reflect.StructField) []interface{} {
	vv := []interface{}{}
	for _, f := range fnames {
		vv = append(vv, v.FieldByName(f.Name).Addr().Interface())
	}
	return vv
}

func FieldsStubs(fnames []string) []string {
	stubs := make([]string, len(fnames))
	for i := range fnames {
		stubs[i] = "?"
	}
	return stubs
}

func MapStrings(ss []string, mapper func(string) string) []string {
	nss := make([]string, len(ss))
	for i, s := range ss {
		nss[i] = mapper(s)
	}
	return nss
}
