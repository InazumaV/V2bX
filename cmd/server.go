package cmd

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	vCore "github.com/Yuzuki616/V2bX/core"

	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/Yuzuki616/V2bX/node"
	"github.com/spf13/cobra"
)

var (
	config string
	watch  bool
)

var serverCommand = cobra.Command{
	Use:   "server",
	Short: "Run V2bX server",
	Run:   serverHandle,
	Args:  cobra.NoArgs,
}

func init() {
	serverCommand.PersistentFlags().
		StringVarP(&config, "config", "c",
			"/etc/V2bX/config.yml", "config file path")
	serverCommand.PersistentFlags().
		BoolVarP(&watch, "watch", "w",
			true, "watch file path change")
	command.AddCommand(&serverCommand)
}

func serverHandle(_ *cobra.Command, _ []string) {
	showVersion()
	c := conf.New()
	err := c.LoadFromPath(config)
	if err != nil {
		log.WithField("err", err).Error("Load config file failed")
		return
	}
	limiter.Init()
	log.Info("Start V2bX...")
	vc, err := vCore.NewCore(&c.CoreConfig)
	if err != nil {
		log.WithField("err", err).Error("new core failed")
		return
	}
	err = vc.Start()
	if err != nil {
		log.WithField("err", err).Error("Start core failed")
		return
	}
	defer vc.Close()
	nodes := node.New()
	err = nodes.Start(c.NodesConfig, vc)
	if err != nil {
		log.WithField("err", err).Error("Run nodes failed")
		return
	}
	dns := os.Getenv("XRAY_DNS_PATH")
	if watch {
		err = c.Watch(config, dns, func() {
			nodes.Close()
			err = vc.Close()
			if err != nil {
				log.WithField("err", err).Error("Restart node failed")
				return
			}
			vc, err = vCore.NewCore(&c.CoreConfig)
			if err != nil {
				log.WithField("err", err).Error("New core failed")
				return
			}
			err = vc.Start()
			if err != nil {
				log.WithField("err", err).Error("Start core failed")
				return
			}
			err = nodes.Start(c.NodesConfig, vc)
			if err != nil {
				log.WithField("err", err).Error("Run nodes failed")
				return
			}
			runtime.GC()
		})
		if err != nil {
			log.WithField("err", err).Error("start watch failed")
			return
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
