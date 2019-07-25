`garbaged` is controlled via an API, so you can integrate it
with other tooling in your infrastructure. There's also the `tt`
command which is an API client for `garbaged`.

# Endpoints

| Endpoint                | Methods      | What's it do?                                 |
|-------------------------|--------------|-----------------------------------------------|
| /v1/trash               | GET          | List trash to be collected on the next run    |
| /v1/trash/all           | GET          | List all trash                                |
| /v1/trash/new           | POST         | Create new trash                              |
| /v1/trash/pickup        | POST         | Run a trash pickup                            |
| /v1/holidays            | GET          | List all trash holidays                       |
| /v1/holiday/role/{name} | POST, DELETE | Create (or delete) a role-based trash holiday |
| /v1/holiday/type/{name} | POST, DELETE | Create (or delete) a type-based trash holiday |

## /v1/trash
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

## /v1/trash/all
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

## /v1/trash/new
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

## /v1/trash/pickup
This is a POST endpoint that will run a trash pickup. (in progress)

## /v1/holidays
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

## /v1/holiday/role/{name} and /v1/holiday/type/{name}
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
