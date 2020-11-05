package util

import (
	"reflect"
	"testing"
)

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    []int64
		wantErr bool
	}{
		{
			name:    "case 1 ",
			args:    "1/2/3/4/5",
			want:    []int64{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "case 2 ",
			args:    "",
			want:    []int64{},
			wantErr: true,
		},
		{
			name:    "case 3 ",
			args:    "abc",
			want:    []int64{},
			wantErr: true,
		},
		{
			name:    "case 4 ",
			args:    "1/2/3/4/5/2/5/4",
			want:    []int64{1, 2, 3, 4, 5, 2, 5, 4},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCategory(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCategory() got = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestGetCategory1(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    []int64
		wantErr bool
	}{
		{
			name:    "case 1 ",
			args:    "1/2/3/4/5",
			want:    []int64{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "case 2 ",
			args:    "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "case 3 ",
			args:    "abc",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "case 4 ",
			args:    "1/2/3/4/5/2/5/4",
			want:    []int64{1, 2, 3, 4, 5, 2, 5, 4},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCategory(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCategory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
