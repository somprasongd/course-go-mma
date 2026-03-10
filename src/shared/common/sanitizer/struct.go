package sanitizer

import (
	"reflect"
	"strings"
)

// SanitizeStruct applies sanitization to all fields with sanitize tags
// It supports recursive sanitization for nested structs, pointers, slices, and maps
func SanitizeStruct(s interface{}) {
	if s == nil {
		return
	}

	val := reflect.ValueOf(s)
	sanitizeValue(val, "")
}

// sanitizeValue recursively sanitizes a reflect.Value
// tag is passed from parent for cases like slice of strings
func sanitizeValue(val reflect.Value, inheritedTag string) {
	if !val.IsValid() {
		return
	}

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		sanitizeStruct(val)
	case reflect.Slice, reflect.Array:
		sanitizeSlice(val, inheritedTag)
	case reflect.Map:
		sanitizeMap(val)
	case reflect.String:
		// String with inherited tag (from slice)
		if inheritedTag != "" && val.CanSet() {
			applySanitization(val, inheritedTag)
		}
	}
}

// sanitizeStruct sanitizes a struct value
func sanitizeStruct(val reflect.Value) {
	if val.Kind() != reflect.Struct {
		return
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the sanitize tag
		tag := fieldType.Tag.Get("sanitize")

		switch field.Kind() {
		case reflect.String:
			if tag != "" {
				applySanitization(field, tag)
			}
		case reflect.Struct:
			// Recursively sanitize nested struct
			sanitizeStruct(field)
		case reflect.Ptr:
			if !field.IsNil() {
				// Check if it's a pointer to struct
				if field.Elem().Kind() == reflect.Struct {
					sanitizeValue(field, "")
				} else if field.Elem().Kind() == reflect.String && tag != "" {
					// Pointer to string with tag
					strVal := field.Elem().String()
					sanitized := sanitizeString(strVal, tag)
					field.Set(reflect.ValueOf(&sanitized))
				}
			}
		case reflect.Slice, reflect.Array:
			sanitizeSlice(field, tag)
		case reflect.Map:
			sanitizeMap(field)
		}
	}
}

// sanitizeSlice sanitizes slice/array elements
func sanitizeSlice(val reflect.Value, tag string) {
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return
	}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		switch elem.Kind() {
		case reflect.String:
			if tag != "" && elem.CanSet() {
				applySanitization(elem, tag)
			}
		case reflect.Struct:
			sanitizeStruct(elem)
		default:
			sanitizeValue(elem, "")
		}
	}
}

// sanitizeMap sanitizes map keys and values
func sanitizeMap(val reflect.Value) {
	if val.Kind() != reflect.Map || val.IsNil() {
		return
	}

	// For maps, we need to create a new map if we're modifying keys or values
	// Note: Map keys cannot be modified directly, so we only sanitize string values
	for _, key := range val.MapKeys() {
		elem := val.MapIndex(key)
		if !elem.IsValid() {
			continue
		}

		// Handle different value types
		switch elem.Kind() {
		case reflect.String:
			// Apply XSS sanitization to map string values by default
			sanitized := SanitizeInput(elem.String())
			if sanitized != elem.String() {
				val.SetMapIndex(key, reflect.ValueOf(sanitized))
			}
		case reflect.Interface:
			// Handle interface{} containing string
			if elem.Elem().Kind() == reflect.String {
				sanitized := SanitizeInput(elem.Elem().String())
				val.SetMapIndex(key, reflect.ValueOf(sanitized))
			} else if elem.Elem().Kind() == reflect.Struct ||
				(elem.Elem().Kind() == reflect.Ptr && elem.Elem().Elem().Kind() == reflect.Struct) {
				// Struct or pointer to struct in interface{}
				sanitizeValue(elem.Elem(), "")
			}
		case reflect.Struct:
			sanitizeStruct(elem)
		case reflect.Ptr:
			if !elem.IsNil() && elem.Elem().Kind() == reflect.Struct {
				sanitizeStruct(elem.Elem())
			}
		}
	}
}

// applySanitization applies sanitization based on tag
func applySanitization(field reflect.Value, tag string) {
	if field.Kind() != reflect.String || !field.CanSet() {
		return
	}

	result := sanitizeString(field.String(), tag)
	field.SetString(result)
}

// sanitizeString sanitizes a string based on tag
func sanitizeString(input, tag string) string {
	switch tag {
	case "email":
		return SanitizeEmail(input)
	case "xss":
		return SanitizeInput(input)
	case "trim":
		return strings.TrimSpace(input)
	case "lower":
		return strings.ToLower(input)
	default:
		return input
	}
}
