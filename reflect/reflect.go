// This package provides some punk-rock reflection which is not in the stdlib.
package reflect

import (
	r "reflect"
)

// StructToMap converts a struct to map. The object's default key string
// is the struct field name but can be specified in the struct field's
// tag value. The "cql" key in the struct field's tag value is the key
// name. Examples:
//
//   // Field appears in the resulting map as key "myName".
//   Field int `cql:"myName"`
//
//   // Field appears in the resulting as key "Field"
//   Field int
//
//   // Field appears in the resulting map as key "myName"
//   Field int "myName"
func StructToMap(val interface{}) (map[string]interface{}, bool) {
	// indirect so function works with both structs and pointers to them
	structVal := r.Indirect(r.ValueOf(val))
	kind := structVal.Kind()
	if kind != r.Struct {
		return nil, false
	}
	structFields := cachedTypeFields(structVal.Type())
	mapVal := make(map[string]interface{}, len(structFields))
	for _, info := range structFields {
		field := fieldByIndex(structVal, info.index)
		mapVal[info.name] = field.Interface()
	}
	return mapVal, true
}

// MapsToStructs converts a slice of maps to a slice of structs using MapToStruct
func MapsToStructs(maps []map[string]interface{}, v interface{}) error {
	val := r.Indirect(r.ValueOf(v))

	val.Set(r.MakeSlice(val.Type(), val.Len(), val.Cap()))

	// Iterate through the slice/array and decode each element before adding it
	// to the dest slice/array
	for i, m := range maps {
		// Get element of array, growing if necessary.
		if i >= val.Cap() {
			newcap := val.Cap() + val.Cap()/2
			if newcap < 4 {
				newcap = 4
			}
			newval := r.MakeSlice(val.Type(), val.Len(), newcap)
			r.Copy(newval, val)
			val.Set(newval)
		}
		if i >= val.Len() {
			val.SetLen(i + 1)
		}

		if i < val.Len() {
			// Decode into element.
			err := MapToStruct(m, val.Index(i))
			if err != nil {
				return err
			}
		}

		i++
	}

	// Ensure that the destination is the correct size
	if len(maps) < val.Len() {
		val.SetLen(len(maps))
	}

	return nil
}

// MapToStruct converts a map to a struct. It is the inverse of the StructToMap
// function. For details see StructToMap.
func MapToStruct(m map[string]interface{}, struc interface{}) error {
	val := r.Indirect(r.ValueOf(struc))
	structFields := cachedTypeFields(val.Type())

	// Create fields map for faster lookup
	fieldsMap := make(map[string]field)
	for _, field := range structFields {
		fieldsMap[field.name] = field
	}

	for k, v := range m {
		if info, ok := fieldsMap[k]; ok {
			structField := fieldByIndex(val, info.index)
			if structField.Type().Name() == r.TypeOf(v).Name() {
				structField.Set(r.ValueOf(v))
			}
		}
	}
	return nil
}

// FieldsAndValues returns a list field names and a corresponing list of values
// for the given struct. For details on how the field names are determined please
// see StructToMap.
func FieldsAndValues(val interface{}) ([]string, []interface{}, bool) {
	// indirect so function works with both structs and pointers to them
	structVal := r.Indirect(r.ValueOf(val))
	kind := structVal.Kind()
	if kind != r.Struct {
		return nil, nil, false
	}
	structFields := cachedTypeFields(structVal.Type())
	fields := make([]string, len(structFields))
	values := make([]interface{}, len(structFields))
	for i, info := range structFields {
		field := fieldByIndex(structVal, info.index)
		fields[i] = info.name
		values[i] = field.Interface()
	}
	return fields, values, true
}

func fieldByIndex(v r.Value, index []int) r.Value {
	for _, i := range index {
		if v.Kind() == r.Ptr {
			if v.IsNil() {
				if v.CanSet() {
					v.Set(r.New(v.Type().Elem()))
				} else {
					return r.Value{}
				}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}

	return v
}
