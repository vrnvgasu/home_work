package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/config"
	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/server/http"
	storageimp "github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/storage/impl"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/calendar_config-dev.yml", "Path to configuration file")
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
	if err = storage.Migrate(); err != nil {
		logg.Error("migrate storage failed" + err.Error())
		panic(err)
	}

	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar)
	grpcServer := internalgrpc.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}

		if err := grpcServer.Stop(); err != nil {
			logg.Error("failed to stop grpc server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	go func() {
		if err := server.Start(ctx); err != nil {
			logg.Error("failed to start http server: " + err.Error())
			cancel()
			os.Exit(1)
		}
	}()
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			logg.Error("failed to start grpc server: " + err.Error())
			cancel()
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	fmt.Println("shutting down http server...")
}
