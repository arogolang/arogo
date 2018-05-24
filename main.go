package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arogolang/arogo/config"
	"github.com/arogolang/arogo/errlog"
	"github.com/arogolang/arogo/pool"
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

	cfg := config.Get()

	pool.NewPoolServer(cfg.PoolWebAddr)
	//pool.NewPoolStratumServer(cfg.PoolStartumAddr)

	tCh := make(chan os.Signal)
	signal.Notify(tCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-tCh:
		errlog.Info("stopping")
	}
}
