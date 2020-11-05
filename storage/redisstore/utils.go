package redisstore

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"strings"
	"sync"
)

const (
	TagRedis        = "json"
	TIME_HALF_MONTH = 3600 * 24 * 15
)

type Args []interface{}

type fieldSpec struct {
	name      string
	index     []int
	omitEmpty bool
}

type structSpec struct {
	m map[string]*fieldSpec
	l []*fieldSpec
}

var (
	structSpecMutex sync.RWMutex
	structSpecCache = make(map[reflect.Type]*structSpec)
)

func (r *redisStore) HMSET(key string, in interface{}) (err error) {
	_, err = r.GetConnection().Do(append([]interface{}{"HMSET"}, Args{key}.AddFlat(in).Json()...)).Result()
	if err != nil {
		return err
	}
	return r.SetExpire(key, TIME_HALF_MONTH)
}

func (r *redisStore) HMGET(key string, fields []interface{}, out interface{}) (err error) {
	v, err := redis.Values(r.GetConnection().Do(append([]interface{}{"HMGET", key}, fields...)...).Result())
	//for _, value := range v {
	//	if value == nil {
	//		return errors.New("exist fields is empty")
	//	}
	//}

	return scanStringField(out, fields, v)
}

func (r *redisStore) SetExpire(key string, time int64) error {
	_, err := r.GetConnection().Do("EXPIRE", key, time).Result()
	return err
}

func scanStringField(out interface{}, fields []interface{}, values []interface{}) (err error) {
	m := make(map[string]bool)
	s := reflect.ValueOf(out).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		k, _ := typeOfT.Field(i).Tag.Lookup("redis")
		if f.Type().String() == "string" {
			m[k] = true
		}
	}

	str := "{ "
	countValidValues := 0
	for key, value := range values {
		if value != nil && fields[key].(string) != "" && string(value.([]byte)) != "" {
			if val, ok := m[fields[key].(string)]; val && ok {
				str += fmt.Sprintf("\"%s\": \"%s\" , ", fields[key].(string), string(value.([]byte)))
			} else {
				str += fmt.Sprintf("\"%s\": %s , ", fields[key].(string), string(value.([]byte)))
			}
			countValidValues++
		}
	}
	if countValidValues > 0 {
		str = str[:len(str)-2] + " }"
	} else {
		str = ""
	}
	if str != "" {
		err = json.Unmarshal([]byte(str), out)
		return err
	}
	return nil
}

func (args Args) Json() Args {
	var list []interface{}
	for _, value := range args {
		js, _ := json.Marshal(value)
		str := string(js)
		if _, ok := value.(string); ok {
			str = str[1 : len(str)-1]
		}
		list = append(list, str)
	}
	return list
}

func (args Args) AddFlat(v interface{}) Args {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Struct:
		args = flattenStruct(args, rv)
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			args = append(args, rv.Index(i).Interface())
		}
	case reflect.Map:
		for _, k := range rv.MapKeys() {
			args = append(args, k.Interface(), rv.MapIndex(k).Interface())
		}
	case reflect.Ptr:
		if rv.Type().Elem().Kind() == reflect.Struct {
			if !rv.IsNil() {
				args = flattenStruct(args, rv.Elem())
			}
		} else {
			args = append(args, v)
		}
	default:
		args = append(args, v)
	}
	return args
}

func flattenStruct(args Args, v reflect.Value) Args {
	ss := structSpecForType(v.Type())
	for _, fs := range ss.l {
		fv := v.FieldByIndex(fs.index)
		if fs.omitEmpty {
			var empty = false
			switch fv.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				empty = fv.Len() == 0
			case reflect.Bool:
				empty = !fv.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				empty = fv.Int() == 0
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				empty = fv.Uint() == 0
			case reflect.Float32, reflect.Float64:
				empty = fv.Float() == 0
			case reflect.Interface, reflect.Ptr:
				empty = fv.IsNil()
			}
			if empty {
				continue
			}
		}

		args = append(args, fs.name, fv.Interface())
	}
	return args
}

func structSpecForType(t reflect.Type) *structSpec {
	structSpecMutex.RLock()
	ss, found := structSpecCache[t]
	structSpecMutex.RUnlock()
	if found {
		return ss
	}

	structSpecMutex.Lock()
	defer structSpecMutex.Unlock()
	ss, found = structSpecCache[t]
	if found {
		return ss
	}

	ss = &structSpec{m: make(map[string]*fieldSpec)}
	compileStructSpec(t, make(map[string]int), nil, ss)
	structSpecCache[t] = ss
	return ss
}

func compileStructSpec(t reflect.Type, depth map[string]int, index []int, ss *structSpec) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch {
		case f.PkgPath != "" && !f.Anonymous:
			// Ignore unexported fields.
		case f.Anonymous:
			// TODO: Handle pointers. Requires change to decoder and
			// protection against infinite recursion.
			if f.Type.Kind() == reflect.Struct {
				compileStructSpec(f.Type, depth, append(index, i), ss)
			}
		default:
			fs := &fieldSpec{name: f.Name}
			tag := f.Tag.Get(TagRedis)
			p := strings.Split(tag, ",")
			if len(p) > 0 {
				if p[0] == "-" {
					continue
				}
				if len(p[0]) > 0 {
					fs.name = p[0]
				}
				for _, s := range p[1:] {
					switch s {
					case "omitempty":
						fs.omitEmpty = true
					default:
						panic(fmt.Errorf("redigo: unknown field tag %s for type %s", s, t.Name()))
					}
				}
			}
			d, found := depth[fs.name]
			if !found {
				d = 1 << 30
			}
			switch {
			case len(index) == d:
				// At same depth, remove from result.
				delete(ss.m, fs.name)
				j := 0
				for i := 0; i < len(ss.l); i++ {
					if fs.name != ss.l[i].name {
						ss.l[j] = ss.l[i]
						j += 1
					}
				}
				ss.l = ss.l[:j]
			case len(index) < d:
				fs.index = make([]int, len(index)+1)
				copy(fs.index, index)
				fs.index[len(index)] = i
				depth[fs.name] = len(index)
				ss.m[fs.name] = fs
				ss.l = append(ss.l, fs)
			}
		}
	}
}

func ConvertRedisValue(in interface{}) string {
	js, _ := json.Marshal(in)
	return strings.Trim(string(js), "\"")
}
