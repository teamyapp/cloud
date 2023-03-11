package transaction

import (
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
)

type Command struct {
	Execute func() *errs.Error
	Undo    func() *errs.Error
}

type Transaction struct {
	sqlTx       *sql.Tx
	commands    []Command
	isCommitted bool
}

func (t *Transaction) SQLTx() *sql.Tx {
	return t.sqlTx
}

func (t *Transaction) ExecuteCommand(command Command) *errs.Error {
	t.commands = append(t.commands, command)
	err := command.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (t *Transaction) Commit() *errs.Error {
	if t.sqlTx != nil {
		err := t.sqlTx.Commit()
		if err != nil {
			return &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
		}
	}

	t.isCommitted = true
	return nil
}

func (t *Transaction) Rollback() *errs.Error {
	if t.isCommitted {
		return nil
	}

	if t.sqlTx != nil {
		err := t.sqlTx.Rollback()
		if err != nil {
			return &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
		}
	}

	for index := len(t.commands) - 1; index >= 0; index-- {
		undoFunc := t.commands[index].Undo
		if undoFunc != nil {
			err := undoFunc()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func newTransaction(sqlTx *sql.Tx) *Transaction {
	return &Transaction{
		sqlTx:    sqlTx,
		commands: []Command{},
	}
}
