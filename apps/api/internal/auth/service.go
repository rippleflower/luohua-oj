package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo       Repository
	sessionTTL time.Duration
	logger     *slog.Logger
	limiter    *loginLimiter
}

func NewService(repo Repository, sessionTTL time.Duration, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:       repo,
		sessionTTL: sessionTTL,
		logger:     logger,
		limiter:    newLoginLimiter(10*time.Minute, 8),
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput, ip string, userAgent string) (AuthResult, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Username = strings.TrimSpace(input.Username)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.Email == "" || input.Username == "" || input.DisplayName == "" {
		return AuthResult{}, errors.New("email, username, and displayName are required")
	}
	if len(input.Password) < 8 {
		return AuthResult{}, errors.New("password must be at least 8 characters")
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.repo.CreateUser(ctx, input.Email, input.Username, passwordHash, input.DisplayName)
	if err != nil {
		return AuthResult{}, err
	}
	return s.issueSession(ctx, user, ip, userAgent)
}

func (s *Service) Login(ctx context.Context, input LoginInput, userAgent string) (AuthResult, error) {
	identifier := strings.TrimSpace(strings.ToLower(input.Identifier))
	if identifier == "" || strings.TrimSpace(input.Password) == "" {
		return AuthResult{}, errors.New("identifier and password are required")
	}
	if !s.limiter.Allow(fmt.Sprintf("%s:%s", input.IP, identifier)) {
		return AuthResult{}, errors.New("too many login attempts")
	}

	stored, err := s.repo.FindUserByIdentifier(ctx, identifier)
	if err != nil {
		return AuthResult{}, err
	}
	if stored.Status != "" && stored.Status != "ACTIVE" {
		return AuthResult{}, errors.New("account is not active")
	}
	if err := verifyPassword(input.Password, stored.PasswordHash); err != nil {
		return AuthResult{}, errors.New("invalid credentials")
	}
	if passwordNeedsUpgrade(stored.PasswordHash) {
		passwordHash, err := hashPassword(input.Password)
		if err != nil {
			return AuthResult{}, err
		}
		if err := s.repo.UpdatePassword(ctx, stored.ID, passwordHash); err != nil {
			return AuthResult{}, err
		}
	}
	return s.issueSession(ctx, stored.User, input.IP, userAgent)
}

func (s *Service) issueSession(ctx context.Context, user User, ip string, userAgent string) (AuthResult, error) {
	token, hash, err := newSessionToken()
	if err != nil {
		return AuthResult{}, err
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	sessionID, err := s.repo.CreateSession(ctx, user.ID, hash, ip, userAgent, expiresAt)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User:         user,
		SessionID:    sessionID,
		SessionToken: token,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, sessionToken string) (AuthenticatedUser, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return AuthenticatedUser{}, errors.New("session token is required")
	}
	_, hash, err := sessionTokenAndHash(sessionToken)
	if err != nil {
		return AuthenticatedUser{}, err
	}
	user, err := s.repo.GetAuthenticatedUserBySessionHash(ctx, hash)
	if err != nil {
		return AuthenticatedUser{}, err
	}
	if err := s.repo.TouchSession(ctx, user.SessionID); err != nil {
		s.logger.WarnContext(ctx, "session touch failed", "event", "auth.session.touch_failed", "sessionId", user.SessionID.String(), "error", err)
	}
	return user, nil
}

func sessionTokenAndHash(sessionToken string) (string, string, error) {
	if _, err := base64.RawURLEncoding.DecodeString(sessionToken); err != nil {
		return "", "", errors.New("invalid session token")
	}
	hash := sha256.Sum256([]byte(sessionToken))
	return sessionToken, base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return nil
	}
	_, hash, err := sessionTokenAndHash(sessionToken)
	if err != nil {
		return nil
	}
	return s.repo.DeleteSessionByHash(ctx, hash)
}

func (s *Service) ListSessions(ctx context.Context, user AuthenticatedUser) ([]Session, error) {
	return s.repo.ListSessions(ctx, user.ID, user.SessionTokenHash)
}

func (s *Service) ChangePassword(ctx context.Context, input ChangePasswordInput, ip string) error {
	stored, err := s.repo.GetStoredUserByID(ctx, input.UserID)
	if err != nil {
		return err
	}
	if err := verifyPassword(input.CurrentPassword, stored.PasswordHash); err != nil {
		return errors.New("current password is invalid")
	}
	passwordHash, err := hashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(ctx, input.UserID, passwordHash); err != nil {
		return err
	}
	if err := s.repo.RevokeOtherSessions(ctx, input.UserID, input.SessionTokenHash); err != nil {
		return err
	}
	return nil
}

