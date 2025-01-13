package main

import (
	"errors"
	"golang.org/x/sys/unix"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func RunAsDaemon(daemon *Daemon) error {
	slog.Info("Running in daemon mode")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, unix.SIGINT, unix.SIGTERM)

	defer func() {
		signal.Stop(signalChannel)
	}()

	period, err := time.ParseDuration(daemon.Config.TimeSync)
	if err != nil {
		return err
	} else if period.Seconds() < 10 {
		return errors.New("time period for checks is smaller than 10 sec")
	}
	slog.Debug("Sync period is set to " + period.String())

	updateTicker := time.NewTicker(period)

	for {
		select {
		case sig := <-signalChannel:
			slog.Debug("Received signal " + sig.String())
			switch sig {
			case unix.SIGINT:
				return nil
			default:
				slog.Warn("Unknown signal", sig.String())
				break
			}
			break

		case tick := <-updateTicker.C:
			slog.Debug("Received tick", "time", tick.String())
			for index := range daemon.Repos {
				slog.Info("Processing sync", "directory", daemon.Repos[index].BaseFolder)
				err := daemon.Repos[index].DoTheThings()
				if err != nil {
					slog.Warn("Error while processing "+daemon.Repos[index].BaseFolder, "msg", err.Error(), "path", daemon.Repos[index].BaseFolder)
				}

			}
			break
		}
	}

}
