# HyGo

[![PkgGoDev](https://pkg.go.dev/badge/github.com/c-f/hygo)](https://pkg.go.dev/github.com/c-f/hygo?tab=doc)

HyGo is a util to test various credentials against a fleet of hosts of various services. It can be (with a few modification) the golang alternative to hydra, medusa or aircrack.

Please see Caveats to choose the right tool for the right job.

Status: `Beta`

<!-- luls -->
<div align="center" style="ign:center">
<img align="center" width="400" src="./hygo.jpg">
</div>

## Usage 

HyGo is **incomplete** by purpose. While HyGo contains easy structures to handle auth testing against a vast amount of services, all go files are missing in the `cmd` directory. 

How to build it then ? 

```bash 
go get github.com/c-f/hygo
```
Then you just need to create a simple script handling:

```
- user arguments / validate (flags)
- read wordlist from file to create parse model.Credential
- read targets from file 
- handle err/out stream to print json (if you want)
- get handler based on service and send creds to bruter
- close chans and wait :) 
```


## Caveats

**Why you should not use this lib**: 

- You hate writing code /golang
- Not all protocols are supported  - so if you're looking for SMB, RDP and HTTP please use the good alternative solutions. 
- Beta phase - while it is tested against container of various vendors and versions it is currently not tested against a lot of live systems, therefore consider it `Beta`
- Currently a new connection/ TCP is made for each attempt. Since auth requests are delayed through Sleep to be more nice to the server this "optimization"/TCP handling is not done.


**A few reasons why**
- You like json ? because i like json :) so input and output is json (except all err msgs)
- While medusa is a good tool, it can't handle none-terminating TCP connections. This is painful since medusa does not provide a timeout option. Additionally new versions of services are not fully supported (e.g. MYSQL 8)
- Hydra, yes you can use hydra. Personally, i have some bad experienced and switched to tools
- Brutespray, great tool, especially a very usefull wrapper for medusa. This is also the main problem, for the reasons above
- Patator, great, solves most of the problems. In our tests the info log was tremendous, which made it necessary to modify the script. Also additional features would need adjustments (e.g. JSON out). 
- Go should handle most of the issues. Good libs exist, which makes it quite easy to implement the behaviour. Great performance and in the end a "slim" binary, which can be easily copied also to a target computer to perform, without the neccessity to install deps and other. No good point - but you know :) 


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

./hygo -s mysql -C ../integration_test/mysql.wordlist --iL ../integration_test/targets.json -delay 1s

./hygo -s ssh -C ../integration_test/ssh.wordlist --iL ../integration_test/targets.json -delay 1s

./hygo -s postgres -C ../integration_test/postgres.wordlist --iL ../integration_test/targets.json -delay 1s

./hygo -s mssql -C ../integration_test/mssql.wordlist --iL ../integration_test/targets.json -delay 1s

./hygo -s mongo -C ../integration_test/mongo.wordlist --iL ../integration_test/targets.json -delay 1s

```

## Supported 
Data Source: https://hub.docker.com/search?q=&type=image&category=database


#### Other Services 
- [x] SSH `ssh`
- [ ] FTP(s)
- [ ] Telnet 
- [ ] SMB
- [ ] RDP
- [ ] IMPA/POP3/SMTP

#### Databases
- [x] MongoDB `mysql`
- [x] MariaDB `mysql`
- [x] Postgres `postgres` 
- [x] Redis `redis`
- [x] MSSQL `mssql`
- [ ] InfluxDB
- [ ] Cassandra
- [ ] CouchDB
- [ ] Neo4j
- [ ] RethinkDB
- [ ] Crate
- [ ] ArangoDB
- [ ] OrientDB
- [ ] Oracle
- [ ] 