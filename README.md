# Trash Taxi
Imagine a world where you, a systems professional, no longer needs to log in to
servers. You have the best tooling around monitoring and observability! Every
service is stateless! Sounds like a dream, right?

We know that the reality for many systems professionals is not quite so dreamy.
Your application broke; you log in and start looking at log files and... make a
few changes to a configuration file here and there. Over time, small changes
compound. Suddenly, the state of your system no longer aligns with the state
represented in your code.

We _also_ know that manual changes introduce security and availability issues.
In this age of configuration management and stateless services, we can do better.

The Trash Taxi system revolves aroudn two tools: `nt` which sends API requests
to `garbaged`.

`garbaged` manages the lifecycle of your infrastructure. It allows systems
professionals to run an arbitrary root commands in an emergency. At a later
time, `garbaged` terminates hosts, keeping your infrastructure fresh.

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
        "AWS": "arn:aws:iam::ORIGIN_ACCT_NUM:root"
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

## Configuration File
Here's a sample configuration file. Subsections will describe the fields for
each configuration area.
```
{
  "aws": {
    "dryrun": false,
    "role_ec2_tag": "Role",
    "type_ec2_tag": "Type",
    "termlimit": 5
  },
  "bind": 443,
  "debug": false,
  "tlscrt": "/secret/cert.pem",
  "tlskey": "/secret/key.pem",
  "iid_verify_cert": "/etc/aws_public_verifier.pem",
  "timebetweenpickup: "2h",
  "database": {
    "host": "",
    "port": 5432,
    "user": "",
    "pass": "",
    "name": "",
    "sslmode: ""
  },
  "accounts": {
    "123412341234": {
      "name": "dev",
      "arn": "arn:aws:iam::123412341234:role/trashtaxi",
      "externalid": "hellohello"
    },
    "432143214321": {
      "name": "prod",
      "arn": "arn:aws:iam::432143214321:role/trashtaxi"
      "externalid": "hellogoodbye"
    }
  },
  "stateful": {
    "roles": ["haproxy"],
    "types": ["database"]
  },
  "graylog": {
    "enabled": true,
    "host": "graylog.tls.zone",
    "port": 12345
  }
  "slack": {
    "enabled": true,
    "apikey": "xoxp...",
    "channels": ["#ops"],
    "timewait": "120s" 
  }
}
```

### Top-level Configuration
| Variable | Type   | Purpose                                    | Possible Value(s)         |
|-------------------|--------|-----------------------------------|---------------------------|
| bind              | Int    | Port for `garbaged` to bind to    | 443                       |
| debug             | Bool   | Some extra logging, queries       | 5432                      |
| iid_verify_cert   | String | Path to the AWS IID Verifier cert | /etc/aws_verifier.pem     |
| timebetweenpickup | String | Rate limiter between pickups      | 60, 1m, 2h, 1w (go times) |
| tlscrt            | String | Path to the certificate file      | /secret/tlscert.pem       |
| tlskey            | String | Path to the key file              | /secret/tlskey.pem        |

### Database Configuration (`database`)

| Variable | Type   | Purpose                    | Possible Value(s)    |
|----------|--------|----------------------------|----------------------|
| host     | String | Path to your database host | database.example.com |
| port     | Int    | Port # for the database    | 5432                 |
| user     | String | Username for the database  | trashtaxi            |
| pass     | String | Password for the database  | hunter7              |
| name     | String | Database name              | garbaged             |
| sslmode  | String | Postgres SSLMode           | disable, enforce     |

### AWS Account Config (`accounts`)
The object key should be the AWS account ID.

| Variable   | Type   | Purpose                               | Possible Value(s)                        |
|------------|--------|---------------------------------------|------------------------------------------|
| name       | String | A friendly name for the account       | dev, prod, testing                       |
| arn        | String | The role ARN that `garbaged` will use | arn:aws:iam::123412341234:role/trashtaxi |
| externalid | String | The External ID used for trust        | bo0kei2caiquoN6T (a random string)       |

### AWS Role/Type Tags (`aws`)
`garbaged` will query an instance's EC2 tags to determine it's role and type.
For example, the role tag could correspond with a particular node's Chef role,
and the type tag could correspond with a broader class of machine.

| Variable     | Type   | Purpose                                          | Possible Value(s)     |
|--------------|--------|--------------------------------------------------|-----------------------|
| dryrun       | Bool   | For debug; adds DryRun flag to Terminate request | true, false           |
| role_ec2_tag | String | The tag that specifies the host's role           | mongodb, flink-worker |
| type_ec2_tag | String | The tag that specifies the host's type           | stateless, database   |

### Defined Stateful Services (`stateful`)
`garbaged` can automatically set up trash holidays based on the role and type tags you defined
above.

| Variable | Type  | Purpose                                        | Possible Value(s)     |
|----------|-------|------------------------------------------------|-----------------------|
| roles    | Array | Roles that should be ignored in trash pickup   | mongodb, jumphost     |
| types    | Array | Types that should be ignored in trash pickup   | database              |

### Slack Configuration (`slack`)
`garbaged` can send a message to a slack channel before actually doing a trash pickup.

| Variable | Type   | Purpose                                        | Possible Value(s)         |
|----------|--------|------------------------------------------------|---------------------------|
| enabled  | Bool   | If slack notification is enabled               | true, false               |
| apikey   | String | The API key for your slack bot                 | xoxp-...                  |
| channels | Array  | An array of channels to notify                 | ["#ops", "#security"]     |
| timewait | String | An amount of time to wait before pickup        | 60, 1m, 2h, 1w (go times) |

