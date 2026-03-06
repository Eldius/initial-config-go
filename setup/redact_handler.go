package setup

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
)

type redactHandler struct {
	h            slog.Handler
	keysToRedact []string
}

// newRedactHandler creates a new slog.Handler that redacts sensitive keys.
// It performs a case-insensitive partial match on keys.
func newRedactHandler(h slog.Handler, keysToRedact []string) slog.Handler {
	loweredKeys := make([]string, len(keysToRedact))
	for i, k := range keysToRedact {
		loweredKeys[i] = strings.ToLower(k)
	}
	return &redactHandler{
		h:            h,
		keysToRedact: loweredKeys,
	}
}

func (r *redactHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return r.h.Enabled(ctx, level)
}

func (r *redactHandler) Handle(ctx context.Context, record slog.Record) error {
	if len(r.keysToRedact) == 0 {
		return r.h.Handle(ctx, record)
	}

	// Create a new record to avoid modifying the original one's attributes
	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		newRecord.AddAttrs(r.redactAttr(attr))
		return true
	})
	return r.h.Handle(ctx, newRecord)
}

func (r *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(r.keysToRedact) == 0 {
		return &redactHandler{h: r.h.WithAttrs(attrs), keysToRedact: r.keysToRedact}
	}

	newAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		newAttrs[i] = r.redactAttr(attr)
	}

	return &redactHandler{h: r.h.WithAttrs(newAttrs), keysToRedact: r.keysToRedact}
}

func (r *redactHandler) WithGroup(name string) slog.Handler {
	return &redactHandler{h: r.h.WithGroup(name), keysToRedact: r.keysToRedact}
}

func (r *redactHandler) shouldRedact(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, rk := range r.keysToRedact {
		if strings.Contains(lowerKey, rk) {
			return true
		}
	}
	return false
}

func (r *redactHandler) redactAttr(attr slog.Attr) slog.Attr {
	if r.shouldRedact(attr.Key) {
		return slog.String(attr.Key, "***")
	}

	switch attr.Value.Kind() {
	case slog.KindGroup:
		groupAttrs := attr.Value.Group()
		newGroupAttrs := make([]slog.Attr, len(groupAttrs))
		for i, a := range groupAttrs {
			newGroupAttrs[i] = r.redactAttr(a)
		}
		// Convert []slog.Attr to []any for slog.Group
		args := make([]any, len(newGroupAttrs))
		for i, v := range newGroupAttrs {
			args[i] = v
		}
		return slog.Group(attr.Key, args...)
	case slog.KindAny:
		val := attr.Value.Any()
		if val != nil {
			return slog.Any(attr.Key, r.redactValue(val))
		}
	}

	return attr
}

func (r *redactHandler) redactValue(v any) any {
	if v == nil {
		return nil
	}

	vVal := reflect.ValueOf(v)
	vType := vVal.Type()

	switch vType.Kind() {
	case reflect.Map:
		if vType.Key().Kind() == reflect.String {
			return r.redactMap(vVal)
		}
	case reflect.Struct:
		return r.redactStruct(vVal)
	case reflect.Ptr:
		if vVal.IsNil() {
			return v
		}
		// Handle pointer by redacting the dereferenced value
		return r.redactValue(vVal.Elem().Interface())
	case reflect.Slice, reflect.Array:
		return r.redactSlice(vVal)
	}

	return v
}

func (r *redactHandler) redactMap(v reflect.Value) any {
	newMap := reflect.MakeMap(v.Type())
	for _, key := range v.MapKeys() {
		kStr := key.String()
		val := v.MapIndex(key)
		if r.shouldRedact(kStr) {
			newMap.SetMapIndex(key, reflect.ValueOf("***"))
		} else {
			redactedVal := r.redactValue(val.Interface())
			newMap.SetMapIndex(key, reflect.ValueOf(redactedVal))
		}
	}
	return newMap.Interface()
}

func (r *redactHandler) redactStruct(v reflect.Value) any {
	vType := v.Type()
	newStruct := reflect.New(vType).Elem()

	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := v.Field(i)

		redacted := r.shouldRedact(field.Name)
		if !redacted {
			if tag := field.Tag.Get("json"); tag != "" {
				tagPart := strings.Split(tag, ",")[0]
				if tagPart != "" && tagPart != "-" {
					redacted = r.shouldRedact(tagPart)
				}
			}
		}

		if redacted {
			newStruct.Field(i).Set(r.zeroValue(fieldVal))
		} else {
			redactedVal := r.redactValue(fieldVal.Interface())
			// Ensure the redacted value is assignable to the field
			rv := reflect.ValueOf(redactedVal)
			if rv.Type().AssignableTo(field.Type) {
				newStruct.Field(i).Set(rv)
			} else {
				// Fallback to original value if types don't match after redaction logic
				newStruct.Field(i).Set(fieldVal)
			}
		}
	}
	return newStruct.Interface()
}

func (r *redactHandler) redactSlice(v reflect.Value) any {
	newSlice := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
	for i := 0; i < v.Len(); i++ {
		redactedVal := r.redactValue(v.Index(i).Interface())
		rv := reflect.ValueOf(redactedVal)
		if rv.Type().AssignableTo(v.Type().Elem()) {
			newSlice.Index(i).Set(rv)
		} else {
			newSlice.Index(i).Set(v.Index(i))
		}
	}
	return newSlice.Interface()
}

func (r *redactHandler) zeroValue(val reflect.Value) reflect.Value {
	vType := val.Type()
	switch vType.Kind() {
	case reflect.String:
		return reflect.ValueOf("***")
	default:
		return reflect.Zero(vType)
	}
}
