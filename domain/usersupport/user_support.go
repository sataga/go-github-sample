// Package usersupport is a domain logic for user support
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package usersupport

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

var (
	jp           = time.FixedZone("Asia/Tokyo", 9*60*60)
	titlePattern = regexp.MustCompile(`.*INC(?P<ServiceID>[0-9]{7}.*)`)
)

// UserSupport is interface for getting user support info
type UserSupport interface {
	GetDailyReportStats(now time.Time, dayAgo int) (*DailyStats, error)
	GetLongTermReportStats(until time.Time, kind string, span int) (*LongTermStats, error)
	GetAnalysisReportStats(since, until time.Time, state string, span int) (*AnalysisStats, error)
	GetKeywordReportStats(until time.Time, kind string, span int) (*KeywordStats, error)
	MethodTest(since, until time.Time) (*AnalysisStats, error)
	// GenMonthlyReport(data map[string]*LongTermStats) string
}

// Repository r/w data which usersupport domain requires
type Repository interface {
	GetUpdatedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetClosedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetCurrentOpenNotUpdatedSupportIssues(until time.Time) ([]*github.Issue, error)
	GetCurrentOpenSupportIssues() ([]*github.Issue, error)
	GetCreatedSupportIssues(since, until time.Time) ([]*github.Issue, error)
	GetLabelsByQuery(query string) ([]*github.LabelResult, error)
}

type userSupport struct {
	repo Repository
}

// DailyStats is stats open data from GitHub
type DailyStats struct {
	dayAgo              int                  `yaml:"day_ago"`
	NumNotUpdatedIssues int                  `yaml:"num_not_updated_issues"`
	NumTeamAResponse    int                  `yaml:"num_team_a_response"`
	NumTeamBResponse    int                  `yaml:"num_team_b_response"`
	NumTeamAHighIssues  int                  `yaml:"num_team_a_high_issues"`
	NumTeamBLowIssues   int                  `yaml:"num_team_b_low_issues"`
	UrgencyHighIssues   int                  `yaml:"num_urgency_high_issues"`
	UrgencyLowIssues    int                  `yaml:"num_urgency_low_issues"`
	DetailStats         map[int]*DetailStats `yaml:"detail_stats"`
}
type LongTermStats struct {
	SummaryStats map[string]*SummaryStats `yaml:"summary_stats"`
	DetailStats  map[int]*DetailStats     `yaml:"detail_stats"`
}

type AnalysisStats struct {
	DetailStats map[int]*DetailStats `yaml:"detail_stats"`
}

type KeywordStats struct {
	KeywordSummary map[string]*KeywordSummary `yaml:"keyword_summary"`
}

type KeywordSummary struct {
	Span                     string         `yaml:"span"`
	KeywordCountAsAll        map[string]int `yaml:"keyword_count_as_all"`
	KeywordCountAsEscalation map[string]int `yaml:"keyword_count_as_escalation"`
}
type SummaryStats struct {
	Span                       string  `yajl:"span"`
	NumCreatedIssues           int     `yaml:"num_created_issues"`
	NumClosedIssues            int     `yaml:"num_closed_issues"`
	NumGenreNormalIssues       int     `yaml:"num_genre_log_survey_issues"`
	NumGenreRequestIssues      int     `yaml:"num_genre_issues"`
	NumGenreFailureIssues      int     `yaml:"num_genre_impact_survey_issues"`
	NumEscalationAllIssues     int     `yaml:"num_escalation_all_issues"`
	NumEscalationNormalIssues  int     `yaml:"num_escalation_normal_issues"`
	NumEscalationRequestIssues int     `yaml:"num_escalation_request_issues"`
	NumEscalationFailureIssues int     `yaml:"num_escalation_failure_issues"`
	NumUrgencyHighIssues       int     `yaml:"num_urgency_high_issues"`
	NumUrgencyLowIssues        int     `yaml:"num_urgency_low_issues"`
	NumScoreA                  int     `yaml:"num_score_A"`
	NumScoreB                  int     `yaml:"num_score_B"`
	NumScoreC                  int     `yaml:"num_score_C"`
	NumScoreD                  int     `yaml:"num_score_D"`
	NumScoreE                  int     `yaml:"num_score_E"`
	NumScoreF                  int     `yaml:"num_score_F"`
	NumTotalScore              float64 `yaml:"num_total_score"`
}

