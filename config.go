package amalgam

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/namsral/flag"
)

var (
	Listen    = ":8001"
	DbName    = "claimator"
	DbHost    = "127.0.0.1"
	DbPort    = 5432
	DbUser    = "user"
	DbPass    = ""
	Verbosity = 4
	Secret    = ""
	Config    = "r2d2.conf"

	FLAGSET *flag.FlagSet = nil
)

func init() {
	filename, _ := os.Executable()
	Config = filepath.Base(filename) + ".conf"

	user, err := user.Current()
	if err == nil {
		DbUser = user.Username
	}

	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	FLAGSET = f

	f.StringVar(&Listen, "listen", Listen, "http address")
	f.IntVar(
		&Verbosity, "verbosity", Verbosity, "logging verbosity, 0: none, 4: all",
	)
	f.StringVar(&Config, "config", Config, "config file")
	f.StringVar(&DbName, "dbname", DbName, "database name")
	f.StringVar(&DbHost, "dbhost", DbHost, "database host")
	f.IntVar(&DbPort, "dbport", DbPort, "database port")
	f.StringVar(&DbUser, "dbuser", DbUser, "database user")
	f.StringVar(&DbPass, "dbpass", DbPass, "database password")
	f.StringVar(&Secret, "secret", Secret, "django secret key")
}

func Init() {
	if err := FLAGSET.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	log.Println("config_parsed", "args", os.Args[1:], "flags", FLAGSET.Args())

}
