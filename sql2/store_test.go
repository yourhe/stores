package sql2

import (
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
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
	_, err = db.Exec(CREATE_SQL)
	check(err)
	defer db.Exec(DROP_SQL)
	sql, _ := NewSqlBackend(db)
	sql.SetPKField("id")
	sql.SetTable("report_demo")

	r := &ReportDemo{
		Name:    "n1",
		Content: "n1",
	}
	err = sql.Write("", r)
	check(err)
	rr := []*ReportDemo{}
	_ = sql.ReadAll(&rr)
	// check(err)
	if len(rr) <= 0 {
		t.Fatal("list fail")
	}
	nr := &ReportDemo{}
	_ = sql.Read("1", nr)
	if nr.ID != 1 || nr.Name != "n1" || nr.Content != "n1" {
		t.Fatal("get fail")
	}
	nr.Content = "new"
	nr.Name = "new"
	_ = sql.Write(strconv.FormatInt(nr.ID, 10), nr)
	nnr := &ReportDemo{}
	_ = sql.Read("1", nnr)
	if nnr.ID != 1 || nnr.Name != "new" || nnr.Content != "new" {
		t.Fatal("update fail")
	}
	err = sql.Write("1", nil)
	check(err)
	nrr := []*ReportDemo{}
	_ = sql.ReadAll(&nrr)
	if len(nrr) != 0 {
		t.Fatal("delete fail")
	}
}