type DetailStats struct {
	Title        string `yaml:"detail_stats_of_title"`
	ServiceID    string `yaml:"detail_stats_of_service_id"`
	HTMLURL      string `yaml:"detail_stats_of_issue_url"`
	CreatedAt    string `yaml:"detail_stats_of_created_at"`
	ClosedAt     string `yaml:"detail_stats_of_closed_at"`
	State        string `yaml:"detail_stats_of_state"`
	TargetSpan   string `yaml:"detail_stats_of_target_span"`
	TeamName     string `yaml:"detail_stats_of_team_name"`
	Urgency      string `yaml:"detail_stats_of_urgency"`
	TeamAResolve bool   `yaml:"detail_stats_of_team_a_resolve"`
	Genre        string `yaml:"detail_stats_of_genre"`
	Labels       string `yaml:"detail_stats_of_labels"`
	Assignee     string `yaml:"detail_stats_of_assign"`
	NumComments  int    `yaml:"detail_stats_of_num_comment"`
	OpenDuration int    `yaml:"detail_stats_of_open_duration"`
	Escalation   bool   `yaml:"detail_stats_of_esalation"`
}

// NewUserSupport creates UserSupport
func NewUserSupport(repo Repository) UserSupport {
	return &userSupport{
		repo: repo,
	}
}

// GetDailryReport
func (us *userSupport) GetDailyReportStats(now time.Time, dayAgo int) (*DailyStats, error) {
	until := now.Add(time.Duration(-24*dayAgo) * time.Hour)
	startEnd := fmt.Sprintf("%s", until.Format("2006-01-02"))
	opi, err := us.repo.GetCurrentOpenNotUpdatedSupportIssues(until)
	if err != nil {
		return nil, fmt.Errorf("get open issues : %s", err)
	}
	DailyStats := &DailyStats{
		dayAgo:              dayAgo,
		NumNotUpdatedIssues: len(opi),
		DetailStats:         make(map[int]*DetailStats, len(opi)),
	}
	for i, issue := range opi {
		DailyStats.DetailStats[i] = &DetailStats{}
		if labelContains(issue.Labels, "緊急度：高") {
			DailyStats.UrgencyHighIssues++
			DailyStats.DetailStats[i].Urgency = "高"
		}
		if labelContains(issue.Labels, "緊急度:中") {
			DailyStats.UrgencyHighIssues++
			DailyStats.DetailStats[i].Urgency = "中"
		}
		if labelContains(issue.Labels, "緊急度:低") {
			DailyStats.UrgencyLowIssues++
			DailyStats.DetailStats[i].Urgency = "低"
		}
		if labelContains(issue.Labels, "CaaS-A 対応中") {
			DailyStats.NumTeamAResponse++
			DailyStats.DetailStats[i].TeamName = "CaaS-A 対応中"
		}
		if labelContains(issue.Labels, "CaaS-B 対応中") {
			DailyStats.NumTeamBResponse++
			DailyStats.DetailStats[i].TeamName = "CaaS-B 対応中"
		}
		DailyStats.DetailStats[i].writeDetailStats(issue, startEnd)
	}
	return DailyStats, nil
}

func (ds *DailyStats) GetDailyReportStats() string {
	var sb strings.Builder

	type kvDetail struct {
		Key int
		Val *DetailStats
	}
	var kvArrForDetail []kvDetail
	for k, v := range ds.DetailStats {
		kvArrForDetail = append(kvArrForDetail, kvDetail{k, v})
	}
	sort.Slice(kvArrForDetail, func(i, j int) bool {
		return kvArrForDetail[i].Val.CreatedAt < kvArrForDetail[j].Val.CreatedAt
	})
	sb.WriteString(fmt.Sprintf("■ *%d日間* 以上更新がなかったチケット一覧\n", ds.dayAgo))
	sb.WriteString(fmt.Sprintf("=== サマリー ===\n"))
	sb.WriteString(fmt.Sprintf("総未更新チケット数: %d 件\n", ds.NumNotUpdatedIssues))
	sb.WriteString(fmt.Sprintf("    緊急度:高・中: %d 件\n", ds.NumTeamAResponse))
	sb.WriteString(fmt.Sprintf("    緊急度:低: %d 件\n", ds.NumTeamBResponse))
	sb.WriteString(fmt.Sprintf("=== 詳細 ===\n"))
	for _, d := range kvArrForDetail {
		dates := d.Val.OpenDuration / 24
		hours := d.Val.OpenDuration % 24
		sb.WriteString(fmt.Sprintf("- <%s|%s> ", d.Val.HTMLURL, d.Val.Title))
		sb.WriteString(fmt.Sprintf("経過時間:%dd%dh ", dates, hours))
		sb.WriteString(fmt.Sprintf("緊急度:%s ", d.Val.Urgency))
		sb.WriteString(fmt.Sprintf("%s\n", d.Val.Assignee))
	}
	return sb.String()
}

