package util

import "testing"

func TestInterfaceToUint(t *testing.T) {
	type args struct {
		n interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantNum uint32
		wantOk  bool
	}{
		// TODO: Add test cases.
		{
			name:    "Float64",
			args:    args{
				n: float64(1.52),
			},
			wantNum: 1,
			wantOk:  true,
		},
		{
			name:    "Float32",
			args:    args{
				n: float32(1.52),
			},
			wantNum: 1,
			wantOk:  true,
		},
		{
			name:    "Int64",
			args:    args{
				n: int64(152),
			},
			wantNum: 152,
			wantOk:  true,
		},
		{
			name:    "Int",
			args:    args{
				n: 152,
			},
			wantNum: 152,
			wantOk:  true,
		},
		{
			name:    "Int32",
			args:    args{
				n: int32(152),
			},
			wantNum: 152,
			wantOk:  true,
		},
		{
			name:    "String",
			args:    args{
				n: "152",
			},
			wantNum: 0,
			wantOk:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNum, gotOk := InterfaceToUint(tt.args.n)
			if gotNum != tt.wantNum {
				t.Errorf("InterfaceToUint() gotNum = %v, want %v", gotNum, tt.wantNum)
			}
			if gotOk != tt.wantOk {
				t.Errorf("InterfaceToUint() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}