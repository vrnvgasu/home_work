package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/rabbitmq/consumer"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/sender"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/sender_config-dev.yml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config.Cfg = config.NewConfig(configFile)
	logg := logger.New(config.Cfg.Logger.Level)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	cons, err := consumer.NewConsumer()
	if err != nil {
		logg.Error("create consumer failed" + err.Error())
		panic(err)
	}
	senderServer := sender.NewSender(cons, logg)

	if err = senderServer.Run(); err != nil {
		logg.Error("senderServer run failed" + err.Error())
		panic(err)
	}

	logg.Info("sender is running...")

	<-ctx.Done()
	logg.Info("shutting down sender...")
	senderServer.Stop()
}
