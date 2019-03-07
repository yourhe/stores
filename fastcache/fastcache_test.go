package fastcache

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/yourhe/stores/aliyunoss"
)

func TestNewFastCacheStore(t *testing.T) {
	defer os.RemoveAll("./users")
	// filestore := file.NewFileStore(file.Dir("./users"))
	filestore := aliyunoss.NewStore()
	var tp string = "a"
	var err error
	got := NewFastCacheStore(filestore, tp)
	var pk = "abc"
	got.Write("abgc", &pk)
	got.Write("abc", &pk)
	var out = ""
	got.Read("abc", &out)

	if !reflect.DeepEqual(out, pk) {
		t.Error(pk, ":", out)
	}
	var outarry = []*string{}
	got.ReadAll(&outarry)
	if len(outarry) != 2 {
		t.Error("ReadAll error", outarry)
	}
	fmt.Println(outarry)
	//test del key
	err = got.Write("abc", nil)
	if err != nil {
		t.Error(err)
	}
	var outarry1 = []*string{}

	err = got.ReadAll(&outarry1)
	if err != nil {
		t.Error(err)
	}
	if len(outarry1) != 1 {
		t.Error("ReadAll error", outarry1)
	}
	err = got.Write("abgc", nil)
	if err != nil {
		t.Error(err)
	}
	var outarry2 = []*string{}
	err = got.ReadAll(&outarry2)
	if err != nil {
		t.Error(err)
	}
	if len(outarry2) != 0 {
		t.Error("ReadAll error", outarry2)
	}
	err = got.Read("abgc", &out)
	if err == nil {
		t.Error(err)
	}
	// fmt.Println(string(outarry[0]))
	// g := outarry[0]
}
