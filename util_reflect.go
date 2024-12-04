package main

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type rflt struct {}

func new_slice_ptr[T any](len, cap int) []*T {
	o  := new(T)
	ot := reflect.TypeOf(o)
	st := reflect.SliceOf(ot)
	return reflect.MakeSlice(st, len, cap).Interface().([]*T)
}
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

func (rfl *rflt) getString(obj any, fcn string) string {
	vs := reflect.ValueOf(&obj).MethodByName(fcn).Call([]reflect.Value{})
	v0 := vs[0]
	sv := ""
	if v0.Kind() == reflect.String {
		sv = v0.String()
	} else if v0.CanFloat() {
		sv = fmt.Sprintf("%f", v0.Float())
	} else if v0.CanInt() {
		sv = fmt.Sprintf("%d", v0.Int())
	} else if v0.CanUint() {
		sv = fmt.Sprintf("%d", v0.Uint())
	} else if v0.Kind() == reflect.Bool {
		sv = fmt.Sprintf("%v", v0.Bool())
	}
	return sv
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
		sb.WriteString("{")
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
		sb.WriteString("}")
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
	}
	return fv
}

func (rfl *rflt) setFieldValue(obj any, fld string, val any) {
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
		default:
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
		case bool:
		default:
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
		case bool:
		default:
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
		case bool:
		default:
		}

	} else if fldV.Bool() {
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
		default:
		}
	}
}
