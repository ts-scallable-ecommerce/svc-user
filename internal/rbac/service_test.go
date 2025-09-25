package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestResolvePermissionsSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name"}).AddRow("read").AddRow("write")
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

	resolver := NewPermissionResolver(db)
	perms, err := resolver.ResolvePermissions(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ResolvePermissions error: %v", err)
	}
	if len(perms) != 2 || perms[0] != "read" || perms[1] != "write" {
		t.Fatalf("unexpected permissions: %v", perms)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestResolvePermissionsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnError(errors.New("db error"))

	resolver := NewPermissionResolver(db)
	if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
		t.Fatalf("expected error from query failure")
	}
}

func TestResolvePermissionsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name"}).AddRow(nil)
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

	resolver := NewPermissionResolver(db)
	if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestResolvePermissionsRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock error: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name"}).AddRow("read").RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

	resolver := NewPermissionResolver(db)
	if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
		t.Fatalf("expected rows error")
	}
}
