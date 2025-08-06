package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_Set(t *testing.T) {
	type fields struct {
		Protocol string
		Host     string
		Port     int
	}
	type args struct {
		flagValue string
	}
	type want struct {
		address *Address
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    want
	}{
		{name: "base", args: args{flagValue: "http://127.0.0.1:8080"}, wantErr: false,
			want: want{address: &Address{
				Protocol: "http",
				Host:     "127.0.0.1",
				Port:     8080,
			}}},
		{name: "empty protocol", args: args{flagValue: "localhost:8080"}, wantErr: false,
			want: want{address: &Address{
				Protocol: "http",
				Host:     "localhost",
				Port:     8080,
			}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Address{
				Protocol: tt.fields.Protocol,
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
			}
			if err := a.Set(tt.args.flagValue); (err != nil) != tt.wantErr {
				t.Errorf("Address.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, *tt.want.address, *a)
		})
	}
}
