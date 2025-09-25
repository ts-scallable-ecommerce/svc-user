package rbac

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestResolvePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	resolver := NewPermissionResolver(db)

	rows := sqlmock.NewRows([]string{"name"}).AddRow("read").AddRow("write")
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

	perms, err := resolver.ResolvePermissions(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ResolvePermissions() error = %v", err)
	}
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(perms))
	}
}

func TestResolvePermissionsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	resolver := NewPermissionResolver(db)
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnError(sql.ErrConnDone)

	if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
		t.Fatal("expected query error")
	}
}

func TestResolvePermissionsRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	resolver := NewPermissionResolver(db)
	rows := sqlmock.NewRows([]string{"name"}).AddRow("read").RowError(0, sql.ErrNoRows)
	mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

	if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
		t.Fatal("expected rows error")
	}
}
