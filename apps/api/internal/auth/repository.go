package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SQLRepository struct {
	db *pgxpool.Pool
}

func NewSQLRepository(db *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateUser(ctx context.Context, email string, username string, passwordHash string, displayName string) (User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)

	var user User
	err = tx.QueryRow(
		ctx,
		`INSERT INTO users (email, username, password_hash, role, status)
		 VALUES ($1, $2, $3, 'USER', 'ACTIVE')
		 RETURNING id, email, username, role::text, status`,
		email,
		username,
		passwordHash,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Role, &user.Status)
	if err != nil {
		return User{}, err
	}

	if _, err := tx.Exec(
		ctx,
		`INSERT INTO user_profiles (user_id, display_name, preferred_locale, preferred_language, bio, avatar_url)
		 VALUES ($1, $2, 'zh', 'CPP17', '', '')`,
		user.ID,
		displayName,
	); err != nil {
		return User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, err
	}
	user.DisplayName = displayName
	return user, nil
}

func (r *SQLRepository) FindUserByIdentifier(ctx context.Context, identifier string) (StoredUser, error) {
	return r.loadStoredUser(
		ctx,
		`SELECT
		   u.id,
		   u.email,
		   u.username,
		   u.password_hash,
		   u.role::text,
		   u.status,
		   COALESCE(up.display_name, u.username),
		   COALESCE(array_agg(apg.permission_key ORDER BY apg.permission_key) FILTER (WHERE apg.permission_key IS NOT NULL), '{}'::text[])
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN admin_permission_grants apg ON apg.user_id = u.id
		 WHERE lower(u.email) = $1 OR lower(u.username) = $1
		 GROUP BY u.id, up.display_name`,
		strings.ToLower(identifier),
	)
}

func (r *SQLRepository) GetStoredUserByID(ctx context.Context, userID uuid.UUID) (StoredUser, error) {
	return r.loadStoredUser(
		ctx,
		`SELECT
		   u.id,
		   u.email,
		   u.username,
		   u.password_hash,
		   u.role::text,
		   u.status,
		   COALESCE(up.display_name, u.username),
		   COALESCE(array_agg(apg.permission_key ORDER BY apg.permission_key) FILTER (WHERE apg.permission_key IS NOT NULL), '{}'::text[])
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN admin_permission_grants apg ON apg.user_id = u.id
		 WHERE u.id = $1
		 GROUP BY u.id, up.display_name`,
		userID,
	)
}

func (r *SQLRepository) loadStoredUser(ctx context.Context, query string, args ...any) (StoredUser, error) {
	var (
		stored      StoredUser
		permissions []string
	)
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&stored.ID,
		&stored.Email,
		&stored.Username,
		&stored.PasswordHash,
		&stored.Role,
		&stored.Status,
		&stored.DisplayName,
		&permissions,
	)
	if err != nil {
		return StoredUser{}, err
	}
	stored.Permissions = permissionKeysFromStrings(permissions)
	return stored, nil
}

func (r *SQLRepository) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, ip string, userAgent string, expiresAt time.Time) (uuid.UUID, error) {
	var sessionID uuid.UUID
	err := r.db.QueryRow(
		ctx,
		`INSERT INTO user_sessions (user_id, session_token_hash, ip, user_agent, expires_at, last_seen_at)
		 VALUES ($1, $2, $3, $4, $5, now())
		 RETURNING id`,
		userID,
		tokenHash,
		ip,
		userAgent,
		expiresAt,
	).Scan(&sessionID)
	return sessionID, err
}

func (r *SQLRepository) GetAuthenticatedUserBySessionHash(ctx context.Context, tokenHash string) (AuthenticatedUser, error) {
	var (
		user        AuthenticatedUser
		permissions []string
	)
	err := r.db.QueryRow(
		ctx,
		`SELECT
		   u.id,
		   u.email,
		   u.username,
		   u.role::text,
		   u.status,
		   COALESCE(up.display_name, u.username),
		   us.id,
		   COALESCE(array_agg(apg.permission_key ORDER BY apg.permission_key) FILTER (WHERE apg.permission_key IS NOT NULL), '{}'::text[])
		 FROM user_sessions us
		 INNER JOIN users u ON u.id = us.user_id
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN admin_permission_grants apg ON apg.user_id = u.id
		 WHERE us.session_token_hash = $1
		   AND us.expires_at > now()
		 GROUP BY u.id, up.display_name, us.id`,
		tokenHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Role,
		&user.Status,
		&user.DisplayName,
		&user.SessionID,
		&permissions,
	)
	if err != nil {
		return AuthenticatedUser{}, err
	}
	user.SessionTokenHash = tokenHash
	user.Permissions = permissionKeysFromStrings(permissions)
	return user, nil
}

