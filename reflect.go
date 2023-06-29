package orm

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var tag string = "column"

// 自定义tag标签
func CustomTag(t string) {
	tag = t
}

// 将 i 转为 数组形式的 interface{}
// i 可以是 struct，pointer to struct, array, pointer to array(array of struct or pointer)
func interfaceToArray(i interface{}) []interface{} {
	tp := reflect.TypeOf(i)
	val := reflect.ValueOf(i)
	var rs []interface{}
	if tp.Kind() == reflect.Ptr { // i 是一个指针
		_val := reflect.Indirect(val)           // 指向的数据
		_tp := reflect.TypeOf(_val.Interface()) // 数据的类型
		if _tp.Kind() == reflect.Struct {       // 指向的数据是一个 struct
			rs = append(make([]interface{}, 0, 1), val.Interface())
		} else if _tp.Kind() == reflect.Array || _tp.Kind() == reflect.Slice { // 指向的数据是一个 数组    i: *[]Type  *[]*Type   _val: []Type  []*Type
			rs = make([]interface{}, 0, _val.Len())
			for idx := 0; idx < _val.Len(); idx++ {
				rs = append(rs, _val.Index(idx).Interface())
			}
		}
	} else if tp.Kind() == reflect.Slice || tp.Kind() == reflect.Array { // i 是一个数组   i: []Type  []*Type
		rs = make([]interface{}, 0, val.Len())
		for idx := 0; idx < val.Len(); idx++ {
			rs = append(rs, val.Index(idx).Interface())
		}
	} else if tp.Kind() == reflect.Struct { // i 是一个struct
		rs = append(make([]interface{}, 0, 1), val.Interface())
	}
	return rs
}

// columns 结构体对应的 column tag
// fn fieldMapping    column -> field name
// dataset 可以是指针数组，也可以是结构体数组  但不能是空数据
// 1. []struct
// 2. []*struct
// 但不支持除结构体以外的数据类型 如 []int  []*bool []**struct 等
// 返回的数据排列方式按照columns的顺序，如果dataset是数组，那么dataset中的每个数据都会按照columns的顺序排列
func ReadValue(columns []string, fn map[string]string, dataset ...interface{}) ([]interface{}, error) {
	if len(dataset) == 0 {
		return nil, errors.New("empty dataset")
	}
	if reflect.TypeOf(dataset[0]).Kind() == reflect.Ptr {
		return readValueOfPtr(columns, fn, dataset...)
	} else if reflect.TypeOf(dataset[0]).Kind() == reflect.Struct {
		return readValueOfStruct(columns, fn, dataset...)
	}
	return nil, errors.New("unsupported data type:" + reflect.TypeOf(dataset[0]).Kind().String())
}

// pts 必须是指针数组
func readValueOfPtr(columns []string, fn map[string]string, pts ...interface{}) ([]interface{}, error) {
	result := make([]interface{}, 0, len(columns)*len(pts))
	for _, t := range pts {
		params := make([]interface{}, 0, len(columns))
		tv := reflect.Indirect(reflect.ValueOf(t))
		for _, c := range columns {
			if _, ok := fn[c]; !ok {
				return nil, errors.New("can not find exported field with tag(" + tag + ") '" + c + "'")
			}
			v := tv.FieldByName(fn[c]).Interface()
			params = append(params, v)
		}
		result = append(result, params...)
	}
	return result, nil
}

// stucts 必须是结构体数组
func readValueOfStruct(columns []string, fn map[string]string, structs ...interface{}) ([]interface{}, error) {
	result := make([]interface{}, 0, len(columns)*len(structs))
	for _, t := range structs {
		params := make([]interface{}, 0, len(columns))
		tv := reflect.ValueOf(t)
		for _, c := range columns {
			if _, ok := fn[c]; !ok {
				return nil, errors.New("can not find exported field with tag(" + tag + ") '" + c + "'")
			}
			v := tv.FieldByName(fn[c]).Interface()
			params = append(params, v)
		}
		result = append(result, params...)
	}
	return result, nil
}

