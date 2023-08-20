package transaction

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
)

type Factory struct {
	db *sql.DB
}

func (f Factory) BeginTx(ct context.Context, sqlTxOpts *sql.TxOptions) (*Transaction, *errs.Error) {
	if f.db == nil {
		return newTransaction(nil), nil
	}

	tx, err := f.db.BeginTx(ct, sqlTxOpts)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return newTransaction(tx), nil
}

func NewFactory(db *sql.DB) Factory {
	return Factory{db: db}
}
