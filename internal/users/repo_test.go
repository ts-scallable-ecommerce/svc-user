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

func newRepo(t *testing.T) (*SQLRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return NewSQLRepository(db), mock, func() { db.Close() }
}

func TestSQLRepositoryCreate(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, first_name, last_name, status) VALUES ($1,$2,$3,$4,$5) RETURNING id,\ncreated_at, updated_at")).
		WithArgs("user@example.com", "hash", sql.NullString{}, sql.NullString{}, "pending").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow("uuid", time.Now(), time.Now()))

	user := &User{
		Email:        "user@example.com",
		PasswordHash: "hash",
		Status:       "pending",
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if user.ID != "uuid" {
		t.Fatalf("expected id to be populated")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLRepositoryFindByEmail(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("uuid", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("user@example.com").
		WillReturnRows(rows)

	user, err := repo.FindByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("FindByEmail error: %v", err)
	}
	if user.Email != "user@example.com" || user.Status != "active" {
		t.Fatalf("unexpected user data: %+v", user)
	}
}

func TestSQLRepositoryFindByEmailNotFound(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	if _, err := repo.FindByEmail(context.Background(), "missing@example.com"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLRepositoryFindByEmailError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("user@example.com").
		WillReturnError(sql.ErrConnDone)

	if _, err := repo.FindByEmail(context.Background(), "user@example.com"); err != sql.ErrConnDone {
		t.Fatalf("expected raw error, got %v", err)
	}
}

func TestSQLRepositoryFindByIDError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1")).
		WithArgs("uuid").
		WillReturnError(sql.ErrConnDone)

	if _, err := repo.FindByID(context.Background(), "uuid"); err != sql.ErrConnDone {
		t.Fatalf("expected raw error, got %v", err)
	}
}

func TestSQLRepositoryFindByIDSuccess(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("uuid", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1")).
		WithArgs("uuid").
		WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), "uuid")
	if err != nil {
		t.Fatalf("FindByID error: %v", err)
	}
	if user.ID != "uuid" {
		t.Fatalf("expected uuid, got %s", user.ID)
	}
}

func TestSQLRepositoryUpdate(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "uuid").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), &User{ID: "uuid", Status: "active"})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
}

func TestSQLRepositoryUpdateNotFound(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "uuid").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Update(context.Background(), &User{ID: "uuid", Status: "active"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLRepositoryUpdateError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "uuid").
		WillReturnError(sql.ErrConnDone)

	if err := repo.Update(context.Background(), &User{ID: "uuid", Status: "active"}); err != sql.ErrConnDone {
		t.Fatalf("expected raw error, got %v", err)
	}
}

func TestSQLRepositoryUpdateRowsAffectedError(t *testing.T) {
	repo, mock, cleanup := newRepo(t)
	defer cleanup()

	result := sqlmock.NewErrorResult(errors.New("rows affected error"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(sql.NullString{}, sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, "uuid").
		WillReturnResult(result)

	if err := repo.Update(context.Background(), &User{ID: "uuid", Status: "active"}); err == nil {
		t.Fatalf("expected rows affected error")
	}
}
