package log

import (
	"fmt"
	"path"
	"strings"

	"github.com/sydneyowl/clh-server/pkg/config"

	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

func _getAboveLevels(level slog.Level) []slog.Level {
	// get above levels
	abLevels := make([]slog.Level, 0)
	for _, lvl := range slog.AllLevels {
		if lvl <= level {
			abLevels = append(abLevels, lvl)
		}
	}
	return abLevels
}

func SetLogger(cfg config.Log, verbose bool) error {
	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	level, err := slog.StringToLevel(cfg.Level)
	if err != nil {
		availNames := make([]string, 0, len(slog.LevelNames))
		for _, v := range slog.LevelNames {
			availNames = append(availNames, v)
		}
		return fmt.Errorf("invalid log level: %s. should choose from: %s", cfg.Level, strings.Join(availNames, ", "))
	}

	dir := cfg.LogFileDirectory
	if dir == "" {
		dir = config.FallbackLogDir
	}
	h1 := handler.MustFileHandler(path.Join(dir, "clh.log"), handler.WithLogLevel(level), handler.WithUseJSON(true))

	slog.Std().Level = slog.InfoLevel
	if verbose {
		slog.Std().Level = slog.TraceLevel
	}

	if cfg.LogToFile {
		slog.PushHandler(h1)
	}

	return nil
}
