package config

import "os"

func GetEnvStringOrDefault(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return val
}

type HttpConfig interface {
	Address() string
}

type PGConfig interface {
	DSN() string
}
