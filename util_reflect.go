package main

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type rflt struct {}

// func new_slice_ptr[T any](len, cap int) []*T {
// 	o  := new(T)
// 	ot := reflect.TypeOf(o)
// 	st := reflect.SliceOf(ot)
// 	return reflect.MakeSlice(st, len, cap).Interface().([]*T)
// }
// func new_slice[T any](len, cap int) []T {
// 	o  := new(T)
// 	ot := reflect.TypeOf(*o)
// 	st := reflect.SliceOf(ot)
// 	return reflect.MakeSlice(st, len, cap).Interface().([]T)
// }

func (rfl *rflt) objValue(obj any) *reflect.Value {
	objV := reflect.ValueOf(obj)
	if objV.Kind() == reflect.Pointer {
		objV = objV.Elem()
	}
	return &objV
}

func (rfl *rflt) fields(obj any) []string {
	flds := []string{}
	objV := rfl.objValue(obj)
	for _, vf := range reflect.VisibleFields(objV.Type()) {
		if vf.IsExported() {
			flds = append(flds, vf.Name)
		}
	}
	return flds
}

func (rfl *rflt) getFieldValueAsString(obj any, name string) string {
	objV := rfl.objValue(obj)
	fld  := objV.FieldByName(name)
	fv   := ""
	if fld.Kind() == reflect.String {
		fv = fld.String()
	} else if fld.CanFloat() {
		fv = fmt.Sprintf("%f", fld.Float())
	} else if fld.CanInt() {
		fv = fmt.Sprintf("%d", fld.Int())
	} else if fld.CanUint() {
		fv = fmt.Sprintf("%d", fld.Uint())
	} else if fld.Kind() == reflect.Bool {
		fv = fmt.Sprintf("%v", fld.Bool())
	} else if fld.Kind() == reflect.Slice {
		var sb bytes.Buffer
		for a := 0; a < fld.Len(); a++ {
			if sb.Len() > 1 {
				sb.WriteString(",")
			}
			e := fld.Index(a).Interface()
			switch val := e.(type) {
			case string:
				sb.WriteString(val)
			case int:
				sb.WriteString(fmt.Sprintf("%d", val))
			case int32:
				sb.WriteString(fmt.Sprintf("%d", val))
			case int64:
				sb.WriteString(fmt.Sprintf("%d", val))
			case bool:
				sb.WriteString(fmt.Sprintf("%v", val))
			default:
				sb.WriteString(" ")
			}
		}
		fv = sb.String()
	}
	return fv
}

func (rfl *rflt) getFieldValueAsInt64(obj any, name string) int64 {
	objV := rfl.objValue(obj)
	fld  := objV.FieldByName(name)
	fv   := int64(0)
	if fld.Kind() == reflect.String {
		fv = int64(0)
	} else if fld.CanFloat() {
		fv = int64(fld.Float())
	} else if fld.CanInt() {
		fv = int64(fld.Int())
	} else if fld.CanUint() {
		fv = int64(fld.Uint())
	} else if fld.Kind() == reflect.Bool {
		if fld.Bool() {
			fv = int64(1)
		}
	} else if fld.Kind() == reflect.Slice {
		fv = int64(0)
	} else {
		fv = 0
	}
	return fv
}

func (rfl *rflt) getFieldValue(obj any, name string) any {
	objV := rfl.objValue(obj)
	fld  := objV.FieldByName(name)
	if fld.Kind() == reflect.String {
		return fld.String()
	} else if fld.CanFloat() {
		return fld.Float()
	} else if fld.CanInt() {
		return fld.Int()
	} else if fld.CanUint() {
		return fld.Uint()
	} else if fld.Kind() == reflect.Bool {
		return fld.Bool()
	} else if fld.Kind() == reflect.Slice {
		return fld.Interface()
	} else {
		panic(fmt.Sprintf("getFieldValue: fld=(%s) field_kind=(%v)", name, fld.Kind()))
	}
}

