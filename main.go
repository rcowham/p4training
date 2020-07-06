package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/rcowham/p4training/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"fmt"
)

const defaultRegion = "eu-west-1"
const defaultSecurityGroup = "secgroup-training-windows"
const defaultAMI = "ami-0121afa3964191d8a"
const defaultInstanceSize = "t2.small" // Better performance than micro, even if not free

func main() {

	var (
		username = kingpin.Flag(
			"username",
			"Username to create for",
		).Short('u').String()
		email = kingpin.Flag(
			"email",
			"Users email",
		).Short('e').String()
		shortcode = kingpin.Flag(
			"shortcode",
			"Shortcode for course",
		).Short('s').String()
	)

	kingpin.Version(version.Print("p4training"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(defaultRegion)},
	)

	// Create EC2 service client
	svc := ec2.New(sess)

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
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:          aws.String(defaultAMI),
		InstanceType:     aws.String(defaultInstanceSize),
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
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")
}
