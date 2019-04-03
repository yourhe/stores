package sql2

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

type Demo1 struct {
	ID      int64
	Name    string
	Content string
}

const CREATE_DEMO1_SQL = `
	CREATE TABLE demo1 (
		id INT NOT NULL AUTO_INCREMENT,
		name VARCHAR(100) NULL,
		content TEXT NULL,
		PRIMARY KEY (id));
`
const DROP_DEMO1_SQL = `DROP TABLE demo1;`

type Demo2 struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"create_at"`
	XXX       string    `json:"-"`
}

const CREATE_DEMO2_SQL = `
CREATE TABLE demo2 (
	ID INT NOT NULL AUTO_INCREMENT,
	Name VARCHAR(100) NULL,
	Content TEXT NULL,
	CreatedAt DATETIME NULL,
	PRIMARY KEY (ID));
`
const DROP_DEMO2_SQL = `DROP TABLE Demo2;`

func Test_Curd(t *testing.T) {
	check := CheckErrFunc(t)
	err := godotenv.Load("../.env")
	check(err)
	db, _ := NewDB(os.Getenv("SQL_TYPE"), os.Getenv("SQL_CONNECTION"))
	_, err = db.Exec(CREATE_DEMO1_SQL)
	check(err)
	defer db.Exec(DROP_DEMO1_SQL)

	sql, _ := NewSqlBackend(db)
	sql.SetFieldMapper(LowerFieldMapper)
	sql.SetPKField("id")
	sql.SetTable("demo1")
	r := &Demo1{
		Name:    "n1",
		Content: "n1",
	}
	id, _ := sql.Insert(r)
	if id != 1 {
		t.Fatal("create fail")
	}
	rr := []*Demo1{}
	_ = sql.List(&rr)
	// check(err)
	if len(rr) <= 0 {
		t.Fatal("list fail")
	}
	nr := &Demo1{}
	_ = sql.Get(1, nr)
	if nr.ID != 1 || nr.Name != "n1" || nr.Content != "n1" {
		t.Fatal("get fail")
	}
	nr.Content = "new"
	nr.Name = "new"
	_ = sql.Update(1, nr)
	nnr := &Demo1{}
	_ = sql.Get(1, nnr)
	if nnr.ID != 1 || nnr.Name != "new" || nnr.Content != "new" {
		t.Fatal("update fail")
	}
	var rp *Demo1
	err = sql.Delete(1, rp)
	check(err)
	nrr := []*Demo1{}
	_ = sql.List(&nrr)
	if len(nrr) != 0 {
		t.Fatal("delete fail")
	}
}

func Test_withProtoFilter(t *testing.T) {
	check := CheckErrFunc(t)
	err := godotenv.Load("../.env")
	check(err)
	db, _ := NewDB(os.Getenv("SQL_TYPE"), os.Getenv("SQL_CONNECTION"))
	_, err = db.Exec(CREATE_DEMO2_SQL)
	check(err)
	defer db.Exec(DROP_DEMO2_SQL)

	sql, _ := NewSqlBackend(db)
	sql.SetPKField("ID")
	sql.SetFieldFilter(ProtoFieldFilter)
	r := &Demo2{
		Name:      "n1",
		Content:   "n1",
		CreatedAt: time.Now(),
	}
	id, err := sql.Insert(r)
	check(err)
	if id != 1 {
		t.Fatal("create fail")
	}
	rr := []*Demo2{}
	err = sql.List(&rr)
	check(err)
	if len(rr) <= 0 {
		t.Fatal("list fail")
	}
	nr := &Demo2{}
	err = sql.Get(1, nr)
	check(err)
	if nr.ID != 1 || nr.Name != "n1" || nr.Content != "n1" {
		t.Fatal("get fail")
	}
	nr.Content = "new"
	nr.Name = "new"
	nr.CreatedAt = time.Now()
	err = sql.Update(1, nr)
	check(err)
	nnr := &Demo2{}
	err = sql.Get(1, nnr)
	check(err)
	if nnr.ID != 1 || nnr.Name != "new" || nnr.Content != "new" {
		t.Fatal("update fail")
	}
	var rp *Demo2
	err = sql.Delete(1, rp)
	check(err)
	nrr := []*Demo2{}
	err = sql.List(&nrr)
	check(err)
	if len(nrr) != 0 {
		t.Fatal("delete fail")
	}
}

func CheckErrFunc(t *testing.T) func(error) {
	return func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
}