func (r *SQLRepository) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE user_sessions SET last_seen_at = now() WHERE id = $1`, sessionID)
	return err
}

func (r *SQLRepository) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_sessions WHERE session_token_hash = $1`, tokenHash)
	return err
}

func (r *SQLRepository) ListSessions(ctx context.Context, userID uuid.UUID, currentHash string) ([]Session, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, ip, user_agent, expires_at, last_seen_at, created_at, session_token_hash
		 FROM user_sessions
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		var (
			item      Session
			tokenHash string
		)
		if err := rows.Scan(&item.ID, &item.IP, &item.UserAgent, &item.ExpiresAt, &item.LastSeenAt, &item.CreatedAt, &tokenHash); err != nil {
			return nil, err
		}
		item.Current = tokenHash == currentHash
		sessions = append(sessions, item)
	}
	return sessions, rows.Err()
}

func (r *SQLRepository) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1 AND id = $2`, userID, sessionID)
	return err
}

func (r *SQLRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, userID, passwordHash)
	return err
}

func (r *SQLRepository) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentHash string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1 AND session_token_hash <> $2`, userID, currentHash)
	return err
}

func (r *SQLRepository) GetMeSummary(ctx context.Context, userID uuid.UUID) (MeSummary, error) {
	user, err := r.loadUser(ctx, userID)
	if err != nil {
		return MeSummary{}, err
	}

	var summary MeSummary
	summary.User = user
	if err := r.db.QueryRow(
		ctx,
		`SELECT
		   COALESCE(solved_count, 0),
		   COALESCE(submission_count, 0),
		   COALESCE(accepted_count, 0),
		   last_active_at
		 FROM user_stats
		 WHERE user_id = $1`,
		userID,
	).Scan(&summary.SolvedCount, &summary.SubmissionCount, &summary.AcceptedCount, &summary.LastActiveAt); err != nil && err != pgx.ErrNoRows {
		return MeSummary{}, err
	}

	submissionRows, err := r.db.Query(
		ctx,
		`SELECT submission_id, status::text, created_at, problem_id, problem_json
		 FROM submission_summaries
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 5`,
		userID,
	)
	if err != nil {
		return MeSummary{}, err
	}
	defer submissionRows.Close()
	for submissionRows.Next() {
		var (
			item        MeSubmission
			problemJSON []byte
		)
		if err := submissionRows.Scan(&item.ID, &item.Status, &item.CreatedAt, &item.ProblemID, &problemJSON); err != nil {
			return MeSummary{}, err
		}
		var problem struct {
			Slug  string `json:"slug"`
			Title string `json:"title"`
		}
		if err := json.Unmarshal(problemJSON, &problem); err != nil {
			return MeSummary{}, err
		}
		item.ProblemSlug = problem.Slug
		item.ProblemTitle = problem.Title
		summary.RecentSubmissions = append(summary.RecentSubmissions, item)
	}
	if err := submissionRows.Err(); err != nil {
		return MeSummary{}, err
	}

	contestRows, err := r.db.Query(
		ctx,
		`SELECT contest_id, slug, title, status, starts_at, ends_at
		 FROM contest_public_summaries
		 WHERE status IN ('RUNNING', 'UPCOMING')
		 ORDER BY starts_at ASC
		 LIMIT 3`,
	)
	if err != nil {
		return MeSummary{}, err
	}
	defer contestRows.Close()
	for contestRows.Next() {
		var item MeContest
		if err := contestRows.Scan(&item.ID, &item.Slug, &item.Title, &item.Status, &item.StartsAt, &item.EndsAt); err != nil {
			return MeSummary{}, err
		}
		summary.Contests = append(summary.Contests, item)
	}
	return summary, contestRows.Err()
}

