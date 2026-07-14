# Getting a shell on the API EC2 instance

There is no SSH into the `glassact-api` instance — by design. The security
group (`apps/infrastructure/ec2.tf`) has no port-22 ingress, there is no key
pair, and there is no bastion. Ingress is limited to CloudFront's
origin-facing prefix list on port 8080.

Access is entirely through **SSM Session Manager**. The instance runs on
AL2023 (SSM agent baked in) with the `AmazonSSMManagedInstanceCore` policy
attached, so the agent connects outbound to SSM and hands you a shell over the
AWS API — no inbound port is ever opened.

For a Postgres tunnel specifically (port-forwarding, not a shell), see
`docs/prod-database-access.md`.

## Prerequisites

- AWS CLI configured with credentials that can `ssm:StartSession` (and
  `ssm:TerminateSession`) on the instance.
- The Session Manager plugin installed locally — `start-session` fails without
  it:
  https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html
- Everything is in `us-west-2` (passed explicitly on each command below).

## Interactive shell

```bash
cd apps/infrastructure
aws ssm start-session --region us-west-2 --target "$(terraform output -raw api_instance_id)"
```

This drops you in as `ssm-user`. Become root with `sudo su -`. The API and
Postgres run as Docker containers, so the usual next steps are:

```bash
sudo docker ps
sudo docker logs -f glassact-api
```

## Confirming the instance is reachable

If `start-session` hangs or reports the target is not connected, check that
the box is registered as a managed node (the agent needs outbound HTTPS to
SSM, which the egress-all rule permits):

```bash
aws ssm describe-instance-information --region us-west-2 \
  --query "InstanceInformationList[?InstanceId=='$(cd apps/infrastructure && terraform output -raw api_instance_id)']"
```

## Running a one-off command without a session

For non-interactive automation you can send a command instead of opening a
session (this does not need the Session Manager plugin):

```bash
aws ssm send-command --region us-west-2 \
  --instance-ids "$(cd apps/infrastructure && terraform output -raw api_instance_id)" \
  --document-name AWS-RunShellScript \
  --parameters 'commands=["docker ps"]'
```

## SSH-over-SSM (optional)

If you specifically need the `ssh` client — e.g. for `scp` — you can use SSM
as the transport by adding this to `~/.ssh/config`:

```
host i-* mi-*
    ProxyCommand sh -c "aws ssm start-session --region us-west-2 --target %h --document-name AWS-StartSSHSession --parameters portNumber=%p"
```

Then `ssh ssm-user@<instance-id>` works. This still requires a public key in
`authorized_keys` on the instance, which is not provisioned by default — so
for interactive access, prefer `start-session` above.
