package yIP

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		ip   int
		mask int
	}
	tests := []struct {
		name    string
		args    args
		wantIps []int
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "1", args: args{3232235777, 31}, wantIps: []int{3232235776, 3232235777}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIps, err := Parse(tt.args.ip, tt.args.mask)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIps, tt.wantIps) {
				t.Errorf("Parse() gotIps = %v, want %v", gotIps, tt.wantIps)
			}
		})
	}
}

func TestIp2int(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		args    args
		wantIip int
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "1", args: args{"192.168.1.1"}, wantIip: 3232235777, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIip, err := Ip2int(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ip2int() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIip != tt.wantIip {
				t.Errorf("Ip2int() gotIip = %v, want %v", gotIip, tt.wantIip)
			}
		})
	}
}

func TestIsIp(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{name: "1", args: args{
			"192.168.1.1",
		}, want: true},
		{name: "2", args: args{
			"192.168.1.1/24",
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIp(tt.args.ip); got != tt.want {
				t.Errorf("IsIp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt2Ip(t *testing.T) {
	type args struct {
		iip int
	}
	tests := []struct {
		name   string
		args   args
		wantIp string
	}{
		// TODO: Add test cases.
		{name: "1", args: args{3232235777}, wantIp: "192.168.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIp := Int2Ip(tt.args.iip); gotIp != tt.wantIp {
				t.Errorf("Int2Ip() = %v, want %v", gotIp, tt.wantIp)
			}
		})
	}
}
