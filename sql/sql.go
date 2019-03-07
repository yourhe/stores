package sql

import (
	"fmt"
	stdlog "log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/yourhe/stores"
	// _ "github.com/go-sql-driver/mysql"
	// sq "github.com/Masterminds/squirrel"
	// _ "github.com/mattn/go-sqlite3"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var _ stores.Store = (*Store)(nil)

const Name = "SQL"

//Store ...
type Store struct {
	migrate map[string]bool
	list    map[string]map[string]interface{}
	db      *gorm.DB
	keys    []string
	mux     sync.RWMutex
	log     *stdlog.Logger
}

func NewStore() *Store {
	db, err := gorm.Open("sqlite3", "locked.sqlite")
	if err != nil {
		panic("failed to connect database")
	}
	// db, err := sql.Open("sqlite3",
	// 	"file:locked.sqlite?cache=shared")
	// if err != nil {
	// 	stdlog.Fatal(err)
	// }
	return &Store{
		db: db,
	}
}

func (s *Store) Write(key string, in interface{}) error {
	// val := stores.IPtrToValue(in)
	// fmt.Println(in, in == nil, reflect.ValueOf(in).IsNil())

	if in == nil || reflect.ValueOf(in).IsNil() {

		ptr := reflect.New(reflect.TypeOf(in).Elem())
		// fmt.Println(ptr.Elem())
		s.db.Where("id = ?", key).Delete(ptr.Elem().Interface())
		return nil
	}
	s.db.AutoMigrate(in)
	db := s.db.Create(in)
	db.Save(in)
	// st, _ := s.db.Prepare("")
	// st.Exec
	// val := stores.IPtrToValue(in)
	// table := val.Type().Name() + "s"
	// fields, types, its := GetIterFieldValues(val)
	// s.db.Exec(DataCreateOf(table, fields, types))
	// // fmt.Println(sq.Insert(table).Columns(names...).Values(its...).ToSql())
	// fmt.Println(sq.Insert(table).Columns(fields...).Values(its...).RunWith(s.db).Exec())
	return db.Error
}

func (s *Store) Read(id string, out interface{}) error {
	db := s.db.Where("id = ?", id).First(out)
	return db.Error
}

func (s *Store) ReadAll(out interface{}) error {
	return s.db.Find(out).Error
	// panic("not implemented")
}

func (s *Store) Keys(...string) []string {
	panic("not implemented")
}
func (s *Store) Perfix() string {
	panic("not implemented")
}

func GenQuery(in interface{}) {

}

func GetIterFieldValues(val reflect.Value) ([]string, []string, []interface{}) {
	names := GetIterFieldNames(val)
	if len(names) < 1 {
		return nil, nil, nil
	}
	its := make([]interface{}, len(names), len(names))
	sqltyps := make([]string, len(names), len(names))
	for i, name := range names {
		v := val.Field(i)
		sf := val.Type().Field(i)

		fmt.Println(i, name, sf.Name, DataTypeOf(v, sf))
		sqltyps[i] = DataTypeOf(v, sf)
		if v.CanInterface() {
			its[i] = v.Interface()
		}

	}

	return names, sqltyps, its
}

func DataCreateOf(name string, fields, typs []string) string {
	tmpsql := fmt.Sprintf("Create TABLE %s ( ", name)
	for i, field := range fields {
		var next = ""
		if i < len(fields)-1 {
			next = ","
		}
		tmpsql = fmt.Sprintf("%s %s %s%s", tmpsql, field, typs[i], next)
	}
	tmpsql = fmt.Sprintf("%s )", tmpsql)
	return tmpsql
}

func fieldCanAutoIncrement(sf reflect.StructField) bool {
	return strings.ToLower(sf.Name) == "id"
}

func DataTypeOf(dataValue reflect.Value, sf reflect.StructField) (sqlType string) {
	pk := fieldCanAutoIncrement(sf)
	switch dataValue.Kind() {
	case reflect.Bool:
		sqlType = "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		if pk {
			sqlType = "integer primary key autoincrement"
		} else {
			sqlType = "integer"
		}
	case reflect.Int64, reflect.Uint64:
		if pk {
			sqlType = "integer primary key autoincrement"
		} else {
			sqlType = "bigint"
		}
	case reflect.Float32, reflect.Float64:
		sqlType = "real"
	case reflect.String:
		sqlType = "text"
		if pk {
			sqlType = "text primary key"
		}

	case reflect.Struct:
		if _, ok := dataValue.Interface().(time.Time); ok {
			sqlType = "datetime"
		}
	default:
		if IsByteArrayOrSlice(dataValue) {
			sqlType = "blob"
		}
	}
	return
}

func IsByteArrayOrSlice(value reflect.Value) bool {
	return (value.Kind() == reflect.Array || value.Kind() == reflect.Slice) && value.Type().Elem() == reflect.TypeOf(uint8(0))
}
func GetIterFieldNames(val reflect.Value) []string {
	typ := val.Type()
	count := typ.NumField()
	if count < 1 {
		return nil
	}
	names := make([]string, count, count)
	for i := 0; i < count; i++ {
		names[i] = strings.ToLower(typ.Field(i).Name)
	}
	return names
}
