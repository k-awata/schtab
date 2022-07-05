package schtab

import (
	"reflect"
	"testing"
)

func Test_parseField(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		want1   int
		wantErr bool
	}{
		{"*", args{"*"}, []int{-1}, 0, false},
		{"*/5", args{"*/5"}, []int{-1}, 5, false},
		{"6-10", args{"6-10"}, []int{6, 7, 8, 9, 10}, 0, false},
		{"jan", args{"jan"}, []int{1}, 0, false},
		{"mon-fri", args{"mon-fri"}, []int{1, 2, 3, 4, 5}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseField(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseField() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseField() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
