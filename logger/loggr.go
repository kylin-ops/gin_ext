package logger

import (
	"log"

	"github.com/kylin-ops/gin_ext"
	"github.com/kylin-ops/gin_ext/logger/jlog"
	"github.com/kylin-ops/gin_ext/logger/slog"
)

type Options struct {
	Console      bool   `yaml:"console" json:"console"`
	File         bool   `yaml:"file" json:"file"`
	Level        string `yaml:"level" json:"level"`
	Path         string `yaml:"path" json:"path"`
	RollbackTime int    `yaml:"rollback_time" json:"rollback_time"`
	Count        int    `yaml:"count" json:"count"`
}

func NewLogger(o *Options) {
	joption := jlog.Options{
		Console:      o.Console,
		File:         o.File,
		Level:        o.Level,
		Path:         o.Path + ".json.log",
		RollbackTime: o.RollbackTime,
		Count:        o.Count,
	}
	soption := slog.LogOption{
		Console:      o.Console,
		File:         o.File,
		Level:        o.Level,
		Path:         o.Path,
		RollbackTime: o.RollbackTime,
		Count:        o.Count,
	}
	jlogger, err := jlog.NewLogger(&joption)
	if err != nil {
		log.Fatalf("load logger error, error info: %s", err.Error())
	}
	slogger, err := slog.NewLogger(&soption)
	if err != nil {
		log.Fatalf("load logger error, error info: %s", err.Error())
	}
	gin_ext.JLogger = jlogger
	gin_ext.SLogger = slogger
}
