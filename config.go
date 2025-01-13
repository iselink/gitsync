package main

import (
	"errors"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"log/slog"
)

type ConfigFile struct {
	Entry    []ConfigDirectory `hcl:"entry,block"`
	TimeSync string            `hcl:"sync"`
	LogLevel string            `hcl:"log_level,optional"`
}

type ConfigDirectory struct {
	Path       string      `hcl:"local_path"`
	Repository string      `hcl:"repository"`
	Auth       *ConfigAuth `hcl:"auth,block"`
}

type ConfigAuth struct {
	Type     string `hcl:"type"`
	Username string `hcl:"username,optional"`
	Password string `hcl:"password,optional"`
}

func LoadConfig(file string) (conf ConfigFile, err error) {
	err = hclsimple.DecodeFile(file, nil, &conf)
	return
}

func CheckAndConfigure(conf *ConfigFile) error {
	if conf.LogLevel != "" {
		err := setLogLevel(conf.LogLevel)
		if err != nil {
			return err
		}
	}

	return nil
}

func setLogLevel(level string) error {
	switch level {
	case "debug":
		{
			slog.SetLogLoggerLevel(slog.LevelDebug)
			slog.Info("Set logging level to debug")
			break
		}
	case "info":
		{
			slog.SetLogLoggerLevel(slog.LevelInfo)
			slog.Info("Set logging level to info")
			break
		}
	case "warn":
		{
			slog.SetLogLoggerLevel(slog.LevelWarn)
			break
		}
	case "error":
		{
			slog.SetLogLoggerLevel(slog.LevelError)
			break
		}
	default:
		{
			return errors.New("unknown value '" + level + "'")
		}
	}

	return nil
}
