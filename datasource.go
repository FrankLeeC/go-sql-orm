package orm

import (
	"database/sql"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

var ds = make(map[string]*sql.DB)

// default datasource   the first registered datasource
var defaultDatasource string

type DatasourceConfig struct {
	name        string
	dns         string
	maxConn     int
	maxIdleConn int
}

func NewDatasourceConfig(name, dns string) DatasourceConfig {
	return DatasourceConfig{name: name, dns: dns}
}

func (a DatasourceConfig) MaxConn(m int) DatasourceConfig {
	a.maxConn = m
	return a
}

func (a DatasourceConfig) MaxIdleConn(m int) DatasourceConfig {
	a.maxIdleConn = m
	return a
}

func RegisterDatsource(config DatasourceConfig) {
	cfg, err := mysql.ParseDSN(config.dns)
	if err != nil {
		panic("mysql dns error:" + err.Error())
	}
	cfg.ParseTime = true
	cfg.Loc = time.Local
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic("connect to database error:" + err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic("ping database error:" + err.Error())
	}
	maxConn := config.maxConn
	maxIdleConn := config.maxIdleConn
	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(maxIdleConn)

	ds[config.name] = db

	if len(ds) == 1 {
		defaultDatasource = config.name
	}
}
