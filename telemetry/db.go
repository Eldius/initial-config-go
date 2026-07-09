package telemetry

import (
	"database/sql"
	"fmt"
	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
)

// GetSqlxDB returns a database connection.
func GetSqlxDB(driver, connStr string) (*sqlx.DB, error) {
	driverName, err := registerOtelSQL(driver)
	if err != nil {
		return nil, fmt.Errorf("failed to register otelsql driver: %w", err)
	}

	db, err := sqlx.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := instrumentOtelSQL(db.DB); err != nil {
		return nil, fmt.Errorf("failed to instrument otelsql: %w", err)
	}

	return db, nil
}

// GetDB returns a database connection.
func GetDB(driver, connStr string) (*sql.DB, error) {
	driverName, err := registerOtelSQL(driver)
	if err != nil {
		return nil, fmt.Errorf("failed to register otelsql driver: %w", err)
	}

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := instrumentOtelSQL(db); err != nil {
		return nil, fmt.Errorf("failed to instrument otelsql: %w", err)
	}
	return db, nil
}

func registerOtelSQL(driver string) (string, error) {
	return otelsql.Register(driver,
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithMeterProvider(otel.GetMeterProvider()),
		otelsql.WithAttributes(getDefaultTelemetryAttributes(cfgCache)...),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
		}),
	)
}

func instrumentOtelSQL(db *sql.DB) error {
	_, err := otelsql.RegisterDBStatsMetrics(db,
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithMeterProvider(otel.GetMeterProvider()),
		otelsql.WithAttributes(getDefaultTelemetryAttributes(cfgCache)...),
	)

	return err
}
