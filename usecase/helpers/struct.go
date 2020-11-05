package helpers

import (
	"reflect"
	"strings"
)

func JsonTagExists(key, outerKey string, r reflect.Type) bool {
	return FieldTagExists(key, outerKey, "json", r)
}

func FieldTagExists(key, outerKey, tagName string, r reflect.Type) bool {
	if r.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		v := outerKey + strings.Split(f.Tag.Get(tagName), ",")[0] // use split to ignore tagName "options" like omitempty, etc.
		if v == key {
			return true
		}
		tags := strings.Split(key, ".")
		if f.Type.Kind() == reflect.Struct && len(tags) > 1 && v == tags[0]{
			return FieldTagExists(key, v+".", tagName, f.Type)
		}
	}
	return false
}

func BuildTagList(in interface{}, blackList []string) []string {
	t := reflect.TypeOf(in)
	var blackListMap = make(map[string]interface{})
	for _, b := range blackList {
		blackListMap[b] = nil
	}
	var list []string
	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get("json")
		_, blocked := blackListMap[tag]
		if blocked || tag == "" || tag == "-" {
			continue
		}

		//typ := field.Type.Kind()

		//if typ == reflect.Struct || typ == reflect.Array || typ == reflect.Ptr || typ == reflect.UnsafePointer || typ == reflect.Slice || typ == reflect.Map {
		//	continue
		//}
		list = append(list, tag)
	}
	return list
}