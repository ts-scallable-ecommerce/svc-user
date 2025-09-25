package users

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	cleanup := func() {
		db.Close()
	}
	return db, mock, cleanup
}

func TestSQLRepositoryCreate(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{Email: "user@example.com", PasswordHash: "hash", Status: "pending"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, first_name, last_name, status) VALUES ($1,$2,$3,$4,$5) RETURNING id,\ncreated_at, updated_at")).
		WithArgs(user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow("123", time.Now(), time.Now()))

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.ID != "123" {
		t.Fatalf("user.ID = %s, want 123", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepositoryFindByEmail(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("1", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("user@example.com").
		WillReturnRows(rows)

	user, err := repo.FindByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("FindByEmail() email = %s", user.Email)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLRepositoryFindByEmailNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	if _, err := repo.FindByEmail(context.Background(), "missing@example.com"); err != ErrNotFound {
		t.Fatalf("FindByEmail() error = %v, want %v", err, ErrNotFound)
	}
}

func TestSQLRepositoryFindByEmailError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1")).
		WithArgs("user@example.com").
		WillReturnError(fmt.Errorf("db error"))

	if _, err := repo.FindByEmail(context.Background(), "user@example.com"); err == nil {
		t.Fatal("expected query error")
	}
}

func TestSQLRepositoryFindByID(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)

	rows := sqlmock.NewRows([]string{"id", "email", "phone", "password_hash", "first_name", "last_name", "status", "email_verified_at", "created_at", "updated_at"}).
		AddRow("1", "user@example.com", sql.NullString{}, "hash", sql.NullString{}, sql.NullString{}, "active", sql.NullTime{}, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1")).
		WithArgs("1").
		WillReturnRows(rows)

	user, err := repo.FindByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if user.ID != "1" {
		t.Fatalf("FindByID() id = %s", user.ID)
	}
}

func TestSQLRepositoryFindByIDNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1")).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	if _, err := repo.FindByID(context.Background(), "missing"); err != ErrNotFound {
		t.Fatalf("FindByID() error = %v, want %v", err, ErrNotFound)
	}
}

func TestSQLRepositoryFindByIDError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1")).
		WithArgs("1").
		WillReturnError(fmt.Errorf("db error"))

	if _, err := repo.FindByID(context.Background(), "1"); err == nil {
		t.Fatal("expected query error")
	}
}

func TestSQLRepositoryUpdate(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{ID: "1", Phone: sql.NullString{String: "123", Valid: true}}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(user.Phone, user.FirstName, user.LastName, user.Status, user.EmailVerifiedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestSQLRepositoryUpdateNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{ID: "missing"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(user.Phone, user.FirstName, user.LastName, user.Status, user.EmailVerifiedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Update(context.Background(), user); err != ErrNotFound {
		t.Fatalf("Update() error = %v, want %v", err, ErrNotFound)
	}
}

func TestSQLRepositoryUpdateExecError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{ID: "1"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(user.Phone, user.FirstName, user.LastName, user.Status, user.EmailVerifiedAt, user.ID).
		WillReturnError(fmt.Errorf("exec error"))

	if err := repo.Update(context.Background(), user); err == nil {
		t.Fatal("expected exec error")
	}
}

func TestSQLRepositoryUpdateRowsAffectedError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{ID: "1"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6")).
		WithArgs(user.Phone, user.FirstName, user.LastName, user.Status, user.EmailVerifiedAt, user.ID).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	if err := repo.Update(context.Background(), user); err == nil {
		t.Fatal("expected rows affected error")
	}
}

func TestSQLRepositoryCreateError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	repo := NewSQLRepository(db)
	user := &User{Email: "user@example.com", PasswordHash: "hash", Status: "pending"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (email, password_hash, first_name, last_name, status) VALUES ($1,$2,$3,$4,$5) RETURNING id,\ncreated_at, updated_at")).
		WithArgs(user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Status).
		WillReturnError(fmt.Errorf("insert error"))

	if err := repo.Create(context.Background(), user); err == nil {
		t.Fatal("expected create error")
	}
}
