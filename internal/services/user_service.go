package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lohanguedes/gobid/internal/store/pgstore"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	// TODO: make this a interface in order to be more idiomatic
	pool *pgxpool.Pool // do not forget to add the pool of connections here.
	db   *pgstore.Queries
}

func NewUserService(pool *pgxpool.Pool) UserService {
	return UserService{
		pool: pool,
		db:   pgstore.New(pool),
	}
}

func (us *UserService) CreateUser(ctx context.Context, userName, email, password, bio string) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return uuid.UUID{}, err
	}

	args := pgstore.CreateUserParams{
		UserName:     userName,
		Email:        email,
		PasswordHash: hash,
		Bio:          bio,
	}

	id, err := us.db.CreateUser(
		ctx,
		args,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.UUID{}, pgstore.ErrDuplicateEmail
		}
	}

	return id, nil
}

func (us *UserService) Authenticate(ctx context.Context, email, password string) (uuid.UUID, error) {
	user, err := us.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, pgstore.ErrInvalidCredentials
		}
	}

	err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return uuid.UUID{}, pgstore.ErrInvalidCredentials
		}

		return uuid.UUID{}, err
	}

	return user.ID, err
}
