package db

type DBConfig struct {
	Type string
	DSN  string
}

func NewDBConfig(dsn string) *DBConfig {
	return &DBConfig{
		// hardcoded for now, can be extended later
		Type: "postgres",
		DSN:  dsn,
	}
}
