package users

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Second)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, first_name, last_name, status) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at")).
		WithArgs("user@example.com", "hash", sql.NullString{}, sql.NullString{}, "pending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow("id-1", createdAt, updatedAt))

	user := &User{Email: "user@example.com", PasswordHash: "hash", Status: "pending"}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID != "id-1" {
		t.Fatalf("expected user ID to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestSQLRepositoryFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("id-1", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, time.Now(), time.Now())
	mock.ExpectQuery("SELECT id, email").WithArgs("user@example.com").WillReturnRows(rows)

	user, err := repo.FindByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("FindByEmail returned error: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", user.Email)
	}
}

func TestSQLRepositoryFindByEmailScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("id-1", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, time.Now(), time.Now()).RowError(0, errors.New("scan"))
	mock.ExpectQuery("SELECT id, email").WithArgs("user@example.com").WillReturnRows(rows)

	if _, err := repo.FindByEmail(context.Background(), "user@example.com"); err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestSQLRepositoryFindByEmailNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	mock.ExpectQuery("SELECT id, email").WithArgs("missing@example.com").WillReturnError(sql.ErrNoRows)

	if _, err := repo.FindByEmail(context.Background(), "missing@example.com"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLRepositoryFindByIDError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	mock.ExpectQuery("SELECT id, email").WithArgs("id-404").WillReturnError(errors.New("boom"))

	if _, err := repo.FindByID(context.Background(), "id-404"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSQLRepositoryFindByIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	mock.ExpectQuery("SELECT id, email").WithArgs("missing").WillReturnError(sql.ErrNoRows)

	if _, err := repo.FindByID(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLRepositoryFindByIDSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("id-1", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, time.Now(), time.Now())
	mock.ExpectQuery("SELECT id, email").WithArgs("id-1").WillReturnRows(rows)

	u, err := repo.FindByID(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if u.ID != "id-1" {
		t.Fatalf("unexpected id: %s", u.ID)
	}
}

func TestSQLRepositoryUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer db.Close()

	repo := NewSQLRepository(db)

	mock.ExpectExec("UPDATE users SET").
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "id-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), &User{ID: "id-1", Status: "active"}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	mock.ExpectExec("UPDATE users SET").
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "id-2").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Update(context.Background(), &User{ID: "id-2", Status: "active"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing user, got %v", err)
	}

	mock.ExpectExec("UPDATE users SET").
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "id-3").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows")))

	if err := repo.Update(context.Background(), &User{ID: "id-3", Status: "active"}); err == nil {
		t.Fatalf("expected rows affected error")
	}

	mock.ExpectExec("UPDATE users SET").
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "id-4").
		WillReturnError(errors.New("exec"))

	if err := repo.Update(context.Background(), &User{ID: "id-4", Status: "active"}); err == nil {
		t.Fatalf("expected exec error")
	}
}
