package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

type InsertContext struct {
	err          error
	sql          string
	params       []interface{}
	retLastIndex bool
	db           *sql.DB
	tx           *sql.Tx
}

func createInsertContext(db *sql.DB, tx *sql.Tx, table string, columns []string, dataset ...interface{}) *InsertContext {
	if len(dataset) == 0 {
		return &InsertContext{}
	}
	dataCount := len(dataset)
	placeholder := fmt.Sprintf("(%s)", placeholder(len(columns)))
	ps := make([]string, 0, dataCount)
	for i := 0; i < dataCount; i++ {
		ps = append(ps, placeholder)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("insert into %s (%s) values %s", table, strings.Join(columns, ","), strings.Join(ps, ",")))
	sql := sb.String()

	params, err := ReadValue(columns, FieldMapping(dataset[0]), dataset...)
	if err != nil {
		return &InsertContext{err: err}
	}
	return &InsertContext{sql: sql, params: params, retLastIndex: false, db: db, tx: tx}
}

func (a *InsertContext) LastIndex() *InsertContext {
	a.retLastIndex = true
	return a
}

func (a *InsertContext) Exec() (int64, error) {
	if a.err != nil { // 如果构建异常，不执行
		return 0, a.err
	}
	if a.params == nil || len(a.params) == 0 { // 无数据，不需要执行
		return 0, nil
	}
	var stat *sql.Stmt
	var err error
	if a.tx == nil {
		stat, err = a.db.Prepare(a.sql)
	} else {
		stat, err = a.tx.Prepare(a.sql)
	}
	if stat != nil {
		defer stat.Close()
	}
	if err != nil {
		return 0, err
	}
	rs, err := stat.Exec(a.params...)
	if err != nil {
		return 0, err
	}
	if a.retLastIndex {
		return rs.LastInsertId()
	}
	return rs.RowsAffected()
}

// 返回 InsertContext 构建过程中的异常
func (a *InsertContext) ContextError() error {
	return a.err
}

func placeholder(count int) string {
	s := make([]string, 0, count)
	for i := 0; i < count; i++ {
		s = append(s, "?")
	}
	return strings.Join(s, ",")
}

// 返回语句和参数
func (a *InsertContext) Desc() (string, []interface{}) {
	return a.sql, a.params
}
