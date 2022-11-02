package main

import (
	"flag"
	"fmt"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/node"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	configFile   = flag.String("config", "/etc/V2bX/config.yml", "Config file for V2bX.")
	watch        = flag.Bool("watch", true, "Watch config file for changes.")
	printVersion = flag.Bool("version", false, "show version")
)

var (
	version  = "TempVersion" //use ldflags replace
	codename = "V2bX"
	intro    = "A V2board backend based on Xray-core"
)

func showVersion() {
	fmt.Printf("%s %s (%s) \n", codename, version, intro)
}

func main() {
	flag.Parse()
	showVersion()
	if *printVersion {
		return
	}
	config := conf.New()
	err := config.LoadFromPath(*configFile)
	if err != nil {
		log.Panicf("can't unmarshal config file: %s \n", err)
	}
	log.Println("Start V2bX...")
	x := core.New(config)
	err = x.Start()
	if err != nil {
		log.Panicf("Failed to start core: %s", err)
	}
	defer x.Close()
	nodes := node.New()
	err = nodes.Start(config.NodesConfig, x)
	if err != nil {
		log.Panicf("run nodes error: %s", err)
	}
	if *watch {
		err = config.Watch(*configFile, func() {
			nodes.Close()
			err = x.Restart(config)
			if err != nil {
				log.Panicf("Failed to restart core: %s", err)
			}
			err = nodes.Start(config.NodesConfig, x)
			if err != nil {
				log.Panicf("run nodes error: %s", err)
			}
			runtime.GC()
		})
		if err != nil {
			log.Panicf("watch config file error: %s", err)
		}
	}
	//Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()
	// Running backend
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-osSignals
	}
}
