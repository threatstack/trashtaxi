# Getting Started
## Concepts and Terms Used in this Documentation

You'll want to be familiar with REST APIs and AWS to use `garbaged`. `garbaged`
is a garbage collector. The API and program functions follow that idiom.

  * *Trash* refers to hosts that have been marked for collection.
  * A *trash pickup* is the process of terminating hosts marked as _trash_.
  * Scheduling a *Trash Holiday* for a _role_ or _type_ will temporarily suspend
    the _trash pickup_ process for that role or type.

Got it? Great! If there are things that you think should be here but aren't,
open an issue so we can fix it.

## Building and installing

You can use `go build` and drop the output wherever you wish. `garbaged`'s
configuration file (by default) lives in `/etc/garbaged.json` but you can
reassign it using the `GARBAGED_CONFIG` environment variable.

There is a command line utility called `tt` which is in `cmd/tt`. It's a
command line interface to the API. You can run `go build` in that directory.

## Prerequisites

*Important*: `garbaged` is tightly coupled to AWS. You'll need to do some coding
work to use it with other cloud providers.

Before you get started with `garbaged` you'll want to have a few things in
place:

  0. Build and distribute [nt](cmd/nt) in your infrastructure. `nt` sends an API 
    call to `garbaged` before it starts a root shell.
  1. *A small Postgres instance* - `garbaged` stores state in a couple Postgres
     tables.
  2. *Cross-account Roles* - `garbaged` assumes access to other accounts, so
     you can use it across multiple AWS accounts. The next section will describe
     what permissions you need to grant.
  3. *Role Tagging* - `garbaged` can be configured to ignore certain EC2 tags -
     something like `Role` or `Type`. In our infrastructure, a Cassandra host
     may have `Role` set to `cassandra` and `Type` set to `database`. We'll
     have `garbaged` ignore `database` "Type" hosts, so you don't end up taking
     out a database host automatically.

### Cross-Account Role Permissions
In your *ORIGIN* account (the account or role that will `garbaged` will use to
execute) configure permissions so that `garbaged` may AssumeRole into another
account. In this example, the role in the *TARGET* account is named `trashtaxi`:
```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": [
        "arn:aws:iam::TARGET_ACCOUNT_NUM:role/trashtaxi"
      ],
      "Effect": "Allow"
    }
  ]
}
```
In the *TARGET* account, configure an IAM role that can read tags and terminate
instances, and add a trust relationship to the *ORIGIN* account. The ExternalID
is a good idea.

*TARGET* Account Role Policy:
```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:TerminateInstances",
        "ec2:DescribeInstance*",
        "ec2:DescribeTags"
      ],
      "Resource": "*"
    }
  ]
}
```
*TARGET* Account Role Trust:
```
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::ORIGIN_ACCT_NUM:role/ORIGIN_IAM_ROLE"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "RANDOM_GENERATED_STRING"
        }
      }
    }
  ]
}
```
