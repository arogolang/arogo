package mysqldb

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/arogolang/arogo/config"
	"github.com/arogolang/arogo/errlog"
)

type MySqlDB struct {
	db *sqlx.DB
}

func NewMySqlDB(conf *config.SqlDBConf) (*MySqlDB, error) {
	mysqlconn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
		conf.UserName, conf.Password, conf.Address, conf.DBName)

	conn, err := sqlx.Connect("mysql", mysqlconn)
	if err != nil {
		errlog.Errorf("sql.Open failed: %v", err)
		return nil, err
	}

	db := &MySqlDB{
		db: conn,
	}
	errlog.Info("Mysql connect success")

	return db, nil
}

func (d *MySqlDB) GetDB() *sqlx.DB {
	return d.db
}

func (d *MySqlDB) Close() error {
	return d.db.Close()
}
