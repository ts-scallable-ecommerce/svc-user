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
                t.Fatalf("sqlmock.New error: %v", err)
        }
        defer db.Close()

        rows := sqlmock.NewRows([]string{"name"}).AddRow("user.read").AddRow("user.write")
        mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

        resolver := NewPermissionResolver(db)
        perms, err := resolver.ResolvePermissions(context.Background(), "user-1")
        if err != nil {
                t.Fatalf("ResolvePermissions returned error: %v", err)
        }
        if len(perms) != 2 {
                t.Fatalf("expected 2 permissions, got %d", len(perms))
        }

        if err := mock.ExpectationsWereMet(); err != nil {
                t.Fatalf("expectations were not met: %v", err)
        }
}

func TestResolvePermissionsQueryError(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("sqlmock.New error: %v", err)
        }
        defer db.Close()

        mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnError(errors.New("boom"))

        resolver := NewPermissionResolver(db)
        if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
                t.Fatalf("expected error from ResolvePermissions")
        }
}

func TestResolvePermissionsRowError(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
                t.Fatalf("sqlmock.New error: %v", err)
        }
        defer db.Close()

        rows := sqlmock.NewRows([]string{"name"}).AddRow("user.read").RowError(0, errors.New("scan"))
        mock.ExpectQuery("SELECT p.name").WithArgs("user-1").WillReturnRows(rows)

        resolver := NewPermissionResolver(db)
        if _, err := resolver.ResolvePermissions(context.Background(), "user-1"); err == nil {
                t.Fatalf("expected scan error")
        }
}
