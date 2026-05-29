package contest_makeup

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/example/oj3/apps/api/internal/auth"
	"github.com/google/uuid"
)

var (
	ErrContestNotFound = errors.New("contest not found")
	ErrContestNotEnded = errors.New("contest is not ended")
)

type List struct {
	ContestSlug string
	GeneratedAt time.Time
	Items       []Item
}

type Item struct {
	ProblemID       uuid.UUID
	ProblemCode     string
	ProblemSlug     string
	ProblemTitle    string
	Difficulty      string
	Category        string
	LastStatus      string
	AttemptCount    int
	SeverityRank    int
	ReasonSummary   string
	SuggestedAction string
}

type Reader interface {
	GetMakeupList(ctx context.Context, actor auth.AuthenticatedUser, contestSlug string) (List, error)
}

type Repository interface {
	GetContestMeta(ctx context.Context, slug string) (ContestMeta, error)
	ListContestProblems(ctx context.Context, contestID uuid.UUID) ([]ContestProblem, error)
	ListUserProblemAttempts(ctx context.Context, userID uuid.UUID, contestID uuid.UUID) (map[uuid.UUID]ProblemAttempt, error)
}

type ContestMeta struct {
	ID     uuid.UUID
	Slug   string
	Status string
}

type ContestProblem struct {
	ProblemID    uuid.UUID
	ProblemCode  string
	ProblemSlug  string
	ProblemTitle string
	Difficulty   string
	Position     int
}

type ProblemAttempt struct {
	ProblemID    uuid.UUID
	HasAccepted  bool
	LatestStatus string
	AttemptCount int
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMakeupList(ctx context.Context, actor auth.AuthenticatedUser, contestSlug string) (List, error) {
	normalizedSlug := strings.TrimSpace(contestSlug)
	if normalizedSlug == "" {
		return List{}, ErrContestNotFound
	}

	meta, err := s.repo.GetContestMeta(ctx, normalizedSlug)
	if err != nil {
		return List{}, err
	}
	if strings.ToUpper(strings.TrimSpace(meta.Status)) != "ENDED" {
		return List{}, ErrContestNotEnded
	}

	problems, err := s.repo.ListContestProblems(ctx, meta.ID)
	if err != nil {
		return List{}, err
	}
	attempts, err := s.repo.ListUserProblemAttempts(ctx, actor.ID, meta.ID)
	if err != nil {
		return List{}, err
	}

	attemptedUnsolved := make([]Item, 0, len(problems))
	unattempted := make([]Item, 0, len(problems))

	for _, problem := range problems {
		attempt, ok := attempts[problem.ProblemID]
		if ok {
			if attempt.HasAccepted {
				continue
			}
			attemptedUnsolved = append(attemptedUnsolved, Item{
				ProblemID:       problem.ProblemID,
				ProblemCode:     problem.ProblemCode,
				ProblemSlug:     problem.ProblemSlug,
				ProblemTitle:    problem.ProblemTitle,
				Difficulty:      strings.ToUpper(problem.Difficulty),
				Category:        "ATTEMPTED_UNSOLVED",
				LastStatus:      attempt.LatestStatus,
				AttemptCount:    attempt.AttemptCount,
				SeverityRank:    statusSeverityRank(attempt.LatestStatus),
				ReasonSummary:   summarizeAttemptStatus(attempt.LatestStatus, attempt.AttemptCount),
				SuggestedAction: suggestAction(attempt.LatestStatus, true),
			})
			continue
		}
		unattempted = append(unattempted, Item{
			ProblemID:       problem.ProblemID,
			ProblemCode:     problem.ProblemCode,
			ProblemSlug:     problem.ProblemSlug,
			ProblemTitle:    problem.ProblemTitle,
			Difficulty:      strings.ToUpper(problem.Difficulty),
			Category:        "UNATTEMPTED_RECOMMENDED",
			LastStatus:      "",
			SeverityRank:    statusSeverityRank(""),
			ReasonSummary:   "比赛中未尝试该题",
			SuggestedAction: suggestAction("", false),
		})
	}

	slices.SortFunc(attemptedUnsolved, compareItemOrder)
	slices.SortFunc(unattempted, compareItemOrder)

	items := append(attemptedUnsolved, unattempted...)
	return List{
		ContestSlug: meta.Slug,
		GeneratedAt: time.Now().UTC(),
		Items:       items,
	}, nil
}

func compareItemOrder(left Item, right Item) int {
	if leftRank, rightRank := difficultyRank(left.Difficulty), difficultyRank(right.Difficulty); leftRank != rightRank {
		return leftRank - rightRank
	}
	if left.SeverityRank != right.SeverityRank {
		return left.SeverityRank - right.SeverityRank
	}
	if left.ProblemCode == right.ProblemCode {
		return 0
	}
	if left.ProblemCode < right.ProblemCode {
		return -1
	}
	return 1
}

func difficultyRank(difficulty string) int {
	switch strings.ToUpper(strings.TrimSpace(difficulty)) {
	case "EASY":
		return 0
	case "MEDIUM":
		return 1
	case "HARD":
		return 2
	default:
		return 3
	}
}

func statusSeverityRank(status string) int {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "RUNTIME_ERROR":
		return 0
	case "MEMORY_LIMIT_EXCEEDED":
		return 1
	case "TIME_LIMIT_EXCEEDED":
		return 2
	case "WRONG_ANSWER":
		return 3
	case "COMPILE_ERROR":
		return 4
	case "SYSTEM_ERROR":
		return 5
	case "RUNNING":
		return 6
	case "PENDING":
		return 7
	default:
		return 99
	}
}