### Graylog Configuration (`graylog`)
`garbaged` can send its log data to Graylog over UDP, if you use that kind of thing.

| Variable | Type   | Purpose                                        | Possible Value(s) |
|----------|--------|------------------------------------------------|-------------------|
| enabled  | Bool   | If graylog message sending is enabled          | true, false       |
| host     | String | The hostname of your graylog instance          | graylog.tls.zone  |
| port     | Int    | Graylog listener port                          | 12345             |

## API
You can use `garbaged` via its API.

| Endpoint                | Methods      | What's it do?                                 |
|-------------------------|--------------|-----------------------------------------------|
| /v1/trash               | GET          | List trash to be collected on the next run    |
| /v1/trash/all           | GET          | List all trash                                |
| /v1/trash/new           | POST         | Create new trash                              |
| /v1/trash/pickup        | POST         | Run a trash pickup                            |
| /v1/holidays            | GET          | List all trash holidays                       |
| /v1/holiday/role/{name} | POST, DELETE | Create (or delete) a role-based trash holiday |
| /v1/holiday/type/{name} | POST, DELETE | Create (or delete) a type-based trash holiday |

### /v1/trash
This is a GET endpoint that returns the trash to be collected on the next trash
pickup.

Sample Output:
```
[
  {
    "ID": 34,
    "CreatedAt": "2018-04-27T12:03:55.33722-04:00",
    "UpdatedAt": "2018-04-27T12:03:55.33722-04:00",
    "DeletedAt": null,
    "Host": "i-00d79a1b897a4bc2a",
    "Region": "us-east-1",
    "Account": "123412341234",
    "Role": "indexer",
    "Type": "compute"
  },
  {
    "ID": 37,
    "CreatedAt": "2018-04-27T12:03:55.826185-04:00",
    "UpdatedAt": "2018-04-27T12:03:55.826185-04:00",
    "DeletedAt": null,
    "Host": "i-045777ea462046671",
    "Region": "us-east-1",
    "Account": "432143214321",
    "Role": "flink-worker",
    "Type": "compute"
  }
]
```

### /v1/trash/all
This is a GET endpoint that returns ALL hosts marked as trash.

Sample Output:
```
[
  {
    "ID": 34,
    "CreatedAt": "2018-04-27T12:03:55.33722-04:00",
    "UpdatedAt": "2018-04-27T12:03:55.33722-04:00",
    "DeletedAt": null,
    "Host": "i-00d79a1b897a4bc2a",
    "Region": "us-east-1",
    "Account": "123412341234",
    "Role": "indexer",
    "Type": "compute"
  },
  {
    "ID": 37,
    "CreatedAt": "2018-04-27T12:03:55.826185-04:00",
    "UpdatedAt": "2018-04-27T12:03:55.826185-04:00",
    "DeletedAt": null,
    "Host": "i-045777ea462046671",
    "Region": "us-east-1",
    "Account": "432143214321",
    "Role": "flink-worker",
    "Type": "compute"
  },
  {
    "ID": 38,
    "CreatedAt": "2018-04-27T12:03:56.012163-04:00",
    "UpdatedAt": "2018-04-27T12:03:56.012163-04:00",
    "DeletedAt": null,
    "Host": "i-015a06f0f5be3b4ee",
    "Region": "us-east-1",
    "Account": "432143214321",
    "Role": "jumphost",
    "Type": "support"
  }
]
```

### /v1/trash/new
This is a POST endpoint that takes a pkcs7-encoded AWS [Instance Identity Document](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-identity-documents.html)
(IID), verifies it, then adds data to the database.

You can obtain the AWS IID from any EC2 host by running:
```
curl http://169.254.169.254/latest/dynamic/instance-identity/pkcs7 | tr -d '\n'
```

Sample Input:
```
curl -XPOST -d '{"iid":"IID_OUTPUT_FROM_ABOVE"}' http://localhost:3000/v1/trash/new
```

Sample Output:
```
{
  "accepted": true,
  "context": ""
}
```

Sample Error:
```
{
  "accepted": false,
  "context": "Unable to verify PKCS7 Document"
}
```

### /v1/trash/pickup
This is a POST endpoint that will run a trash pickup. (in progress)

### /v1/holidays
This is a GET endpoint that lists all the trash holidays by role and type. It
includes permanent holidays that are specified in the configuration file, too.

Sample Output:
```
{
  "types": [
    "database"
  ],
  "roles": [
    "jumphost"
  ]
}
```

### /v1/holiday/role/{name} and /v1/holiday/type/{name}
This is a POST (and DELETE) endpoint that will create (or delete) a role holiday
The POST takes no parameters.

Sample Creation Input:
```
curl -XPOST http://localhost:3000/v1/holiday/role/flink-worker
````

Sample Creation Output:
```
{
  "accepted": true,
  "context": ""
}
```

Sample Creation Error:
```
{
  "accepted": false,
  "context": "pq: duplicate key value violates unique constraint \"role_holidays_role_key\""
}
```

Sample Deletion Input:
```
curl -XDELETE http://localhost:3000/v1/holiday/role/flink-worker
```

Sample Deletion Output:
```
{
  "accepted": true,
  "context": "Deleted role if it existed"
}
```

## Contributing, Questions, etc.

Questions, suggestions, any other thoughts - please open an issue. Fortunately,
this project is small:  You'll find the bulk of the server code and endpoint
handlers under `server/`. If you need to add configuration flags or variables
the configuration structs and such under `config/`.
