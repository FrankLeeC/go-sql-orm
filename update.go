package orm

import (
	"database/sql"
	"errors"
	"strings"
)

type UpdateContext struct {
	err       error
	sql       string
	setCols   []string
	whereCols []string
	build     bool
	params    []interface{}
	db        *sql.DB
	tx        *sql.Tx
}

func createUpdateContext(db *sql.DB, tx *sql.Tx, table string, setCols []string, where string) *UpdateContext {
	if len(setCols) == 0 {
		ctx := &UpdateContext{build: false, err: errors.New(`no [set] columns to update`)}
		return ctx
	}
	where = strings.TrimSpace(where)
	if where == "" {
		ctx := &UpdateContext{build: false, err: errors.New(`for security. can't update without [where] parameter. to update all dataset, pass "1=1" to [where] parameter`)}
		return ctx
	}

	var sb strings.Builder
	for i, e := range setCols {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(e + " = ?")
	}
	sql := "update " + table + " set " + sb.String() + " where " + where

	return &UpdateContext{build: false, db: db, tx: tx, sql: sql, setCols: setCols}
}

// 直接传递所有参数
func (a *UpdateContext) Params(params ...interface{}) *UpdateContext {
	a.params = params
	a.build = true
	return a
}

// 通过反射data获取所有参数，将按照 setCols, whereCols 的顺序组成
// whereCols where后的字段，顺序必须与where中的字段完全一致
// 如 update table set a=? where b=? and c=? and d=?   那么 whereCols 必须是 []string{"b", "c", "d"}  否则参数顺序将错乱！
func (a *UpdateContext) ReflectParamsFrom(data interface{}, whereCols []string) *UpdateContext {
	a.whereCols = whereCols
	cols := append(a.setCols, a.whereCols...)
	params, err := ReadValue(cols, FieldMapping(data), data)
	if err != nil {
		return &UpdateContext{build: false, err: err}
	}
	a.params = params
	a.build = true
	return a
}

func execute(db *sql.DB, tx *sql.Tx, updelSQL string, params ...interface{}) (int64, error) {
	var stat *sql.Stmt
	var err error
	if tx == nil {
		stat, err = db.Prepare(updelSQL)
	} else {
		stat, err = tx.Prepare(updelSQL)
	}
	if stat != nil {
		defer stat.Close()
	}
	if err != nil {
		return 0, err
	}
	var rs sql.Result
	if len(params) == 0 {
		rs, err = stat.Exec()
	} else {
		rs, err = stat.Exec(params...)
	}
	if err != nil {
		return 0, err
	}
	return rs.RowsAffected()
}

func (a *UpdateContext) Exec() (int64, error) {
	if a.err != nil { // 如果构建异常，不执行
		return 0, a.err
	}
	if !a.build {
		return 0, errors.New("build failed because of no params passed")
	}
	if a.sql == "" {
		return 0, nil
	}
	return execute(a.db, a.tx, a.sql, a.params...)
}

// 返回 UpdateContext 构建过程中的异常
func (a *UpdateContext) ContextError() error {
	return a.err
}

// 返回语句和参数
func (a *UpdateContext) Desc() (string, []interface{}) {
	return a.sql, a.params
}
