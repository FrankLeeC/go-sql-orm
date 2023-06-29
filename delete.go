package orm

import (
	"database/sql"
	"errors"
	"strings"
)

type DeleteContext struct {
	err       error
	sql       string
	whereCols []string
	build     bool
	params    []interface{}
	db        *sql.DB
	tx        *sql.Tx
}

func createDeleteContext(db *sql.DB, tx *sql.Tx, table string, where string) *DeleteContext {
	if strings.TrimSpace(where) == "" {
		return &DeleteContext{build: false, err: errors.New(`for security. can't delete without [where] parameter. to delete all dataset, pass "1=1" to [where] parameter`)}
	}
	sql := "delete from " + table + " where " + where
	return &DeleteContext{build: false, db: db, tx: tx, sql: sql}
}

// 直接传递所有参数
func (a *DeleteContext) Params(params ...interface{}) *DeleteContext {
	a.params = params
	a.build = true
	return a
}

// 通过反射data获取所有参数，将按照 setCols, whereCols 的顺序组成
// whereCols where后的字段，顺序必须与where中的字段完全一致
// 如 update table set a=? where b=? and c=? and d=?   那么 whereCols 必须是 []string{"b", "c", "d"}  否则参数顺序将错乱
func (a *DeleteContext) ReflectParamsFrom(data interface{}, whereCols []string) *DeleteContext {
	a.whereCols = whereCols
	if len(a.whereCols) > 0 {
		params, err := ReadValue(a.whereCols, FieldMapping(data), data)
		if err != nil {
			return &DeleteContext{build: false, err: err}
		}
		a.params = params
	}
	a.build = true
	return a
}

func (a *DeleteContext) Exec() (int64, error) {
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

// 返回 DeleteContext 构建过程中的异常
func (a *DeleteContext) ContextError() error {
	return a.err
}

// 返回语句和参数
func (a *DeleteContext) Desc() (string, []interface{}) {
	return a.sql, a.params
}