// result must be a pointer to slice or struct
func writeValue(columns []string, p []interface{}, result interface{}) error {
	m := make(map[string]interface{})
	for i := 0; i < len(columns); i++ {
		m[columns[i]] = p[i]
	}
	val := reflect.ValueOf(result) // pointer to T
	ind := reflect.Indirect(val)   // T
	var v reflect.Value
	kind := ind.Kind()
	if reflect.Slice == kind { // 数组
		if ind.Type().Elem().Kind() == reflect.Ptr { // 指针数组
			v = reflect.New(ind.Type().Elem().Elem())
		} else { // 结构体数组
			v = reflect.New(ind.Type().Elem())
		}
	} else if reflect.Struct == kind { // 结构体
		v = reflect.New(ind.Type())
	} else if reflect.Ptr == kind { // 结构体的指针
		v = reflect.New(ind.Type().Elem())
		// 不是结构体数组，也不是结构体
	} else { // 不是数组，也不是结构体，也不是结构体的指针
		return errors.New("result is not struct, or pointer to struct, or slice")
	}
	_v := reflect.TypeOf(v.Interface())
	num := _v.Elem().NumField()
	for index := 0; index < num; index++ {
		col := strings.TrimSpace(_v.Elem().Field(index).Tag.Get(tag))
		if col == "" { // this field is not tag-mapping
			continue
		}
		cv, ok := m[col]
		if !ok { // this field is not selected
			continue
		}
		if cv == nil { // 结果是空值
			continue
		}
		// fmt.Println(reflect.TypeOf(cv), cv, reflect.TypeOf([]uint8{}))
		field := v.Elem().Field(index)
		if field.CanSet() {
			switch v.Elem().Field(index).Type() {
			case type_string:
				if reflect.TypeOf(cv) == type_uint8_slice {
					_cv := reflect.ValueOf(cv)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
					}
					field.SetString(string(_b))
				} else if reflect.TypeOf(cv) == type_byte_slice {
					_cv := reflect.ValueOf(cv)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, _cv.Index(i).Interface().(byte))
					}
					field.SetString(string(_b))
				} else {
					return errors.New("only []uint8/[]byte type can be converted to string, provided type is " + reflect.TypeOf(cv).Name())
				}
			case type_uint8_slice, type_byte_slice:
				if reflect.TypeOf(cv) == type_uint8_slice {
					_cv := reflect.ValueOf(cv)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
					}
					field.Set(reflect.ValueOf(_b))
				} else if reflect.TypeOf(cv) == type_byte_slice {
					_cv := reflect.ValueOf(cv)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, _cv.Index(i).Interface().(byte))
					}
					field.Set(reflect.ValueOf(_b))
				} else {
					return errors.New("only []uint8/[]byte type can be converted to []uint8, provided type is " + reflect.TypeOf(cv).Name())
				}
			case type_int, type_int8, type_int16, type_int32, type_int64: // 字段定义的精度和实际的精度不一定相符合
				if val_int, ok := cv.(int); ok {
					field.SetInt(int64(val_int))
				} else if val_8, ok := cv.(int8); ok {
					field.SetInt(int64(val_8))
				} else if val_16, ok := cv.(int16); ok {
					field.SetInt(int64(val_16))
				} else if val_32, ok := cv.(int32); ok {
					field.SetInt(int64(val_32))
				} else if val_64, ok := cv.(int64); ok {
					field.SetInt(val_64)
				} else if val_u, ok := cv.(uint); ok {
					field.SetInt(int64(val_u))
				} else if val_u8, ok := cv.(uint8); ok {
					field.SetInt(int64(val_u8))
				} else if val_u16, ok := cv.(uint16); ok {
					field.SetInt(int64(val_u16))
				} else if val_u32, ok := cv.(uint32); ok {
					field.SetInt(int64(val_u32))
				} else if val_u64, ok := cv.(uint64); ok {
					field.SetInt(int64(val_u64))
				} else if val_uint8_slice, ok := cv.([]uint8); ok {
					_cv := reflect.ValueOf(val_uint8_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
					}
					val_64, err := strconv.ParseInt(string(_b), 10, 64)
					if err != nil {
						return errors.New("convert value from []uint8 to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetInt(val_64)
				} else if val_byte_slice, ok := cv.([]byte); ok {
					_cv := reflect.ValueOf(val_byte_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, _cv.Index(i).Interface().(byte))
					}
					val_64, err := strconv.ParseInt(string(_b), 10, 64)
					if err != nil {
						return errors.New("convert value from []byte to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetInt(val_64)
				}
			case type_float32, type_float64: // 字段定义的精度和实际的精度不一定相符合
				if _val32, ok := cv.(float32); ok {
					field.SetFloat(float64(_val32))
				} else if _val64, ok := cv.(float64); ok {
					field.SetFloat(float64(_val64))
				} else if val_uint8_slice, ok := cv.([]uint8); ok {
					_cv := reflect.ValueOf(val_uint8_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
					}
					fv, err := strconv.ParseFloat(string(_b), 64)
					if err != nil {
						return errors.New("convert value from []uint8 to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetFloat(fv)
				} else if val_byte_slice, ok := cv.([]byte); ok {
					_cv := reflect.ValueOf(val_byte_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, _cv.Index(i).Interface().(byte))
					}
					fv, err := strconv.ParseFloat(string(_b), 64)
					if err != nil {
						return errors.New("convert value from []byte to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetFloat(fv)
				}
			case type_uint, type_uint8, type_uint16, type_uint32, type_uint64: // 字段定义的精度和实际的精度不一定相符合
				if val_int, ok := cv.(int); ok {
					field.SetUint(uint64(val_int))
				} else if val_8, ok := cv.(int8); ok {
					field.SetUint(uint64(val_8))
				} else if val_16, ok := cv.(int16); ok {
					field.SetUint(uint64(val_16))
				} else if val_32, ok := cv.(int32); ok {
					field.SetUint(uint64(val_32))
				} else if val_64, ok := cv.(int64); ok {
					field.SetUint(uint64(val_64))
				} else if val_u, ok := cv.(uint); ok {
					field.SetUint(uint64(val_u))
				} else if val_u8, ok := cv.(uint8); ok {
					field.SetUint(uint64(val_u8))
				} else if val_u16, ok := cv.(uint16); ok {
					field.SetUint(uint64(val_u16))
				} else if val_u32, ok := cv.(uint32); ok {
					field.SetUint(uint64(val_u32))
				} else if val_u64, ok := cv.(uint64); ok {
					field.SetUint(uint64(val_u64))
				} else if val_uint8_slice, ok := cv.([]uint8); ok {
					_cv := reflect.ValueOf(val_uint8_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
					}
					val_u64, err := strconv.ParseUint(string(_b), 10, 64)
					if err != nil {
						return errors.New("convert value from []uint8 to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetUint(val_u64)
				} else if val_byte_slice, ok := cv.([]byte); ok {
					_cv := reflect.ValueOf(val_byte_slice)
					_b := make([]byte, 0, _cv.Len())
					for i := 0; i < _cv.Len(); i++ {
						_b = append(_b, _cv.Index(i).Interface().(byte))
					}
					val_u64, err := strconv.ParseUint(string(_b), 10, 64)
					if err != nil {
						return errors.New("convert value from []byte to " + v.Elem().Field(index).Type().Name() + " error:" + err.Error())
					}
					field.SetUint(val_u64)
				}
			case type_time:
				_cv := reflect.ValueOf(cv)
				field.Set(_cv)
			}
		}
	}
	if reflect.Slice == kind {
		if ind.Type().Elem().Kind() == reflect.Ptr {
			val = reflect.Append(reflect.Indirect(val), v.Elem().Addr())
			ind.Set(val)
		} else {
			val = reflect.Append(reflect.Indirect(val), v.Elem())
			ind.Set(val)
		}
	} else if reflect.Struct == kind {
		ind.Set(v.Elem())
	} else if reflect.Ptr == kind {
		ind.Set(v.Elem().Addr())
	}
	return nil
}

var (
	const_string      string    = ""
	const_int         int       = math.MaxInt
	const_int8        int8      = math.MaxInt8
	const_int16       int16     = math.MaxInt16
	const_int32       int32     = math.MaxInt32
	const_int64       int64     = math.MaxInt64
	const_float32     float32   = math.MaxFloat32
	const_float64     float64   = math.MaxFloat64
	const_uint        uint      = math.MaxUint
	const_uint8       uint8     = math.MaxUint8
	const_uint16      uint16    = math.MaxUint16
	const_uint32      uint32    = math.MaxUint32
	const_uint64      uint64    = math.MaxUint64
	const_time        time.Time = time.Now()
	const_uint8_slice []uint8   = []uint8{}
	const_byte_slice  []byte    = []byte{}
)

var (
	type_string      = reflect.TypeOf(const_string)
	type_int         = reflect.TypeOf(const_int)
	type_int8        = reflect.TypeOf(const_int8)
	type_int16       = reflect.TypeOf(const_int16)
	type_int32       = reflect.TypeOf(const_int32)
	type_int64       = reflect.TypeOf(const_int64)
	type_float32     = reflect.TypeOf(const_float32)
	type_float64     = reflect.TypeOf(const_float64)
	type_uint        = reflect.TypeOf(const_uint)
	type_uint8       = reflect.TypeOf(const_uint8)
	type_uint16      = reflect.TypeOf(const_uint16)
	type_uint32      = reflect.TypeOf(const_uint32)
	type_uint64      = reflect.TypeOf(const_uint64)
	type_time        = reflect.TypeOf(const_time)
	type_uint8_slice = reflect.TypeOf(const_uint8_slice)
	type_byte_slice  = reflect.TypeOf(const_byte_slice)
)
