# AWST (Amazon Web Services Tool)

Small CLI for everyday AWS chores: list EC2 instances and load balancers, and open a shell on an EC2 instance through SSM Session Manager. No SSH keys or open ports needed.

## Requirements

- Go 1.22+ (to build)
- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html) (for `connect`). On macOS: `brew install --cask session-manager-plugin`.
  Check that it is on your company's approved software list first.
- Target instances must run the SSM agent and have an instance profile that allows SSM.
- Your IAM identity needs `ec2:DescribeInstances`, `ec2:DescribeInstanceStatus`, `elasticloadbalancing:DescribeLoadBalancers` and `ssm:StartSession`.

## Install

```bash
go build -o awst .
sudo cp awst /usr/local/bin/awst
```

Or run it from the repo without installing: `go run . <command> [flags]` (don't add `awst` after `go run .`).

## Global flags

| Flag | Default | |
|---|---|---|
| `-p, --profile` | SDK credential chain (env vars, SSO, default profile) | AWS profile from `~/.aws/config` |
| `-r, --region` | `ap-south-1` | AWS region |
| `-f, --filter` | none | Case-insensitive substring match. Omit it (or quote `"*"`) to show everything |

> Quote wildcards: an unquoted `*` gets expanded by your shell into file names.

## Commands

### `awst list ec2`

```bash
awst list ec2 -p myprofile -f web
```

Lists instances sorted by name. The filter matches name, ID, state, type, IPs and status checks.

```
Name   ID    State    Type       Private IP  Public IP  Checks
api    i-0a  running  t3.micro   10.0.1.5    13.1.1.1   Ok
db     i-0c  running  r6g.large  10.0.2.9    -          Failed
web-2  i-0b  stopped  t3.small   10.0.1.6    -          -

3 instances · 2 running · 1 stopped
```

- **State**: green running, yellow stopped/stopping, cyan pending, red terminated.
- **Checks**: `Ok` when both system and instance status checks pass, `Failed` otherwise, `-` when there's no check data (e.g. the instance isn't running).

### `awst list lb`

```bash
awst list lb -p myprofile -f api
```

Lists classic, application, network and gateway load balancers. The filter matches name and DNS name.

### `awst connect`

```bash
awst connect -p myprofile -f platform
```

1. Finds **running** instances that match `--filter`. The filter is required.
2. Shows a picker: `↑`/`↓` to move, `/` to search, Enter to select. If only one instance matches, it connects straight away.
3. Asks which user to log in as (`ubuntu`, `root`, `ec2-user`, …).
4. Runs `aws ssm start-session --document-name AWS-StartInteractiveCommand` with `sudo su - <user>` and prints the full command.

## Troubleshooting

| Error | Fix |
|---|---|
| `session-manager-plugin not found on PATH` | Install the Session Manager plugin (see Requirements). |
| `aws not found on PATH` | Install AWS CLI v2. |
| `expected 1 arg, got N ...` | You passed an unquoted `*`. Quote it (`"*"`) or drop `--filter`. |
| `unknown command "awst"` | With `go run`, use `go run . connect ...`, not `go run main.go awst connect ...`. |
| `TargetNotConnected` from SSM | The instance's SSM agent isn't registered. Check its instance profile and network access to SSM. |

Colors turn off automatically when output is piped or `NO_COLOR` is set.

## Development

```bash
go vet ./... && go test ./...
```
