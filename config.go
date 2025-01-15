package main

import (
	"errors"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"log/slog"
	"os"
	"os/user"
)

type ConfigFile struct {
	Entry    []ConfigDirectory `hcl:"entry,block"`
	TimeSync string            `hcl:"sync"`
	LogLevel string            `hcl:"log_level,optional"`
}

type ConfigDirectory struct {
	Path       string        `hcl:"local_path"`
	Repository string        `hcl:"repository"`
	Auth       *ConfigAuth   `hcl:"auth,block"`
	Author     *ConfigAuthor `hcl:"author,block"`
}

type ConfigAuth struct {
	Type     string `hcl:"type"`
	Username string `hcl:"username,optional"`
	Password string `hcl:"password,optional"`
}

type ConfigAuthor struct {
	Name  *string `hcl:"name,optional"`
	Email *string `hcl:"email,optional"`
}

func LoadConfig(file string) (conf ConfigFile, err error) {
	var funcs = make(map[string]function.Function)

	funcs["hostname"] = function.New(&function.Spec{
		Description: "Return hostname of the machine",
		Type:        function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			hostname, err := os.Hostname()
			return cty.StringVal(hostname), err
		},
	})
	funcs["username"] = function.New(&function.Spec{
		Description: "Return username of the running user",
		Type:        function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			usr, err := user.Current()
			return cty.StringVal(usr.Username), err
		},
	})

	ec := hcl.EvalContext{
		Functions: funcs,
	}

	err = hclsimple.DecodeFile(file, &ec, &conf)
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