func (us *userSupport) GetLongTermReportStats(until time.Time, kind string, span int) (*LongTermStats, error) {
	var since time.Time
	loc, _ := time.LoadLocation("Asia/Tokyo")
	switch kind {
	case "weekly-report":
		since = until.AddDate(0, 0, -7)
	case "monthly-report":
		since = time.Date(until.Year(), until.Month(), 1, 0, 0, 0, 0, loc)
		until = since.AddDate(0, +1, -1)
	}

	LongTermStats := &LongTermStats{
		SummaryStats: make(map[string]*SummaryStats, span),
		DetailStats:  make(map[int]*DetailStats, span),
	}
	cnt := 0
	for i := 1; i <= span; i++ {
		startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		cri, err := us.repo.GetCreatedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get open issues : %s", err)
		}
		upi, err := us.repo.GetUpdatedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get open issues : %s", err)
		}
		numClosed := 0
		for _, issue := range upi {
			if issue.State != nil && *issue.State == "closed" {
				numClosed++
			}
		}

		cli, err := us.repo.GetClosedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get updated issues : %s", err)
		}

		LongTermStats.SummaryStats[startEnd] = &SummaryStats{
			Span:             startEnd,
			NumClosedIssues:  numClosed,
			NumCreatedIssues: len(cri),
		}
		for _, issue := range cli {
			LongTermStats.DetailStats[cnt] = &DetailStats{
				TeamAResolve: false,
			}
			if labelContains(issue.Labels, "genre:通常問合せ") {
				LongTermStats.SummaryStats[startEnd].NumGenreNormalIssues++
			}
			if labelContains(issue.Labels, "genre:要望") {
				LongTermStats.SummaryStats[startEnd].NumGenreRequestIssues++
			}
			if labelContains(issue.Labels, "genre:サービス障害") {
				LongTermStats.SummaryStats[startEnd].NumGenreFailureIssues++
			}
			if labelContains(issue.Labels, "Escalation") {
				LongTermStats.SummaryStats[startEnd].NumEscalationAllIssues++
				if labelContains(issue.Labels, "genre:通常問合せ") {
					LongTermStats.SummaryStats[startEnd].NumEscalationNormalIssues++
				}
				if labelContains(issue.Labels, "genre:要望") {
					LongTermStats.SummaryStats[startEnd].NumEscalationRequestIssues++
				}
				if labelContains(issue.Labels, "genre:サービス障害") {
					LongTermStats.SummaryStats[startEnd].NumEscalationFailureIssues++
				}
			}
			if labelContains(issue.Labels, "緊急度：高") || labelContains(issue.Labels, "緊急度:中") {
				LongTermStats.SummaryStats[startEnd].NumUrgencyHighIssues++
			}
			if labelContains(issue.Labels, "緊急度:低") {
				LongTermStats.SummaryStats[startEnd].NumUrgencyLowIssues++
			}

			var totalTime int
			if issue.State != nil && *issue.State == "closed" {
				totalTime = int(issue.ClosedAt.Sub(*issue.CreatedAt).Hours())
			} else {
				totalTime = int(issue.UpdatedAt.Sub(*issue.CreatedAt).Hours())
			}
			switch {
			case totalTime <= 2*24:
				LongTermStats.SummaryStats[startEnd].NumScoreA++
			case totalTime <= 5*24:
				LongTermStats.SummaryStats[startEnd].NumScoreB++
			case totalTime <= 10*24:
				LongTermStats.SummaryStats[startEnd].NumScoreC++
			case totalTime <= 20*24:
				LongTermStats.SummaryStats[startEnd].NumScoreD++
			case totalTime <= 30*24:
				LongTermStats.SummaryStats[startEnd].NumScoreE++
			default:
				LongTermStats.SummaryStats[startEnd].NumScoreF++
			}

			LongTermStats.DetailStats[cnt].writeDetailStats(issue, startEnd)
			cnt++
		}
		LongTermStats.SummaryStats[startEnd].NumClosedIssues = len(cli)
		if LongTermStats.SummaryStats[startEnd].NumClosedIssues == 0 {
			LongTermStats.SummaryStats[startEnd].NumTotalScore = 0
		} else {
			tmpTotal := (LongTermStats.SummaryStats[startEnd].NumScoreA * 1) + (LongTermStats.SummaryStats[startEnd].NumScoreB * 2) + (LongTermStats.SummaryStats[startEnd].NumScoreC * 3) + (LongTermStats.SummaryStats[startEnd].NumScoreD * 4) + (LongTermStats.SummaryStats[startEnd].NumScoreE * 5) + (LongTermStats.SummaryStats[startEnd].NumScoreF * 6)
			LongTermStats.SummaryStats[startEnd].NumTotalScore = float64(tmpTotal) / float64(LongTermStats.SummaryStats[startEnd].NumClosedIssues)
		}

		switch kind {
		case "weekly-report":
			since = since.AddDate(0, 0, -7)
			until = until.AddDate(0, 0, -7)
		case "monthly-report":
			since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
			until = since.AddDate(0, +1, -1)
		}

	}
	return LongTermStats, nil
}

