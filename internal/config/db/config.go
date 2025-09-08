package db

type DBConf struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func NewDBConf() *DBConf {
	return &DBConf{
		DatabaseDSN: "",
	}
}