func (rfl *rflt) setFieldValue(obj any, fld string, val any) {
	pgn_float64 := func(v pgtype.Numeric) float64 {
		if flt, err := v.Float64Value(); err == nil {
			if flt.Valid {
				return flt.Float64
			}
		}
		return 0.0
	}
	pgn_int64 := func(v pgtype.Numeric) int64 {
		if int, err := v.Int64Value(); err == nil {
			if int.Valid {
				return int.Int64
			}
		}
		return 0
	}
	pgn_time := func(v pgtype.Time) int64 {
		if v.Valid {
			if v.Microseconds < 0 {
				return 0
			}
			return v.Microseconds
		}
		return 0
	}
	pgn_date := func(v pgtype.Date) int64 {
		if v.Valid {
			if v.Time.UnixMicro() < 0 {
				return 0
			}
			return v.Time.UnixMicro()
		}
		return 0
	}
	pgn_ts := func(v pgtype.Timestamp) int64 {
		if v.Valid {
			if v.Time.UnixMicro() < 0 {
				return 0
			}
			return v.Time.UnixMicro()
		}
		return 0
	}
	pgn_tstz := func(v pgtype.Timestamptz) int64 {
		if v.Valid {
			if v.Time.UnixMicro() < 0 {
				return 0
			}
			return v.Time.UnixMicro()
		}
		return 0
	}
	if val == nil {
		return
	}
	flds := rfl.fields(obj)
	_fld := ""
	if slices.Contains(flds, fld) {			// First check for exact case match.
		_fld = fld
	} else {								// If no exact case match, look for case-insensitive.
		for _, f := range flds {
			if strings.EqualFold(f, fld) {
				_fld = f
				break
			}
		}
	}
	if _fld == "" {
		return	// Should probably return error or log.
	}
	objV := rfl.objValue(obj)
	fldV := objV.FieldByName(_fld)

	if fldV.Kind() == reflect.String {
		switch v := val.(type) {
		case string:
			fldV.SetString(v)
		case float32:
			fldV.SetString(fmt.Sprintf("%f", v))
		case float64:
			fldV.SetString(fmt.Sprintf("%f", v))
		case uint:
			fldV.SetString(fmt.Sprintf("%d", v))
		case uint16:
			fldV.SetString(fmt.Sprintf("%d", v))
		case uint32:
			fldV.SetString(fmt.Sprintf("%d", v))
		case uint64:
			fldV.SetString(fmt.Sprintf("%d", v))
		case int:
			fldV.SetString(fmt.Sprintf("%d", v))
		case int8:
			fldV.SetString(fmt.Sprintf("%d", v))
		case int16:
			fldV.SetString(fmt.Sprintf("%d", v))
		case int32:
			fldV.SetString(fmt.Sprintf("%d", v))
		case int64:
			fldV.SetString(fmt.Sprintf("%d", v))
		case bool:
			fldV.SetString(fmt.Sprintf("%v", v))
		case pgtype.Numeric:
			fldV.SetString(fmt.Sprintf("%f", pgn_float64(v)))
		case pgtype.Date:
			fldV.SetString(fmt.Sprintf("%d", pgn_date(v)))
		case pgtype.Time:
			fldV.SetString(time.UnixMicro(pgn_time(v)).Format("15:04:05.000000"))
		case time.Time:
			if v.UnixMicro() > 0 {
				fldV.SetString(v.Format("15:04:05.000000"))
			}
		case pgtype.Timestamp:
			fldV.SetString(time.UnixMicro(pgn_ts(v)).Format("2006-01-02 15:04:05.000000"))
		case pgtype.Timestamptz:
			fldV.SetString(time.UnixMicro(pgn_tstz(v)).Format("2006-01-02 15:04:05.000000"))
		default:
			panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(string) value_type=(%T)", fld, _fld, fldV.Kind(), val))
		}
		
	} else if fldV.CanFloat() {
		switch v := val.(type) {
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				fldV.SetFloat(f)
			}
		case float32:
			fldV.SetFloat(float64(v))
		case float64:
			fldV.SetFloat(v)
		case uint:
			fldV.SetFloat(float64(v))
		case uint16:
			fldV.SetFloat(float64(v))
		case uint32:
			fldV.SetFloat(float64(v))
		case uint64:
			fldV.SetFloat(float64(v))
		case int:
			fldV.SetFloat(float64(v))
		case int8:
			fldV.SetFloat(float64(v))
		case int16:
			fldV.SetFloat(float64(v))
		case int32:
			fldV.SetFloat(float64(v))
		case int64:
			fldV.SetFloat(float64(v))
		case pgtype.Numeric:
			fldV.SetFloat(pgn_float64(v))
		case bool:
		case pgtype.Date:
			fldV.SetFloat(float64(pgn_date(v)))
		case pgtype.Time:
			fldV.SetFloat(float64(pgn_time(v)))
		case time.Time:
			if v.UnixMicro() > 0 {
				fldV.SetFloat(float64(v.UnixMicro()))
			}
		case pgtype.Timestamp:
			fldV.SetFloat(float64(pgn_ts(v)))
		case pgtype.Timestamptz:
			fldV.SetFloat(float64(pgn_tstz(v)))
		default:
			panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(floatable) value_type=(%T)", fld, _fld, fldV.Kind(), val))
		}

	} else if fldV.CanUint() {
		switch v := val.(type) {
		case string:
			if u, err := strconv.ParseUint(v, 10, 64); err == nil {
				fldV.SetUint(u)
			}
		case float32:
			fldV.SetUint(uint64(v))
		case float64:
			fldV.SetUint(uint64(v))
		case uint:
			fldV.SetUint(uint64(v))
		case uint16:
			fldV.SetUint(uint64(v))
		case uint32:
			fldV.SetUint(uint64(v))
		case uint64:
			fldV.SetUint(uint64(v))
		case int:
			fldV.SetUint(uint64(v))
		case int8:
			fldV.SetUint(uint64(v))
		case int16:
			fldV.SetUint(uint64(v))
		case int32:
			fldV.SetUint(uint64(v))
		case int64:
			fldV.SetUint(uint64(v))
		case pgtype.Numeric:
			fldV.SetUint(uint64(pgn_int64(v)))
		case bool:
		case time.Time:
			fldV.SetUint(uint64(v.UnixMicro()))
		case pgtype.Date:
			fldV.SetUint(uint64(pgn_date(v)))
		case pgtype.Time:
			fldV.SetUint(uint64(pgn_time(v)))
		case pgtype.Timestamp:
			fldV.SetUint(uint64(pgn_ts(v)))
		case pgtype.Timestamptz:
			fldV.SetUint(uint64(pgn_tstz(v)))
		default:
			panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(uintable) value_type=(%T)", fld, _fld, fldV.Kind(), val))
		}

	} else if fldV.CanInt() {
		switch v := val.(type) {
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				fldV.SetInt(i)
			}
		case float32:
			fldV.SetInt(int64(v))
		case float64:
			fldV.SetInt(int64(v))
		case uint:
			fldV.SetInt(int64(v))
		case uint16:
			fldV.SetInt(int64(v))
		case uint32:
			fldV.SetInt(int64(v))
		case uint64:
			fldV.SetInt(int64(v))
		case int:
			fldV.SetInt(int64(v))
		case int8:
			fldV.SetInt(int64(v))
		case int16:
			fldV.SetInt(int64(v))
		case int32:
			fldV.SetInt(int64(v))
		case int64:
			fldV.SetInt(int64(v))
		case pgtype.Numeric:
			fldV.SetInt(pgn_int64(v))
		case bool:
		case time.Time:
			if v.UnixMicro() > 0 {
				fldV.SetInt(v.UnixMicro())
			}
		case pgtype.Date:
			fldV.SetInt(pgn_date(v))
		case pgtype.Time:
			fldV.SetInt(pgn_time(v))
		case pgtype.Timestamp:
			fldV.SetInt(pgn_ts(v))
		case pgtype.Timestamptz:
			fldV.SetInt(pgn_tstz(v))
		default:
			panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(intable) value_type=(%T)", fld, _fld, fldV.Kind(), val))
		}

	} else if fldV.Kind() == reflect.Bool {
		switch v := val.(type) {
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				fldV.SetBool(b)
			}
		case float32:
			fldV.SetBool(v != 0)
		case float64:
			fldV.SetBool(v != 0)
		case uint:
			fldV.SetBool(v != 0)
		case uint16:
			fldV.SetBool(v != 0)
		case uint32:
			fldV.SetBool(v != 0)
		case uint64:
			fldV.SetBool(v != 0)
		case int:
			fldV.SetBool(v != 0)
		case int8:
			fldV.SetBool(v != 0)
		case int16:
			fldV.SetBool(v != 0)
		case int32:
			fldV.SetBool(v != 0)
		case int64:
			fldV.SetBool(v != 0)
		case bool:
			fldV.SetBool(v)
		case pgtype.Numeric:
			fldV.SetBool(pgn_int64(v) != 0)
		case time.Time:
			fldV.SetBool(v.UnixMicro() > 0)
		case pgtype.Date:
			fldV.SetBool(pgn_date(v) > 0)
		case pgtype.Time:
			fldV.SetBool(pgn_time(v) > 0)
		case pgtype.Timestamp:
			fldV.SetBool(pgn_ts(v) > 0)
		case pgtype.Timestamptz:
			fldV.SetBool(pgn_tstz(v) > 0)
		default:
			panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(boolable) value_type=(%T)", fld, _fld, fldV.Kind(), val))
		}
	} else if fldV.Kind() == reflect.Slice {
		panic(fmt.Sprintf("setFieldValue: field_kind=(%v)(slice) value_type=(%T)", fldV.Kind(), val))
	} else {
		panic(fmt.Sprintf("setFieldValue: fld=(%s) _fld=(%s) field_kind=(%v)(????) value_type=(%T)", fld, _fld, fldV.Kind(), val))
	}
}
