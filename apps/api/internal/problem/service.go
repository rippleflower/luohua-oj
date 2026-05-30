package problem

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
)

type Summary struct {
	ID           uuid.UUID
	Slug         string
	Title        string
	Difficulty   string
	Tags         []string
	AcceptedRate float64
	UpdatedAt    time.Time
}

type Detail struct {
	ID            uuid.UUID
	Slug          string
	Title         string
	Difficulty    string
	StatementJSON json.RawMessage
	SamplesJSON   json.RawMessage
	LimitsJSON    json.RawMessage
	MetadataJSON  json.RawMessage
	UpdatedAt     time.Time
}

type AdminProblem struct {
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

type CreateAdminInput struct {
	Slug          string
	Title         string
	Difficulty    string
	TimeLimitMs   int
	MemoryLimitKb int
	Reason        string
	IP            string
}

type UpdateAdminInput struct {
	ProblemID     uuid.UUID
	Slug          string
	Title         string
	Difficulty    string
	TimeLimitMs   int
	MemoryLimitKb int
	Reason        string
	IP            string
}

type PublishAdminInput struct {
	ProblemID uuid.UUID
	Reason    string
	IP        string
}

type Reader interface {
	List(ctx context.Context) ([]Summary, error)
	GetBySlug(ctx context.Context, slug string) (Detail, error)
}

type AdminManager interface {
	CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error)
	UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error)
	PublishAdmin(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error)
}

type Repository interface {
	ListProblems(ctx context.Context) ([]Summary, error)
	GetProblemBySlug(ctx context.Context, slug string) (Detail, error)
	CreateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error)
	UpdateAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error)
	PublishAdminProblem(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Summary, error) {
	return s.repo.ListProblems(ctx)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Detail, error) {
	if strings.TrimSpace(slug) == "" {
		return Detail{}, errors.New("slug is required")
	}
	return s.repo.GetProblemBySlug(ctx, strings.TrimSpace(slug))
}

func (s *Service) CreateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input CreateAdminInput) (AdminProblem, error) {
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Difficulty = strings.TrimSpace(strings.ToUpper(input.Difficulty))
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminInput(input.Slug, input.Title, input.Difficulty, input.TimeLimitMs, input.MemoryLimitKb); err != nil {
		return AdminProblem{}, err
	}
	return s.repo.CreateAdminProblem(ctx, actor, input)
}

func (s *Service) UpdateAdmin(ctx context.Context, actor auth.AuthenticatedUser, input UpdateAdminInput) (AdminProblem, error) {
	if input.ProblemID == uuid.Nil {
		return AdminProblem{}, errors.New("problem id is required")
	}
	input.Slug = normalizeSlug(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Difficulty = strings.TrimSpace(strings.ToUpper(input.Difficulty))
	input.Reason = strings.TrimSpace(input.Reason)
	if err := validateAdminInput(input.Slug, input.Title, input.Difficulty, input.TimeLimitMs, input.MemoryLimitKb); err != nil {
		return AdminProblem{}, err
	}
	return s.repo.UpdateAdminProblem(ctx, actor, input)
}

func (s *Service) PublishAdmin(ctx context.Context, actor auth.AuthenticatedUser, input PublishAdminInput) (AdminProblem, error) {
	if input.ProblemID == uuid.Nil {
		return AdminProblem{}, errors.New("problem id is required")
	}
	input.Reason = strings.TrimSpace(input.Reason)
	return s.repo.PublishAdminProblem(ctx, actor, input)
}

func validateAdminInput(slug string, title string, difficulty string, timeLimitMs int, memoryLimitKb int) error {
	if slug == "" {
		return errors.New("slug is required")
	}
	if title == "" {
		return errors.New("title is required")
	}
	switch difficulty {
	case "EASY", "MEDIUM", "HARD":
	default:
		return errors.New("invalid difficulty")
	}
	if timeLimitMs <= 0 {
		return errors.New("timeLimitMs must be positive")
	}
	if memoryLimitKb <= 0 {
		return errors.New("memoryLimitKb must be positive")
	}
	return nil
}

func normalizeSlug(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
