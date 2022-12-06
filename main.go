package main

import (
	"fmt"
	"github.com/dream-mo/prom-elastic-alert/boot"
	"github.com/dream-mo/prom-elastic-alert/conf"
	"github.com/dream-mo/prom-elastic-alert/utils/logger"
	"github.com/dream-mo/prom-elastic-alert/utils/redis"
	"github.com/dream-mo/prom-elastic-alert/utils/xtime"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
)

func main() {

	var opts conf.FlagOption

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s\n", e)
			if opts.Debug {
				debug.PrintStack()
			}
		}
	}()

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic(err)
	}

	logger.SetLogLevel(opts.GetLogLevel())
	xtime.FixedZone(opts.Zone)

	c := conf.GetAppConfig(opts.ConfigPath)
	redis.Setup()
	ea := boot.NewElasticAlert(c, &opts)
	ea.Start()

	if c.Exporter.Enabled {
		metrics := boot.NewRuleStatusCollector(ea)
		reg := prometheus.NewPedanticRegistry()
		err := reg.Register(metrics)
		if err != nil {
			t := fmt.Sprintf("Register prometheus collector error: %s", err.Error())
			panic(t)
		}
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			reg,
		}
		h := promhttp.HandlerFor(gatherers,
			promhttp.HandlerOpts{
				ErrorHandling: promhttp.ContinueOnError,
			})
		http.Handle("/metrics", h)
		http.HandleFunc("/alert/message", boot.RenderAlertMessage)
		e := http.ListenAndServe(c.Exporter.ListenAddr, nil)

		if e != nil {
			t := fmt.Sprintf("Prometheus exporter start error: %s", e.Error())
			panic(t)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		s := <-quit
		switch s {
		case syscall.SIGHUP:
			c := conf.GetAppConfig(opts.ConfigPath)
			ea.SetAppConf(c)
			logger.Logger.Infoln("Reload application config success!")
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			ea.Stop()
			logger.Logger.Infoln("exiting...")
			return
		}
	}
}
