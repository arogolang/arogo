package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arogolang/arogo/config"
	"github.com/arogolang/arogo/errlog"
	"github.com/arogolang/arogo/mysqldb"
	"github.com/arogolang/arogo/pool"
	"github.com/arogolang/arogo/util"
	"github.com/arogolang/arogo/vars"
)

var configFile *string

func usage() {
	fmt.Printf("Usage: %s [-c CONFIG_PATH] \n", os.Args[0])
	flag.PrintDefaults()
}

func setOptions() {
	configFile = flag.String("c", "", "JSON file from which to read configuration values")
	flag.Parse()

	config.File = *configFile
}

func main() {
	flag.Usage = usage

	setOptions()

	if args := flag.Args(); len(args) > 1 && (args[1] == "help" || args[1] == "-h") {
		flag.Usage()
		return
	}

	if config.File == "" || !util.FileExists(config.File) {
		if util.FileExists("config.json") {
			config.File = "config.json"
		}
	}

	cfg := config.Get()
	dbMgr := &vars.PoolDBMgr{}

	if cfg.LogFile != "" {
		logfile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer logfile.Close()

		if !cfg.Debug {
			errlog.SetLogLevel(cfg.LogLevel)
		}

		errlog.AddLogOutput(logfile)
	}

	var err error

	dbMgr.NodeDB, err = mysqldb.NewMySqlDB(&cfg.NodeDB)
	if err != nil {
		errlog.Fatalf("cannot init mysql", err)
	}

	dbMgr.PoolDB, err = mysqldb.NewMySqlDB(&cfg.PoolDB)
	if err != nil {
		errlog.Fatalf("cannot init mysql", err)
	}

	tableExists, err := dbMgr.PoolDB.CheckTables(cfg.PoolDB.DBName, "miners")
	if err != nil || tableExists == false {
		err = dbMgr.PoolDB.InitTables()
		if err != nil {
			errlog.Fatalf("cannot init mysql", err)
		}
	}

	pool.NewPoolServer(cfg.PoolWebAddr, dbMgr)
	//pool.NewPoolStratumServer(cfg.PoolStartumAddr, mysqlDB)

	tCh := make(chan os.Signal)
	signal.Notify(tCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-tCh:
		errlog.Info("stopping")
	}
}
