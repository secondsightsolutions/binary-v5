package main

import (
	"bytes"
	"fmt"
	"reflect"
)

type rflt struct {}

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