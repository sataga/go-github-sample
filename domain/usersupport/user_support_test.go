// Package usersupport is a domain logic for user support
//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

package usersupport

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"
)

var (
	loc, _          = time.LoadLocation("Asia/Tokyo")
	now             = time.Now()
	oneHourAgo      = now.Add(-1 * time.Hour)
	threeDayAgo     = now.Add(-3 * 24 * time.Hour)
	fiveDayAgo      = now.Add(-5 * 24 * time.Hour)
	sevenDayAgo     = now.Add(-7 * 24 * time.Hour)
	tenDayAgo       = now.Add(-10 * 24 * time.Hour)
	firstDayOfMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	lastDayOfMonth  = firstDayOfMonth.AddDate(0, +1, -1)
	startEnd        = fmt.Sprintf("%s~%s", firstDayOfMonth.Format("2006-01-02"), lastDayOfMonth.Format("2006-01-02"))
	issuePatterns   = []*github.Issue{
		{
			ID:        github.Int64(1),
			Title:     github.String("issue 1"),
			CreatedAt: &tenDayAgo,
			ClosedAt:  &threeDayAgo,
			State:     github.String("closed"),
			Body:      github.String("test 1"),
			Comments:  github.Int(1),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：低")},
				{Name: github.String("CaaS-A 対応中")},
				{Name: github.String("Escalation")},
				{Name: github.String("genre:通常問合せ")},
				{Name: github.String("keyword:Kubernetes")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/1"),
		},
		{
			ID:        github.Int64(2),
			Title:     github.String("issue 2"),
			CreatedAt: &sevenDayAgo,
			ClosedAt:  &threeDayAgo,
			State:     github.String("closed"),
			Body: github.String(`
				test2
				hogehoge
			`),
			Comments: github.Int(2),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：中")},
				{Name: github.String("CaaS-A 対応中")},
				{Name: github.String("genre:要望")},
				{Name: github.String("keyword:Openstack")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/2"),
		},
		{
			ID:        github.Int64(3),
			Title:     github.String("issue 3"),
			CreatedAt: &fiveDayAgo,
			UpdatedAt: &oneHourAgo,
			State:     github.String("open"),
			Body:      github.String("test 3"),
			Comments:  github.Int(3),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：高")},
				{Name: github.String("CaaS-A 対応中")},
				{Name: github.String("genre:サービス障害")},
				{Name: github.String("keyword:Network")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/3"),
		},
		{
			ID:        github.Int64(4),
			Title:     github.String("issue 4"),
			CreatedAt: &threeDayAgo,
			UpdatedAt: &oneHourAgo,
			State:     github.String("open"),
			Body:      github.String("test 4"),
			Comments:  github.Int(4),
			Labels: []github.Label{
				{Name: github.String("PF_Support")},
				{Name: github.String("緊急度：低")},
				{Name: github.String("CaaS-B 対応中")},
				{Name: github.String("genre:通常問合せ")},
				{Name: github.String("keyword:Kubernetes")},
			},
			HTMLURL: github.String("https://github.com/sataga/issue-warehouse/issues/4"),
		},
	}
	keywordPatterns = []*github.LabelResult{
		{Name: github.String("keyword:Network")},
		{Name: github.String("keyword:Openstack")},
		{Name: github.String("keyword:Kubernetes")},
	}
)

func Test_userSupport_GetDailyReportStats(t *testing.T) {
	var c *gomock.Controller

	updatedIssues := []*github.Issue{
		issuePatterns[2],
		issuePatterns[3],
	}

	type fields struct {
		repo Repository
	}
	type args struct {
		dayAgo int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *DailyStats
		wantErr    bool
		beforefunc func(f *fields)
		afterFunc  func()
	}{
		// TODO: Add test cases.
		{
			name: "since < until",
			args: args{
				dayAgo: 5,
			},
			want: &DailyStats{
				NumNotUpdatedIssues: 2,
				NumTeamAResponse:    1,
				NumTeamBResponse:    1,
				NumTeamAHighIssues:  0,
				NumTeamBLowIssues:   0,
				UrgencyHighIssues:   1,
				UrgencyLowIssues:    1,
				dayAgo:              5,
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 3",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/3",
						CreatedAt:    fiveDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-A",
						Urgency:      "高",
						Genre:        "サービス障害",
						Labels:       "Network",
						NumComments:  3,
						OpenDuration: 119,
						Escalation:   false,
					},
					1: {
						Title:        "issue 4",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/4",
						CreatedAt:    threeDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-B",
						Urgency:      "低",
						Genre:        "通常問合せ",
						Labels:       "Kubernetes",
						NumComments:  4,
						OpenDuration: 71,
						Escalation:   false,
					},
				},
			},
			wantErr: false,
			beforefunc: func(f *fields) {
				c = gomock.NewController(t)
				musr := NewMockRepository(c)
				musr.EXPECT().GetCurrentOpenNotUpdatedSupportIssues(gomock.Any()).Return(updatedIssues, nil)
				f.repo = musr
			},
			afterFunc: func() {
				c.Finish()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforefunc != nil {
				tt.beforefunc(&tt.fields)
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetDailyReportStats(now, tt.args.dayAgo)
			// fmt.Printf("got: %+v %+v\n ", got.DetailStats[0], got.DetailStats[1])
			// fmt.Printf("want: %+v %+v\n ", tt.want.DetailStats[0], tt.want.DetailStats[1])
			if (err != nil) != tt.wantErr {
				t.Errorf("userSupport.GetDailyReportStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userSupport.GetDailyReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDailyStats_GetDailyReportStats(t *testing.T) {
	type fields struct {
		dayAgo              int
		NumNotUpdatedIssues int
		NumTeamAResponse    int
		NumTeamBResponse    int
		NumTeamAHighIssues  int
		NumTeamBLowIssues   int
		UrgencyHighIssues   int
		UrgencyLowIssues    int
		DetailStats         map[int]*DetailStats
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "print daily-report",
			fields: fields{
				NumNotUpdatedIssues: 2,
				NumTeamAResponse:    1,
				NumTeamBResponse:    1,
				NumTeamAHighIssues:  0,
				NumTeamBLowIssues:   0,
				UrgencyHighIssues:   1,
				UrgencyLowIssues:    1,
				dayAgo:              5,
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 3",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/3",
						CreatedAt:    fiveDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-A",
						Urgency:      "高",
						Labels:       "Network",
						NumComments:  3,
						OpenDuration: 119,
						Escalation:   false,
					},
					1: {
						Title:        "issue 4",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/4",
						CreatedAt:    threeDayAgo.Format("2006-01-02"),
						State:        "open",
						TargetSpan:   fiveDayAgo.Format("2006-01-02"),
						TeamName:     "CaaS-B",
						Urgency:      "低",
						Labels:       "Kubernetes",
						NumComments:  4,
						OpenDuration: 71,
						Escalation:   false,
					},
				},
			},
			want: `■ *5日間* 以上更新がなかったチケット一覧
=== サマリー ===
総未更新チケット数: 2 件
    緊急度：高・中: 1 件
    緊急度：低: 1 件
=== 詳細 ===
- <https://github.com/sataga/issue-warehouse/issues/3|issue 3> 経過時間:4d23h 緊急度：高 
- <https://github.com/sataga/issue-warehouse/issues/4|issue 4> 経過時間:2d23h 緊急度：低 
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DailyStats{
				dayAgo:              tt.fields.dayAgo,
				NumNotUpdatedIssues: tt.fields.NumNotUpdatedIssues,
				NumTeamAResponse:    tt.fields.NumTeamAResponse,
				NumTeamBResponse:    tt.fields.NumTeamBResponse,
				NumTeamAHighIssues:  tt.fields.NumTeamAHighIssues,
				NumTeamBLowIssues:   tt.fields.NumTeamBLowIssues,
				UrgencyHighIssues:   tt.fields.UrgencyHighIssues,
				UrgencyLowIssues:    tt.fields.UrgencyLowIssues,
				DetailStats:         tt.fields.DetailStats,
			}
			// fmt.Printf("got: %+v\n ", ds.GetDailyReportStats())
			// fmt.Printf("want:%+v\n ", tt.want)
			if got := ds.GetDailyReportStats(); got != tt.want {
				t.Errorf("DailyStats.GetDailyReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userSupport_GetLongTermReportStats(t *testing.T) {
	var c *gomock.Controller

	closeIssues := []*github.Issue{
		issuePatterns[0],
		issuePatterns[1],
	}
	openIssues := []*github.Issue{
		issuePatterns[2],
		issuePatterns[3],
	}
	type fields struct {
		repo Repository
	}
	type args struct {
		since time.Time
		until time.Time
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *LongTermStats
		wantErr    bool
		beforefunc func(f *fields)
		afterfunc  func()
	}{
		// TODO: Add test cases.
		{
			name: "Normal operation for monthly",
			args: args{
				since: firstDayOfMonth,
				until: lastDayOfMonth,
			},
			want: &LongTermStats{
				SummaryStats: map[string]*SummaryStats{
					startEnd: {
						Span:                       startEnd,
						NumCreatedIssues:           2,
						NumClosedIssues:            2,
						NumGenreNormalIssues:       1,
						NumGenreRequestIssues:      1,
						NumGenreFailureIssues:      0,
						NumEscalationAllIssues:     1,
						NumEscalationNormalIssues:  1,
						NumEscalationRequestIssues: 0,
						NumEscalationFailureIssues: 0,
						NumUrgencyHighIssues:       1,
						NumUrgencyLowIssues:        1,
						NumScoreA:                  0,
						NumScoreB:                  1,
						NumScoreC:                  1,
						NumScoreD:                  0,
						NumScoreE:                  0,
						NumScoreF:                  0,
						NumTotalScore:              2.5,
					},
				},
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 1",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/1",
						CreatedAt:    tenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "低",
						Genre:        "通常問合せ",
						Labels:       "Kubernetes",
						NumComments:  1,
						OpenDuration: 168,
						Escalation:   true,
					},
					1: {
						Title:        "issue 2",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/2",
						CreatedAt:    sevenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "中",
						Genre:        "要望",
						Labels:       "Openstack",
						NumComments:  2,
						OpenDuration: 96,
						Escalation:   false,
					},
				},
			},
			wantErr: false,
			beforefunc: func(f *fields) {
				c = gomock.NewController(t)
				musr := NewMockRepository(c)
				musr.EXPECT().GetCreatedSupportIssues(gomock.Any(), gomock.Any()).Return(openIssues, nil)
				musr.EXPECT().GetClosedSupportIssues(gomock.Any(), gomock.Any()).Return(closeIssues, nil)
				f.repo = musr
			},
			afterfunc: func() {
				c.Finish()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforefunc != nil {
				tt.beforefunc(&tt.fields)
			}
			if tt.afterfunc != nil {
				defer tt.afterfunc()
			}
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetLongTermReportStats(tt.args.since, tt.args.until)
			// for k, v := range tt.want.SummaryStats {
			// 	fmt.Println(k)
			// 	fmt.Println(v)
			// }
			// for k, v := range tt.want.DetailStats {
			// 	fmt.Println(k)
			// 	fmt.Println(v)
			// }
			// for k, v := range got.SummaryStats {
			// 	fmt.Println(k)
			// 	fmt.Println(v)
			// }
			if (err != nil) != tt.wantErr {
				t.Errorf("userSupport.GetLongTermReportStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userSupport.GetLongTermReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLongTermStats_GenLongTermReport(t *testing.T) {
	type fields struct {
		SummaryStats map[string]*SummaryStats
		DetailStats  map[int]*DetailStats
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "print monthly-report",
			fields: fields{
				SummaryStats: map[string]*SummaryStats{
					startEnd: {
						Span:                       startEnd,
						NumCreatedIssues:           2,
						NumClosedIssues:            2,
						NumGenreNormalIssues:       1,
						NumGenreRequestIssues:      1,
						NumGenreFailureIssues:      0,
						NumEscalationAllIssues:     1,
						NumEscalationNormalIssues:  1,
						NumEscalationRequestIssues: 0,
						NumEscalationFailureIssues: 0,
						NumUrgencyHighIssues:       1,
						NumUrgencyLowIssues:        1,
						NumScoreA:                  0,
						NumScoreB:                  1,
						NumScoreC:                  1,
						NumScoreD:                  0,
						NumScoreE:                  0,
						NumScoreF:                  0,
						NumTotalScore:              2.5,
					},
				},
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 1",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/1",
						CreatedAt:    tenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "低",
						Labels:       "Kubernetes",
						Genre:        "通常問合せ",
						NumComments:  1,
						OpenDuration: 168,
						Escalation:   true,
					},
					1: {
						Title:        "issue 2",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/2",
						CreatedAt:    sevenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "中",
						Labels:       "Openstack",
						Genre:        "要望",
						NumComments:  2,
						OpenDuration: 96,
						Escalation:   false,
					},
				},
			},
			want: `## サマリー 
|項目|startEnd|
|----|----|
|起票件数|2|
|クローズ件数|2|
|緊急度：高・中|1|
|緊急度：低|1|
|全体エスカレーション件数|1|
|全体CaaS-A完結率(％)|50.0|
|通常エスカレーション件数|1|
|通常CaaS-A完結率(％)|50.0|
|ジャンル:通常問合せ件数|1|
|ジャンル:要望件数|1|
|ジャンル:サービス障害件数|0|
|合計スコア|2.50|
|スコアA|0|
|スコアB|1|
|スコアC|1|
|スコアD|0|
|スコアE|0|
|スコアF|0|

## 詳細 
- [issue 1](https://github.com/sataga/issue-warehouse/issues/1),低,通常問合せ,CaaS-A/,comment数:1,経過時間(hour):168,解決フラグ:true,(startEnd)
- [issue 2](https://github.com/sataga/issue-warehouse/issues/2),中,要望,CaaS-A/,comment数:2,経過時間(hour):96,解決フラグ:false,(startEnd)
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lts := &LongTermStats{
				SummaryStats: tt.fields.SummaryStats,
				DetailStats:  tt.fields.DetailStats,
			}
			tt.want = strings.Replace(tt.want, "startEnd", startEnd, -1)
			if got := lts.GenLongTermReport(); got != tt.want {
				t.Errorf("LongTermStats.GenLongTermReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userSupport_GetAnalysisReportStats(t *testing.T) {
	var c *gomock.Controller

	testIssues := []*github.Issue{
		issuePatterns[0],
		issuePatterns[1],
	}
	type fields struct {
		repo Repository
	}
	type args struct {
		since time.Time
		until time.Time
		state string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *AnalysisStats
		wantErr    bool
		beforefunc func(f *fields, state string)
		afterfunc  func()
	}{
		// TODO: Add test cases.
		{
			name: "Normal operation for created",
			args: args{
				since: firstDayOfMonth,
				until: lastDayOfMonth,
				state: "created",
			},
			wantErr: false,
			want: &AnalysisStats{
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 1",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/1",
						CreatedAt:    tenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "低",
						Labels:       "Kubernetes",
						Genre:        "通常問合せ",
						NumComments:  1,
						OpenDuration: 168,
						Escalation:   true,
					},
					1: {
						Title:        "issue 2",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/2",
						CreatedAt:    sevenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "中",
						Labels:       "Openstack",
						Genre:        "要望",
						NumComments:  2,
						OpenDuration: 96,
						Escalation:   false,
					},
				},
			},
			beforefunc: func(f *fields, state string) {
				c = gomock.NewController(t)
				musr := NewMockRepository(c)
				if state == "created" {
					musr.EXPECT().GetCreatedSupportIssues(gomock.Any(), gomock.Any()).Return(testIssues, nil)
				} else {
					musr.EXPECT().GetClosedSupportIssues(gomock.Any(), gomock.Any()).Return(testIssues, nil)
				}
				f.repo = musr
			},
			afterfunc: func() {
				c.Finish()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforefunc != nil {
				tt.beforefunc(&tt.fields, tt.args.state)
			}
			if tt.afterfunc != nil {
				defer tt.afterfunc()
			}
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetAnalysisReportStats(tt.args.since, tt.args.until, tt.args.state)
			// for k, v := range tt.want.DetailStats {
			// 	fmt.Printf("want:%d", k)
			// 	fmt.Println(v)
			// }
			// for k, v := range got.DetailStats {
			// 	fmt.Printf("got :%d", k)
			// 	fmt.Println(v)
			// }
			if (err != nil) != tt.wantErr {
				t.Errorf("userSupport.GetAnalysisReportStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userSupport.GetAnalysisReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalysisStats_GenAnalysisReport(t *testing.T) {
	type fields struct {
		DetailStats map[int]*DetailStats
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "print analysis-report",
			fields: fields{
				DetailStats: map[int]*DetailStats{
					0: {
						Title:        "issue 1",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/1",
						CreatedAt:    tenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "低",
						Labels:       "Kubernetes",
						Genre:        "通常問合せ",
						NumComments:  1,
						OpenDuration: 168,
						Escalation:   true,
					},
					1: {
						Title:        "issue 2",
						HTMLURL:      "https://github.com/sataga/issue-warehouse/issues/2",
						CreatedAt:    sevenDayAgo.Format("2006-01-02"),
						ClosedAt:     threeDayAgo.Format("2006-01-02"),
						State:        "closed",
						TargetSpan:   startEnd,
						TeamName:     "CaaS-A",
						Urgency:      "中",
						Labels:       "Openstack",
						Genre:        "要望",
						NumComments:  2,
						OpenDuration: 96,
						Escalation:   false,
					},
				},
			},
			want: `期間,Title,起票日,クローズ日,ステータス,担当チーム,担当アサイン,緊急度,問い合わせ種別,エスカレ有無,コメント数,経過時間,Keywordラベル,URL
startEnd,issue 1,tenDayAgo,threeDayAgo,closed,CaaS-A,,低,通常問合せ,true,1,168,Kubernetes,https://github.com/sataga/issue-warehouse/issues/1 
startEnd,issue 2,sevenDayAgo,threeDayAgo,closed,CaaS-A,,中,要望,false,2,96,Openstack,https://github.com/sataga/issue-warehouse/issues/2 
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := &AnalysisStats{
				DetailStats: tt.fields.DetailStats,
			}
			tt.want = strings.Replace(tt.want, "startEnd", startEnd, -1)
			tt.want = strings.Replace(tt.want, "threeDayAgo", threeDayAgo.Format("2006-01-02"), -1)
			tt.want = strings.Replace(tt.want, "sevenDayAgo", sevenDayAgo.Format("2006-01-02"), -1)
			tt.want = strings.Replace(tt.want, "tenDayAgo", tenDayAgo.Format("2006-01-02"), -1)
			if got := as.GenAnalysisReport(); got != tt.want {
				t.Errorf("AnalysisStats.GenAnalysisReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userSupport_GetKeywordReportStats(t *testing.T) {
	var c *gomock.Controller

	closeIssues := []*github.Issue{
		issuePatterns[2],
		issuePatterns[3],
	}

	keywords := []*github.LabelResult{
		keywordPatterns[0],
		keywordPatterns[1],
		keywordPatterns[2],
	}
	type fields struct {
		repo Repository
	}
	type args struct {
		since time.Time
		until time.Time
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *KeywordStats
		wantErr    bool
		beforefunc func(f *fields)
		afterfunc  func()
	}{
		// TODO: Add test cases.
		{
			name: "Normal operation for keyword-monthly-report",
			args: args{
				since: firstDayOfMonth,
				until: lastDayOfMonth,
			},
			want: &KeywordStats{
				KeywordSummary: map[string]*KeywordSummary{
					startEnd: {
						Span: startEnd,
						KeywordCountAsAll: map[string]int{
							"keyword:Network":    1,
							"keyword:Openstack":  0,
							"keyword:Kubernetes": 1,
						},
						KeywordCountAsEscalation: map[string]int{
							"keyword:Network":    0,
							"keyword:Openstack":  0,
							"keyword:Kubernetes": 0,
						},
					},
				},
			},
			wantErr: false,
			beforefunc: func(f *fields) {
				c = gomock.NewController(t)
				musr := NewMockRepository(c)
				musr.EXPECT().GetLabelsByQuery(gomock.Any()).Return(keywords, nil)
				musr.EXPECT().GetClosedSupportIssues(gomock.Any(), gomock.Any()).Return(closeIssues, nil)
				f.repo = musr
			},
			afterfunc: func() {
				c.Finish()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforefunc != nil {
				tt.beforefunc(&tt.fields)
			}
			if tt.afterfunc != nil {
				defer tt.afterfunc()
			}
			us := &userSupport{
				repo: tt.fields.repo,
			}
			got, err := us.GetKeywordReportStats(tt.args.since, tt.args.until)
			if (err != nil) != tt.wantErr {
				t.Errorf("userSupport.GetKeywordReportStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userSupport.GetKeywordReportStats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeywordStats_GenKeywordReport(t *testing.T) {
	type fields struct {
		KeywordSummary map[string]*KeywordSummary
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "print keyword-report",
			fields: fields{
				map[string]*KeywordSummary{
					startEnd: {
						Span: startEnd,
						KeywordCountAsAll: map[string]int{
							"keyword:Network":    0,
							"keyword:Openstack":  1,
							"keyword:Kubernetes": 1,
						},
						KeywordCountAsEscalation: map[string]int{
							"keyword:Network":    0,
							"keyword:Openstack":  0,
							"keyword:Kubernetes": 0,
						},
					},
				},
			},
			want: `## サマリー(全体) 
|項目|startEnd|Total|
|----|----|----|
|keyword:Kubernetes|1|1|
|keyword:Network|0|0|
|keyword:Openstack|1|1|
## サマリー(Escalationのみ計上) 
|項目|startEnd|Total|
|----|----|----|
|keyword:Kubernetes|0|0|
|keyword:Network|0|0|
|keyword:Openstack|0|0|
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KeywordStats{
				KeywordSummary: tt.fields.KeywordSummary,
			}
			tt.want = strings.Replace(tt.want, "startEnd", startEnd, -1)
			if got := ks.GenKeywordReport(); got != tt.want {
				t.Errorf("KeywordStats.GenKeywordReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
