package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser       Role = "USER"
	RoleAdmin      Role = "ADMIN"
	RoleSuperAdmin Role = "SUPER_ADMIN"
)

type PermissionKey string

const (
	PermissionDashboardView   PermissionKey = "dashboard.view"
	PermissionUsersView       PermissionKey = "users.view"
	PermissionUsersEdit       PermissionKey = "users.edit"
	PermissionUsersRoles      PermissionKey = "users.roles"
	PermissionProblemsView    PermissionKey = "problems.view"
	PermissionProblemsEdit    PermissionKey = "problems.edit"
	PermissionProblemsPublish PermissionKey = "problems.publish"
	PermissionContestsView    PermissionKey = "contests.view"
	PermissionContestsEdit    PermissionKey = "contests.edit"
	PermissionContestsPublish PermissionKey = "contests.publish"
	PermissionSubmissionsView PermissionKey = "submissions.view"
	PermissionSubmissionsRedo PermissionKey = "submissions.rejudge"
	PermissionAnnView         PermissionKey = "announcements.view"
	PermissionAnnEdit         PermissionKey = "announcements.edit"
	PermissionSystemView      PermissionKey = "system.view"
	PermissionSystemEdit      PermissionKey = "system.edit"
	PermissionAuditView       PermissionKey = "audit.view"
)

var AllPermissions = []PermissionKey{
	PermissionDashboardView,
	PermissionUsersView,
	PermissionUsersEdit,
	PermissionUsersRoles,
	PermissionProblemsView,
	PermissionProblemsEdit,
	PermissionProblemsPublish,
	PermissionContestsView,
	PermissionContestsEdit,
	PermissionContestsPublish,
	PermissionSubmissionsView,
	PermissionSubmissionsRedo,
	PermissionAnnView,
	PermissionAnnEdit,
	PermissionSystemView,
	PermissionSystemEdit,
	PermissionAuditView,
}

type User struct {
	ID          uuid.UUID
	Email       string
	Username    string
	Role        Role
	Permissions []PermissionKey
	DisplayName string
	Status      string
}

func (u User) HasPermission(permission PermissionKey) bool {
	if u.Role == RoleSuperAdmin {
		return true
	}
	for _, granted := range u.Permissions {
		if granted == permission {
			return true
		}
	}
	return false
}

type Session struct {
	ID         uuid.UUID
	Current    bool
	IP         string
	UserAgent  string
	ExpiresAt  time.Time
	LastSeenAt *time.Time
	CreatedAt  time.Time
}

type AuthenticatedUser struct {
	User
	SessionID        uuid.UUID
	SessionTokenHash string
}

type AuthResult struct {
	User         User
	SessionID    uuid.UUID
	SessionToken string
	ExpiresAt    time.Time
}

type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Identifier string
	Password   string
	IP         string
}

type ChangePasswordInput struct {
	UserID           uuid.UUID
	CurrentPassword  string
	NewPassword      string
	SessionTokenHash string
}

type UpdateProfileInput struct {
	UserID      uuid.UUID
	DisplayName string
	Bio         string
	AvatarURL   string
	IP          string
}

type UpdatePreferencesInput struct {
	UserID            uuid.UUID
	PreferredLocale   string
	PreferredLanguage string
	IP                string
}

type MeSummary struct {
	User              User
	SolvedCount       int
	SubmissionCount   int
	AcceptedCount     int
	LastActiveAt      *time.Time
	RecentSubmissions []MeSubmission
	Contests          []MeContest
}

type MeSubmission struct {
	ID           uuid.UUID
	Status       string
	CreatedAt    time.Time
	ProblemID    uuid.UUID
	ProblemSlug  string
	ProblemTitle string
}

type MeContest struct {
	ID       uuid.UUID
	Slug     string
	Title    string
	Status   string
	StartsAt time.Time
	EndsAt   time.Time
}

type MeSettings struct {
	User              User
	Username          string
	Email             string
	DisplayName       string
	Bio               string
	AvatarURL         string
	PreferredLocale   string
	PreferredLanguage string
	Sessions          []Session
}

