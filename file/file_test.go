package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var testDir = "./testdata"

func TestFileStore_getFileName(t *testing.T) {
	f := &FileStore{
		dir:    "/xxx",
		suffix: ".json",
	}
	got := f.getFileName("abc")
	want := "/xxx/abc.json"
	if got != want {
		t.Errorf("FileStore.getFileName() = %v, want %v", got, want)

	}

}

func TestFileStore_Write(t *testing.T) {
	defer os.RemoveAll(testDir)
	fs := NewFileStore(Dir(testDir), Suffix(".json"))
	fs.Write("test", "asdfasdf")

	bs, err := ioutil.ReadFile(filepath.Join(testDir, "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "\"asdfasdf\"" {
		t.Error(string(bs), "asdfasdf")
	}

}

func TestFileStore_Read(t *testing.T) {
	defer os.RemoveAll(testDir)
	type args struct {
		in0 string
		// in1 interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{"test"},
			wantErr: false,
		},
		{
			args:    args{""},
			wantErr: true,
		},
	}
	f := NewFileStore(Dir(testDir), Suffix(".json"))
	f.Write("test", "test")
	for _, tt := range tests {
		o := ""
		t.Run(tt.name, func(t *testing.T) {
			if err := f.Read(tt.args.in0, &o); (err != nil) != tt.wantErr || !reflect.DeepEqual(o, tt.args.in0) {
				t.Errorf("FileStore.Read() in = %v, out %v", tt.args.in0, o)
				t.Error(err)
			}
		})
	}
}

func TestFileStore_ReadAll(t *testing.T) {
	defer os.RemoveAll(testDir)

	tests := []struct {
		name    string
		args    interface{}
		expect  interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    []*string{},
			expect:  []*string{Strings("test")},
			wantErr: false,
		},
	}

	f := NewFileStore(Dir(testDir), Suffix(".json"))
	f.Write("test", "test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := []*string{}
			if err := f.ReadAll(&to); (err != nil) != tt.wantErr {
				t.Errorf("FileStore.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.expect, to) {
				t.Errorf("FileStore.ReadAll() have = %v, want %v", to, tt.expect)
			}
		})
	}
}

func TestFileStore_Delete(t *testing.T) {
	defer os.RemoveAll(testDir)
	fs := NewFileStore(Dir(testDir), Suffix(".json"))
	fs.Write("test", "asdfasdf")

	bs, err := ioutil.ReadFile(filepath.Join(testDir, "test.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "\"asdfasdf\"" {
		t.Error(string(bs), "asdfasdf")
	}
	fs.Write("test", nil)

}

func Strings(s string) *string {
	return &s
}
