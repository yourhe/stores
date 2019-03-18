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
	sql.SetPKField("id")
	sql.SetTable("report_demo")

	var sqlEx stores.SqlExecutor = sql
	_, err = sqlEx.Exec(CREATE_SQL)
	check(err)
	defer sqlEx.Exec(DROP_SQL)
	var sqlStore stores.Store = sql
	r := &ReportDemo{
		Name:    "n1",
		Content: "n1",
	}
	err = sqlStore.Write("", r)
	check(err)
	if r.ID != 1 {
		t.Fatalf("write new fail")
	}
	rr := []*ReportDemo{}
	_ = sqlStore.ReadAll(&rr)
	// check(err)
	if len(rr) <= 0 {
		t.Fatal("list fail")
	}
	nr := &ReportDemo{}
	_ = sqlStore.Read("1", nr)
	if nr.ID != 1 || nr.Name != "n1" || nr.Content != "n1" {
		t.Fatal("get fail")
	}
	nr.Content = "new"
	nr.Name = "new"
	_ = sqlStore.Write(strconv.FormatInt(nr.ID, 10), nr)
	nnr := &ReportDemo{}
	_ = sqlStore.Read("1", nnr)
	if nnr.ID != 1 || nnr.Name != "new" || nnr.Content != "new" {
		t.Fatal("update fail")
	}
	var sqlQueryer stores.Queryer = sql
	li := []*ReportDemo{}
	_ = sqlQueryer.Find("select id,name from report_demo where name='new'", &li)
	if len(li) != 1 {
		t.Fatal("find fail")
	}
	err = sqlStore.Write("1", nil)
	check(err)
	nrr := []*ReportDemo{}
	_ = sqlStore.ReadAll(&nrr)
	if len(nrr) != 0 {
		t.Fatal("delete fail")
	}
}
