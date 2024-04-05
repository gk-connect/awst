package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var errorColor = color.New(color.FgRed)
var successColor = color.New(color.FgGreen)
var region string
var profile string
var filter string

type LoadBalancer struct {
	Name    string
	DNSName string
	Type    string
	Scheme  string
	VPC     string
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS resources",
	Long:  `list/get ec2/rds`,
	Run: func(cmd *cobra.Command, args []string) {

		if region == "" || profile == "" {
			errorColor.Println("Usage: awst list ec2 --region <region> --profile <profile>")
			return
		}
		if len(args) > 0 {
			if args[0] == "ec2" {
				getEc2List()
			} else if args[0] == "rds" {
				successColor.Printf("Invoked rds")
			} else if args[0] == "lb" {
				getLbList()
			} else {
				errorColor.Printf("Usage: awst list <ec2/rds> <region> <profile>")
			}
		} else {
			errorColor.Printf("Usage: awst list <ec2/rds> <region> <profile>")
		}
	},
}

func init() {
	defaultRegion := "ap-south-1"
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&region, "region", "r", defaultRegion, "AWS region")
	listCmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS profile")
	listCmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter")

}

func getLbList() {
	loadBalancerDetails, err := getAllLoadBalancers(profile, region)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Name", "DNS Name", "Type", "Scheme", "VPC")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, lb := range loadBalancerDetails {
		if len(filter) > 0 {
			if strings.Contains(strings.ToLower(lb.Name), strings.ToLower(filter)) {
				tbl.AddRow(lb.Name, lb.DNSName, lb.Type, lb.Scheme, lb.VPC)
			}
		} else {
			tbl.AddRow(lb.Name, lb.DNSName, lb.Type, lb.Scheme, lb.VPC)
		}

	}
	tbl.Print()

}

func getAllLoadBalancers(profile, region string) ([]LoadBalancer, error) {
	// Create a new session with AWS credentials and configuration
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Initialize slice to store load balancers
	var lbDetails []LoadBalancer

	// Describe Classic Load Balancers
	classicLBs, err := describeClassicLoadBalancers(sess)
	if err != nil {
		return nil, err
	}
	lbDetails = append(lbDetails, classicLBs...)

	// Describe Application Load Balancers
	albList, err := describeApplicationLoadBalancers(sess)
	if err != nil {
		return nil, err
	}
	lbDetails = append(lbDetails, albList...)

	// Describe Network Load Balancers
	nlbList, err := describeNetworkLoadBalancers(sess)
	if err != nil {
		return nil, err
	}
	lbDetails = append(lbDetails, nlbList...)

	return lbDetails, nil
}

func describeClassicLoadBalancers(sess *session.Session) ([]LoadBalancer, error) {
	// Create a new ELB client
	svc := elb.New(sess)

	// Describe Classic Load Balancers
	input := &elb.DescribeLoadBalancersInput{}
	result, err := svc.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}

	// Extract Classic Load Balancer details from the result
	var lbDetails []LoadBalancer
	for _, lb := range result.LoadBalancerDescriptions {
		lbDetails = append(lbDetails, LoadBalancer{
			Name:    *lb.LoadBalancerName,
			DNSName: *lb.DNSName,
			Type:    "Classic",
			Scheme:  *lb.Scheme,
			VPC:     *lb.VPCId,
		})
	}

	return lbDetails, nil
}

func describeApplicationLoadBalancers(sess *session.Session) ([]LoadBalancer, error) {
	// Create a new ELBV2 client
	svc := elbv2.New(sess)

	// Describe Application Load Balancers
	input := &elbv2.DescribeLoadBalancersInput{}
	result, err := svc.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}

	// Extract Application Load Balancer details from the result
	var lbDetails []LoadBalancer
	for _, lb := range result.LoadBalancers {
		lbDetails = append(lbDetails, LoadBalancer{
			Name:    *lb.LoadBalancerName,
			DNSName: *lb.DNSName,
			Type:    "Application",
			Scheme:  *lb.Scheme,
			VPC:     *lb.VpcId,
		})
	}

	return lbDetails, nil
}

