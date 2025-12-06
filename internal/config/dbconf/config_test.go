// Package dbconf provide database configurations.
package dbconf

import (
	"reflect"
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
		wantErr error
		fields  fields
		name    string
		args    args
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

func Test_getDialect(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "postgres", args: args{s: "postgresql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, want: "postgresql"},
		{name: "mysql", args: args{s: "mysql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, want: "mysql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDialect(tt.args.s)
			if assert.NoError(t, err) {
				if !assert.Equal(t, tt.want, got) {
					t.Errorf("getDialect() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_getCredentials(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name     string
		args     args
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{name: "postgres", args: args{s: "postgresql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, wantUser: "test_user", wantPass: "test_password"},
		{name: "mysql", args: args{s: "mysql://user:password@localhost:5432/test_db?search_path=test_schema"}, wantUser: "user", wantPass: "password"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCredentials(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got.user, tt.wantUser) {
				t.Errorf("getCredentials() user = %v, want user %v", got.user, tt.wantUser)
			}
			if !assert.Equal(t, got.pass, tt.wantPass) {
				t.Errorf("getCredentials() password = %v, want password %v", got.pass, tt.wantPass)
			}
		})
	}
}

// func Test_getDBSpec(t *testing.T) {
// 	type args struct {
// 		s string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want dbSpec
// 	}{
// 		{name: "postgres", args: args{s: "postgresql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, want: dbSpec{host: "localhost", port: "5432", database: "test_db"}},
// 		{name: "mysql", args: args{s: "mysql://user:password@127.0.0.1:/_db?search_path=test_schema"}, want: dbSpec{host: "127.0.0.1", port: "", database: "_db"}},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got, err := getDBSpec(tt.args.s); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("getDBSpec() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_getOptions(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		want Options
		name string
		args args
	}{
		{name: "postgres", args: args{s: "postgresql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, want: Options{"search_path": "test_schema"}},
		{name: "mysql", args: args{s: "mysql://user:password@127.0.0.1:/_db?search_path=test_schema?option1=value2"}, want: Options{"search_path": "test_schema", "option1": "value2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOptions(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDBSpec(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    dbSpec
		wantErr bool
	}{
		{name: "postgres", args: args{s: "postgresql://test_user:test_password@localhost:5432/test_db?search_path=test_schema"}, want: dbSpec{host: "localhost", port: "5432", database: "test_db"}},
		{name: "mysql", args: args{s: "mysql://user:password@127.0.0.1:/_db?search_path=test_schema"}, want: dbSpec{host: "127.0.0.1", port: "", database: "_db"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDBSpec(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDBSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDBSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
