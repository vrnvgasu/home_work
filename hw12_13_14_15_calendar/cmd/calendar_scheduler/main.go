package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/rabbitmq/producer"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/scheduler"
	storageimp "github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/storage/impl"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config-dev.yml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	ctx := context.Background()

	config.Cfg = config.NewConfig(configFile)
	logg := logger.New(config.Cfg.Logger.Level)

	// storage
	storage, err := storageimp.NewIStorage()
	if err != nil {
		logg.Error("create storage failed" + err.Error())
		panic(err)
	}
	if err = storage.Connect(ctx); err != nil {
		logg.Error("connect storage failed" + err.Error())
		panic(err)
	}
	defer func() {
		if err = storage.Close(ctx); err != nil {
			logg.Error("close storage failed" + err.Error())
		}
	}()

	calendar := app.New(logg, storage)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	schedulerServer := scheduler.NewScheduler(producer.NewProducer(), logg, calendar)

	go func() {
		if err = schedulerServer.Run(ctx); err != nil {
			logg.Error("schedulerServer run failed" + err.Error())
			panic(err)
		}
	}()

	logg.Info("scheduler is running...")

	<-ctx.Done()
	fmt.Println("shutting down scheduler...")
}
