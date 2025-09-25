package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnectSuccess(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	original := openSQL
	openSQL = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	defer func() { openSQL = original }()

	mock.ExpectPing()

	ctx := context.Background()
	conn, err := Connect(ctx, "postgres://user:pass@localhost/db")
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if conn != db {
		t.Fatalf("expected returned db to match mock")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestConnectPingError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	original := openSQL
	openSQL = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	defer func() { openSQL = original }()

	mock.ExpectPing().WillReturnError(errors.New("unreachable"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := Connect(ctx, "postgres://user:pass@localhost/db"); err == nil {
		t.Fatalf("expected ping error")
	}
}
