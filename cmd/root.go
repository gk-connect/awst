package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	region  string
	profile string
	filter  string
)

var rootCmd = &cobra.Command{
	Use:          "awst",
	Short:        "Quick AWS helpers: list EC2/LBs, connect to EC2 via SSM",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile (default: SDK credential chain)")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "ap-south-1", "AWS region")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "Case-insensitive filter")
}

func newSession() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region), CredentialsChainVerboseErrors: aws.Bool(true)},
		SharedConfigState: session.SharedConfigEnable,
	})
}

func matches(fields ...string) bool {
	if filter == "" || filter == "*" {
		return true
	}
	f := strings.ToLower(filter)
	for _, s := range fields {
		if strings.Contains(strings.ToLower(s), f) {
			return true
		}
	}
	return false
}

func nameTag(tags []*ec2.Tag) string {
	for _, t := range tags {
		if aws.StringValue(t.Key) == "Name" {
			return aws.StringValue(t.Value)
		}
	}
	return "-"
}

func orDash(s *string) string {
	if v := aws.StringValue(s); v != "" {
		return v
	}
	return "-"
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleWidth ignores color codes so colored cells don't break table alignment.
func visibleWidth(s string) int { return utf8.RuneCountInString(ansi.ReplaceAllString(s, "")) }

var (
	dim  = color.New(color.Faint).SprintFunc()
	bold = color.New(color.Bold).SprintFunc()
)

func stateColor(state string) string {
	switch state {
	case "running", "active", "Ok":
		return color.GreenString(state)
	case "stopped", "stopping", "provisioning":
		return color.YellowString(state)
	case "pending":
		return color.CyanString(state)
	case "-":
		return dim(state)
	default: // terminated, shutting-down, failed, Failed
		return color.RedString(state)
	}
}

// loading prints a status line on stderr (TTY only) and returns a func that erases it.
func loading(what string) func() {
	if color.NoColor {
		return func() {}
	}
	who := profile
	if who == "" {
		who = "default creds"
	}
	fmt.Fprintf(os.Stderr, "%s %s %s\n", dim("⋯ fetching"), what, dim("("+region+", "+who+")"))
	return func() { fmt.Fprint(os.Stderr, "\033[1A\033[2K") }
}

// footer prints "N things" plus a filter hint, or a no-match message.
func footer(n int, noun, extra string) {
	f := ""
	if filter != "" && filter != "*" {
		f = fmt.Sprintf(" matching %q", filter)
	}
	if n == 0 {
		fmt.Println(color.YellowString("No %s%s in %s.", noun, f, region))
		return
	}
	fmt.Printf("\n%s%s%s\n", bold(fmt.Sprintf("%d %s", n, noun)), f, extra)
}
