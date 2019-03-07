package sql

import (
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test(t *testing.T) {
	type Token struct {
		// gorm.Model
		ID           string
		TttID        string
		Clientid     string `protobuf:"bytes,1,opt,name=clientid" json:"clientid,omitempty"`
		Clientsecret string `protobuf:"bytes,2,opt,name=clientsecret" json:"clientsecret,omitempty"`
	}
	type Ttt struct {
		// gorm.Model
		ID    string
		Name  string
		Token Token
		// TokenID string
	}
	var u = Ttt{
		// Model: gorm.Model{ID: 2},
		ID:   "2",
		Name: "save to sql",
		Token: Token{
			ID:           "1",
			Clientid:     "cid",
			Clientsecret: "secret",
		},
	}

	db := NewStore()
	db.db.AutoMigrate(&Token{})
	db.Write("2", &u)
	result := Ttt{}
	// db.Read("1", &result)
	// // fmt.Println(result)
	// // var del *Ttt
	// db.Write("2", del)
	// result = Ttt{}
	db.Read("2", &result)
	// fmt.Println(result)
	// var results = []*Ttt{}
	// fmt.Println(db.ReadAll(&results))
	tt := Token{}
	db.db.Model(result).Related(&tt)
	fmt.Println(tt)
	result.Token = tt
	fmt.Println(result)
}
