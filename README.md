# p4training

Utilities to work with p4training instances - creation etc.

You must be setup with an AWS account under `p4consulting` parent. Contact Tom Tyler if you do not have this setup.

# Pre-requisites

Please ensure you have the [AWS CLI installed](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

Having logged in to the AWS console, click on your username > My Security Credentials. You need to have created and saved an "Access Key for CLI, SDK and API access"

You need to make sure you have your [AWS credentials also installed](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

The above suggests running the following, together with information from your Access Key:

    aws configure

For now, AMI and instances are published in the `eu-west-1` region (Ireland) - so specify that along with the other 

The following command should run successfully after you have installed AWS CLI:

    aws ec2 describe-instances

# Download p4training

Check the [releases tab](https://github.com/rcowham/p4training/releases) for the project and download the latest executable for your operating system.

Execute help:

```bash
 ./p4training -h
usage: p4training --email=EMAIL --shortcode=SHORTCODE [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
  -u, --username=USERNAME    Username to create for (defaults to email if not specified)
  -e, --email=EMAIL          Users email
  -s, --shortcode=SHORTCODE  Shortcode for course
  -i, --instance=INSTANCE    AMI instance ID to use (note these vary per region and per OS type). Defaults Windows (ami-099c3bc9f4d739cae), Linux
                             (ami-07f014eead871d2d6)
  -l, --linux                Create Linux VM (otherwise Windows by default)
      --version              Show application version.
```

To create an AWS instance for a particular user (first is Windows, second Linux):

    ./p4training -s SOME-COURSE -e fred@example.com -u "Fred Bloggs"
    ./p4training -s SOME-COURSE --linux -e jim@example.com

Note that you only need to quote arguments when they have spaces in them - such as username above.

These will create EC2 instances with a names `SOME-COURSE#fred@example.com` with tag showing username and `SOME-COURSE#joe@example.com`

It is easy to wrap the above in a Bash or Windows to create multiple users. E.g.

```bash
echo "fred@example.com,Fred Bloggs" >> users.csv
echo "jim@example.com,Jim Jones" >> users.csv

cat users.csv | while IFS="," read -e email user; do ./p4training -s "MY-COURSE" -u "$user" -e "$email"; done
```

# Finding created instances

This command will find all instances with tag value of "MY-COURSE" as specified in shortcode above:

    aws ec2 describe-instances --filters Name=tag-value,Values=MY-COURSE --query "Reservations[*].Instances[*].[InstanceId,PublicIpAddress,Tags[*]]" --output text > instances.txt

It is then easy to parse out the results, for example:

    grep -E "^i\-|Username" instances.txt | paste - - | awk '{print $4, $2}'
