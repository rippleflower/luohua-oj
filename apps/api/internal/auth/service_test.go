package auth

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChangePasswordKeepsCurrentSessionAndRevokesOthers(t *testing.T) {
	userID := uuid.New()
	currentHash := "current-session-hash"
	storedHash, err := hashForTest("old-password")
	require.NoError(t, err)

	repo := &changePasswordRepo{
		stored: authStoredUser{
			ID:           userID,
			PasswordHash: storedHash,
		},
	}
	service := NewService(repo, time.Hour, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	err = service.ChangePassword(context.Background(), ChangePasswordInput{
		UserID:           userID,
		CurrentPassword:  "old-password",
		NewPassword:      "new-password",
		SessionTokenHash: currentHash,
	}, "127.0.0.1")

	require.NoError(t, err)
	require.Equal(t, userID, repo.updatedUserID)
	require.NotEmpty(t, repo.updatedPasswordHash)
	require.NotEqual(t, storedHash, repo.updatedPasswordHash)
	require.Equal(t, userID, repo.revokedUserID)
	require.Equal(t, currentHash, repo.keptHash)
	require.NoError(t, verifyHashForTest("new-password", repo.updatedPasswordHash))
}

type authStoredUser struct {
	ID           uuid.UUID
	PasswordHash string
}

type changePasswordRepo struct {
	stored              authStoredUser
	updatedUserID       uuid.UUID
	updatedPasswordHash string
	revokedUserID       uuid.UUID
	keptHash            string
}

func (c *changePasswordRepo) CreateUser(ctx context.Context, email string, username string, passwordHash string, displayName string) (User, error) {
	return User{}, nil
}

func (c *changePasswordRepo) FindUserByIdentifier(ctx context.Context, identifier string) (StoredUser, error) {
	return StoredUser{}, nil
}

func (c *changePasswordRepo) GetStoredUserByID(ctx context.Context, userID uuid.UUID) (StoredUser, error) {
	return StoredUser{
		User: User{ID: c.stored.ID},
		PasswordHash: c.stored.PasswordHash,
	}, nil
}

func (c *changePasswordRepo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, ip string, userAgent string, expiresAt time.Time) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (c *changePasswordRepo) GetAuthenticatedUserBySessionHash(ctx context.Context, tokenHash string) (AuthenticatedUser, error) {
	return AuthenticatedUser{}, nil
}

func (c *changePasswordRepo) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (c *changePasswordRepo) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	return nil
}

func (c *changePasswordRepo) ListSessions(ctx context.Context, userID uuid.UUID, currentHash string) ([]Session, error) {
	return nil, nil
}

func (c *changePasswordRepo) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	return nil
}

func (c *changePasswordRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	c.updatedUserID = userID
	c.updatedPasswordHash = passwordHash
	return nil
}

func (c *changePasswordRepo) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentHash string) error {
	c.revokedUserID = userID
	c.keptHash = currentHash
	return nil
}

func (c *changePasswordRepo) GetMeSummary(ctx context.Context, userID uuid.UUID) (MeSummary, error) {
	return MeSummary{}, nil
}

func (c *changePasswordRepo) GetMeSettings(ctx context.Context, userID uuid.UUID, currentHash string) (MeSettings, error) {
	return MeSettings{}, nil
}

func (c *changePasswordRepo) UpdateProfile(ctx context.Context, input UpdateProfileInput) (MeSettings, error) {
	return MeSettings{}, nil
}

func (c *changePasswordRepo) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (MeSettings, error) {
	return MeSettings{}, nil
}

func (c *changePasswordRepo) GetAdminDashboard(ctx context.Context, actor User) (Dashboard, error) {
	return Dashboard{}, nil
}

func (c *changePasswordRepo) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	return nil, nil
}

func (c *changePasswordRepo) GetAdminUser(ctx context.Context, targetID uuid.UUID) (AdminUserDetail, error) {
	return AdminUserDetail{}, nil
}

func (c *changePasswordRepo) GetAdminPermissions(ctx context.Context, targetID uuid.UUID) ([]PermissionKey, error) {
	return nil, nil
}

func (c *changePasswordRepo) UpdateAdminUser(ctx context.Context, input UpdateAdminUserInput) (AdminUserDetail, error) {
	return AdminUserDetail{}, nil
}

func (c *changePasswordRepo) SetUserRole(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, role Role, reason string, ip string) error {
	return nil
}

func (c *changePasswordRepo) ReplacePermissions(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, permissions []PermissionKey, reason string, ip string) error {
	return nil
}

func (c *changePasswordRepo) ListAdminProblems(ctx context.Context) ([]AdminProblemSummary, error) {
	return nil, nil
}

func (c *changePasswordRepo) ListAdminContests(ctx context.Context) ([]AdminContestSummary, error) {
	return nil, nil
}

func (c *changePasswordRepo) ListAdminSubmissions(ctx context.Context) ([]AdminSubmissionSummary, error) {
	return nil, nil
}

func (c *changePasswordRepo) ListAuditEvents(ctx context.Context) ([]AuditEvent, error) {
	return nil, nil
}

func (c *changePasswordRepo) GetSystemSettings(ctx context.Context, sourceRoot string, redisAddr string) (SystemSettings, error) {
	return SystemSettings{}, nil
}

func (c *changePasswordRepo) UpdateSystemSettings(ctx context.Context, actor AuthenticatedUser, settings SystemSettings, ip string) (SystemSettings, error) {
	return SystemSettings{}, nil
}

func (c *changePasswordRepo) ListAnnouncements(ctx context.Context) ([]Announcement, error) {
	return nil, nil
}

func (c *changePasswordRepo) CreateAnnouncement(ctx context.Context, actor AuthenticatedUser, title string, content string, status string, audience string, ip string) (Announcement, error) {
	return Announcement{}, nil
}

func (c *changePasswordRepo) UpdateAnnouncement(ctx context.Context, actor AuthenticatedUser, announcementID uuid.UUID, title string, content string, status string, audience string, ip string) (Announcement, error) {
	return Announcement{}, nil
}

func hashForTest(password string) (string, error) {
	return hashPassword(password)
}

func verifyHashForTest(password string, encoded string) error {
	return verifyPassword(password, encoded)
}
