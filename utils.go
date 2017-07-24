package amalgam

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"

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

type NullPgJson struct {
	PgJson PgJson
	Valid  bool
}

func (p NullPgJson) Value() (driver.Value, error) {
	if !p.Valid {
		return nil, nil
	}
	j, err := json.Marshal(p)
	return j, err
}

func (p *NullPgJson) Scan(src interface{}) error {
	if src == nil {
		p.Valid = false
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	js := PgJson{}
	err := json.Unmarshal(source, &js)
	if err != nil {
		return err
	}

	p.Valid = true
	p.PgJson = js

	return nil
}

type AError struct {
	Human string `json:"human"`
}

func GetIPFromRequest(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", errors.Trace(err)
	}

	return ip, nil
}

func EncodeID(id int64, model string) string {
	crc := crc32.ChecksumIEEE(make([]byte, id)) & 0xffffffff
	message := make([]byte, 0)

	mm := make([]byte, 4)
	binary.LittleEndian.PutUint32(mm, crc)
	message = append(message, mm...)

	mm = make([]byte, 8)
	binary.LittleEndian.PutUint64(mm, uint64(id))
	message = append(message, mm...)

	mm = make([]byte, 4)
	message = append(message, mm...)

	block, err := aes.NewCipher([]byte(Secret[:32]))
	if err != nil {
		panic(err)
	}

	t := sha256.Sum256([]byte(Secret + model))
	iv := t[:16]

	blockmode := cipher.NewCBCEncrypter(block, iv)

	blockmode.CryptBlocks(message, message)

	tt := base64.URLEncoding.EncodeToString(message)

	eid := strings.Replace(string(tt), "=", "", -1)

	return eid

}

func DecodeEID(eid string, model string) (string, error) {
	if len(eid)%3 != 0 {
		rem := len(eid) % 3
		for i := 3; i > rem; i-- {
			eid = eid + "="
		}
	}

	e, err := base64.URLEncoding.DecodeString(eid)
	if err != nil {
		return "", errors.Trace(err)
	}

	block, err := aes.NewCipher([]byte(Secret[:32]))
	if err != nil {
		return "", errors.Trace(err)
	}

	t := sha256.Sum256([]byte(Secret + model))
	iv := t[:16]

	blockmode := cipher.NewCBCDecrypter(block, iv)

	blockmode.CryptBlocks(e, e)

	b := int(e[4])
	var exp = 1
	for i := 5; i < len(e); i++ {
		if int(e[i]) != 0 {
			b += int(math.Pow(float64(256), float64(exp))) * int(e[i])
		}
		exp++
	}

	tracker := strconv.Itoa(b)

	return tracker, nil
}

func GetTrackerFromRequest(r *http.Request) string {
	var tracker string = ""
	cookies := r.Cookies()
	for i := 0; i < len(cookies); i++ {
		cookie := cookies[i]
		if cookie.Name == "trackerid" {
			tracker = cookie.Value
			break
		}
	}

	return tracker
}
