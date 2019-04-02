package sql2

import (
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/yourhe/stores"
)

// const TABLE_FOR_TEST = "report_demo"

// const CREATE_SQL = `
// 	CREATE TABLE report_demo (
// 		id INT NOT NULL AUTO_INCREMENT,
// 		name VARCHAR(100) NULL,
// 		content TEXT NULL,
// 		PRIMARY KEY (id));
// `
// const DROP_SQL = `DROP TABLE report_demo;`

func Test_IStore(t *testing.T) {
	check := CheckErrFunc(t)
	err := godotenv.Load("../.env")
	check(err)
	db, _ := NewDB(os.Getenv("SQL_TYPE"), os.Getenv("SQL_CONNECTION"))

	sql, _ := NewSqlBackend(db)

	var sqlEx stores.SqlExecutor = sql
	_, err = sqlEx.Exec(CREATE_DEMO1_SQL)
	check(err)
	defer sqlEx.Exec(DROP_DEMO1_SQL)
	sql.SetPKField("id")
	sql.SetTable("demo1")
	var sqlStore stores.Store = sql
	r := &Demo1{
		Name:    "n1",
		Content: "n1",
	}
	err = sqlStore.Write("", r)
	check(err)
	if r.ID != 1 {
		t.Fatalf("write new fail")
	}
	rr := []*Demo1{}
	err = sqlStore.ReadAll(&rr)
	check(err)
	if len(rr) <= 0 {
		t.Fatal("list fail")
	}
	nr := &Demo1{}
	err = sqlStore.Read("1", nr)
	check(err)
	if nr.ID != 1 || nr.Name != "n1" || nr.Content != "n1" {
		t.Fatal("get fail")
	}
	nr.Content = "new"
	nr.Name = "new"
	_ = sqlStore.Write(strconv.FormatInt(nr.ID, 10), nr)
	nnr := &Demo1{}
	err = sqlStore.Read("1", nnr)
	check(err)
	if nnr.ID != 1 || nnr.Name != "new" || nnr.Content != "new" {
		t.Fatal("update fail")
	}
	var sqlQueryer stores.Queryer = sql
	li := []*Demo1{}
	err = sqlQueryer.Find("select id,name from demo1 where name='new'", &li)
	check(err)
	if len(li) != 1 {
		t.Fatal("find fail")
	}
	var d *Demo1
	err = sqlStore.Write("1", d)
	check(err)
	nrr := []*Demo1{}
	err = sqlStore.ReadAll(&nrr)
	check(err)
	if len(nrr) != 0 {
		t.Fatal("delete fail")
	}
}
