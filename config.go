package amalgam

import (
    "log"
    "os"
    "os/user"
    "path/filepath"
    "fmt"

    _ "github.com/lib/pq"
    "github.com/namsral/flag"
)

var (
    Listen = ":8000"
    DbName = ""
    DbHost = "127.0.0.1"
    DbPort = 5432
    DbUser = "user"
    DbPass = ""
    Verbosity = 4
    Secret = ""
    Config = ""

    FLAGSET *flag.FlagSet = nil

    Confs map[string]interface{}
)

func CreateFlag(
varName interface{},
name string,
defVal interface{},
description string,
) {
    var intVal int
    var stringVal string
    var ok = false

    stringVal, ok = defVal.(string)
    if !ok {
        intVal, ok = defVal.(int)
        if !ok {
            panic("Unhandled flag type!")
        }

        Confs[name] = defVal

        v, ok := varName.(*int)
        if !ok {
            panic("wrong pointer type")
        }
        FLAGSET.IntVar(v, name, intVal, description)

    } else {
        Confs[name] = defVal

        v, ok := varName.(*string)
        if !ok {
            panic("wrong pointer type")
        }
        FLAGSET.StringVar(v, name, stringVal, description)
    }
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

    CreateFlag(&Listen, "listen", Listen, "http address")
    CreateFlag(
        &Verbosity, "verbosity", Verbosity,
        "logging verbosity, 0: none, 4: all",
    )
    CreateFlag(&Config, "config", Config, "config file")
    CreateFlag(&DbName, "dbname", DbName, "database name")
    CreateFlag(&DbHost, "dbhost", DbHost, "database host")
    CreateFlag(&DbPort, "dbport", DbPort, "database port")
    CreateFlag(&DbUser, "dbuser", DbUser, "database user")
    CreateFlag(&DbPass, "dbpass", DbPass, "database password")
    CreateFlag(&Secret, "secret", Secret, "django secret key")
}

func Init() {
    if 2 == len(os.Args) && os.Args[1] == "confs" {
        n, err := filepath.Abs(filepath.Dir(os.Args[0]))
        if err != nil {
            log.Fatal(err)
        }

        prodDir := filepath.Dir(n)

        exName, _ := os.Executable()

        confFile := filepath.Join(prodDir, filepath.Base(exName) + ".conf")

        if _, err := os.Stat(confFile); err == nil {
            panic("conf files already present!")
        } else {
            writeConfFile(confFile)
        }
    }

    if err := FLAGSET.Parse(os.Args[1:]); err != nil {
        log.Fatal(err)
    }

    log.Println("config_parsed", "args", os.Args[1:], "flags", FLAGSET.Args())

}

func writeConfFile(confFile string ) {
    f, err := os.Create(confFile)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    for key, val := range(Confs) {
        var stringVal string
        var intVal int
        var ok bool

        stringVal, ok = val.(string)
        if ok {
            f.WriteString(fmt.Sprintf("%s %s\n", key, stringVal))
        }

        intVal, ok = val.(int)
        if ok {
            f.WriteString(fmt.Sprintf("%s %d\n", key, intVal))
        }
    }

}
