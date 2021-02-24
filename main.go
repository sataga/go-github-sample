package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	dus "github.com/sataga/go-github-sample/domain/usersupport"
	igh "github.com/sataga/go-github-sample/infra/github"
	"github.com/sataga/go-github-sample/infra/slack"
	ius "github.com/sataga/go-github-sample/infra/usersupport"
)

var (
	ghURL   = flag.String("ghurl", "https://api.github.com", "GitHub API base URL")
	ghUser  = flag.String("ghuser", "sataga", "Github user name")
	ghMail  = flag.String("ghmail", "", "Github user email")
	ghToken = flag.String("ghtoken", "", "GitHub Personal access token")

	now             = time.Now()
	oneWeekBefore   = now.Add(-7 * 24 * time.Hour)
	userSupportFlag = flag.NewFlagSet("us", flag.ExitOnError)
	sinceStr        = userSupportFlag.String("since", oneWeekBefore.Format("2006-01-02"), "Date since listing issues from")
	untilStr        = userSupportFlag.String("until", now.Format("2006-01-02"), "Date until listing issues from")

	dailyReportFlag = flag.NewFlagSet("daily-report", flag.ExitOnError)
	dayAgoInt       = dailyReportFlag.Int("day-ago", 7, "Please specify a date that has not been updated")

	analysisReportFlag = flag.NewFlagSet("analysys-report", flag.ExitOnError)
	stateStr           = analysisReportFlag.String("state", "created", "Please choose on (created , closed)")
	spanInt            = analysisReportFlag.Int("span", 4, "Please enter the span you want to get")

	keywordReportFlag = flag.NewFlagSet("keyword-report", flag.ExitOnError)
	keywordKindStr    = keywordReportFlag.String("kind", "monthly", "Please choose on (weekly , monthly)")
	keywordSpanInt    = keywordReportFlag.Int("span", 4, "Please enter the span you want to get")
	keywordUntilStr   = keywordReportFlag.String("until", now.Format("2006-01-02"), "Date until listing issue from")
)

func printDefaultsAll() {
	fmt.Println("usage: go-github-sample [global options] subcommand [subcommand options]")
	flag.PrintDefaults()
	fmt.Println("\nsubcommands:")
	fmt.Println("us:    getting user support info")
	userSupportFlag.PrintDefaults()
	fmt.Println("daily-report:    Notify slack of tickets that have not been updated since the specified date")
	dailyReportFlag.PrintDefaults()
	fmt.Println("analysis-report:    Output user information in CSV format")
	analysisReportFlag.PrintDefaults()
}

func main() {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	flag.Parse()
	if os.Getenv("GITHUB_TOKEN") != "" {
		*ghToken = os.Getenv("GITHUB_TOKEN")
	}
	if os.Getenv("GITHUB_MAIL") != "" {
		*ghMail = os.Getenv("GITHUB_MAIL")
	}
	ghcli, err := igh.NewGitHubClient(*ghURL, *ghToken, *ghUser, *ghMail)
	if err != nil {
		log.Fatalf("github client: %s", err)
	}
	subCommandArgs := os.Args[1+flag.NFlag():]
	if len(subCommandArgs) == 0 {
		printDefaultsAll()
		log.Fatalln("specify subcommand")
	}
	switch subCommand := subCommandArgs[0]; subCommand {
	case "daily-report":
		if err := dailyReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing daily report flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		dairyStats, err := us.GetDailyReportStats(now, *dayAgoInt)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%s", dairyStats.GetDailyReportStats())
		// channel := "times_t-sataga"
		// username := "t-sataga"
		// _, err = slack.PostMessage(channel, username, dairyStats.GetDailyReportStats())
		// if err != nil {
		// 	log.Fatalf("slack post message failed: %s", err)
		// }
	case "monthly-report":
		if err := userSupportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing user support flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var until time.Time
		var err error
		if until, err = time.ParseInLocation("2006-01-02", *untilStr, jst); err != nil {
			log.Fatalf("could not parse: %s", *untilStr)
		}
		span := 4
		LongTermStats, err := us.GetLongTermReportStats(until, subCommandArgs[0], span)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%s", LongTermStats.GenLongTermReport())
	case "analysis-report":
		if err := analysisReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing analysis support flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if since, err = time.Parse("2006-01-02", *sinceStr); err != nil {
			log.Fatalf("could not parse: %s", *sinceStr)
		}
		if until, err = time.Parse("2006-01-02", *untilStr); err != nil {
			log.Fatalf("could not parse: %s", *untilStr)
		}
		// fmt.Printf("Reporting Stats From: %s, Until: %s\n", since, until)
		AnalysisStats, err := us.GetAnalysisReportStats(since, until, *stateStr, *spanInt)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%s", AnalysisStats.GenAnalysisReport())
	case "keyword-report":
		if err := keywordReportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing keyword report flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var until time.Time
		var err error
		if until, err = time.ParseInLocation("2006-01-02", *keywordUntilStr, jst); err != nil {
			log.Fatalf("could not parse: %s", err)
		}
		KeywordStats, err := us.GetKeywordReportStats(until, *keywordKindStr, *keywordSpanInt)
		if err != nil {
			log.Fatalf("get keyword stats: %s", err)
		}
		fmt.Printf("%s", KeywordStats.GenKeywordReport())

	case "slacktest":
		channel := "times_t-sataga"
		username := "t-sataga"
		text := "hogehoge"
		_, err := slack.PostMessage(channel, username, text)
		if err != nil {
			log.Fatalf("slack post message failed: %s", err)
		}
	case "methodtest":
		if err := userSupportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing user support flag: %s", err)
		}
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if since, err = time.Parse("2006-01-02", *sinceStr); err != nil {
			log.Fatalf("could not parse: %s", *sinceStr)
		}
		if until, err = time.Parse("2006-01-02", *untilStr); err != nil {
			log.Fatalf("could not parse: %s", *untilStr)
		}
		fmt.Printf("Reporting Stats From: %s, Until: %s\n", since, until)
		testStats, err := us.MethodTest(since, until)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("%v", testStats)
	case "get-pj-board":
		usrepo := ius.NewUserSupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		result := us.GetProjectBoard()
		fmt.Printf("%s", result)
	}

}