func (r *SQLRepository) GetMeSettings(ctx context.Context, userID uuid.UUID, currentHash string) (MeSettings, error) {
	user, err := r.loadUser(ctx, userID)
	if err != nil {
		return MeSettings{}, err
	}
	var settings MeSettings
	settings.User = user
	err = r.db.QueryRow(
		ctx,
		`SELECT u.username, u.email, COALESCE(up.display_name, u.username), COALESCE(up.bio, ''), COALESCE(up.avatar_url, ''), COALESCE(up.preferred_locale, 'zh'), COALESCE(up.preferred_language, 'CPP17')
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 WHERE u.id = $1`,
		userID,
	).Scan(
		&settings.Username,
		&settings.Email,
		&settings.DisplayName,
		&settings.Bio,
		&settings.AvatarURL,
		&settings.PreferredLocale,
		&settings.PreferredLanguage,
	)
	if err != nil {
		return MeSettings{}, err
	}
	sessions, err := r.ListSessions(ctx, userID, currentHash)
	if err != nil {
		return MeSettings{}, err
	}
	settings.Sessions = sessions
	return settings, nil
}

func (r *SQLRepository) UpdateProfile(ctx context.Context, input UpdateProfileInput) (MeSettings, error) {
	_, err := r.db.Exec(
		ctx,
		`UPDATE user_profiles
		 SET display_name = $2, bio = $3, avatar_url = $4, updated_at = now()
		 WHERE user_id = $1`,
		input.UserID,
		input.DisplayName,
		input.Bio,
		input.AvatarURL,
	)
	if err != nil {
		return MeSettings{}, err
	}
	return r.GetMeSettings(ctx, input.UserID, "")
}

func (r *SQLRepository) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (MeSettings, error) {
	_, err := r.db.Exec(
		ctx,
		`UPDATE user_profiles
		 SET preferred_locale = $2, preferred_language = $3, updated_at = now()
		 WHERE user_id = $1`,
		input.UserID,
		input.PreferredLocale,
		input.PreferredLanguage,
	)
	if err != nil {
		return MeSettings{}, err
	}
	return r.GetMeSettings(ctx, input.UserID, "")
}

