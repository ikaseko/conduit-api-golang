package repository

import (
	"context"
	"log/slog"
	"rwa/internal/model"
	"rwa/internal/security"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type UserStorage interface {
	AddUser(u model.UserTableDB) error
	GetUserForApi(id string) (model.User, error)
	GetUserForAuth(email string, username string) (model.UserTableDB, error)
	GetUserForUpdate(id string) (model.UserTableDB, error)
	DeleteUser(email string, username string) error
	UpdateUser(u model.UserTableDB) error
	AddToken(token string, uid string) error
	GetToken(token string) (model.UserAuthToken, error)
	DeleteToken(token string) error
	FindTokenByUID(uid string) (model.UserAuthToken, error)
	FollowUser(followerId string, followedId string) error
	UnFollowUser(followerId string, followedId string) error
	CheckFollow(followerId string, followedId string) (bool, error)
}

type PostgresUserStorage struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewPostgresUserStorage(db *pgxpool.Pool, log *slog.Logger) *PostgresUserStorage {
	return &PostgresUserStorage{db: db, log: log}
}

func (s PostgresUserStorage) RegisterUser(username string, email string, password string) error {
	const op = "PostgresUserStorage.RegisterUser"

	passwd, err := security.GeneratePasswd(password)
	if err != nil {
		s.log.Error("failed to generate password", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to generate password")
	}

	user := model.UserTableDB{
		Username:     username,
		Email:        email,
		PasswordHash: passwd.Hash,
		PasswordSalt: passwd.Salt,
		Bio:          "",
		Image:        "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.AddUser(user)
	if err != nil {
		s.log.Error("failed to add user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to add user")
	}

	s.log.Info("user registered successfully", slog.String("op", op), slog.String("username", username))
	return nil
}

func (s PostgresUserStorage) AddUser(u model.UserTableDB) error {
	const op = "PostgresUserStorage.AddUser"

	ctx := context.Background()
	query := `INSERT INTO users (username, email, password_hash, password_salt, bio, image) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.db.Exec(ctx, query, u.Username, u.Email, u.PasswordHash, u.PasswordSalt, u.Bio, u.Image)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "users_email_key" {
				s.log.Warn("email already exists", slog.String("op", op), slog.String("email", u.Email))
				return ErrEmailAlreadyExists
			}
			if pgErr.ConstraintName == "users_username_key" {
				s.log.Warn("username already exists", slog.String("op", op), slog.String("username", u.Username))
				return ErrUsernameAlreadyExists
			}
			s.log.Error("unique constraint violation", slog.String("op", op), slog.String("constraint", pgErr.ConstraintName))
			return errors.Wrap(err, "unique constraint violation")
		}
		s.log.Error("failed to add user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to execute query")
	}

	s.log.Info("user added successfully", slog.String("op", op), slog.String("username", u.Username))
	return nil
}

func (s PostgresUserStorage) GetUserForApi(uid string) (model.User, error) {
	const op = "PostgresUserStorage.GetUserForApi"

	if uid == "" {
		s.log.Warn("empty user ID provided", slog.String("op", op))
		return model.User{}, errors.New("user ID not provided")
	}

	ctx := context.Background()
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`
	row := s.db.QueryRow(ctx, query, uid)

	var userDB model.UserTableDB
	err := row.Scan(&userDB.ID, &userDB.Username, &userDB.Email, &userDB.PasswordHash, &userDB.PasswordSalt, &userDB.Bio, &userDB.Image, &userDB.CreatedAt, &userDB.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn("user not found", slog.String("op", op), slog.String("userID", uid))
			return model.User{}, ErrUserNotFound
		}
		s.log.Error("failed to scan user data", slog.String("op", op), slog.String("error", err.Error()))
		return model.User{}, errors.Wrap(err, "failed to scan user data")
	}

	user := model.User{
		ID:        userDB.ID,
		Username:  userDB.Username,
		Email:     userDB.Email,
		Bio:       userDB.Bio,
		Image:     userDB.Image,
		CreatedAt: userDB.CreatedAt,
		UpdatedAt: userDB.UpdatedAt,
	}

	s.log.Debug("user found", slog.String("op", op), slog.String("userID", uid))
	return user, nil
}

func (s PostgresUserStorage) GetUserForAuth(email string, username string) (model.UserTableDB, error) {
	const op = "PostgresUserStorage.GetUserForAuth"

	if email == "" && username == "" {
		s.log.Warn("empty credentials provided", slog.String("op", op))
		return model.UserTableDB{}, errors.New("email or username not provided")
	}

	var query string
	ctx := context.Background()
	var row pgx.Rows
	var err error

	if email == "" {
		query = `SELECT * FROM users WHERE username = $1 LIMIT 1`
		row, err = s.db.Query(ctx, query, username)
		s.log.Debug("querying by username", slog.String("op", op), slog.String("username", username))
	} else {
		query = `SELECT * FROM users WHERE email = $1 LIMIT 1`
		row, err = s.db.Query(ctx, query, email)
		s.log.Debug("querying by email", slog.String("op", op), slog.String("email", email))
	}

	if err != nil {
		s.log.Error("failed to query user", slog.String("op", op), slog.String("error", err.Error()))
		return model.UserTableDB{}, errors.Wrap(err, "failed to query user")
	}

	var userDB model.UserTableDB
	if row.Next() {
		err = row.Scan(&userDB.ID, &userDB.Username, &userDB.Email, &userDB.PasswordHash, &userDB.PasswordSalt, &userDB.Bio, &userDB.Image, &userDB.CreatedAt, &userDB.UpdatedAt)
		if err != nil {
			s.log.Error("failed to scan user data", slog.String("op", op), slog.String("error", err.Error()))
			return model.UserTableDB{}, errors.Wrap(err, "failed to scan user data")
		}
	} else {
		s.log.Warn("user not found", slog.String("op", op))
		return model.UserTableDB{}, ErrUserNotFound
	}

	s.log.Debug("user found for auth", slog.String("op", op), slog.String("userID", userDB.ID))
	return userDB, nil
}

func (s PostgresUserStorage) GetUserForUpdate(id string) (model.UserTableDB, error) {
	const op = "PostgresUserStorage.GetUserForUpdate"

	if id == "" {
		s.log.Warn("empty user ID provided", slog.String("op", op))
		return model.UserTableDB{}, errors.New("user ID not provided")
	}

	ctx := context.Background()
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`
	row := s.db.QueryRow(ctx, query, id)

	var userDB model.UserTableDB
	err := row.Scan(&userDB.ID, &userDB.Username, &userDB.Email, &userDB.PasswordHash, &userDB.PasswordSalt, &userDB.Bio, &userDB.Image, &userDB.CreatedAt, &userDB.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn("user not found", slog.String("op", op), slog.String("userID", id))
			return model.UserTableDB{}, ErrUserNotFound
		}
		s.log.Error("failed to scan user data", slog.String("op", op), slog.String("error", err.Error()))
		return model.UserTableDB{}, errors.Wrap(err, "failed to scan user data")
	}

	s.log.Debug("user found for update", slog.String("op", op), slog.String("userID", id))
	return userDB, nil
}

func (s PostgresUserStorage) DeleteUser(email string, username string) error {
	const op = "PostgresUserStorage.DeleteUser"

	if email == "" && username == "" {
		s.log.Warn("empty credentials provided", slog.String("op", op))
		return errors.New("email or username not provided")
	}

	query := `DELETE FROM users WHERE email = $1 OR username = $2`
	ctx := context.Background()
	result, err := s.db.Exec(ctx, query, email, username)
	if err != nil {
		s.log.Error("failed to delete user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to delete user")
	}

	if result.RowsAffected() == 0 {
		s.log.Warn("no user deleted", slog.String("op", op))
		return ErrUserNotFound
	}

	s.log.Info("user deleted", slog.String("op", op), slog.String("email", email), slog.String("username", username))
	return nil
}

func (s PostgresUserStorage) UpdateUser(u model.UserTableDB) error {
	const op = "PostgresUserStorage.UpdateUser"

	query := `UPDATE users SET username = $1, email = $2, password_hash = $3, password_salt = $4, bio = $5, image = $6, updated_at = $7 WHERE id = $8`
	ctx := context.Background()

	result, err := s.db.Exec(ctx, query, u.Username, u.Email, u.PasswordHash, u.PasswordSalt, u.Bio, u.Image, time.Now(), u.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "users_email_key" {
				s.log.Warn("email already exists", slog.String("op", op), slog.String("email", u.Email))
				return ErrEmailAlreadyExists
			}
			if pgErr.ConstraintName == "users_username_key" {
				s.log.Warn("username already exists", slog.String("op", op), slog.String("username", u.Username))
				return ErrUsernameAlreadyExists
			}
			s.log.Error("unique constraint violation", slog.String("op", op), slog.String("constraint", pgErr.ConstraintName))
			return errors.Wrap(err, "unique constraint violation")
		}
		s.log.Error("failed to update user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to update user")
	}

	if result.RowsAffected() == 0 {
		s.log.Warn("no user updated", slog.String("op", op), slog.String("userID", u.ID))
		return ErrUserNotFound
	}

	s.log.Info("user updated", slog.String("op", op), slog.String("userID", u.ID))
	return nil
}

func (s PostgresUserStorage) AddToken(token string, uid string) error {
	const op = "PostgresUserStorage.AddToken"

	if token == "" || uid == "" {
		s.log.Warn("empty token or user ID", slog.String("op", op))
		return errors.New("token or user ID not provided")
	}

	query := `INSERT INTO tokens (token, user_id) VALUES ($1, $2)`
	ctx := context.Background()
	exec, err := s.db.Exec(ctx, query, token, uid)
	if err != nil {
		s.log.Error("failed to add token", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to add token")
	}

	if exec.RowsAffected() == 0 {
		s.log.Warn("token not added", slog.String("op", op))
		return errors.New("token not added")
	}

	s.log.Info("token added", slog.String("op", op), slog.String("userID", uid))
	return nil
}

func (s PostgresUserStorage) GetToken(token string) (model.UserAuthToken, error) {
	const op = "PostgresUserStorage.GetToken"

	if token == "" {
		s.log.Warn("empty token provided", slog.String("op", op))
		return model.UserAuthToken{}, errors.New("token not provided")
	}

	query := `SELECT * FROM tokens WHERE token = $1`
	ctx := context.Background()
	row, err := s.db.Query(ctx, query, token)
	if err != nil {
		s.log.Error("failed to query token", slog.String("op", op), slog.String("error", err.Error()))
		return model.UserAuthToken{}, errors.Wrap(err, "failed to query token")
	}

	var tokenDB model.UserAuthToken
	if row.Next() {
		err = row.Scan(&tokenDB.ID, &tokenDB.Token, &tokenDB.CreatedAt, &tokenDB.EndDate, &tokenDB.UID)
		if err != nil {
			s.log.Error("failed to scan token data", slog.String("op", op), slog.String("error", err.Error()))
			return model.UserAuthToken{}, errors.Wrap(err, "failed to scan token data")
		}
	} else {
		s.log.Warn("token not found", slog.String("op", op))
		return model.UserAuthToken{}, ErrTokenIsNotFound
	}

	s.log.Debug("token found", slog.String("op", op), slog.String("tokenID", tokenDB.ID))
	return tokenDB, nil
}

func (s PostgresUserStorage) DeleteToken(token string) error {
	const op = "PostgresUserStorage.DeleteToken"

	if token == "" {
		s.log.Warn("empty token provided", slog.String("op", op))
		return errors.New("token not provided")
	}

	query := `DELETE FROM tokens WHERE token = $1`
	ctx := context.Background()
	result, err := s.db.Exec(ctx, query, token)
	if err != nil {
		s.log.Error("failed to delete token", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to delete token")
	}

	if result.RowsAffected() == 0 {
		s.log.Warn("no token deleted", slog.String("op", op))
		return ErrTokenIsNotFound
	}

	s.log.Info("token deleted", slog.String("op", op))
	return nil
}

func (s PostgresUserStorage) FindTokenByUID(uid string) (model.UserAuthToken, error) {
	const op = "PostgresUserStorage.FindTokenByUID"

	if uid == "" {
		s.log.Warn("empty user ID provided", slog.String("op", op))
		return model.UserAuthToken{}, errors.New("user ID not provided")
	}

	query := `SELECT * FROM tokens WHERE user_id = $1 LIMIT 1`
	ctx := context.Background()
	row := s.db.QueryRow(ctx, query, uid)

	var tokenDB model.UserAuthToken
	err := row.Scan(&tokenDB.ID, &tokenDB.Token, &tokenDB.CreatedAt, &tokenDB.EndDate, &tokenDB.UID)
	if errors.Is(err, pgx.ErrNoRows) {
		s.log.Warn("token not found for user", slog.String("op", op), slog.String("userID", uid))
		return model.UserAuthToken{}, ErrNoResult
	}

	if err != nil {
		s.log.Error("failed to scan token data", slog.String("op", op), slog.String("error", err.Error()))
		return model.UserAuthToken{}, errors.Wrap(err, "failed to scan token data")
	}

	s.log.Debug("token found for user", slog.String("op", op), slog.String("userID", uid))
	return tokenDB, nil
}

func (s PostgresUserStorage) FollowUser(followerId string, followedId string) error {
	const op = "PostgresUserStorage.FollowUser"

	if followerId == "" || followedId == "" {
		s.log.Warn("empty user ID(s) provided", slog.String("op", op))
		return errors.New("follower ID or followed ID not provided")
	}

	query := `INSERT INTO subscriptions (sub_id, target_user_id) VALUES ($1, $2)`
	ctx := context.Background()
	_, err := s.db.Exec(ctx, query, followerId, followedId)
	if err != nil {
		s.log.Error("failed to follow user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to follow user")
	}

	s.log.Info("user followed", slog.String("op", op), slog.String("followerID", followerId), slog.String("followedID", followedId))
	return nil
}

func (s PostgresUserStorage) UnFollowUser(followerId string, followedId string) error {
	const op = "PostgresUserStorage.UnFollowUser"

	if followerId == "" || followedId == "" {
		s.log.Warn("empty user ID(s) provided", slog.String("op", op))
		return errors.New("follower ID or followed ID not provided")
	}

	query := `DELETE FROM subscriptions WHERE sub_id = $1 AND target_user_id = $2`
	ctx := context.Background()
	result, err := s.db.Exec(ctx, query, followerId, followedId)
	if err != nil {
		s.log.Error("failed to unfollow user", slog.String("op", op), slog.String("error", err.Error()))
		return errors.Wrap(err, "failed to unfollow user")
	}

	if result.RowsAffected() == 0 {
		s.log.Warn("no subscription found to delete", slog.String("op", op))
		return errors.New("subscription not found")
	}

	s.log.Info("user unfollowed", slog.String("op", op), slog.String("followerID", followerId), slog.String("followedID", followedId))
	return nil
}

func (s PostgresUserStorage) CheckFollow(followerId string, followedId string) (bool, error) {
	const op = "PostgresUserStorage.CheckFollow"

	if followerId == "" || followedId == "" {
		s.log.Warn("empty user ID(s) provided", slog.String("op", op))
		return false, errors.New("follower ID or followed ID not provided")
	}

	query := `SELECT * FROM subscriptions WHERE sub_id = $1 AND target_user_id = $2`
	ctx := context.Background()
	row, err := s.db.Query(ctx, query, followerId, followedId)
	if err != nil {
		s.log.Error("failed to check follow status", slog.String("op", op), slog.String("error", err.Error()))
		return false, errors.Wrap(err, "failed to check follow status")
	}

	followed := row.Next()
	s.log.Debug("follow status checked",
		slog.String("op", op),
		slog.String("followerID", followerId),
		slog.String("followedID", followedId),
		slog.Bool("isFollowing", followed))

	return followed, nil
}
