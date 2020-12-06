package usersupport

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

// UserSupport is interface for getting user support info
type UserSupport interface {
	GetUserSupportStats(since, until time.Time) (*Stats, error)
	GetDailyReportStats(until time.Time) (*dailyStats, error)
}

// Repository r/w data which usersupport domain requires
type Repository interface {
	GetUpdatedSupportIssues(state string, since, until time.Time) ([]*github.Issue, error)
	GetCurrentOpenNotUpdatedSupportIssues(until time.Time) ([]*github.Issue, error)
	GetCurrentOpenSupportIssues() ([]*github.Issue, error)
	GetCurrentOpenAnyLabelsSupportIssues(state string, labels []string) ([]*github.Issue, error)
	GetCurrentRepoLabels() ([]*github.Label, error)
	GetLabelsByQuery(repoID int64, query string) (*github.LabelsSearchResult, *github.Response, error)
}

type userSupport struct {
	repo Repository
}

// Stats is stats data from GitHub
type Stats struct {
	NumCreatedIssues                int                      `yaml:"num_created_issues"`
	NumClosedIssues                 int                      `yaml:"num_closed_issues"`
	UrgencyHighIssues               int                      `yaml:"num_urgency_high_issues"`
	UrgencyLowIssues                int                      `yaml:"num_urgency_low_issues"`
	UrgencyHighDifficultyHighIssues int                      `yaml:"num_urgency_high_difficulty_high_issues"`
	UrgencyHighDifficultyLowIssues  int                      `yaml:"num_urgency_high_difficulty_low_issues"`
	UrgencyLowDifficultyHighIssues  int                      `yaml:"num_urgency_low_difficulty_high_issues"`
	UrgencyLowDifficultyLowIssues   int                      `yaml:"num_urgency_low_difficulty_low_issues"`
	NumOpenIssues                   int                      `yaml:"num_open_issues"`
	NumUpdatedIssues                int                      `yaml:"num_updated_issues"`
	NumCommentsPerIssue             map[string]int           `yaml:"num_comments_per_issue"`
	OpenDurationPerIssue            map[string]time.Duration `yaml:"open_duration_per_issue"`
}

// dailyStats is stats open data from GitHub
type dailyStats struct {
	NumNotUpdatedIssues int                       `yaml:"num_not_updated_issues"`
	NumTeamAResponse    int                       `yaml:"num_team_a_response"`
	NumTeamBResponse    int                       `yaml:"num_team_b_response"`
	UrgencyHighIssues   int                       `yaml:"num_urgency_high_issues"`
	UrgencyLowIssues    int                       `yaml:"num_urgency_low_issues"`
	NotUpdatedIssues    map[int]*NotUpdatedIssues `yaml:"not_updated_issues"`
}

// NotUpdatedIssues is dailyStats of datail
type NotUpdatedIssues struct {
	Title        string        `yaml:"not_updated_issues_of_title"`
	URL          string        `yaml:"not_updated_issues_of_issue_url"`
	NumComments  int           `yaml:"not_updated_issues_of_num_comment"`
	OpenDuration time.Duration `yaml:"not_updated_issues_of_open_duration"`
}

// NewUserSupport creates UserSupport
func NewUserSupport(repo Repository) UserSupport {
	return &userSupport{
		repo: repo,
	}
}

// GetDailryReport
func (us *userSupport) GetDailyReportStats(until time.Time) (*dailyStats, error) {
	opi, err := us.repo.GetCurrentOpenNotUpdatedSupportIssues(until)
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	dailyStats := &dailyStats{
		NumNotUpdatedIssues: len(opi),
		NotUpdatedIssues:    make(map[int]*NotUpdatedIssues, len(opi)),
	}
	for i, issue := range opi {
		if labelContains(issue.Labels, "緊急度:高") || labelContains(issue.Labels, "緊急度:中") {
			dailyStats.UrgencyHighIssues++
		}
		if labelContains(issue.Labels, "緊急度:低") {
			dailyStats.UrgencyLowIssues++
		}
		if labelContains(issue.Labels, "Team-A") {
			dailyStats.NumTeamAResponse++
		}
		if labelContains(issue.Labels, "Team-B") {
			dailyStats.NumTeamBResponse++
		}

		var duration time.Duration
		if issue.State != nil && *issue.State == "closed" {
			duration = issue.ClosedAt.Sub(*issue.CreatedAt)
		} else {
			duration = issue.UpdatedAt.Sub(*issue.CreatedAt)
		}
		if issue.Title != nil && issue.URL != nil && issue.Comments != nil {
			dailyStats.NotUpdatedIssues[i] = &NotUpdatedIssues{
				Title:        *issue.Title,
				URL:          *issue.URL,
				NumComments:  *issue.Comments,
				OpenDuration: duration,
			}
		}
	}
	// fmt.Printf("%v", dailyStats)
	return dailyStats, nil
}

// GenReport generate report
func (s *dailyStats) GetDailyReportStats() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== サマリー ===\n"))
	sb.WriteString(fmt.Sprintf("総未更新チケット数, %d\n", s.NumNotUpdatedIssues))
	sb.WriteString(fmt.Sprintf("Team-A 未更新チケット数, %d\n", s.NumTeamAResponse))
	sb.WriteString(fmt.Sprintf("Team-B 未更新チケット数, %d\n", s.NumTeamBResponse))
	sb.WriteString(fmt.Sprintf("=== 詳細((チケットリンク/Openからの経過時間)) ===\n"))
	for _, issue := range s.NotUpdatedIssues {
		totalHours := int(issue.OpenDuration.Hours())
		dates := totalHours / 24
		hours := totalHours % 24
		sb.WriteString(fmt.Sprintf("- <%s|%s> ", issue.URL, issue.Title))
		sb.WriteString(fmt.Sprintf("%dd %dh\n", dates, hours))
	}

	return sb.String()
}

