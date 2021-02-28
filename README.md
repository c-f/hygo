# HyGo
HyGo is a util to test various credentials against a fleet. It can be (with a few modification) the golang alternative to hydra, medusa or aircrack.

Please see Caveats to choose the right tool for the right job.

## Usage 

HyGo is **incomplete** by purpose. While HyGo contains easy structures to handle auth testing against a vast amount of services, all go files are missing in the `cmd` directory. 



## Caveats


## Depencencies: 

- [mysql Driver](https://github.com/go-sql-driver/mysql)
- [mssql Driver](https://github.com/denisenkom/go-mssqldb)
- [postgres Driver](https://github.com/lib/pq)
- [SSH](https://pkg.go.dev/golang.org/x/crypto/ssh) official


## Data Option

- **mongo**: `auth_mechanism` [docu](https://docs.mongodb.com/manual/reference/connection-string/#urioption.authMechanism)
- **ssh**: `ssh_key` [docu](https://www.ssh.com/ssh/key/)


## Integration tests


```bash

go run *go -s mysql -C ../integration_test/mysql.wordlist --iL ../integration_test/targets.json -delay 1s

go run *go -s ssh -C ../integration_test/ssh.wordlist --iL ../integration_test/targets.json -delay 1s

go run *go -s postgres -C ../integration_test/postgres.wordlist --iL ../integration_test/targets.json -delay 1s

go run *go -s mssql -C ../integration_test/mssql.wordlist --iL ../integration_test/targets.json -delay 1s

go run *go -s mongo -C ../integration_test/mongo.wordlist --iL ../integration_test/targets.json -delay 1s

```

## TODO 

## Mongo | redis | cassandra | couchdb