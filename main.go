package main

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"log/slog"
	"os"
	"path/filepath"
)

type DataRepo struct {
	BaseFolder      string
	GitAddress      string
	Repository      *git.Repository
	authFunc        func() transport.AuthMethod
	configDirectory *ConfigDirectory
}

type Daemon struct {
	Config ConfigFile
	Repos  []DataRepo
}

var daemon Daemon

func main() {
	var err error

	confPath, err := DetermineConfigLocation()
	if err != nil {
		slog.Error("Unable to determine configuration location", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	daemon = Daemon{}
	daemon.Config, err = LoadConfig(confPath)
	if err != nil {
		slog.Error("Error while loading config file", slog.String("msg", err.Error()))
		return
	}

	err = CheckAndConfigure(&daemon.Config)
	if err != nil {
		slog.Error("Error while loading config file", slog.String("msg", err.Error()))
		return
	}

	err = createDataRepos(&daemon)
	if err != nil {
		slog.Error("Error while opening repositories on FS", slog.String("msg", err.Error()))
		os.Exit(1)
		return
	}

	err = RunAsDaemon(&daemon)
	if err != nil {
		slog.Error("Error while running in daemon", "msg", err.Error())
		os.Exit(1)
		return
	}

}

func DetermineConfigLocation() (string, error) {
	envPath, envExist := os.LookupEnv("GITSYNC_CONFIG")
	if envExist {
		return envPath, nil
	}

	confDir, err := os.UserConfigDir()
	if err == nil {
		confPath := filepath.Join(confDir, "gitsync", "config.hcl")
		_, err := os.Stat(confPath)
		if err == nil {
			return confPath, nil
		}

		confPath = filepath.Join(confDir, "gitsync", "config.json")
		_, err = os.Stat(confPath)
		if err == nil {
			return confPath, nil
		}
	}

	_, errHcl := os.Stat("config.hcl")
	_, errJson := os.Stat("config.json")

	if errHcl == nil {
		return "config.hcl", nil
	} else if errJson == nil {
		return "config.hcl", nil
	}

	return "", errors.New("unable to determine config file location")
}
