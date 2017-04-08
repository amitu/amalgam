package acko

import (
	"database/sql/driver"
	"encoding/json"

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