func (lts *LongTermStats) GenLongTermReport() string {
	var sb strings.Builder
	var Span []string
	var NumCreatedIssues []string
	var NumClosedIssues []string
	var NumGenreRequestIssues []string
	var NumGenreNormalIssues []string
	var NumGenreFailureIssues []string
	var NumUrgencyHighIssues []string
	var NumUrgencyLowIssues []string
	var NumEscalationAllIssues []string
	var NumTeamAResolveAllPercentage []string
	var NumEscalationNormalIssues []string
	var NumTeamAResolveNormalPercentage []string
	var NumScoreA []string
	var NumScoreB []string
	var NumScoreC []string
	var NumScoreD []string
	var NumScoreE []string
	var NumScoreF []string
	var NumTotalScore []string

	type kvSummary struct {
		Key string
		Val *SummaryStats
	}
	var kvArrForSummary []kvSummary
	for k, v := range lts.SummaryStats {
		kvArrForSummary = append(kvArrForSummary, kvSummary{k, v})
	}
	// sort by Span
	sort.Slice(kvArrForSummary, func(i, j int) bool {
		return kvArrForSummary[i].Val.Span < kvArrForSummary[j].Val.Span
	})

	type kvDetail struct {
		Key int
		Val *DetailStats
	}

	var kvArrForDetail []kvDetail
	for k, v := range lts.DetailStats {
		kvArrForDetail = append(kvArrForDetail, kvDetail{k, v})
	}
	// sort by Span
	sort.Slice(kvArrForDetail, func(i, j int) bool {
		return kvArrForDetail[i].Val.TargetSpan < kvArrForDetail[j].Val.TargetSpan
	})

	for _, d := range kvArrForSummary {
		Span = append(Span, d.Val.Span)
		NumCreatedIssues = append(NumCreatedIssues, strconv.Itoa(d.Val.NumCreatedIssues))
		NumClosedIssues = append(NumClosedIssues, strconv.Itoa(d.Val.NumClosedIssues))
		NumGenreRequestIssues = append(NumGenreRequestIssues, strconv.Itoa(d.Val.NumGenreRequestIssues))
		NumGenreNormalIssues = append(NumGenreNormalIssues, strconv.Itoa(d.Val.NumGenreNormalIssues))
		NumGenreFailureIssues = append(NumGenreFailureIssues, strconv.Itoa(d.Val.NumGenreNormalIssues))
		NumUrgencyHighIssues = append(NumUrgencyHighIssues, strconv.Itoa(d.Val.NumUrgencyHighIssues))
		NumUrgencyLowIssues = append(NumUrgencyLowIssues, strconv.Itoa(d.Val.NumUrgencyLowIssues))
		NumEscalationAllIssues = append(NumEscalationAllIssues, strconv.Itoa(d.Val.NumEscalationAllIssues))
		if d.Val.NumEscalationAllIssues != 0 {
			if d.Val.NumClosedIssues != 0 {
				NumTeamAResolveAllPercentage = append(NumTeamAResolveAllPercentage, fmt.Sprintf("%.1f", (float64(d.Val.NumEscalationAllIssues)/float64(d.Val.NumClosedIssues)*100)))
			} else {
				NumTeamAResolveAllPercentage = append(NumTeamAResolveAllPercentage, "0")
			}
		} else {
			NumTeamAResolveAllPercentage = append(NumTeamAResolveAllPercentage, "0")
		}

		NumEscalationNormalIssues = append(NumEscalationNormalIssues, strconv.Itoa(d.Val.NumEscalationNormalIssues))
		if d.Val.NumEscalationNormalIssues != 0 {
			if d.Val.NumClosedIssues != 0 {
				NumTeamAResolveNormalPercentage = append(NumTeamAResolveNormalPercentage, fmt.Sprintf("%.1f", (float64(d.Val.NumEscalationNormalIssues)/float64(d.Val.NumClosedIssues)*100)))
			} else {
				NumTeamAResolveNormalPercentage = append(NumTeamAResolveNormalPercentage, "0")
			}
		} else {
			NumTeamAResolveNormalPercentage = append(NumTeamAResolveNormalPercentage, "0")
		}

		NumScoreA = append(NumScoreA, strconv.Itoa(d.Val.NumScoreA))
		NumScoreB = append(NumScoreB, strconv.Itoa(d.Val.NumScoreB))
		NumScoreC = append(NumScoreC, strconv.Itoa(d.Val.NumScoreC))
		NumScoreD = append(NumScoreD, strconv.Itoa(d.Val.NumScoreD))
		NumScoreE = append(NumScoreE, strconv.Itoa(d.Val.NumScoreE))
		NumScoreF = append(NumScoreF, strconv.Itoa(d.Val.NumScoreF))
		NumTotalScore = append(NumTotalScore, strconv.FormatFloat(float64(d.Val.NumTotalScore), 'f', 2, 64))
	}
	sb.WriteString(fmt.Sprintf("## サマリー \n"))
	sb.WriteString(fmt.Sprintf("|項目|"))
	sb.WriteString(fmt.Sprintf("%s|\n", strings.Join(Span, "|")))
	sb.WriteString(fmt.Sprintf("|----|"))
	for i := 0; i < len(kvArrForSummary); i++ {
		sb.WriteString(fmt.Sprintf("----|"))
	}
	sb.WriteString(fmt.Sprintf("\n"))
	sb.WriteString(fmt.Sprintf("|起票件数|%s|\n", strings.Join(NumCreatedIssues, "|")))
	sb.WriteString(fmt.Sprintf("|クローズ件数|%s|\n", strings.Join(NumClosedIssues, "|")))
	sb.WriteString(fmt.Sprintf("|緊急度:高・中|%s|\n", strings.Join(NumUrgencyHighIssues, "|")))
	sb.WriteString(fmt.Sprintf("|緊急度:低|%s|\n", strings.Join(NumUrgencyLowIssues, "|")))
	sb.WriteString(fmt.Sprintf("|全体エスカレーション件数|%s|\n", strings.Join(NumEscalationAllIssues, "|")))
	sb.WriteString(fmt.Sprintf("|全体CaaS-A完結率(％)|%s|\n", strings.Join(NumTeamAResolveAllPercentage, "|")))
	sb.WriteString(fmt.Sprintf("|通常エスカレーション件数|%s|\n", strings.Join(NumEscalationNormalIssues, "|")))
	sb.WriteString(fmt.Sprintf("|通常CaaS-A完結率(％)|%s|\n", strings.Join(NumTeamAResolveNormalPercentage, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:通常問合せ件数|%s|\n", strings.Join(NumGenreNormalIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:要望件数|%s|\n", strings.Join(NumGenreRequestIssues, "|")))
	sb.WriteString(fmt.Sprintf("|ジャンル:サービス障害件数|%s|\n", strings.Join(NumGenreFailureIssues, "|")))
	sb.WriteString(fmt.Sprintf("|合計スコア|%s|\n", strings.Join(NumTotalScore, "|")))
	sb.WriteString(fmt.Sprintf("|スコアA|%s|\n", strings.Join(NumScoreA, "|")))
	sb.WriteString(fmt.Sprintf("|スコアB|%s|\n", strings.Join(NumScoreB, "|")))
	sb.WriteString(fmt.Sprintf("|スコアC|%s|\n", strings.Join(NumScoreC, "|")))
	sb.WriteString(fmt.Sprintf("|スコアD|%s|\n", strings.Join(NumScoreD, "|")))
	sb.WriteString(fmt.Sprintf("|スコアE|%s|\n", strings.Join(NumScoreE, "|")))
	sb.WriteString(fmt.Sprintf("|スコアF|%s|\n", strings.Join(NumScoreF, "|")))
	sb.WriteString(fmt.Sprintf("\n"))

	sb.WriteString(fmt.Sprintf("## 詳細 \n"))
	for _, d := range kvArrForDetail {
		sb.WriteString(fmt.Sprintf("- [%s](%s),%s,%s,%s/%s,comment数:%d,経過時間(hour):%d,解決フラグ:%t,(%s)\n", d.Val.Title, d.Val.HTMLURL, d.Val.Urgency, d.Val.Genre, d.Val.TeamName, d.Val.Assignee, d.Val.NumComments, d.Val.OpenDuration, d.Val.Escalation, d.Val.TargetSpan))
	}
	return sb.String()
}

func (us *userSupport) GetAnalysisReportStats(since, until time.Time, state string, span int) (*AnalysisStats, error) {

	loc, _ := time.LoadLocation("Asia/Tokyo")
	since = time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, loc)
	until = since.AddDate(0, +1, -1)
	AnalysisStats := &AnalysisStats{
		DetailStats: make(map[int]*DetailStats, span),
	}

	cnt := 0
	for i := 1; i <= span; i++ {
		startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		var iss []*github.Issue
		var err error
		if state == "created" {
			iss, err = us.repo.GetCreatedSupportIssues(since, until)
			if err != nil {
				return nil, fmt.Errorf("get created issue : %s", err)
			}

		} else {
			iss, err = us.repo.GetClosedSupportIssues(since, until)
			if err != nil {
				return nil, fmt.Errorf("get closed issue : %s", err)
			}
		}
		for _, issue := range iss {
			AnalysisStats.DetailStats[cnt] = &DetailStats{
				Escalation: false,
			}
			AnalysisStats.DetailStats[cnt].writeDetailStats(issue, startEnd)
			cnt++
		}
		since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
		until = since.AddDate(0, +1, -1)
	}

	return AnalysisStats, nil
}

