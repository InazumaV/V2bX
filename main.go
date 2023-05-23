package main

import (
	"flag"
	"fmt"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/limiter"
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
	version   = "TempVersion" //use ldflags replace
	codename  = "V2bX"
	intro     = "A V2board backend based on Xray-core"
	warnColor = "\033[0;31m"
)

func showVersion() {
	fmt.Printf("%s %s (%s) \n", codename, version, intro)
	// Warning
	fmt.Printf("%sThis version need V2board version >= 1.7.0.\n", warnColor)
	fmt.Printf("%sThis version changed config file. Please check config file before running.\n", warnColor)
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
	limiter.Init()
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
	// clear memory
	runtime.GC()
	// wait exit signal
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-osSignals
	}
}
