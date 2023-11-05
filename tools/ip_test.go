package tools

import (
	"testing"
)

func TestListIpsInNetwork(t *testing.T) {
	type args struct {
		cidrAddress string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{cidrAddress: "192.168.0.0/24"},
			want:    256,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListIpsInNetwork(tt.args.cidrAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("Subnets() got = %v ips, want %v", len(got), tt.want)
			}
		})
	}
}
