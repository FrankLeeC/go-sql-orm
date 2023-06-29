package orm

import (
	"database/sql"
)

type TransactionContext struct {
	tx *sql.Tx
}

func (a *TransactionContext) Insert(table string, columns []string, dataset interface{}) *InsertContext {
	return createInsertContext(nil, a.tx, table, columns, interfaceToArray(dataset)...)
}

func (a *TransactionContext) Delete(table string, where string) *DeleteContext {
	return createDeleteContext(nil, a.tx, table, where)
}

func (a *TransactionContext) Update(table string, setCols []string, where string) *UpdateContext {
	return createUpdateContext(nil, a.tx, table, setCols, where)
}

func (a *TransactionContext) Select(table string, columns []string, where string, params ...interface{}) *SelectContext {
	return createSelectContext(nil, a.tx, table, columns, where, params...)
}

func (a *TransactionContext) Search(sql string, params ...interface{}) *SelectContext {
	return &SelectContext{advanced: true, sql: sql, params: params, tx: a.tx}
}

func (a *TransactionContext) Rollback() error {
	return a.tx.Rollback()
}

func (a *TransactionContext) Commit() error {
	return a.tx.Commit()
}