func (s *Service) RevokeSession(ctx context.Context, user AuthenticatedUser, sessionID uuid.UUID) error {
	return s.repo.RevokeSession(ctx, user.ID, sessionID)
}

func (s *Service) MeSummary(ctx context.Context, user AuthenticatedUser) (MeSummary, error) {
	return s.repo.GetMeSummary(ctx, user.ID)
}

func (s *Service) MeSettings(ctx context.Context, user AuthenticatedUser) (MeSettings, error) {
	return s.repo.GetMeSettings(ctx, user.ID, user.SessionTokenHash)
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (MeSettings, error) {
	return s.repo.UpdateProfile(ctx, input)
}

func (s *Service) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (MeSettings, error) {
	return s.repo.UpdatePreferences(ctx, input)
}

func (s *Service) AdminDashboard(ctx context.Context, actor AuthenticatedUser) (Dashboard, error) {
	return s.repo.GetAdminDashboard(ctx, actor.User)
}

func (s *Service) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	return s.repo.ListAdminUsers(ctx)
}

func (s *Service) GetAdminUser(ctx context.Context, targetID uuid.UUID) (AdminUserDetail, error) {
	return s.repo.GetAdminUser(ctx, targetID)
}

func (s *Service) GetAdminPermissions(ctx context.Context, targetID uuid.UUID) ([]PermissionKey, error) {
	return s.repo.GetAdminPermissions(ctx, targetID)
}

func (s *Service) UpdateAdminUser(ctx context.Context, input UpdateAdminUserInput) (AdminUserDetail, error) {
	return s.repo.UpdateAdminUser(ctx, input)
}

func (s *Service) SetUserRole(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, role Role, reason string, ip string) error {
	if role != RoleUser && role != RoleAdmin && role != RoleSuperAdmin {
		return errors.New("invalid role")
	}
	return s.repo.SetUserRole(ctx, actor, targetID, role, reason, ip)
}

func (s *Service) ReplacePermissions(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, permissions []PermissionKey, reason string, ip string) error {
	return s.repo.ReplacePermissions(ctx, actor, targetID, permissions, reason, ip)
}

func (s *Service) ListAdminProblems(ctx context.Context) ([]AdminProblemSummary, error) {
	return s.repo.ListAdminProblems(ctx)
}

func (s *Service) ListAdminContests(ctx context.Context) ([]AdminContestSummary, error) {
	return s.repo.ListAdminContests(ctx)
}

func (s *Service) ListAdminSubmissions(ctx context.Context) ([]AdminSubmissionSummary, error) {
	return s.repo.ListAdminSubmissions(ctx)
}

func (s *Service) ListAuditEvents(ctx context.Context) ([]AuditEvent, error) {
	return s.repo.ListAuditEvents(ctx)
}

func (s *Service) GetSystemSettings(ctx context.Context, sourceRoot string, redisAddr string) (SystemSettings, error) {
	return s.repo.GetSystemSettings(ctx, sourceRoot, redisAddr)
}

func (s *Service) UpdateSystemSettings(ctx context.Context, actor AuthenticatedUser, settings SystemSettings, ip string) (SystemSettings, error) {
	return s.repo.UpdateSystemSettings(ctx, actor, settings, ip)
}

func (s *Service) ListAnnouncements(ctx context.Context) ([]Announcement, error) {
	return s.repo.ListAnnouncements(ctx)
}

func (s *Service) CreateAnnouncement(ctx context.Context, actor AuthenticatedUser, title string, content string, status string, audience string, ip string) (Announcement, error) {
	return s.repo.CreateAnnouncement(ctx, actor, title, content, status, audience, ip)
}

func (s *Service) UpdateAnnouncement(ctx context.Context, actor AuthenticatedUser, announcementID uuid.UUID, title string, content string, status string, audience string, ip string) (Announcement, error) {
	return s.repo.UpdateAnnouncement(ctx, actor, announcementID, title, content, status, audience, ip)
}

type loginLimiter struct {
	mu        sync.Mutex
	window    time.Duration
	maxEvents int
	attempts  map[string][]time.Time
}

func newLoginLimiter(window time.Duration, maxEvents int) *loginLimiter {
	return &loginLimiter{
		window:    window,
		maxEvents: maxEvents,
		attempts:  make(map[string][]time.Time),
	}
}

func (l *loginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)
	events := l.attempts[key][:0]
	for _, event := range l.attempts[key] {
		if event.After(windowStart) {
			events = append(events, event)
		}
	}
	if len(events) >= l.maxEvents {
		l.attempts[key] = events
		return false
	}
	events = append(events, now)
	l.attempts[key] = events
	return true
}
