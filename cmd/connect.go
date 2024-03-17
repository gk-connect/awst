package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var connectIp string
var connectKey string
var connectKeyPath string

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "A brief description of your command",
	Long:  "connect to aws ec2 isntances",
	Run: func(cmd *cobra.Command, args []string) {
		if connectIp != "private" && connectIp != "public" {
			errorColor.Println("Available SSH Ip options are private/public")
			return
		}

		if !(len(filter) > 0) {
			errorColor.Println("Use --filter flag with valid data")
			return
		}
		getEc2Instances()

	},
}

func init() {
	defaultRegion := "ap-south-1"
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVarP(&profile, "profile", "p", "", "AWS profile")
	connectCmd.Flags().StringVarP(&region, "region", "r", defaultRegion, "AWS region")
	connectCmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter")
	connectCmd.Flags().StringVarP(&connectIp, "ssh-ip-type", "i", "private", "SSH Ip Adress")
	connectCmd.Flags().StringVarP(&connectKey, "key", "k", "NONE", "SSH Key")
	connectCmd.Flags().StringVarP(&connectKeyPath, "key-path", "l", "~/.ssh/", "SSH Key Path")
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
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String("*" + strings.ToUpper(filter) + "*"),
					aws.String("*" + strings.ToLower(filter) + "*"),
					aws.String("*" + strings.ToTitle(filter) + "*"),
				},
			},
		},
	}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}

	// fmt.Println(result)

	ec2List := []string{}
	ec2SshKey := []string{}
	ec2Ip := []string{}
	ec2Data := make(map[string][]string)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceName := "-"
			// fmt.Println(*instance.InstanceId)
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

			if connectIp == "public" {
				ec2Ip = append(ec2Ip, publicIp)
			} else {
				ec2Ip = append(ec2Ip, privateIp)
			}
			if instance.KeyName != nil {
				ec2SshKey = append(ec2SshKey, *instance.KeyName)
			} else {
				ec2SshKey = append(ec2SshKey, connectKey)
			}

			ec2List = append(ec2List, instanceName+" ["+*instance.InstanceId+"]  => "+publicIp+" / "+privateIp)

		}
	}
	ec2Data["ec2_list"] = ec2List
	ec2Data["ec2_ip"] = ec2Ip
	ec2Data["ec2_key"] = ec2SshKey

	interactivePrompt(ec2Data)

}

func interactivePrompt(ec2Data map[string][]string) {
	ec2_prompt := promptui.Select{
		Label: "Select your instance to SSH",
		Items: ec2Data["ec2_list"],
	}

	_, ec2Item, err := ec2_prompt.Run()
	if err != nil {
		errorColor.Printf("Prompt failed %v\n", err)
		return
	}

	selectItemIndex := findIndex(ec2Data["ec2_list"], ec2Item)

	user_prompt := promptui.Select{
		Label: "Select your instance to SSH",
		Items: []string{"ubuntu", "root", "ec2-user", "admin", "centos", "debian"},
	}

	_, ec2User, err := user_prompt.Run()
	if err != nil {
		errorColor.Printf("Prompt failed %v\n", err)
		return
	}

	initateSshSession(ec2Data["ec2_key"][selectItemIndex]+".pem", ec2User, ec2Data["ec2_ip"][selectItemIndex])

}

func initateSshSession(sshKey string, sshUser string, sshIp string) {

	if sshKey == "NONE.pem" {
		fmt.Println("No Valid Key Found")
		return
	}

	successColor.Println("Executing.. ssh -i " + connectKeyPath + sshKey + " " + sshUser + "@" + sshIp)

	cmd := "ssh"

	// Arguments for the command
	args := []string{"-i", connectKeyPath + sshKey, sshUser + "@" + sshIp}

	// Execute the command
	session := exec.Command(cmd, args...)
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	err := session.Run()
	if err != nil {
		errorColor.Println("Error:", err)
		return
	}
}
func findIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1 // Item not found
}
