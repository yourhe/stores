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
	fields := []string{}
	vstubs := []string{}
	values := []interface{}{}
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		if name != s.pk {
			fields = append(fields, name)
			vstubs = append(vstubs, "?")
			values = append(values, v.Field(i).Interface())
		}
	}
	fs := "(" + strings.Join(fields, ",") + ")"
	vs := "(" + strings.Join(vstubs, ",") + ")"
	query := "INSERT INTO " + s.table + " " + fs + " VALUES " + vs
	res, err := s.DB.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (s *SqlBackend) Update(id int64, model interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	stubs := []string{}
	values := []interface{}{}
	t := reflect.TypeOf(model).Elem()
	v := reflect.Indirect(reflect.ValueOf(model))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		if name != s.pk {
			stubs = append(stubs, name+"=?")
			values = append(values, v.Field(i).Interface())
		}
	}
	vs := strings.Join(stubs, ",")
	query := "UPDATE " + s.table + " SET " + vs + " where " + s.pk + "=" + strconv.FormatInt(id, 10)
	_, err := s.DB.Exec(query, values...)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqlBackend) Get(key int64, dest interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	fields := []string{}
	values := []interface{}{}
	t := reflect.TypeOf(dest).Elem()
	v := reflect.Indirect(reflect.ValueOf(dest))
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		fields = append(fields, name)
		values = append(values, v.Field(i).Addr().Interface()) //获取结构体的字段指针
	}
	fs := strings.Join(fields, ",")
	query := "select " + fs + " from " + s.table + " where " + s.pk + " = ?"
	err := s.DB.QueryRow(query, key).Scan(values...)
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

func getStructFields(t reflect.Type) []string {
	ss := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := strings.ToLower(f.Name)
		ss = append(ss, name)
	}
	return ss
}

func newStructValuePtrs(t reflect.Type) (reflect.Value, []interface{}) {
	vptr := reflect.New(t)
	v := reflect.Indirect(vptr)
	vv := []interface{}{}
	for i := 0; i < t.NumField(); i++ {
		vv = append(vv, v.Field(i).Addr().Interface())
	}
	return vptr, vv
}

func (s *SqlBackend) List(list interface{}) error {
	if err := s.IsReady(); err != nil {
		return err
	}
	dest := reflect.ValueOf(list).Elem()
	t := reflect.TypeOf(list).Elem().Elem().Elem()
	fields := getStructFields(t)
	fs := strings.Join(fields, ",")
	query := "select " + fs + " from " + s.table
	rows, err := s.DB.Query(query)
	defer rows.Close()
	for rows.Next() {
		elem, vs := newStructValuePtrs(t)
		err := rows.Scan(vs...)
		dest.Set(reflect.Append(dest, elem))
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}
