package orm

import (
	"reflect"
	"strings"
)

// 返回所有的column字段，除了 excepts
// t 可以是 struct 也可以是 *struct
func ColumnsExcept(t interface{}, excepts ...string) []string {
	m := make(map[string]int, len(excepts))
	for i := 0; i < len(excepts); i++ {
		m[excepts[i]] = 1
	}
	rs := make([]string, 0)
	tp := reflect.TypeOf(t)
	if tp.Kind() == reflect.Struct {
		p := reflect.PtrTo(reflect.TypeOf(t))
		for i := 0; i < p.Elem().NumField(); i++ {
			column := strings.TrimSpace(p.Elem().Field(i).Tag.Get(tag))
			if _, ok := m[column]; ok {
				continue
			}
			rs = append(rs, column)
		}
	} else if tp.Kind() == reflect.Ptr {
		p := reflect.TypeOf(t)
		for i := 0; i < p.Elem().NumField(); i++ {
			column := strings.TrimSpace(p.Elem().Field(i).Tag.Get(tag))
			if _, ok := m[column]; ok {
				continue
			}
			rs = append(rs, column)
		}
	}
	return rs
}

// 获取 结构体/结构体的指针 t 的 column tag 与 field name 对应关系
// t 必须是 strcut 或者 指向 strcut 的指针
func FieldMapping(t interface{}) map[string]string {
	fn := make(map[string]string)
	tp := reflect.TypeOf(t)
	if tp.Kind() == reflect.Struct {
		p := reflect.PtrTo(reflect.TypeOf(t))
		for i := 0; i < p.Elem().NumField(); i++ {
			column := strings.TrimSpace(p.Elem().Field(i).Tag.Get(tag))
			fn[column] = p.Elem().Field(i).Name
		}
	} else if tp.Kind() == reflect.Ptr {
		p := reflect.TypeOf(t)
		for i := 0; i < p.Elem().NumField(); i++ {
			column := strings.TrimSpace(p.Elem().Field(i).Tag.Get(tag))
			fn[column] = p.Elem().Field(i).Name
		}
	}
	return fn
}
