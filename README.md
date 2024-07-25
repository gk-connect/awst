

```markdown
# AWST (Amazon Web Services Tool)

AWST is a command-line tool for managing AWS resources. It provides various commands to interact with AWS services.

## Installation

To install AWST, run:
```
```bash
go build -o awst
sudo cp awst /usr/local/bin/awst
```

## Usage

### List EC2 Instances

To list EC2 instances, use the following command:

```bash
awst list ec2 --profile <aws_config_profile> --filter <filter-text>
```

### Connect to an EC2 Instance

To connect to an EC2 instance, use the following command:

```bash
awst connect --profile <aws_config_profile> --filter <filter-text>
```

#### Optional Args:

- `--region`: Specify the AWS region. Default is `ap-south-1`.
- `--ssh-ip-type`: Specify the type of IP to use for SSH. Default is `private`.
- `--key`: Specify the keypair to use for SSH authentication. Only required if the instance doesn't have a keypair attached during launch or you want to use a different key. Default is the key pair provided during instance launch.
- `--key-path`: Specify the folder where your SSH key resides. Default is `~/.ssh/`.

## Examples

List all EC2 instances:

```bash
awst list ec2 --profile myprofile --filter *
```

Connect to an EC2 instance using a specific key:

```bash
awst connect --profile myprofile --filter myinstance --key mykey --key-path /path/to/key
```
### List Load Balancers
```bash
awst list lb --profile <myprofile> --region <region> --filter <filter>
```

## Contributing

Contributions are welcome! Please fork this repository and open a pull request with your changes.

#Gopu Krishnan

