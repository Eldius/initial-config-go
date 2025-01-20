package setup

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strings"
)

type redactHandler struct {
	slog.Handler
	h            slog.Handler
	keysToRedact []string
}

func newRedactHandler(h slog.Handler, keysToRedact []string) slog.Handler {
	return &redactHandler{
		h:            h,
		keysToRedact: keysToRedact,
	}
}

func (r *redactHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return r.h.Enabled(ctx, level)
}

func (r *redactHandler) Handle(ctx context.Context, record slog.Record) error {
	return r.h.Handle(ctx, record)
}

func (r *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(r.keysToRedact) == 0 {
		return &redactHandler{h: r.h.WithAttrs(attrs)}
	}

	var mapValueIDs []int
	//var structValueIDs []int
	for i, attr := range attrs {
		for _, key := range r.keysToRedact {
			if strings.Contains(strings.ToLower(attr.Key), key) {
				attrs[i].Value = slog.StringValue("***")
				continue
			}

			if attr.Value.Kind() == slog.KindAny {
				av := attr.Value.Any()
				attrs[i].Value = slog.AnyValue(parseValue(av, r.keysToRedact))
			}
		}
	}

	for _, i := range mapValueIDs {
		if m, ok := attrs[i].Value.Any().(map[string]any); ok {
			attrs[i].Value = slog.AnyValue(parseMap(m, r.keysToRedact))
		}
	}

	return &redactHandler{h: r.h.WithAttrs(attrs), keysToRedact: r.keysToRedact}
}

func (r *redactHandler) WithGroup(name string) slog.Handler {
	return &redactHandler{h: r.h.WithGroup(name), keysToRedact: r.keysToRedact}
}

func parseMap(v map[string]any, redactedKeyList []string) map[string]any {
	v = maps.Clone(v)
	fmt.Printf("parsingMap: %#v\n", v)
	for k := range maps.Keys(v) {
		switch valueType := reflect.TypeOf(v[k]); valueType.Kind() {
		case reflect.Struct:
			v[k] = parseValue(v[k], redactedKeyList)
		case reflect.Map:
			v[k] = parseValue(v[k], redactedKeyList)
		}
		fmt.Printf("key: %s/%s, value: %v (%s) is equal: %v\n", k, strings.ToLower(k), v[k], redactedKeyList, slices.Contains(redactedKeyList, strings.ToLower(k)))
		for _, rk := range redactedKeyList {
			if strings.Contains(strings.ToLower(k), strings.ToLower(rk)) {
				v[k] = "***"
			}
		}
	}
	return v
}

const (
	stringRedactedValue = "***"
)

func parseValue[T any](v T, redactedKeyList []string) any {

	vType := reflect.TypeOf(v)
	tmpV := reflect.ValueOf(v)

	switch vType.Kind() {
	case reflect.Map:
		fmt.Printf("parsing map %#v\n", v)
		if m, ok := tmpV.Interface().(map[string]any); ok {
			return parseMap(m, redactedKeyList)
		}
	case reflect.Struct:
		fmt.Printf("parsing struct %#v\n", v)
		return parseStruct(tmpV.Interface(), redactedKeyList)
	default:
		fmt.Printf("parsing value %#v\n", v)
		return v
	}

	return v
}

func zeroValue(val reflect.Value) reflect.Value {
	v := val.Interface()
	vType := val.Type()
	newVal := reflect.New(vType)
	fmt.Printf("zeroValue init[%T/%T]: %#v\n", v, val, v)
	switch vType.Kind() {
	case reflect.String:
		newVal = reflect.New(vType)
		newVal.Elem().SetString(stringRedactedValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		newVal.Elem().SetInt(-1)
	default:
		newVal.Elem().SetZero()
	}
	return newVal
}

func parseStruct[T any](v T, redactedKeyList []string) T {
	vType := reflect.TypeOf(v)
	vVal := reflect.ValueOf(v)
	nVal := reflect.New(vType)
	for _, key := range redactedKeyList {
		for i := range vType.NumField() {
			fieldValue := vVal.Field(i)
			field := vType.Field(i)
			if strings.ToLower(field.Name) == key || strings.ToLower(field.Tag.Get("json")) == key {
				fmt.Printf("field name: %#v\n", field.Name)
				newValue := zeroValue(fieldValue)
				fmt.Printf("assignable 0: %+v\n\n", newValue.Type().AssignableTo(field.Type))
				if !newValue.Type().AssignableTo(field.Type) && newValue.Kind() == reflect.Pointer && fieldValue.Kind() != newValue.Kind() {
					fmt.Printf("new field value 0: %#v => %#v [%T/%T]\n", fieldValue, newValue, fieldValue.Interface(), newValue.Interface())
					newValue = newValue.Elem()
					fmt.Printf("new field value 1: %#v => %#v [%T/%T]\n", fieldValue, newValue, fieldValue.Interface(), newValue.Interface())
				}
				fmt.Printf("new field value: %#v [%T/%T] => %v/%v\n", fieldValue.Interface(), newValue.Interface(), newValue, newValue.Kind(), newValue.Kind())
				fmt.Printf("attr type: %T => %T => %T [kind: %s => %s => %s]\n", vVal.Field(i).Interface(), newValue.Interface(), fieldValue.Interface(), vVal.Field(i).Kind(), newValue.Kind(), fieldValue.Kind())
				fmt.Printf("assignable 1: %+v (converted: %v)\n\n", newValue.Type().AssignableTo(field.Type), newValue.Convert(field.Type))
				nVal.Elem().Field(i).Set(newValue.Convert(field.Type))
				fmt.Printf("zeroVal[%T]: %#v(%#v) => %#v\n", fieldValue.Interface(), newValue.Interface(), field.Name, key)
				continue
			}
			nVal.Elem().Field(i).Set(fieldValue)
			//vVal.Field(i).Set(fieldValue)
		}
	}
	fmt.Printf("parsedStruct: %#v\n", v)
	return nVal.Elem().Interface().(T)
}
