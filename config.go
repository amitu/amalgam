package amalgam

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/getsentry/raven-go"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

var (
	Listen                       = ":8000"
	DbName                       = ""
	DbHost                       = "127.0.0.1"
	DbPort                       = 5432
	DbUser                       = "user"
	DbPass                       = ""
	Verbosity                    = 4
	Secret                       = ""
	Config                       = ""
	CreateConf                   = false
	Debug                        = true
	Sentry                       = ""
	UseSession                   = true
	UseTransaction               = true
	StatsD                       = ""
	App                          = "acko"
	FLAGSET        *flag.FlagSet = nil

	Confs map[string]interface{}
)

func StringFlag(
	varName *string,
	name string,
	defVal string,
	description string,
) {
	Confs[name] = defVal
	FLAGSET.StringVar(varName, name, defVal, description)
}

func IntFlag(
	varName *int,
	name string,
	defVal int,
	description string,
) {
	Confs[name] = defVal
	FLAGSET.IntVar(varName, name, defVal, description)
}

func BoolFlag(
	varName *bool,
	name string,
	defVal bool,
	description string,
) {
	Confs[name] = defVal
	FLAGSET.BoolVar(varName, name, defVal, description)
}

func init() {
	filename, _ := os.Executable()
	Config = filepath.Base(filename) + ".conf"

	Confs = make(map[string]interface{})

	user, err := user.Current()
	if err == nil {
		DbUser = user.Username
	}

	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FLAGSET = f

	StringFlag(&Listen, "listen", Listen, "http address")
	IntFlag(
		&Verbosity, "verbosity", Verbosity,
		"logging verbosity, 0: none, 4: all",
	)
	StringFlag(&DbName, "dbname", DbName, "database name")
	StringFlag(&DbHost, "dbhost", DbHost, "database host")
	IntFlag(&DbPort, "dbport", DbPort, "database port")
	StringFlag(&DbUser, "dbuser", DbUser, "database user")
	StringFlag(&DbPass, "dbpass", DbPass, "database password")
	StringFlag(&Secret, "secret", Secret, "django secret key")
	BoolFlag(&CreateConf, "create-conf", CreateConf, "")
	BoolFlag(&Debug, "debug", Debug, "")
	BoolFlag(
		&UseSession, "session",
		UseSession, "Should sessions be a part of middleware",
	)
	BoolFlag(
		&UseTransaction, "transaction",
		UseTransaction, "Should transactions be handled",
	)
	StringFlag(&Sentry, "sentry", Sentry, "sentry endpoint")
	StringFlag(&StatsD, "statsd", StatsD, "statsD endpoint")
	StringFlag(&App, "app", App, "the app in use")
}

func Init() {
	if err := FLAGSET.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if CreateConf {
		n, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}

		prodDir := filepath.Dir(n)

		exName, _ := os.Executable()

		confFile := filepath.Join(prodDir, filepath.Base(exName)+".conf")

		if _, err := os.Stat(confFile); err == nil {
			panic("conf files already present!")
		} else {
			writeConfFile(confFile)
		}
	}

	StringFlag(&Config, "config", Config, "config file")

	if err := FLAGSET.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	log.Println("config_parsed", "args", os.Args[1:], "flags", FLAGSET.Args())

	raven.SetDSN(Sentry)
	statsdInit()

}

func writeConfFile(confFile string) {
	f, err := os.Create(confFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for key, val := range Confs {
		var stringVal string
		var intVal int
		var boolVal bool
		var ok bool

		stringVal, ok = val.(string)
		if ok {
			f.WriteString(fmt.Sprintf("%s %s\n", key, stringVal))
		}

		intVal, ok = val.(int)
		if ok {
			f.WriteString(fmt.Sprintf("%s %d\n", key, intVal))
		}

		boolVal, ok = val.(bool)
		if ok {
			f.WriteString(fmt.Sprintf("%s %t\n", key, boolVal))
		}
	}

}
