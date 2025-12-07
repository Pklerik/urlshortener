package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/config/audit"
	"github.com/Pklerik/urlshortener/internal/config/dbconf"
)

type EnvMap map[string]string

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name    string
		envVars EnvMap
		want    config.StartupFlagsParser
	}{
		{name: "base",
			envVars: EnvMap{"DATABASE_DSN": "postgresql://test_user:test_pass@localhost:5432/test_db?search_path=test_schema", "SECRET_KEY": "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs="},
			want: &config.StartupFlags{ServerAddress: &config.Address{Protocol: "http", Host: "localhost", Port: 8080},
				BaseURL: "http://localhost:8080", LogLevel: "info", LocalStorage: "local_storage.json", Timeout: 600,
				SecretKey: "fH72anZI1e6YFLN+Psh6Dv308js8Ul+q3mfPe8E36Qs=",
				DBConf: &dbconf.Conf{
					RawString: "postgresql://test_user:test_pass@localhost:5432/test_db?search_path=test_schema",
					Dialect:   "postgresql",
					User:      "test_user",
					Password:  "test_pass",
					Host:      "localhost",
					Port:      "5432",
					Database:  "test_db",
					Options:   dbconf.Options{"search_path": "test_schema"},
				},
				Audit: &audit.Audit{
					LogFilePath: "",
					LogURLPath:  "",
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			if got := parseArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFlags() = %v, want %v", got, tt.want)
			}
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