func summarizeAttemptStatus(status string, attempts int) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "WRONG_ANSWER":
		return formatReason("最近一次结果是 WA，题意或边界条件仍有偏差，建议复查边界样例", attempts)
	case "TIME_LIMIT_EXCEEDED":
		return formatReason("最近一次结果是 TLE，当前实现复杂度不足，建议先降复杂度", attempts)
	case "MEMORY_LIMIT_EXCEEDED":
		return formatReason("最近一次结果是 MLE，空间使用超限，建议收紧数据结构", attempts)
	case "RUNTIME_ERROR":
		return formatReason("最近一次结果是 RE，存在非法访问或状态异常，建议先定位崩溃点", attempts)
	case "COMPILE_ERROR":
		return formatReason("最近一次结果是 CE，代码无法通过编译，建议先修复构建错误", attempts)
	case "SYSTEM_ERROR":
		return formatReason("最近一次结果是 SYSTEM_ERROR，建议先重提确认环境因素", attempts)
	case "PENDING", "RUNNING":
		return formatReason("比赛结束时仍有未完成判题记录", attempts)
	default:
		return formatReason("比赛内有尝试但尚未通过", attempts)
	}
}

func formatReason(message string, attempts int) string {
	if attempts <= 1 {
		return message
	}
	return message + "（共尝试 " + strconv.Itoa(attempts) + " 次）"
}

func suggestAction(status string, attempted bool) string {
	if !attempted {
		return "先完成一次可运行提交，再根据结果迭代优化"
	}
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "WRONG_ANSWER":
		return "重读题意并补 3 组极端样例，优先修正判定条件"
	case "TIME_LIMIT_EXCEEDED":
		return "先降复杂度到可过范围，再做常数优化"
	case "MEMORY_LIMIT_EXCEEDED":
		return "改用更紧凑的数据结构并减少拷贝"
	case "RUNTIME_ERROR":
		return "定位越界与空指针路径，先加断言再重提"
	case "COMPILE_ERROR":
		return "先修复编译错误并最小化改动验证"
	default:
		return "从最后一次提交出发，先保证正确性再提速"
	}
}