func describeNetworkLoadBalancers(sess *session.Session) ([]LoadBalancer, error) {
	// Create a new ELBV2 client
	svc := elbv2.New(sess)

	// Describe Network Load Balancers
	input := &elbv2.DescribeLoadBalancersInput{}
	result, err := svc.DescribeLoadBalancers(input)
	if err != nil {
		return nil, err
	}

	// Extract Network Load Balancer details from the result
	var lbDetails []LoadBalancer
	for _, lb := range result.LoadBalancers {
		lbDetails = append(lbDetails, LoadBalancer{
			Name:    *lb.LoadBalancerName,
			DNSName: *lb.DNSName,
			Type:    "Network",
			Scheme:  *lb.Scheme,
			VPC:     *lb.VpcId,
		})
	}

	return lbDetails, nil
}

func getEc2List() {
	awsProfile := profile
	awsRegion := region
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: awsProfile,
		Config: aws.Config{
			Region:                        aws.String(awsRegion),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{}
	var allInstances []*ec2.Instance

	// Iterate through pages until there are no more pages left
	err := svc.DescribeInstancesPages(input,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			// Append instances from current page to allInstances slice
			for _, reservation := range page.Reservations {
				allInstances = append(allInstances, reservation.Instances...)

			}
			return !lastPage
		})
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}

	// Create a map to store instance IDs
	instanceIds := make([]*string, 0, len(allInstances))

	// fmt.Println(len(allInstances))
	for _, instance := range allInstances {
		instanceIds = append(instanceIds, instance.InstanceId)
	}
	// Process instance statuses
	instanceStatusMap := make(map[string]bool)
	// Paginate through instance IDs to describe their statuses
	var allRows [][]string
	for i := 0; i < len(instanceIds); i += 100 {
		end := i + 100
		if end > len(instanceIds) {
			end = len(instanceIds)
		}
		sliceOfIDs := instanceIds[i:end]

		// fmt.Println(len(sliceOfIDs))

		// Describe instance status
		statusInput := &ec2.DescribeInstanceStatusInput{
			InstanceIds: sliceOfIDs,
		}
		statusResult, err := svc.DescribeInstanceStatus(statusInput)
		if err != nil {
			fmt.Println("Error describing instance status:", err)
			return
		}

		for _, status := range statusResult.InstanceStatuses {
			// fmt.Println(*status.InstanceId)
			instanceStatusMap[*status.InstanceId] = (*status.SystemStatus.Status == "ok" && *status.InstanceStatus.Status == "ok")
		}

		status := "Failed"
		// Append instance details with their statuses to rows
		// fmt.Print("Instances length: ")
		// fmt.Println(len(allInstances))
		allRows = make([][]string, 0)
		for _, instance := range allInstances {

			instanceName := "-"
			if len(instance.Tags) > 0 {
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" {
						instanceName = *tag.Value
						break
					}
				}
			}
			publicIp := "-"
			if instance.PublicIpAddress != nil {
				publicIp = *instance.PublicIpAddress
			}
			privateIp := "-"
			if instance.PrivateIpAddress != nil {
				privateIp = *instance.PrivateIpAddress
			}
			status = "Failed"
			if _, ok := instanceStatusMap[*instance.InstanceId]; ok {
				status = "Ok"
			}
			statusColor := color.New(color.FgRed) // Default color is red for status "false"
			if status == "Ok" {
				statusColor = color.New(color.FgGreen) // Change color to green if status is "true"
			}

			if len(filter) > 0 {
				if (strings.Contains(strings.ToLower(*instance.InstanceId), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(instanceName), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(*instance.State.Name), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(*instance.InstanceType), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(privateIp), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(publicIp), strings.ToLower(filter))) ||
					(strings.Contains(strings.ToLower(status), strings.ToLower(filter))) {
					row := []string{*instance.InstanceId, instanceName, *instance.State.Name, *instance.InstanceType, privateIp, publicIp, statusColor.Sprint(status)}
					allRows = append(allRows, row)
				}

			} else {
				row := []string{*instance.InstanceId, instanceName, *instance.State.Name, *instance.InstanceType, privateIp, publicIp, statusColor.Sprint(status)}
				allRows = append(allRows, row)
			}

		}

		// fmt.Println(len(allRows))
	}

	// Print the table after all instances have been processed
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ID", "Name", "State", "Type", "Private IP", "Public IP", "Status")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, row := range allRows {
		// Convert the []string slice to []interface{} before adding it to the table
		interfaceRow := make([]interface{}, len(row))
		for i, v := range row {
			interfaceRow[i] = v
		}
		tbl.AddRow(interfaceRow...)
	}
	tbl.Print()
}
