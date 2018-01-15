package main

import (
    "os"
    "go.uber.org/zap"
    "net/http"
    "os/signal"
    "context"
    "syscall"
    "time"
)

const devEnv = "dev"
const prodEnv = "prod"
const shutDownTimeOut = 2 * time.Second

func main() {
    var config Config
    var configFile string
    var logger *zap.Logger

    if len(os.Args) < 2 {
        panic("No config file specified!")
    } else {
        configFile = os.Args[1]
    }

    config = LoadConfiguration(configFile)

    if config.Env == prodEnv {
        logger, _ = zap.NewProduction()
    } else if config.Env == devEnv {
        logger, _ = zap.NewDevelopment()
    }

    sugaredLogger := logger.Sugar()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    httpServer := &http.Server{
        Addr: ":" + config.Port,
        Handler: &server{
            Config:        config,
            Logger: sugaredLogger,
        },
    }

    go func() {
        sugaredLogger.Info("Server started on port " + config.Port)
        httpServer.ListenAndServe()
    }()

    //Graceful shutdown
    <-stop
    ctx, _ := context.WithTimeout(context.Background(), shutDownTimeOut)
    sugaredLogger.Infof("Shutdown with timeout: %s", shutDownTimeOut)
    httpServer.Shutdown(ctx)
    sugaredLogger.Info("Server stopped")
}
