package orm

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

type SelectContext struct {
	err      error
	sql      string
	params   []interface{}
	db       *sql.DB
	tx       *sql.Tx
	step     int  // 构建过程步骤
	advanced bool // search 模式
	ordered  bool // 是否设置过order by
}

func createSelectContext(db *sql.DB, tx *sql.Tx, table string, columns []string, where string, params ...interface{}) *SelectContext {
	var cs string
	if len(columns) == 0 {
		cs = " * "
	} else {
		cs = strings.Join(columns, ",")
	}
	sql := "select " + cs + " from " + table
	where = strings.TrimSpace(where)
	if where != "" {
		sql += " where " + where
	}
	return &SelectContext{advanced: false, step: 1, sql: sql, params: params, db: db, tx: tx}
}

func (a *SelectContext) GroupBy(cols ...string) *SelectContext {
	if a.err != nil {
		return a
	}
	if a.advanced {
		a.err = errors.New("can not use GroupBy when build by Search")
		return a
	}
	if a.step >= 2 {
		a.err = errors.New("GroupBy can only be used after createSelectContext")
		return a
	}
	a.step = 2
	if len(cols) > 0 {
		a.sql += " group by " + strings.Join(cols, ",")
	}
	return a
}

func (a *SelectContext) OrderByAsc(col string) *SelectContext {
	if a.err != nil {
		return a
	}
	if a.advanced {
		a.err = errors.New("can not use OrderByAsc when build by Search")
		return a
	}
	if a.step > 3 {
		a.err = errors.New("can not use OrderByAsc after Limit")
		return a
	}
	a.step = 3
	if !a.ordered {
		a.sql += " order by " + col + " asc "
		a.ordered = true
	} else {
		a.sql += ", " + col + " asc "
	}
	return a
}

func (a *SelectContext) OrderByDesc(col string) *SelectContext {
	if a.err != nil {
		return a
	}
	if a.advanced {
		a.err = errors.New("can not use OrderByDesc when build by Search")
		return a
	}
	if a.step > 3 {
		a.err = errors.New("can not use OrderByDesc after Limit")
		return a
	}
	a.step = 3
	if !a.ordered {
		a.sql += " order by " + col + " desc "
		a.ordered = true
	} else {
		a.sql += ", " + col + " desc "
	}
	return a
}

// 通过 Search 创建的无法使用 limit
func (a *SelectContext) Limit(offset, size int) *SelectContext {
	if a.err != nil {
		return a
	}
	if a.advanced {
		a.err = errors.New("can not use Limit when build by Search")
		return a
	}
	a.step = 4
	a.sql += " limit ?, ?"
	a.params = append(a.params, offset, size)
	return a
}

func (a *SelectContext) Result(r interface{}) error {
	if a.err != nil { // 如果构建异常，不执行
		return a.err
	}
	if reflect.TypeOf(r).Kind() != reflect.Ptr {
		a.err = new(InvalidResultTypeError)
		return a.err
	}
	val := reflect.ValueOf(r)    // pointer to T
	ind := reflect.Indirect(val) // T
	kind := ind.Kind()
	if kind != reflect.Slice && kind != reflect.Struct && kind != reflect.Ptr {
		a.err = new(InvalidResultTypeError)
		return a.err
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
		return err
	}

	var rows *sql.Rows
	if a.params == nil || len(a.params) == 0 {
		rows, err = stat.Query()
	} else {
		rows, err = stat.Query(a.params...)
	}
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return err
	}
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}
		res := make([]interface{}, len(columns))
		v := make([]interface{}, len(columns))
		for i := range v {
			res[i] = &v[i]
		}
		rows.Scan(res...)
		err = writeValue(columns, v, r)
		if err != nil {
			a.err = err
			return err
		}
	}
	return nil
}

// 返回 SelectContext 构建过程中的异常
func (a *SelectContext) ContextError() error {
	return a.err
}

// 返回语句和参数
func (a *SelectContext) Desc() (string, []interface{}) {
	return a.sql, a.params
}
