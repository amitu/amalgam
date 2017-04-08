package acko

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/juju/errors"
)

const (
	KeyWG = "wg"
)

func GetContext() (context.Context, error) {
	ctx := context.Background()

	wg := sync.WaitGroup{}
	defer wg.Wait()

	ctx = context.WithValue(ctx, KeyWG, wg)
	conninfo := fmt.Sprintf(
		"user=%s password='%s' host=%s port=%d dbname=%s sslmode=disable",
		DbUser, DbPass, DbHost, DbPort, DbName,
	)

	db, err := sqlx.Connect("postgres", conninfo)
	if err != nil {
		LOGGER.Error("db_connect_failed", "err", errors.ErrorStack(err))
		return nil, errors.Trace(err)
	}

	tx, err := db.Beginx()
	if err != nil {
		LOGGER.Error("db_tx_failed", "err", errors.ErrorStack(err))
		return nil, errors.Trace(err)
	}

	ctx = context.WithValue(ctx, KeyDBTransaction, tx)
	ctx = context.WithValue(ctx, KeyDB, db)
	ctx = context.WithValue(ctx, KeyConnInfo, conninfo)

	return ctx, nil
}
