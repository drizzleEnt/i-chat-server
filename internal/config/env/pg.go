package env

import (
	"chatsrv/internal/config"
	"fmt"
)

type pgConfig struct {
	dsn string
}

const (
	dbname     = "POSTGRES_DB"
	dbuser     = "POSTGRES_USER"
	dbpassword = "POSTGRES_PASSWORD"
	dbport     = "PG_PORT"
	dbhost     = "PG_HOST"
	dbssl      = "PG_SSL"
)

func NewPGConfig() (*pgConfig, error) {
	host := config.GetEnvStringOrDefault(dbhost, "0.0.0.0")
	port := config.GetEnvStringOrDefault(dbport, "5432")
	name := config.GetEnvStringOrDefault(dbname, "db")
	user := config.GetEnvStringOrDefault(dbuser, "postgres")
	password := config.GetEnvStringOrDefault(dbpassword, "postgres")
	ssl := config.GetEnvStringOrDefault(dbssl, "disable")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, name, user, password, ssl)

	return &pgConfig{
		dsn: dsn,
	}, nil
}

func (cfg *pgConfig) DSN() string {
	return cfg.dsn
}
