package orm

import (
	"database/sql"
)

type Context struct {
	db *sql.DB
}

func CreateContext() *Context {
	if defaultDatasource != "" {
		return &Context{db: ds[defaultDatasource]}
	}
	// no datasource registered
	return nil
}

func CreateContextOf(datasource string) (*Context, error) {
	db, ok := ds[datasource]
	if !ok {
		return nil, InvalidDatasourceError{datasource: datasource}
	}
	return &Context{db: db}, nil
}

// dataset 支持指针/结构体/结构体数组/结构体指针数组
// 1. struct
// 2. *struct
// 3. []struct
// 4. []*struct
// 5. *[]struct
// 6. *[]*struct
// 不支持除结构体之外的类型 如 int, bool, float 等 也不支持多重指针如 **struct []**struct **[]struct 等
func (a *Context) Insert(table string, columns []string, dataset interface{}) *InsertContext {
	return createInsertContext(a.db, nil, table, columns, interfaceToArray(dataset)...)
}

func (a *Context) Delete(table string, where string) *DeleteContext {
	return createDeleteContext(a.db, nil, table, where)
}

func (a *Context) Update(table string, setCols []string, where string) *UpdateContext {
	return createUpdateContext(a.db, nil, table, setCols, where)
}

func (a *Context) Select(table string, columns []string, where string, params ...interface{}) *SelectContext {
	return createSelectContext(a.db, nil, table, columns, where, params...)
}

// 直接传入语句和参数的查询
func (a *Context) Search(sql string, params ...interface{}) *SelectContext {
	return &SelectContext{advanced: true, sql: sql, params: params, db: a.db}
}

func (a *Context) Begin() (*TransactionContext, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	return &TransactionContext{tx: tx}, nil
}
