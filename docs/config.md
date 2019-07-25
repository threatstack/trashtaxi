# Configuration Reference

Here's a sample configuration file. Take a look at the subsections for 
configuration variable information.

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
