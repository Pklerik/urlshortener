// Package dbconf provide database configurations.
package dbconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConf_Set(t *testing.T) {
	type fields struct {
		User      string
		Password  string
		Host      string
		Port      string
		Database  string
		Options   Options
		RawString string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{name: "empty conf", fields: fields{}, args: args{s: ""}, wantErr: ErrEmptyDatabaseConfig},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http"}, wantErr: ErrIncorrectDatabaseURL},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http://"}, wantErr: ErrIncorrectDatabaseURL},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http://asdd:"}, wantErr: ErrIncorrectDatabaseURL},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http://asdd:asdad@"}, wantErr: ErrIncorrectDatabaseURL},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http://asdd:asdad@asda:"}, wantErr: ErrIncorrectDatabaseURL},
		{name: "incorrect DB URL", fields: fields{}, args: args{s: "http://asdd:asdad@asda:1234/"}, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbc := &Conf{
				User:      tt.fields.User,
				Password:  tt.fields.Password,
				Host:      tt.fields.Host,
				Port:      tt.fields.Port,
				Database:  tt.fields.Database,
				Options:   tt.fields.Options,
				RawString: tt.fields.RawString,
			}
			if err := dbc.Set(tt.args.s); err != nil {
				if !assert.ErrorAs(t, err, &tt.wantErr) {
					t.Errorf("Conf.Set() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
