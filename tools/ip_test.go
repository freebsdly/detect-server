package tools

import (
	"testing"
)

func TestSubnetIps(t *testing.T) {
	type args struct {
		subnet string
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name:      "192.168.1.0/24",
			args:      args{subnet: "192.168.1.0/24"},
			wantErr:   false,
			wantCount: 256,
		},
		{
			name:      "192.168.0.0/23",
			args:      args{subnet: "192.168.1.0/23"},
			wantErr:   false,
			wantCount: 1024,
		},
		{
			name:      "192.168.0.0/22",
			args:      args{subnet: "192.168.1.0/22"},
			wantErr:   false,
			wantCount: 1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SubnetIps(tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubnetIps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("SubnetIps() got ip count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}

func TestSubnets(t *testing.T) {
	type args struct {
		subnet string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "192.168.1.1/23",
			args: args{subnet: "192.168.1.1/23"},
		},
		{
			name: "192.168.0.1/23",
			args: args{subnet: "192.168.0.1/23"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Subnets(tt.args.subnet)
		})
	}
}

func Test_listSubnets(t *testing.T) {
	type args struct {
		cidrAddress  string
		newPrefixLen int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				cidrAddress:  "192.168.1.0/22",
				newPrefixLen: 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listSubnets(tt.args.cidrAddress, tt.args.newPrefixLen)
		})
	}
}
