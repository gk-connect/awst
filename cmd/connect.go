package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "A brief description of your command",
	Long:  "connect to aws ec2 isntances",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("connect called")
		fmt.Println(profile)
		getEc2Instances()
	},
}

func init() {
	defaultRegion := "ap-south-1"
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS profile")
	connectCmd.Flags().StringVarP(&region, "region", "r", defaultRegion, "AWS region")
	connectCmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter")
}

func getEc2Instances() {
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
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("*JENKINS*")},
			},
		},
	}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}

	ec2List := []string{}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Println("Instance ID:", *instance.InstanceId)
			fmt.Println("Instance State:", *instance.State.Name)

			ec2List = append(ec2List, *instance.InstanceId)
			// Add more details if needed
		}
	}

	interactivePrompt(ec2List)

}

func interactivePrompt(ec2List []string) {
	prompt := promptui.Select{
		Label: "Select your favorite programming language",
		Items: ec2List,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)

		return
	}

	fmt.Printf("You selected: %s\n", result)
	initateSshSession("PROD.pem", "ubuntu", "110.130.345.221")

}

func initateSshSession(sshKey string, sshUser string, sshIp string) {
	cmd := "ssh"

	// Arguments for the command
	args := []string{"-i", "~/.ssh/" + sshKey, sshUser + "@" + sshIp}

	// Execute the command
	session := exec.Command(cmd, args...)
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err := session.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