// GenReport generate analysis report
func (as *AnalysisStats) GenAnalysisReport() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("期間,Title,起票日,クローズ日,ステータス,担当チーム,担当アサイン,緊急度,問い合わせ種別,エスカレ有無,コメント数,経過時間,Keywordラベル,URL\n"))

	type kvDetail struct {
		Key int
		Val *DetailStats
	}

	var kvArrForDetail []kvDetail
	for k, v := range as.DetailStats {
		kvArrForDetail = append(kvArrForDetail, kvDetail{k, v})
	}
	sort.Slice(kvArrForDetail, func(i, j int) bool {
		return kvArrForDetail[i].Val.CreatedAt < kvArrForDetail[j].Val.CreatedAt
	})

	for _, d := range kvArrForDetail {
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%t,%d,%d,%s,%s \n", d.Val.TargetSpan, d.Val.Title, d.Val.CreatedAt, d.Val.ClosedAt, d.Val.State, d.Val.TeamName, d.Val.Assignee, d.Val.Urgency, d.Val.Genre, d.Val.Escalation, d.Val.NumComments, d.Val.OpenDuration, d.Val.Labels, d.Val.HTMLURL))
	}
	return sb.String()
}

func (us *userSupport) GetKeywordReportStats(until time.Time, kind string, span int) (*KeywordStats, error) {
	var since time.Time
	loc, _ := time.LoadLocation("Asia/Tokyo")
	switch kind {
	case "weekly":
		since = until.AddDate(0, 0, -7)
	case "monthly":
		since = time.Date(until.Year(), until.Month(), 1, 0, 0, 0, 0, loc)
		until = since.AddDate(0, +1, -1)
	}
	keywords, _ := us.repo.GetLabelsByQuery("keyword:")
	KeywordStats := &KeywordStats{
		KeywordSummary: make(map[string]*KeywordSummary, span),
	}

	for i := 1; i <= span; i++ {
		startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
		cli, err := us.repo.GetClosedSupportIssues(since, until)
		if err != nil {
			return nil, fmt.Errorf("get closed issues : %s", err)
		}

		KeywordStats.KeywordSummary[startEnd] = &KeywordSummary{
			Span:                     startEnd,
			KeywordCountAsAll:        make(map[string]int, len(keywords)),
			KeywordCountAsEscalation: make(map[string]int, len(keywords)),
		}
		for _, label := range keywords {
			if _, ok := KeywordStats.KeywordSummary[startEnd].KeywordCountAsAll[*label.Name]; !ok {
				KeywordStats.KeywordSummary[startEnd].KeywordCountAsAll[*label.Name] = 0
				KeywordStats.KeywordSummary[startEnd].KeywordCountAsEscalation[*label.Name] = 0
			}

			for _, issue := range cli {
				if labelContains(issue.Labels, *label.Name) {
					if val, ok := KeywordStats.KeywordSummary[startEnd].KeywordCountAsAll[*label.Name]; ok {
						KeywordStats.KeywordSummary[startEnd].KeywordCountAsAll[*label.Name] = val + 1
					}
					if labelContains(issue.Labels, "Escalation") {
						if val, ok := KeywordStats.KeywordSummary[startEnd].KeywordCountAsEscalation[*label.Name]; ok {
							KeywordStats.KeywordSummary[startEnd].KeywordCountAsEscalation[*label.Name] = val + 1
						}
					}
				}
			}
		}
		switch kind {
		case "weekly":
			since = until.AddDate(0, 0, -7)
			until = until.AddDate(0, 0, -7)
		case "monthly":
			since = time.Date(since.Year(), since.Month()-1, 1, 0, 0, 0, 0, loc)
			until = since.AddDate(0, +1, -1)
		}
	}
	return KeywordStats, nil

}