type AdminUserSummary struct {
	User
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AdminUserDetail struct {
	AdminUserSummary
	Bio               string
	AvatarURL         string
	PreferredLocale   string
	PreferredLanguage string
	SolvedCount       int
	SubmissionCount   int
	AcceptedCount     int
	LastActiveAt      *time.Time
}

type UpdateAdminUserInput struct {
	Actor       AuthenticatedUser
	TargetID    uuid.UUID
	Status      string
	DisplayName string
	Bio         string
	AvatarURL   string
	Reason      string
	IP          string
}

type AdminProblemSummary struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	Difficulty       string
	TimeLimitMs      int
	MemoryLimitKb    int
	Status           string
	CurrentVersionNo int
	IsPublished      bool
	SubmissionCount  int
	AcceptedRate     float64
	UpdatedAt        time.Time
}

type AdminContestSummary struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	Status           string
	StartsAt         time.Time
	EndsAt           time.Time
	ParticipantCount int
	ProblemCount     int
	LatestSnapshotNo int
	UpdatedAt        time.Time
}

type AdminSubmissionSummary struct {
	ID           uuid.UUID
	Username     string
	ProblemTitle string
	Language     string
	Status       string
	CreatedAt    time.Time
}

type AuditEvent struct {
	ID            uuid.UUID
	ActorUserID   *uuid.UUID
	ActorUsername *string
	ActorRole     *Role
	Action        string
	TargetType    string
	TargetID      string
	Diff          map[string]any
	Reason        string
	IP            string
	CreatedAt     time.Time
}

type Dashboard struct {
	Actor              User
	TotalUsers         int
	ActiveProblems     int
	RunningContests    int
	Submissions24h     int
	PendingSubmissions int
	Alerts             []DashboardAlert
}

type DashboardAlert struct {
	Title       string
	Tone        string
	Description string
}

type SystemSettings struct {
	RegistrationEnabled bool
	JudgeQueuePaused    bool
	StorageMode         string
	SourceRoot          string
	RedisAddr           string
	UpdatedAt           time.Time
}

type Announcement struct {
	ID        uuid.UUID
	Title     string
	Content   string
	Status    string
	Audience  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	CreateUser(ctx context.Context, email string, username string, passwordHash string, displayName string) (User, error)
	FindUserByIdentifier(ctx context.Context, identifier string) (StoredUser, error)
	GetStoredUserByID(ctx context.Context, userID uuid.UUID) (StoredUser, error)
	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, ip string, userAgent string, expiresAt time.Time) (uuid.UUID, error)
	GetAuthenticatedUserBySessionHash(ctx context.Context, tokenHash string) (AuthenticatedUser, error)
	TouchSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSessionByHash(ctx context.Context, tokenHash string) error
	ListSessions(ctx context.Context, userID uuid.UUID, currentHash string) ([]Session, error)
	RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentHash string) error
	GetMeSummary(ctx context.Context, userID uuid.UUID) (MeSummary, error)
	GetMeSettings(ctx context.Context, userID uuid.UUID, currentHash string) (MeSettings, error)
	UpdateProfile(ctx context.Context, input UpdateProfileInput) (MeSettings, error)
	UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (MeSettings, error)
	GetAdminDashboard(ctx context.Context, actor User) (Dashboard, error)
	ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error)
	GetAdminUser(ctx context.Context, targetID uuid.UUID) (AdminUserDetail, error)
	GetAdminPermissions(ctx context.Context, targetID uuid.UUID) ([]PermissionKey, error)
	UpdateAdminUser(ctx context.Context, input UpdateAdminUserInput) (AdminUserDetail, error)
	SetUserRole(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, role Role, reason string, ip string) error
	ReplacePermissions(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, permissions []PermissionKey, reason string, ip string) error
	ListAdminProblems(ctx context.Context) ([]AdminProblemSummary, error)
	ListAdminContests(ctx context.Context) ([]AdminContestSummary, error)
	ListAdminSubmissions(ctx context.Context) ([]AdminSubmissionSummary, error)
	ListAuditEvents(ctx context.Context) ([]AuditEvent, error)
	GetSystemSettings(ctx context.Context, sourceRoot string, redisAddr string) (SystemSettings, error)
	UpdateSystemSettings(ctx context.Context, actor AuthenticatedUser, settings SystemSettings, ip string) (SystemSettings, error)
	ListAnnouncements(ctx context.Context) ([]Announcement, error)
	CreateAnnouncement(ctx context.Context, actor AuthenticatedUser, title string, content string, status string, audience string, ip string) (Announcement, error)
	UpdateAnnouncement(ctx context.Context, actor AuthenticatedUser, announcementID uuid.UUID, title string, content string, status string, audience string, ip string) (Announcement, error)
}

type StoredUser struct {
	User
	PasswordHash string
}
