package id

import "testing"

func Test_gen(t *testing.T) {
	tests := []struct {
		name string
		want uint
	}{
		// TODO: Add test cases.
		{name: "1", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gen(); got != tt.want {
				t.Errorf("gen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		args   args
		wantId uint
	}{
		// TODO: Add test cases.
		{
			name:   "1",
			args:   args{"id"},
			wantId: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotId := New(tt.args.key); gotId != tt.wantId {
				t.Errorf("New() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}