func (ks *KeywordStats) GenKeywordReport() string {
	type kvSummary struct {
		Key string
		Val *KeywordSummary
	}

	type kvCount struct {
		Key string
		Val int
	}
	var sb strings.Builder
	var Span []string
	var kvArrForSummary []kvSummary
	var kvArrForCountAsAll []kvCount
	var kvArrForCountAsEscalation []kvCount
	kvArrForResultAsAll := make(map[string][]int, len(ks.KeywordSummary))
	kvArrForResultAsEscalation := make(map[string][]int, len(ks.KeywordSummary))

	for k, v := range ks.KeywordSummary {
		kvArrForSummary = append(kvArrForSummary, kvSummary{k, v})
	}
	sort.Slice(kvArrForSummary, func(i, j int) bool {
		return kvArrForSummary[i].Val.Span < kvArrForSummary[j].Val.Span
	})

	for _, d := range kvArrForSummary {
		Span = append(Span, d.Val.Span)
		for k, v := range d.Val.KeywordCountAsAll {
			kvArrForCountAsAll = append(kvArrForCountAsAll, kvCount{k, v})
		}
		sort.Slice(kvArrForCountAsAll, func(i, j int) bool {
			return kvArrForCountAsAll[i].Key < kvArrForCountAsAll[j].Key
		})
		for k, v := range d.Val.KeywordCountAsEscalation {
			kvArrForCountAsEscalation = append(kvArrForCountAsEscalation, kvCount{k, v})
		}
		sort.Slice(kvArrForCountAsEscalation, func(i, j int) bool {
			return kvArrForCountAsEscalation[i].Key < kvArrForCountAsEscalation[j].Key
		})
		for _, v := range kvArrForCountAsAll {
			kvArrForResultAsAll[v.Key] = append(kvArrForResultAsAll[v.Key], v.Val)
		}
		for _, v := range kvArrForCountAsEscalation {
			kvArrForResultAsEscalation[v.Key] = append(kvArrForResultAsEscalation[v.Key], v.Val)
		}
		kvArrForCountAsAll = nil
		kvArrForCountAsEscalation = nil
	}
	sb.WriteString(fmt.Sprintf("## サマリー(全体) \n"))
	sb.WriteString(fmt.Sprintf("|項目|"))
	sb.WriteString(fmt.Sprintf("%s|Total|\n", strings.Join(Span, "|")))
	sb.WriteString(fmt.Sprintf("|----|"))
	for i := 0; i < len(kvArrForSummary); i++ {
		sb.WriteString(fmt.Sprintf("----|"))
	}
	sb.WriteString(fmt.Sprintf("----|\n"))

	for k, v := range kvArrForResultAsAll {
		sb.WriteString(fmt.Sprintf("|%s", k))
		total := 0
		for _, d := range v {
			sb.WriteString(fmt.Sprintf("|%d", d))
			total = total + d
		}
		sb.WriteString(fmt.Sprintf("|%d|\n", total))
	}
	sb.WriteString(fmt.Sprintf("## サマリー(Escalationのみ計上) \n"))
	sb.WriteString(fmt.Sprintf("|項目|"))
	sb.WriteString(fmt.Sprintf("%s|Total|\n", strings.Join(Span, "|")))
	sb.WriteString(fmt.Sprintf("|----|"))
	for i := 0; i < len(kvArrForSummary); i++ {
		sb.WriteString(fmt.Sprintf("----|"))
	}
	sb.WriteString(fmt.Sprintf("----|\n"))

	for k, v := range kvArrForResultAsEscalation {
		sb.WriteString(fmt.Sprintf("|%s", k))
		total := 0
		for _, d := range v {
			sb.WriteString(fmt.Sprintf("|%d", d))
			total = total + d
		}
		sb.WriteString(fmt.Sprintf("|%d|\n", total))
	}

	return sb.String()

}

