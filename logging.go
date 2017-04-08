package acko

import (
	"fmt"
	"log/syslog"
	"os"
	"sync"

	"github.com/inconshreveable/log15"
)

type LoggerStatsType struct {
	Warns  int `json:"warns"`
	Errors int `json:"errors"`
	Crits  int `json:"crits"`
}

var (
	LOGGER            log15.Logger
	LOGGER_STATS_Lock sync.Mutex
	LOGGER_STATS      = &LoggerStatsType{}
)

func LoggerStatter(r *log15.Record) error {
	LOGGER_STATS_Lock.Lock()
	defer LOGGER_STATS_Lock.Unlock()

	switch r.Lvl {
	case log15.LvlWarn:
		LOGGER_STATS.Warns += 1
	case log15.LvlError:
		LOGGER_STATS.Errors += 1
	case log15.LvlCrit:
		LOGGER_STATS.Crits += 1
	default:
	}

	return nil
}

func LvlFilterHandler(exactLvl log15.Lvl, h log15.Handler) log15.Handler {
	return log15.FilterHandler(func(r *log15.Record) (pass bool) {
		return r.Lvl == exactLvl
	}, h)
}

// Logging verbosity level, from 0 (nothing) upwards.
func SetLoggingVerbosity(level int) {
	// TODO: make this guy pick the binary name instead of hardcoding
	logfile_error := fmt.Sprintf("/tmp/claimator__errorLogger.log")
	logfile_info := fmt.Sprintf("/tmp/claimator__infoLogger.log")
	logfile_general := fmt.Sprintf("/tmp/claimator__generalLogger.log")

	LOGGER = log15.Root()

	LOGGER.SetHandler(
		log15.MultiHandler(
			log15.LvlFilterHandler(
				log15.Lvl(level),
				log15.CallerFuncHandler(
					log15.CallerStackHandler(
						"%+v",
						log15.MultiHandler(
							log15.Must.FileHandler(logfile_general, log15.LogfmtFormat()),
							log15.Must.SyslogNetHandler("udp",
								"logs5.papertrailapp.com:53174",
								syslog.LOG_DEBUG|syslog.LOG_USER, "claimator",
								log15.LogfmtFormat()),
							log15.StreamHandler(os.Stderr, log15.TerminalFormat()),
							log15.FuncHandler(LoggerStatter),
						),
					),
				),
			),
			log15.LvlFilterHandler(
				log15.Lvl(log15.LvlError),
				log15.CallerFuncHandler(
					log15.CallerStackHandler(
						"%+v",
						log15.Must.FileHandler(logfile_error, log15.LogfmtFormat()),
					),
				),
			),
			LvlFilterHandler(
				log15.Lvl(log15.LvlInfo),
				log15.Must.FileHandler(logfile_info, log15.LogfmtFormat()),
			)),
	)

	LOGGER.Debug(
		"logger_initialized", log15.Ctx{
			"logfile": logfile_general,
			"pid":     os.Getpid(),
			"level":   log15.Lvl(level).String(),
		},
	)
}
