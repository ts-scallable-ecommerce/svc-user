package users

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// User represents the core user entity persisted in PostgreSQL.
type User struct {
	ID              string
	Email           string
	Phone           sql.NullString
	PasswordHash    string
	FirstName       sql.NullString
	LastName        sql.NullString
	Status          string
	EmailVerifiedAt sql.NullTime
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Repository describes read/write operations for the user aggregate.
type Repository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, u *User) error
}

var ErrNotFound = errors.New("user not found")

// SQLRepository is a simple implementation backed by database/sql.
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a repository instance.
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// Create inserts a new user record.
func (r *SQLRepository) Create(ctx context.Context, u *User) error {
	query := `INSERT INTO users (email, password_hash, first_name, last_name, status) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Status).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// FindByEmail returns a user by email.
func (r *SQLRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE email=$1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Status,
		&u.EmailVerifiedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindByID returns a user by identifier.
func (r *SQLRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, phone, password_hash, first_name, last_name, status, email_verified_at, created_at, updated_at FROM users WHERE id=$1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Status,
		&u.EmailVerifiedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Update persists modified fields of a user.
func (r *SQLRepository) Update(ctx context.Context, u *User) error {
	query := `UPDATE users SET phone=$1, first_name=$2, last_name=$3, status=$4, email_verified_at=$5, updated_at=now() WHERE id=$6`
	res, err := r.db.ExecContext(ctx, query, u.Phone, u.FirstName, u.LastName, u.Status, u.EmailVerifiedAt, u.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
