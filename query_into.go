package amalgam

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/juju/errors"
)

const (
	KeyDBTransaction = "dbtx"
	KeyDB            = "db"
	KeySession       = "http-session-id"
	KeyConnInfo      = "conninfo"
)

func Ctx2SessionKey(ctx context.Context) (string, error) {
	val := ctx.Value(KeySession)
	if val == nil {
		return "", errors.New("session not in context")
	}
	sessionkey, ok := val.(string)
	if !ok {
		LOGGER.Error(
			"value_is_not_a_string",
			"value", val, "type", fmt.Sprintf("%T", val),
		)
		return "", errors.New("value is not a string")
	}

	return sessionkey, nil
}

func Ctx2Db(ctx context.Context) (*sqlx.DB, error) {
	val := ctx.Value(KeyDB)
	if val == nil {
		return nil, errors.New("db not in context")
	}
	db, ok := val.(*sqlx.DB)
	if !ok {
		LOGGER.Error(
			"value_is_not_a_DB",
			"value", val, "type", fmt.Sprintf("%T", val),
		)
		return nil, errors.New("value is not a DB")
	}
	return db, nil
}

func Ctx2ConnInfo(ctx context.Context) (string, error) {
	val := ctx.Value(KeyConnInfo)
	if val == nil {
		return "", errors.New("db not in context")
	}
	info, ok := val.(string)
	if !ok {
		LOGGER.Error(
			"value_is_not_a_string",
			"value", val, "type", fmt.Sprintf("%T", val),
		)
		return "", errors.New("value is not a string")
	}
	return info, nil
}

func Ctx2Tx(ctx context.Context) (*sqlx.Tx, error) {
	val := ctx.Value(KeyDBTransaction)
	if val == nil {
		return nil, errors.New("transaction not in context")
	}
	tx, ok := val.(*sqlx.Tx)
	if !ok {
		LOGGER.Error(
			"value_is_not_a_transaction",
			"value", val, "type", fmt.Sprintf("%T", val),
		)
		return nil, errors.New("value is not a transaction")
	}
	return tx, nil
}

func QueryIntoInt(
	ctx context.Context, q string, args ...interface{},
) (int, error) {
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return 0, errors.Trace(err)
	}

	i := 0
	err = tx.QueryRow(q, args...).Scan(&i)
	return i, errors.Trace(err)
}

func QueryIntoString(
	ctx context.Context, q string, args ...interface{},
) (string, error) {
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return "", errors.Trace(err)
	}

	var s string
	err = tx.QueryRow(q, args...).Scan(&s)
	return s, errors.Trace(err)
}

func QueryIntoStruct(
	ctx context.Context, v interface{}, q string, args ...interface{},
) error {
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(tx.QueryRowx(q, args...).StructScan(v))
}

// QueryIntoMap will return a map[string]interface{} representation of exactly
// one row only. This is because you cannot have a single map object which
// contains data for multiple rows
func QueryIntoMap(
	ctx context.Context, q string, args ...interface{},
) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return m, errors.Trace(err)
	}

	rows := tx.QueryRowx(q, args...)

	cols, err := rows.Columns()
	if err != nil {
		return m, errors.Trace(err)
	}

	// Create a slice of interface{}'s to represent each column,
	// and a second slice to contain pointers to each item in the
	// columns slice.
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i, _ := range columns {
		columnPointers[i] = &columns[i]
	}

	// Scan the result into the column pointers...
	if err := rows.Scan(columnPointers...); err != nil {
		return m, err
	}

	// Create our map, and retrieve the value for each column from the
	// pointers slice, storing it in the map with the name of the column
	// as the key.
	for i, colName := range cols {
		val := columnPointers[i].(*interface{})
		m[colName] = *val
	}

	return m, nil

}

func QueryIntoSlice(
	ctx context.Context, v interface{}, q string, args ...interface{},
) error {
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(tx.Select(v, q, args...))
}

func Exec(
	ctx context.Context, q string, args ...interface{},
) error {
	tx, err := Ctx2Tx(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	_, err = tx.Exec(q, args...)
	return errors.Trace(err)
}
