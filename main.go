package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/rcowham/p4training/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"fmt"
)

const defaultRegion = "eu-west-1"
const defaultSecurityGroupWindows = "secgroup-training-windows"
const defaultSecurityGroupLinux = "secgroup-training-linux"
const defaultAMIWindows = "ami-099c3bc9f4d739cae"
const defaultAMILinux = "ami-07f014eead871d2d6"
const defaultInstanceSizeWindows = "t2.medium" // Better performance than micro, even if not free
const defaultInstanceSizeLinux = "t3.micro"
const defaultKeynameWindows = ""
const defaultKeynameLinux = "p4training2"

func main() {

	var (
		amiDescription = fmt.Sprintf("AMI instance ID to use (note these vary per region and per OS type). Defaults Windows (%s), Linux (%s)", defaultAMIWindows, defaultAMILinux)
		username       = kingpin.Flag(
			"username",
			"Username to create for (defaults to email if not specified)",
		).Short('u').String()
		email = kingpin.Flag(
			"email",
			"Users email",
		).Short('e').Required().String()
		shortcode = kingpin.Flag(
			"shortcode",
			"Shortcode for course",
		).Short('s').Required().String()
		AMI = kingpin.Flag(
			"instance",
			amiDescription,
		).Short('i').String()
		uselinux = kingpin.Flag(
			"linux",
			"Create Linux VM (otherwise Windows by default)",
		).Short('l').Bool()
	)

	kingpin.Version(version.Print("p4training"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(defaultRegion)},
	)

	// Create EC2 service client
	svc := ec2.New(sess)

	defaultSecurityGroup := defaultSecurityGroupWindows
	defaultAMI := defaultAMIWindows
	defaultInstanceSize := defaultInstanceSizeWindows
	defaultKeyname := defaultKeynameWindows
	if *uselinux {
		defaultSecurityGroup = defaultSecurityGroupLinux
		defaultAMI = defaultAMILinux
		defaultInstanceSize = defaultInstanceSizeLinux
		defaultKeyname = defaultKeynameLinux
	}
	if *AMI != "" {
		defaultAMI = *AMI
	}
	// Retrieve the security group descriptions
	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupNames: aws.StringSlice([]string{defaultSecurityGroup}),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidGroupId.Malformed":
				fallthrough
			case "InvalidGroup.NotFound":
				fmt.Printf("%s.\n", aerr.Message())
			}
		}
		fmt.Printf("Unable to get descriptions for security groups, %v", err)
	}

	var grpID string
	fmt.Println("Security Group:")
	for _, group := range result.SecurityGroups {
		fmt.Printf("Group: %v\n", group)
		grpID = *group.GroupId
	}
	fmt.Printf("GroupID: %v\n", grpID)

	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(defaultAMI),
		InstanceType:     aws.String(defaultInstanceSize),
		KeyName:          aws.String(defaultKeyname),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{grpID}),
	})
	if err != nil {
		fmt.Println("Could not create instance", err)
		return
	}

	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)
	fmt.Printf("Details: %v\n", *runResult)
	// fmt.Println("IP address", *runResult.Instances[0].PublicIpAddress)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(fmt.Sprintf("%s#%s", *shortcode, *email)),
			},
			{
				Key:   aws.String("P4Training"),
				Value: aws.String("Yes"),
			},
			{
				Key:   aws.String("Username"),
				Value: aws.String(*username),
			},
			{
				Key:   aws.String("Course"),
				Value: aws.String(*shortcode),
			},
			{
				Key:   aws.String("Owner"),
				Value: aws.String(os.Getenv("USER")),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")
}
