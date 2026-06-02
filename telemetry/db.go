package telemetry

import (
	"database/sql"
	"fmt"
	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
)

func GetSQLXDB(driver, connStr string) (*sqlx.DB, error) {
	driverName, err := otelsql.Register(driver,
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register otelsql driver: %w", err)
	}

	db, err := sqlx.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func GetDB(driver, connStr string) (*sql.DB, error) {
	driverName, err := otelsql.Register(driver,
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register otelsql driver: %w", err)
	}

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
