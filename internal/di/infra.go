package di

import (
	"context"
	"fmt"

	"github.com/mickamy/employee-management/internal/config"
	"github.com/mickamy/employee-management/internal/infra/storage/db"
	"github.com/mickamy/employee-management/internal/infra/storage/tx"
)

type Infra struct {
	_              context.Context   `inject:"arg"` //nolint:containedctx // required by injector
	_              *Config           `inject:"embed"`
	Writer         db.Writer         `inject:"with=provideWriter"`
	Reader         db.Reader         `inject:"with=provideReader"`
	Transactor     tx.Transactor     `inject:"with=provideTransactor"`
	ReadTransactor tx.ReadTransactor `inject:"with=provideReadTransactor"`
}

func (infra *Infra) Close() error {
	infra.Writer.Close()
	infra.Reader.Close()
	return nil
}

func provideWriter(ctx context.Context, cfg config.Database) (db.Writer, error) {
	writer, err := db.NewWriter(ctx, cfg.WriterURL)
	if err != nil {
		return db.Writer{}, fmt.Errorf("new writer: %w", err)
	}
	return writer, nil
}

func provideReader(ctx context.Context, cfg config.Database, writer db.Writer) (db.Reader, error) {
	reader, err := db.NewReader(ctx, cfg.ReaderURL)
	if err != nil {
		writer.Close()
		return db.Reader{}, fmt.Errorf("new reader: %w", err)
	}
	return reader, nil
}

func provideTransactor(writer db.Writer) tx.Transactor {
	return tx.NewTransactor(writer)
}

func provideReadTransactor(reader db.Reader) tx.ReadTransactor {
	return tx.NewReadTransactor(reader)
}
