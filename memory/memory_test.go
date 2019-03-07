package memory

import (
	"sync"
	"testing"
)

func TestSplitKeys(t *testing.T) {
	path, key := SplitKeys("abc/sss")
	if path != "abc" {
		t.Error(path)
	}
	if key != "sss" {
		t.Error(key)
	}

	path, key = SplitKeys("abc/asdfa/ggg")
	if path != "abc/asdfa" {
		t.Error(path)
	}
	if key != "ggg" {
		t.Error(key)
	}

	path, key = SplitKeys("ggg")
	if path != "" {
		t.Error(path)
	}
	if key != "ggg" {
		t.Error(key)
	}
}

func TestMemory_Write(t *testing.T) {
	type fields struct {
		list map[string]map[string]interface{}
		keys []string
		mux  sync.RWMutex
	}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    int
	}{
		{
			args:    args{key: "abc", value: "abc"},
			wantErr: false,
			want:    1,
		},
		{
			args:    args{key: "abc", value: nil},
			wantErr: false,
			want:    0,
		},
		// TODO: Add test cases.
	}
	m := NewMemoryStore()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Write(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Memory.Write() error = %v, wantErr %v\n%v", err, tt.wantErr, m.list)
			}
			var out []*string
			err := m.Read("", &out)
			if len(out) != tt.want {
				t.Errorf("Memory.Write() error = %v, wantErr %v\n%v", err, len(out), out)
			}

		})
	}

}