func (us *userSupport) MethodTest(since, until time.Time) (*AnalysisStats, error) {
	startEnd := fmt.Sprintf("%s~%s", since.Format("2006-01-02"), until.Format("2006-01-02"))
	cli, err := us.repo.GetUpdatedSupportIssues(since, until)
	if err != nil {
		return nil, fmt.Errorf("get updated issues : %s", err)
	}
	AnalysisStats := &AnalysisStats{
		DetailStats: make(map[int]*DetailStats, len(cli)),
	}
	for i, issue := range cli {
		AnalysisStats.DetailStats[i] = &DetailStats{
			TeamAResolve: false,
		}
		AnalysisStats.DetailStats[i].writeDetailStats(issue, startEnd)
	}
	return AnalysisStats, nil
}

func (ds *DetailStats) writeDetailStats(issue *github.Issue, startEnd string) {
	var labels []string
	if issue.Labels != nil {
		for _, label := range issue.Labels {
			if strings.Contains(*label.Name, "keyword") {
				labels = append(labels, strings.Replace(*label.Name, "keyword:", "", -1))
			}
			if strings.Contains(*label.Name, "緊急度") {
				ds.Urgency = strings.Replace(*label.Name, "緊急度:", "", -1)
			}
			if strings.Contains(*label.Name, "CaaS-") {
				ds.TeamName = strings.Replace(*label.Name, " 対応中", "", -1)
			}
			if strings.Contains(*label.Name, "Escalation") {
				ds.TeamAResolve = true
			}
			if strings.Contains(*label.Name, "genre") {
				ds.Genre = strings.Replace(*label.Name, "genre:", "", -1)
			}
		}
	}
	var assigns []string
	if issue.Assignees != nil {
		for _, assign := range issue.Assignees {
			assigns = append(assigns, "@"+*assign.Login)
		}
	}

	var totalTime int
	if issue.State != nil && *issue.State == "closed" {
		totalTime = int(issue.ClosedAt.Sub(*issue.CreatedAt).Hours())
		ds.ClosedAt = issue.ClosedAt.In(jp).Format("2006-01-02")
	} else {
		totalTime = int(issue.UpdatedAt.Sub(*issue.CreatedAt).Hours())
	}

	titleMatches := titlePattern.FindStringSubmatch(*issue.Title)
	if len(titleMatches) == 2 {
		ds.ServiceID = "INC" + titleMatches[1]
	}

	ds.Assignee = strings.Join(assigns, " ")
	ds.Title = *issue.Title
	ds.HTMLURL = *issue.HTMLURL
	ds.NumComments = *issue.Comments
	ds.State = *issue.State
	ds.CreatedAt = issue.CreatedAt.In(jp).Format("2006-01-02")
	ds.OpenDuration = totalTime
	ds.Assignee = strings.Join(assigns, " ")
	ds.Labels = strings.Join(labels, " ")
	ds.TargetSpan = startEnd
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
