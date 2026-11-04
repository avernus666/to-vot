package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/avernus666/averna/apps/api/internal/security"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh token")
	ErrEmailTaken         = errors.New("email is already registered")
)

const (
	accessLifetime  = 15 * time.Minute
	refreshLifetime = 30 * 24 * time.Hour
)

type Service struct {
	db *pgxpool.Pool
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
	OrgName     string
	OrgSlug     string
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (User, TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	displayName := strings.TrimSpace(in.DisplayName)
	orgName := strings.TrimSpace(in.OrgName)
	slug := strings.ToLower(strings.TrimSpace(in.OrgSlug))

	if email == "" || !strings.Contains(email, "@") || len(email) > 320 {
		return User{}, TokenPair{}, errors.New("a valid email address is required")
	}
	if len(in.Password) < 12 || len(in.Password) > 1024 {
		return User{}, TokenPair{}, errors.New("password must be between 12 and 1024 bytes")
	}
	if displayName == "" || len(displayName) > 120 {
		return User{}, TokenPair{}, errors.New("display name must be between 1 and 120 characters")
	}
	if orgName == "" || len(orgName) > 120 {
		return User{}, TokenPair{}, errors.New("organization name must be between 1 and 120 characters")
	}
	if !validSlug(slug) {
		return User{}, TokenPair{}, errors.New("organization slug must contain 3-48 lowercase letters, numbers, or hyphens")
	}

	passwordHash, err := security.HashPassword(in.Password)
	if err != nil {
		return User{}, TokenPair{}, err
	}
	pair, accessHash, refreshHash, err := newTokenPair()
	if err != nil {
		return User{}, TokenPair{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return User{}, TokenPair{}, fmt.Errorf("begin registration: %w", err)
	}
	defer tx.Rollback(ctx)

	user := User{}
	err = tx.QueryRow(ctx, `
		INSERT INTO users(email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name
	`, email, passwordHash, displayName).Scan(&user.ID, &user.Email, &user.DisplayName)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, TokenPair{}, ErrEmailTaken
		}
		return User{}, TokenPair{}, fmt.Errorf("create user: %w", err)
	}

	var orgID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO organizations(name, slug)
		VALUES ($1, $2)
		RETURNING id
	`, orgName, slug).Scan(&orgID)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, TokenPair{}, fmt.Errorf("organization slug is already taken")
		}
		return User{}, TokenPair{}, fmt.Errorf("create organization: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members(organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, orgID, user.ID); err != nil {
		return User{}, TokenPair{}, fmt.Errorf("create organization owner membership: %w", err)
	}

	if err := insertSession(ctx, tx, user.ID, accessHash, refreshHash, pair); err != nil {
		return User{}, TokenPair{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, TokenPair{}, fmt.Errorf("commit registration: %w", err)
	}
	return user, pair, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var user User
	var passwordHash string
	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, TokenPair{}, ErrInvalidCredentials
		}
		return User{}, TokenPair{}, fmt.Errorf("load user: %w", err)
	}

	ok, err := security.VerifyPassword(password, passwordHash)
	if err != nil {
		return User{}, TokenPair{}, fmt.Errorf("verify password hash: %w", err)
	}
	if !ok {
		return User{}, TokenPair{}, ErrInvalidCredentials
	}

	pair, accessHash, refreshHash, err := newTokenPair()
	if err != nil {
		return User{}, TokenPair{}, err
	}
	if err := insertSession(ctx, s.db, user.ID, accessHash, refreshHash, pair); err != nil {
		return User{}, TokenPair{}, err
	}
	return user, pair, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (TokenPair, error) {
	if rawRefresh == "" {
		return TokenPair{}, ErrInvalidRefresh
	}
	newPair, accessHash, refreshHash, err := newTokenPair()
	if err != nil {
		return TokenPair{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return TokenPair{}, fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE sessions
		SET access_token_hash = $1,
		    access_expires_at = $2,
		    refresh_token_hash = $3,
		    expires_at = $4,
		    last_used_at = now()
		WHERE refresh_token_hash = $5
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`, accessHash, newPair.AccessExpiresAt, refreshHash, newPair.RefreshExpiresAt, security.TokenHash(rawRefresh))
	if err != nil {
		return TokenPair{}, fmt.Errorf("rotate refresh token: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return TokenPair{}, ErrInvalidRefresh
	}
	if err := tx.Commit(ctx); err != nil {
		return TokenPair{}, fmt.Errorf("commit refresh: %w", err)
	}
	return newPair, nil
}

func (s *Service) Logout(ctx context.Context, rawRefresh string) error {
	if rawRefresh == "" {
		return nil
	}
	_, err := s.db.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE refresh_token_hash = $1 AND revoked_at IS NULL
	`, security.TokenHash(rawRefresh))
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *Service) AuthenticateAccess(ctx context.Context, rawAccess string) (User, error) {
	if rawAccess == "" {
		return User{}, ErrInvalidCredentials
	}
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.access_token_hash = $1
		  AND s.access_expires_at > now()
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()
	`, security.TokenHash(rawAccess)).Scan(&user.ID, &user.Email, &user.DisplayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("authenticate access token: %w", err)
	}
	return user, nil
}

func newTokenPair() (TokenPair, []byte, []byte, error) {
	access, err := security.NewOpaqueToken()
	if err != nil {
		return TokenPair{}, nil, nil, err
	}
	refresh, err := security.NewOpaqueToken()
	if err != nil {
		return TokenPair{}, nil, nil, err
	}
	now := time.Now().UTC()
	pair := TokenPair{
		AccessToken: access, RefreshToken: refresh, TokenType: "Bearer",
		AccessExpiresAt: now.Add(accessLifetime), RefreshExpiresAt: now.Add(refreshLifetime),
	}
	return pair, security.TokenHash(access), security.TokenHash(refresh), nil
}

func insertSession(ctx context.Context, q interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, userID uuid.UUID, accessHash, refreshHash []byte, pair TokenPair) error {
	_, err := q.Exec(ctx, `
		INSERT INTO sessions(user_id, access_token_hash, access_expires_at, refresh_token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, accessHash, pair.AccessExpiresAt, refreshHash, pair.RefreshExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func validSlug(value string) bool {
	if len(value) < 3 || len(value) > 48 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
			return false
		}
	}
	return true
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
