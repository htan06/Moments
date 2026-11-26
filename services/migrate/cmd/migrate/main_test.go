package main

import "testing"

func TestModuleFlags(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		values  []string
		wantErr bool
	}{
		{name: "keeps order", values: []string{"user=" + dir, "account=" + dir}},
		{name: "missing dir", values: []string{"user"}, wantErr: true},
		{name: "empty dir", values: []string{"user="}, wantErr: true},
		{name: "dir does not exist", values: []string{"user=" + dir + "/nope"}, wantErr: true},
		{name: "bad name", values: []string{"User-1=" + dir}, wantErr: true},
		{name: "duplicate name", values: []string{"user=" + dir, "user=" + dir}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f moduleFlags
			var err error
			for _, v := range tt.values {
				if err = f.Set(v); err != nil {
					break
				}
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", f)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(f) != 2 || f[0].Name != "user" || f[1].Name != "account" {
				t.Fatalf("got %v, want user then account", f)
			}
		})
	}
}
