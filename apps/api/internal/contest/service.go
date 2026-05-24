package contest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
)

type Summary struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	Status           string
	StartsAt         time.Time
	EndsAt           time.Time
	DurationLabel    string
	ProblemCount     int
	ParticipantCount int
	Blurb            string
}

type RecentSubmission struct {
	ID          string `json:"id"`
	ProblemCode string `json:"problemCode"`
	Status      string `json:"status"`
	At          string `json:"at"`
}

type ProblemSnapshot struct {
	Code       string `json:"code"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Status     string `json:"status"`
	FirstSolve string `json:"firstSolve,omitempty"`
}

type Detail struct {
	Summary
	RankSummary       string
	Remaining         string
	RecentSubmissions []RecentSubmission
	Problems          []ProblemSnapshot
	UpdatedAt         time.Time
}

type AdminProblemBinding struct {
	ProblemID    uuid.UUID
	ProblemSlug  string
	ProblemTitle string
	Code         string
	Position     int
}

type AdminSnapshotSummary struct {
	ID           uuid.UUID
	SnapshotNo   int
	FrozenAt     time.Time
	ProblemCount int
}

type AdminContest struct {
	ID               uuid.UUID
	Slug             string
	Title            string
	Description      string
	Status           string
	StartsAt         time.Time
	EndsAt           time.Time
	ParticipantCount int
	ProblemCount     int
	LatestSnapshotNo int
	UpdatedAt        time.Time
	Problems         []AdminProblemBinding
	Snapshots        []AdminSnapshotSummary
}

type CreateAdminInput struct {
	Slug        string
	Title       string
	Description string
	Status      string
	StartsAt    time.Time
	EndsAt      time.Time
	Reason      string
	IP          string
}

type UpdateAdminInput struct {
	ContestID   uuid.UUID
	Slug        string
	Title       string
	Description string
	Status      string
	StartsAt    time.Time
	EndsAt      time.Time
	Reason      string
	IP          string
}

type ReplaceProblemsInput struct {
	ContestID uuid.UUID
	Problems  []AdminProblemBindingInput
	Reason    string
	IP        string
}

type AdminProblemBindingInput struct {
	ProblemID uuid.UUID
	Code      string
	Position  int
}

type FreezeAdminInput struct {
	ContestID uuid.UUID
	Reason    string
	IP        string
}

type Reader interface {
	List(ctx context.Context) ([]Summary, error)
	GetBySlug(ctx context.Context, slug string) (Detail, error)
}

type AdminManager interface {
	ListAdmin(ctx context.Context) ([]AdminContest, error)
	CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminContest, error)
	UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminContest, error)
	ReplaceProblemsAdmin(ctx context.Context, actor auth.AuthenticatedUser, input ReplaceProblemsInput) (AdminContest, error)
	FreezeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input FreezeAdminInput) (AdminContest, error)
}

type Repository interface {
	ListContests(ctx context.Context) ([]Summary, error)
	GetContestBySlug(ctx context.Context, slug string) (Detail, error)
	ListAdminContests(ctx context.Context) ([]AdminContest, error)
	CreateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminContest, error)
	UpdateAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminContest, error)
	ReplaceAdminContestProblems(ctx context.Context, actor auth.AuthenticatedUser, input ReplaceProblemsInput) (AdminContest, error)
	FreezeAdminContest(ctx context.Context, actor auth.AuthenticatedUser, input FreezeAdminInput) (AdminContest, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Summary, error) {
	return s.repo.ListContests(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Detail, error) {
	if strings.TrimSpace(slug) == "" {
		return Detail{}, errors.New("slug is required")
	}
	return s.repo.GetContestBySlug(ctx, strings.TrimSpace(slug))
}

func (s *Service) ListAdmin(ctx context.Context) ([]AdminContest, error) {
	return s.repo.ListAdminContests(ctx)
}

func (s *Service) CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminContest, error) {
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeContestStatus(input.Status)
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminContestInput(input.Slug, input.Title, input.Status, input.StartsAt, input.EndsAt); err != nil {
		return AdminContest{}, err
	}
	return s.repo.CreateAdminContest(ctx, actor, input)
}

func (s *Service) UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminContest, error) {
	if input.ContestID == uuid.Nil {
		return AdminContest{}, errors.New("contest id is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = normalizeContestStatus(input.Status)
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminContestInput(input.Slug, input.Title, input.Status, input.StartsAt, input.EndsAt); err != nil {
		return AdminContest{}, err
	}
	return s.repo.UpdateAdminContest(ctx, actor, input)
}

func (s *Service) ReplaceProblemsAdmin(ctx context.Context, actor auth.AuthenticatedUser, input ReplaceProblemsInput) (AdminContest, error) {
	if input.ContestID == uuid.Nil {
		return AdminContest{}, errors.New("contest id is required")
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateProblemBindings(input.Problems); err != nil {
		return AdminContest{}, err
	}
	return s.repo.ReplaceAdminContestProblems(ctx, actor, input)
}

func (s *Service) FreezeAdmin(ctx context.Context, actor auth.AuthenticatedUser, input FreezeAdminInput) (AdminContest, error) {
	if input.ContestID == uuid.Nil {
		return AdminContest{}, errors.New("contest id is required")
	}
	input.Reason = strings.TrimSpace(input.Reason)
	return s.repo.FreezeAdminContest(ctx, actor, input)
}

func validateAdminContestInput(slug string, title string, status string, startsAt time.Time, endsAt time.Time) error {
	if slug == "" {
		return errors.New("slug is required")
	}
	if title == "" {
		return errors.New("title is required")
	}
	switch status {
	case "UPCOMING", "RUNNING", "ENDED":
	default:
		return errors.New("invalid status")
	}
	if startsAt.IsZero() {
		return errors.New("startsAt is required")
	}
	if endsAt.IsZero() {
		return errors.New("endsAt is required")
	}
	if !endsAt.After(startsAt) {
		return errors.New("endsAt must be after startsAt")
	}
	return nil
}

func validateProblemBindings(problems []AdminProblemBindingInput) error {
	seenCodes := make(map[string]struct{}, len(problems))
	seenPositions := make(map[int]struct{}, len(problems))
	seenProblems := make(map[uuid.UUID]struct{}, len(problems))
	for index := range problems {
		problems[index].Code = strings.ToUpper(strings.TrimSpace(problems[index].Code))
		if problems[index].ProblemID == uuid.Nil {
			return fmt.Errorf("problemId is required at index %d", index)
		}
		if problems[index].Code == "" {
			return fmt.Errorf("code is required at index %d", index)
		}
		if problems[index].Position <= 0 {
			return fmt.Errorf("position must be positive at index %d", index)
		}
		if _, exists := seenCodes[problems[index].Code]; exists {
			return fmt.Errorf("duplicate code: %s", problems[index].Code)
		}
		if _, exists := seenPositions[problems[index].Position]; exists {
			return fmt.Errorf("duplicate position: %d", problems[index].Position)
		}
		if _, exists := seenProblems[problems[index].ProblemID]; exists {
			return fmt.Errorf("duplicate problemId: %s", problems[index].ProblemID)
		}
		seenCodes[problems[index].Code] = struct{}{}
		seenPositions[problems[index].Position] = struct{}{}
		seenProblems[problems[index].ProblemID] = struct{}{}
	}
	slices.SortFunc(problems, func(left, right AdminProblemBindingInput) int {
		switch {
		case left.Position < right.Position:
			return -1
		case left.Position > right.Position:
			return 1
		default:
			return strings.Compare(left.Code, right.Code)
		}
	})
	return nil
}

func normalizeContestStatus(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
