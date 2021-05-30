package hygo

import (
	"strconv"

	"github.com/c-f/hygo/modules/db"
	"github.com/c-f/hygo/modules/ssh"
)

// GetBruter is the Factory, which creates a bruter for a given target and service
func GetBruter(service string, conf *Config, target string, port string) Bruter {

	switch service {
	/* --[SSH]-- */
	case ssh.Name:
		module := ssh.New(target, GetPortOrDefault(port, 22), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
	/* --[DBs]-- */
	case db.MysqlName:
		module := db.NewMysql(target, GetPortOrDefault(port, 3306), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
	case db.PostgresName:
		module := db.NewPostgres(target, GetPortOrDefault(port, 5432), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
	case db.MssqlName:
		module := db.NewMssql(target, GetPortOrDefault(port, 1443), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
	case db.MongoDBName:
		module := db.NewMongoDB(target, GetPortOrDefault(port, 27017), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
	case db.RedisName:
		module := db.NewRedis(target, GetPortOrDefault(port, 6379), conf.Sleep, conf.Timeout, conf.LogFailedAttempts)
		module.StopIfNetErr = !conf.Force // ignore network error
		return module
		// couchdb
		// cassandra
	}
	return nil
}

// GetPortOrDefault uses the default port if no port is available or not a number
func GetPortOrDefault(in string, alt int) int {
	prt, err := strconv.Atoi(in)
	if err != nil {
		prt = alt
	}
	return prt
}
