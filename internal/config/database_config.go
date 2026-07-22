package config

import (
	"context"

	"github.com/caarlos0/env/v11"

	"github.com/mickamy/employee-management/internal/lib/validator"
)

type Database struct {
	WriterURL string `env:"DATABASE_WRITER_URL" validate:"required"`
	ReaderURL string `env:"DATABASE_READER_URL" validate:"required"`
}

func ParseDatabase() Database {
	var config Database
	if err := env.Parse(&config); err != nil {
		panic(err)
	}
	if err := validator.Struct(context.Background(), &config); err != nil {
		panic(err)
	}
	return config
}
