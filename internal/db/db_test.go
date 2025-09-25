package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"database/sql"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestConnectFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := Connect(ctx, "postgres://invalid"); err == nil {
		t.Fatal("expected connection failure")
	}
}

func TestConnectSuccess(t *testing.T) {
	original := openDB
	defer func() { openDB = original }()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	openDB = func(string, string) (*sql.DB, error) { return db, nil }
	mock.ExpectPing()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, err := Connect(ctx, "postgres://example")
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if got != db {
		t.Fatal("expected returned db to equal mock instance")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestConnectPingError(t *testing.T) {
	original := openDB
	defer func() { openDB = original }()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	openDB = func(string, string) (*sql.DB, error) { return db, nil }
	mock.ExpectPing().WillReturnError(errors.New("unavailable"))

	if _, err := Connect(context.Background(), "postgres://example"); err == nil {
		t.Fatal("expected ping error")
	}
}

func TestConnectOpenError(t *testing.T) {
	original := openDB
	defer func() { openDB = original }()

	openDB = func(string, string) (*sql.DB, error) { return nil, errors.New("open error") }

	if _, err := Connect(context.Background(), "postgres://example"); err == nil {
		t.Fatal("expected open error")
	}
}
