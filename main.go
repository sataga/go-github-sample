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
)

func printDefaultsAll() {
	fmt.Println("usage: go-github-sample [global options] subcommand [subcommand options]")
	flag.PrintDefaults()
	fmt.Println("\nsubcommands:")
	fmt.Println("us:    getting user support info")
	userSupportFlag.PrintDefaults()
}

func main() {
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
	case "us":
		if err := userSupportFlag.Parse(subCommandArgs[1:]); err != nil {
			log.Fatalf("parsing user support flag: %s", err)
		}
		usrepo := ius.NewUsersupportRepository(ghcli)
		us := dus.NewUserSupport(usrepo)
		var since, until time.Time
		var err error
		if since, err = time.Parse("2006-01-02", *sinceStr); err != nil {
			log.Fatalf("could not parse: %s", *sinceStr)
		}
		if until, err = time.Parse("2006-01-02", *untilStr); err != nil {
			log.Fatalf("could not parse: %s", *untilStr)
		}
		// 終日までのIssueをカウントするための下処理
		until = until.AddDate(0, 0, 1)
		until = until.Add(-time.Minute)
		usStats, err := us.GetUserSupportStats(since, until)
		if err != nil {
			log.Fatalf("get user support stats: %s", err)
		}
		fmt.Printf("UserSupportStats From: %s, Until: %s\n", since, until)
		fmt.Printf("%s", usStats.GenReport())
	case "slacktest":
		channel := "times_t-sataga"
		username := "t-sataga"
		text := "hogehoge"
		_, err := slack.PostMessage(channel, username, text)
		if err != nil {
			log.Fatalf("slack post message failed: %s", err)
		}
	}
}
