package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:       "list {ec2|lb}",
	Short:     "List AWS resources",
	ValidArgs: []string{"ec2", "lb"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("expected 1 arg, got %d %v; an unquoted * gets expanded by your shell: quote it (\"*\") or drop --filter", len(args), args)
		}
		return cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] == "lb" {
			return listLBs()
		}
		return listEC2()
	},
}

func init() { rootCmd.AddCommand(listCmd) }

// printTable sorts rows by the sort column (case-insensitive), colors the color columns, and prints.
func printTable(header []interface{}, rows [][]string, sortCol int, colorCols ...int) {
	sort.SliceStable(rows, func(a, b int) bool {
		return strings.ToLower(rows[a][sortCol]) < strings.ToLower(rows[b][sortCol])
	})
	tbl := table.New(header...).
		WithHeaderFormatter(color.New(color.FgCyan, color.Bold).SprintfFunc()).
		WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc()).
		WithWidthFunc(visibleWidth)
	for _, r := range rows {
		cells := make([]interface{}, len(r))
		for i, v := range r {
			cells[i] = v
			if v == "-" {
				cells[i] = dim(v)
			}
		}
		for _, c := range colorCols {
			cells[c] = stateColor(r[c])
		}
		tbl.AddRow(cells...)
	}
	tbl.Print()
}

func listLBs() error {
	sess, err := newSession()
	if err != nil {
		return err
	}
	done := loading("load balancers")
	var rows [][]string
	add := func(name, dns, typ, scheme, vpc *string) {
		if matches(aws.StringValue(name), aws.StringValue(dns)) {
			rows = append(rows, []string{orDash(name), orDash(typ), orDash(scheme), orDash(vpc), orDash(dns)})
		}
	}
	err = elb.New(sess).DescribeLoadBalancersPages(&elb.DescribeLoadBalancersInput{},
		func(p *elb.DescribeLoadBalancersOutput, _ bool) bool {
			for _, lb := range p.LoadBalancerDescriptions {
				add(lb.LoadBalancerName, lb.DNSName, aws.String("classic"), lb.Scheme, lb.VPCId)
			}
			return true
		})
	if err == nil {
		// One elbv2 call covers application, network and gateway LBs; Type says which.
		err = elbv2.New(sess).DescribeLoadBalancersPages(&elbv2.DescribeLoadBalancersInput{},
			func(p *elbv2.DescribeLoadBalancersOutput, _ bool) bool {
				for _, lb := range p.LoadBalancers {
					add(lb.LoadBalancerName, lb.DNSName, lb.Type, lb.Scheme, lb.VpcId)
				}
				return true
			})
	}
	done()
	if err != nil {
		return fmt.Errorf("describing load balancers: %w", err)
	}
	if len(rows) > 0 {
		printTable([]interface{}{"Name", "Type", "Scheme", "VPC", "DNS Name"}, rows, 0)
	}
	footer(len(rows), "load balancers", "")
	return nil
}

func listEC2() error {
	sess, err := newSession()
	if err != nil {
		return err
	}
	svc := ec2.New(sess)
	done := loading("EC2 instances")

	// Status checks only exist for running instances; anything absent shows "-".
	healthy := map[string]bool{}
	err = svc.DescribeInstanceStatusPages(&ec2.DescribeInstanceStatusInput{},
		func(p *ec2.DescribeInstanceStatusOutput, _ bool) bool {
			for _, s := range p.InstanceStatuses {
				healthy[aws.StringValue(s.InstanceId)] =
					aws.StringValue(s.SystemStatus.Status) == "ok" && aws.StringValue(s.InstanceStatus.Status) == "ok"
			}
			return true
		})
	var rows [][]string
	states := map[string]int{}
	if err == nil {
		err = svc.DescribeInstancesPages(&ec2.DescribeInstancesInput{},
			func(p *ec2.DescribeInstancesOutput, _ bool) bool {
				for _, r := range p.Reservations {
					for _, i := range r.Instances {
						id := aws.StringValue(i.InstanceId)
						status := "-"
						if ok, found := healthy[id]; found {
							status = map[bool]string{true: "Ok", false: "Failed"}[ok]
						}
						row := []string{nameTag(i.Tags), id, aws.StringValue(i.State.Name), aws.StringValue(i.InstanceType),
							orDash(i.PrivateIpAddress), orDash(i.PublicIpAddress), status}
						if matches(row...) {
							rows = append(rows, row)
							states[row[2]]++
						}
					}
				}
				return true
			})
	}
	done()
	if err != nil {
		return fmt.Errorf("describing instances: %w", err)
	}
	if len(rows) > 0 {
		printTable([]interface{}{"Name", "ID", "State", "Type", "Private IP", "Public IP", "Checks"}, rows, 0, 2, 6)
	}
	var parts []string
	for _, s := range []string{"running", "pending", "stopping", "stopped", "shutting-down", "terminated"} {
		if n := states[s]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, stateColor(s)))
		}
	}
	extra := ""
	if len(parts) > 0 {
		extra = dim(" · ") + strings.Join(parts, dim(" · "))
	}
	footer(len(rows), "instances", extra)
	return nil
}
