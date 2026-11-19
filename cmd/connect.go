package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Pick a running EC2 instance and open an SSM shell on it",
	RunE: func(cmd *cobra.Command, args []string) error {
		if filter == "" {
			return errors.New("--filter is required")
		}
		return connect()
	},
}

func init() { rootCmd.AddCommand(connectCmd) }

type host struct{ Name, ID, Type, AZ, Private, Public string }

func connect() error {
	for bin, hint := range map[string]string{
		"aws":                    "install AWS CLI v2: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html",
		"session-manager-plugin": "install the Session Manager plugin: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html",
	} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("%s not found on PATH; %s", bin, hint)
		}
	}
	sess, err := newSession()
	if err != nil {
		return err
	}
	input := &ec2.DescribeInstancesInput{Filters: []*ec2.Filter{
		{Name: aws.String("instance-state-name"), Values: aws.StringSlice([]string{"running"})},
	}}
	var hosts []host
	done := loading("running instances")
	err = ec2.New(sess).DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, _ bool) bool {
		for _, r := range page.Reservations {
			for _, i := range r.Instances {
				h := host{Name: nameTag(i.Tags), ID: aws.StringValue(i.InstanceId), Type: aws.StringValue(i.InstanceType),
					Private: orDash(i.PrivateIpAddress), Public: orDash(i.PublicIpAddress)}
				if i.Placement != nil {
					h.AZ = aws.StringValue(i.Placement.AvailabilityZone)
				}
				if matches(h.Name, h.ID, h.Private, h.Public) {
					hosts = append(hosts, h)
				}
			}
		}
		return true
	})
	done()
	if err != nil {
		return fmt.Errorf("describing instances: %w", err)
	}
	if len(hosts) == 0 {
		return fmt.Errorf("no running instances match %q in %s", filter, region)
	}
	sort.SliceStable(hosts, func(a, b int) bool { return strings.ToLower(hosts[a].Name) < strings.ToLower(hosts[b].Name) })

	h := hosts[0]
	if len(hosts) == 1 {
		fmt.Printf("%s %s %s\n", color.GreenString("✔"), bold(h.Name), dim(h.ID+"  "+h.Private))
	} else {
		w := 0
		for _, x := range hosts {
			w = max(w, len(x.Name))
		}
		row := fmt.Sprintf(`{{ printf "%%-%ds" .Name }}  {{ .ID | faint }}  {{ .Private | faint }}`, w)
		idx, _, err := (&promptui.Select{
			Label:    fmt.Sprintf("%d instances · ↑/↓ move · / search · enter select", len(hosts)),
			Items:    hosts,
			Size:     15,
			HideHelp: true,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . | bold }}",
				Active:   "▸ " + strings.Replace(row, ".Name }}", ".Name | cyan | bold }}", 1),
				Inactive: "  " + row,
				Selected: `{{ "✔" | green }} {{ .Name | bold }} {{ .ID | faint }}`,
				Details: `
{{ "Type:" | faint }} {{ .Type }}   {{ "AZ:" | faint }} {{ .AZ }}   {{ "Private:" | faint }} {{ .Private }}   {{ "Public:" | faint }} {{ .Public }}`,
			},
			Searcher: func(in string, i int) bool {
				x := hosts[i]
				return strings.Contains(strings.ToLower(x.Name+" "+x.ID+" "+x.Private+" "+x.Public), strings.ToLower(in))
			},
		}).Run()
		if err != nil {
			return err
		}
		h = hosts[idx]
	}
	_, user, err := (&promptui.Select{
		Label:     "Log in as",
		Items:     []string{"ubuntu", "root", "ec2-user", "admin", "centos", "debian"},
		Templates: &promptui.SelectTemplates{Label: "{{ . | bold }}", Active: "▸ {{ . | cyan | bold }}", Inactive: "  {{ . }}", Selected: `{{ "✔" | green }} {{ . | bold }}`},
	}).Run()
	if err != nil {
		return err
	}

	args := []string{"ssm", "start-session",
		"--target", h.ID,
		"--document-name", "AWS-StartInteractiveCommand",
		"--parameters", fmt.Sprintf(`command=["sudo su - %s"]`, user),
		"--region", region}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	for {
		fmt.Println(dim("$ aws " + strings.Join(args, " ")))
		c := exec.Command("aws", args...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			fmt.Println(dim("session ended: " + err.Error()))
		}
		// Can't intercept exit inside the remote shell, so offer a one-keystroke way back instead.
		_, err := (&promptui.Prompt{Label: fmt.Sprintf("Reconnect to %s as %s", h.Name, user), IsConfirm: true, Default: "y"}).Run()
		if err == promptui.ErrAbort || err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reconnect prompt: %w", err)
		}
	}
}
