package main

import (
	"flag"
	"fmt"
	"github.com/aylei/alert-shim/pkg/api"
	"github.com/aylei/alert-shim/pkg/config"
	"os"
)

var (
	Version   string
	BuildTime string
)

var (
	configFile string
)

func main() {
	flag.StringVar(&configFile, "config", "conf.yaml", "config file location")
	printVersion := flag.Bool("v", false, "print build version")
	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		os.Exit(0)
	}

	err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}
	conf := config.GetConfig()

	g, err := api.New(conf)
	if err != nil {
		panic(err)
	}
	err = g.Run(fmt.Sprintf("%s:%d", conf.Addr, conf.Port))
	if err != nil {
		panic(err)
	}
}
