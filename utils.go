package amalgam

import (
	"database/sql/driver"
	"encoding/json"
    "net"
    "net/http"
    "strings"
    "crypto/cipher"
    "crypto/aes"
    "encoding/base64"
    "strconv"

	"github.com/juju/errors"
)

//Custom data type for easy mapping of jsonB DB column to a GO struct member.
type PgJson map[string]interface{}

// This is the only method of the interface Valuer under the sql/driver package.
// Types implementing this interface can convert themselves to a driver
// acceptable value.
func (p PgJson) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan is the only method of the Scanner interface under the sql package.
// Scan assigns a value from the DB driver to the object that calls it
func (p *PgJson) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	err := json.Unmarshal(source, p)
	if err != nil {
		return err
	}

	return nil
}

func GetIPFromRequest(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	return ip, nil
}

func DecodeTracker(et string) (string, error) {
    et = strings.Replace(et, ".", "=", -1)

    e, err := base64.URLEncoding.DecodeString(et)
    if err != nil {
        return "", errors.Trace(err)
    }

    block, err := aes.NewCipher([]byte(Secret[:24]))
    if err != nil {
        return "", errors.Trace(err)
    }

    blockmode := cipher.NewCBCDecrypter(
        block, []byte(Secret[len(Secret)-16:]),
    )

    blockmode.CryptBlocks(e, e)

    tracker := strconv.Itoa(int(e[4]))

    return tracker, nil
}

func GetTrackerFromRequest(r *http.Request) (string){
    var tracker string = ""
    cookies := r.Cookies()
    for i := 0 ; i < len(cookies); i++ {
        cookie := cookies[i]
        if cookie.Name == "trackerid" {
            tracker = cookie.Value
            break
        }
    }

    return tracker
}