func (r *SQLRepository) GetAdminDashboard(ctx context.Context, actor User) (Dashboard, error) {
	dashboard := Dashboard{Actor: actor}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&dashboard.TotalUsers); err != nil {
		return Dashboard{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM problems WHERE is_published = true`).Scan(&dashboard.ActiveProblems); err != nil {
		return Dashboard{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contests WHERE status = 'RUNNING'`).Scan(&dashboard.RunningContests); err != nil {
		return Dashboard{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM submissions WHERE created_at >= now() - interval '24 hours'`).Scan(&dashboard.Submissions24h); err != nil {
		return Dashboard{}, err
	}
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM submissions WHERE status IN ('PENDING', 'RUNNING')`).Scan(&dashboard.PendingSubmissions); err != nil {
		return Dashboard{}, err
	}
	if dashboard.PendingSubmissions > 0 {
		dashboard.Alerts = append(dashboard.Alerts, DashboardAlert{
			Title:       "Judge queue backlog",
			Tone:        "warning",
			Description: fmt.Sprintf("%d submissions are still pending or running.", dashboard.PendingSubmissions),
		})
	}
	if dashboard.RunningContests == 0 {
		dashboard.Alerts = append(dashboard.Alerts, DashboardAlert{
			Title:       "No running contest",
			Tone:        "info",
			Description: "Contest workspace is currently idle.",
		})
	}
	return dashboard, nil
}

func (r *SQLRepository) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT u.id, u.email, u.username, u.role::text, u.status, COALESCE(up.display_name, u.username), u.created_at, u.updated_at
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 ORDER BY u.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AdminUserSummary, 0)
	for rows.Next() {
		var item AdminUserSummary
		if err := rows.Scan(&item.ID, &item.Email, &item.Username, &item.Role, &item.Status, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetAdminUser(ctx context.Context, targetID uuid.UUID) (AdminUserDetail, error) {
	var (
		item        AdminUserDetail
		permissions []string
	)
	err := r.db.QueryRow(
		ctx,
		`SELECT
		   u.id, u.email, u.username, u.role::text, u.status, COALESCE(up.display_name, u.username), u.created_at, u.updated_at,
		   COALESCE(up.bio, ''), COALESCE(up.avatar_url, ''), COALESCE(up.preferred_locale, 'zh'), COALESCE(up.preferred_language, 'CPP17'),
		   COALESCE(us.solved_count, 0), COALESCE(us.submission_count, 0), COALESCE(us.accepted_count, 0), us.last_active_at,
		   COALESCE(array_agg(apg.permission_key ORDER BY apg.permission_key) FILTER (WHERE apg.permission_key IS NOT NULL), '{}'::text[])
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN user_stats us ON us.user_id = u.id
		 LEFT JOIN admin_permission_grants apg ON apg.user_id = u.id
		 WHERE u.id = $1
		 GROUP BY u.id, up.user_id, us.user_id`,
		targetID,
	).Scan(
		&item.ID, &item.Email, &item.Username, &item.Role, &item.Status, &item.DisplayName, &item.CreatedAt, &item.UpdatedAt,
		&item.Bio, &item.AvatarURL, &item.PreferredLocale, &item.PreferredLanguage,
		&item.SolvedCount, &item.SubmissionCount, &item.AcceptedCount, &item.LastActiveAt,
		&permissions,
	)
	if err != nil {
		return AdminUserDetail{}, err
	}
	item.Permissions = permissionKeysFromStrings(permissions)
	return item, nil
}

func (r *SQLRepository) GetAdminPermissions(ctx context.Context, targetID uuid.UUID) ([]PermissionKey, error) {
	rows, err := r.db.Query(ctx, `SELECT permission_key FROM admin_permission_grants WHERE user_id = $1 ORDER BY permission_key`, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	permissions := make([]PermissionKey, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		permissions = append(permissions, PermissionKey(key))
	}
	return permissions, rows.Err()
}

func (r *SQLRepository) UpdateAdminUser(ctx context.Context, input UpdateAdminUserInput) (AdminUserDetail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return AdminUserDetail{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(
		ctx,
		`UPDATE users SET status = $2, updated_at = now() WHERE id = $1`,
		input.TargetID,
		input.Status,
	); err != nil {
		return AdminUserDetail{}, err
	}
	if _, err := tx.Exec(
		ctx,
		`UPDATE user_profiles
		 SET display_name = $2, bio = $3, avatar_url = $4, updated_at = now()
		 WHERE user_id = $1`,
		input.TargetID,
		input.DisplayName,
		input.Bio,
		input.AvatarURL,
	); err != nil {
		return AdminUserDetail{}, err
	}
	if err := recordAudit(ctx, tx, input.Actor, "admin.user.updated", "user", input.TargetID.String(), map[string]any{
		"status":      input.Status,
		"displayName": input.DisplayName,
	}, input.Reason, input.IP); err != nil {
		return AdminUserDetail{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminUserDetail{}, err
	}
	return r.GetAdminUser(ctx, input.TargetID)
}

func (r *SQLRepository) SetUserRole(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, role Role, reason string, ip string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE users SET role = $2, updated_at = now() WHERE id = $1`, targetID, role); err != nil {
		return err
	}
	if err := recordAudit(ctx, tx, actor, "admin.user.role_changed", "user", targetID.String(), map[string]any{"role": role}, reason, ip); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SQLRepository) ReplacePermissions(ctx context.Context, actor AuthenticatedUser, targetID uuid.UUID, permissions []PermissionKey, reason string, ip string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM admin_permission_grants WHERE user_id = $1`, targetID); err != nil {
		return err
	}
	for _, permission := range permissions {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO admin_permission_grants (user_id, permission_key, granted_by) VALUES ($1, $2, $3)`,
			targetID,
			string(permission),
			actor.ID,
		); err != nil {
			return err
		}
	}
	if err := recordAudit(ctx, tx, actor, "admin.user.permissions_replaced", "user", targetID.String(), map[string]any{"permissions": permissions}, reason, ip); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SQLRepository) ListAdminProblems(ctx context.Context) ([]AdminProblemSummary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   p.id, p.slug, p.title, p.difficulty::text, p.time_limit_ms, p.memory_limit_kb,
		   COALESCE(pv.status, CASE WHEN p.is_published THEN 'PUBLISHED' ELSE 'DRAFT' END),
		   COALESCE(pv.version_no, 1),
		   p.is_published,
		   COALESCE(ps.submission_count, 0),
		   COALESCE(ps.accepted_rate::float8, 0),
		   p.updated_at
		 FROM problems p
		 LEFT JOIN problem_versions pv ON pv.problem_id = p.id AND pv.id = p.current_public_version_id
		 LEFT JOIN problem_stats ps ON ps.problem_id = p.id
		 ORDER BY p.updated_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AdminProblemSummary, 0)
	for rows.Next() {
		var item AdminProblemSummary
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Difficulty, &item.TimeLimitMs, &item.MemoryLimitKb, &item.Status, &item.CurrentVersionNo, &item.IsPublished, &item.SubmissionCount, &item.AcceptedRate, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListAdminContests(ctx context.Context) ([]AdminContestSummary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
		   c.id, c.slug, c.title, c.status, c.starts_at, c.ends_at, c.participant_count,
		   COALESCE((SELECT COUNT(*) FROM contest_problem_links cpl WHERE cpl.contest_id = c.id), 0),
		   COALESCE((SELECT MAX(snapshot_no) FROM contest_snapshots cs WHERE cs.contest_id = c.id), 0),
		   c.updated_at
		 FROM contests c
		 ORDER BY c.starts_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AdminContestSummary, 0)
	for rows.Next() {
		var item AdminContestSummary
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Status, &item.StartsAt, &item.EndsAt, &item.ParticipantCount, &item.ProblemCount, &item.LatestSnapshotNo, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListAdminSubmissions(ctx context.Context) ([]AdminSubmissionSummary, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT ss.submission_id, ss.username, ss.problem_json, ss.language::text, ss.status::text, ss.created_at
		 FROM submission_summaries ss
		 ORDER BY ss.created_at DESC
		 LIMIT 50`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AdminSubmissionSummary, 0)
	for rows.Next() {
		var (
			item        AdminSubmissionSummary
			problemJSON []byte
		)
		if err := rows.Scan(&item.ID, &item.Username, &problemJSON, &item.Language, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		var problem struct {
			Title string `json:"title"`
		}
		if err := json.Unmarshal(problemJSON, &problem); err != nil {
			return nil, err
		}
		item.ProblemTitle = problem.Title
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) ListAuditEvents(ctx context.Context) ([]AuditEvent, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT al.id, al.actor_user_id, u.username, al.actor_role::text, al.action, al.target_type, al.target_id, al.diff_json, al.reason, al.ip, al.created_at
		 FROM audit_logs al
		 LEFT JOIN users u ON u.id = al.actor_user_id
		 ORDER BY al.created_at DESC
		 LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AuditEvent, 0)
	for rows.Next() {
		var (
			item      AuditEvent
			actorRole *string
			diffJSON  []byte
		)
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.ActorUsername, &actorRole, &item.Action, &item.TargetType, &item.TargetID, &diffJSON, &item.Reason, &item.IP, &item.CreatedAt); err != nil {
			return nil, err
		}
		if actorRole != nil {
			role := Role(*actorRole)
			item.ActorRole = &role
		}
		if err := json.Unmarshal(diffJSON, &item.Diff); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) GetSystemSettings(ctx context.Context, sourceRoot string, redisAddr string) (SystemSettings, error) {
	var (
		settingsJSON []byte
		settings     SystemSettings
	)
	err := r.db.QueryRow(ctx, `SELECT settings_json, updated_at FROM system_settings WHERE singleton = true`).Scan(&settingsJSON, &settings.UpdatedAt)
	if err != nil {
		return SystemSettings{}, err
	}
	var stored struct {
		RegistrationEnabled bool   `json:"registrationEnabled"`
		JudgeQueuePaused    bool   `json:"judgeQueuePaused"`
		StorageMode         string `json:"storageMode"`
	}
	if err := json.Unmarshal(settingsJSON, &stored); err != nil {
		return SystemSettings{}, err
	}
	settings.RegistrationEnabled = stored.RegistrationEnabled
	settings.JudgeQueuePaused = stored.JudgeQueuePaused
	settings.StorageMode = stored.StorageMode
	settings.SourceRoot = sourceRoot
	settings.RedisAddr = redisAddr
	return settings, nil
}

func (r *SQLRepository) UpdateSystemSettings(ctx context.Context, actor AuthenticatedUser, settings SystemSettings, ip string) (SystemSettings, error) {
	payload, err := json.Marshal(map[string]any{
		"registrationEnabled": settings.RegistrationEnabled,
		"judgeQueuePaused":    settings.JudgeQueuePaused,
		"storageMode":         settings.StorageMode,
	})
	if err != nil {
		return SystemSettings{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return SystemSettings{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(
		ctx,
		`UPDATE system_settings SET settings_json = $1, updated_by = $2, updated_at = now() WHERE singleton = true`,
		payload,
		actor.ID,
	); err != nil {
		return SystemSettings{}, err
	}
	if err := recordAudit(ctx, tx, actor, "admin.system.settings_updated", "system_settings", "singleton", map[string]any{
		"registrationEnabled": settings.RegistrationEnabled,
		"judgeQueuePaused":    settings.JudgeQueuePaused,
		"storageMode":         settings.StorageMode,
	}, "system settings updated", ip); err != nil {
		return SystemSettings{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SystemSettings{}, err
	}
	return settings, nil
}

func (r *SQLRepository) ListAnnouncements(ctx context.Context) ([]Announcement, error) {
	rows, err := r.db.Query(ctx, `SELECT id, title, content, status, audience, created_at, updated_at FROM announcements ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Announcement, 0)
	for rows.Next() {
		var item Announcement
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.Status, &item.Audience, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) CreateAnnouncement(ctx context.Context, actor AuthenticatedUser, title string, content string, status string, audience string, ip string) (Announcement, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Announcement{}, err
	}
	defer tx.Rollback(ctx)
	var item Announcement
	err = tx.QueryRow(
		ctx,
		`INSERT INTO announcements (title, content, status, audience, created_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, title, content, status, audience, created_at, updated_at`,
		title, content, status, audience, actor.ID,
	).Scan(&item.ID, &item.Title, &item.Content, &item.Status, &item.Audience, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Announcement{}, err
	}
	if err := recordAudit(ctx, tx, actor, "admin.announcement.created", "announcement", item.ID.String(), map[string]any{"status": status, "audience": audience}, title, ip); err != nil {
		return Announcement{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Announcement{}, err
	}
	return item, nil
}

func (r *SQLRepository) UpdateAnnouncement(ctx context.Context, actor AuthenticatedUser, announcementID uuid.UUID, title string, content string, status string, audience string, ip string) (Announcement, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Announcement{}, err
	}
	defer tx.Rollback(ctx)
	var item Announcement
	err = tx.QueryRow(
		ctx,
		`UPDATE announcements
		 SET title = $2, content = $3, status = $4, audience = $5, updated_at = now()
		 WHERE id = $1
		 RETURNING id, title, content, status, audience, created_at, updated_at`,
		announcementID, title, content, status, audience,
	).Scan(&item.ID, &item.Title, &item.Content, &item.Status, &item.Audience, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Announcement{}, err
	}
	if err := recordAudit(ctx, tx, actor, "admin.announcement.updated", "announcement", item.ID.String(), map[string]any{"status": status, "audience": audience}, title, ip); err != nil {
		return Announcement{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Announcement{}, err
	}
	return item, nil
}

func (r *SQLRepository) loadUser(ctx context.Context, userID uuid.UUID) (User, error) {
	var (
		user        User
		permissions []string
	)
	err := r.db.QueryRow(
		ctx,
		`SELECT u.id, u.email, u.username, u.role::text, u.status, COALESCE(up.display_name, u.username), COALESCE(array_agg(apg.permission_key ORDER BY apg.permission_key) FILTER (WHERE apg.permission_key IS NOT NULL), '{}'::text[])
		 FROM users u
		 LEFT JOIN user_profiles up ON up.user_id = u.id
		 LEFT JOIN admin_permission_grants apg ON apg.user_id = u.id
		 WHERE u.id = $1
		 GROUP BY u.id, up.display_name`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Role, &user.Status, &user.DisplayName, &permissions)
	if err != nil {
		return User{}, err
	}
	user.Permissions = permissionKeysFromStrings(permissions)
	return user, nil
}

func permissionKeysFromStrings(values []string) []PermissionKey {
	result := make([]PermissionKey, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		result = append(result, PermissionKey(value))
	}
	return result
}

type auditExec interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func recordAudit(ctx context.Context, exec auditExec, actor AuthenticatedUser, action string, targetType string, targetID string, diff map[string]any, reason string, ip string) error {
	diffJSON, err := json.Marshal(diff)
	if err != nil {
		return err
	}
	_, err = exec.Exec(
		ctx,
		`INSERT INTO audit_logs (actor_user_id, actor_role, action, target_type, target_id, diff_json, reason, ip)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		actor.ID,
		actor.Role,
		action,
		targetType,
		targetID,
		diffJSON,
		reason,
		ip,
	)
	return err
}