// GetUserSupportStats
func (us *userSupport) GetUserSupportStats(since, until time.Time) (*Stats, error) {
	state := "open"
	upi, err := us.repo.GetUpdatedSupportIssues(state, since, until)
	if err != nil {
		return nil, fmt.Errorf("get updated issues : %s", err)
	}
	opi, err := us.repo.GetCurrentOpenSupportIssues()
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	labels, err := us.repo.GetCurrentRepoLabels()
	if err != nil {
		return nil, fmt.Errorf("get label list : %s", err)
	}
	fmt.Println(labels)

	slabels, _, _ := us.repo.GetLabelsByQuery(233058313, "keyword:")
	fmt.Println(slabels)

	usStats := &Stats{
		NumUpdatedIssues:     len(upi),
		NumOpenIssues:        len(opi),
		NumCommentsPerIssue:  make(map[string]int, len(upi)),
		OpenDurationPerIssue: make(map[string]time.Duration, len(upi)),
	}
	for _, issue := range opi {
		if labelContains(issue.Labels, "緊急度:高") || labelContains(issue.Labels, "緊急度:中") {
			usStats.UrgencyHighIssues++
			if labelContains(issue.Labels, "難易度:高") {
				usStats.UrgencyHighDifficultyHighIssues++
			}
			if labelContains(issue.Labels, "難易度:低") {
				usStats.UrgencyHighDifficultyLowIssues++
			}
		}
		if labelContains(issue.Labels, "緊急度:低") {
			usStats.UrgencyLowIssues++
			if labelContains(issue.Labels, "難易度:高") {
				usStats.UrgencyLowDifficultyHighIssues++
			}
			if labelContains(issue.Labels, "難易度:低") {
				usStats.UrgencyLowDifficultyLowIssues++
			}
		}
	}
	numCreated, numClosed := 0, 0
	for _, issue := range upi {
		if issue.State != nil && *issue.State == "closed" {
			numClosed++
		}
		if issue.CreatedAt != nil && issue.CreatedAt.After(since) && issue.CreatedAt.Before(until) {
			numCreated++
		}
		title := fmt.Sprintf("[%s]%s", *issue.State, *issue.Title)
		if issue.Comments != nil {
			usStats.NumCommentsPerIssue[title] = *issue.Comments
		}
		if issue.State != nil && *issue.State == "closed" {
			usStats.OpenDurationPerIssue[title] = issue.ClosedAt.Sub(*issue.CreatedAt)
		} else {
			usStats.OpenDurationPerIssue[title] = issue.UpdatedAt.Sub(*issue.CreatedAt)
		}
	}
	usStats.NumClosedIssues = numClosed
	usStats.NumCreatedIssues = numCreated
	return usStats, nil
}

// GenReport generate report
func (s *Stats) GenReport() string {
	var sb strings.Builder
	sb.WriteString("項目, 情報\n")
	sb.WriteString(fmt.Sprintf("Openチケット数, %d\n", s.NumOpenIssues))
	sb.WriteString(fmt.Sprintf("新規作成されたチケット数, %d\n", s.NumCreatedIssues))
	sb.WriteString(fmt.Sprintf("情報更新されたチケット数, %d\n", s.NumUpdatedIssues))
	sb.WriteString(fmt.Sprintf("クローズしたチケット数, %d\n", s.NumClosedIssues))
	sb.WriteString(fmt.Sprintf("緊急度:高 のチケット数, %d\n", s.UrgencyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:高 のチケット数, %d\n", s.UrgencyHighDifficultyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:低 のチケット数, %d\n", s.UrgencyHighDifficultyLowIssues))
	sb.WriteString(fmt.Sprintf("緊急度:低 のチケット数, %d\n", s.UrgencyLowIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:高 のチケット数, %d\n", s.UrgencyLowDifficultyHighIssues))
	sb.WriteString(fmt.Sprintf("かつ、難易度:低 のチケット数, %d\n", s.UrgencyLowDifficultyLowIssues))
	{
		// 経過時間を出すロジック
		type kv struct {
			Key string
			Val time.Duration
		}
		var kvArr []kv
		for k, v := range s.OpenDurationPerIssue {
			kvArr = append(kvArr, kv{k, v})
		}
		// sort by duration
		sort.Slice(kvArr, func(i, j int) bool {
			return kvArr[i].Val > kvArr[j].Val
		})
		for _, kv := range kvArr {
			// if >=10 {
			// 	break
			// }
			totalHours := int(kv.Val.Hours())
			dates := totalHours / 24
			hours := totalHours % 24
			sb.WriteString(fmt.Sprintf("%s, %dd %dh\n", kv.Key, dates, hours))
		}
	}

	return sb.String()
}

//配列の中に特定の文字列が含まれるかを返す
func labelContains(arr []github.Label, str string) bool {
	for _, v := range arr {
		if *v.Name == str {
			return true
		}
	}
	return false
}
