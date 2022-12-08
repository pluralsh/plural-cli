package utils

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/mitchellh/mapstructure"
	lua "github.com/yuin/gopher-lua"
)

// Mapper maps a lua table to a Go struct pointer.
type Mapper struct {
}

// MapLua maps the lua table to the given struct pointer with default options.
func MapLua(tbl *lua.LTable, st interface{}) error {
	return NewMapper().Map(tbl, st)
}

// NewMapper returns a new mapper.
func NewMapper() *Mapper {

	return &Mapper{}
}

// Map maps the lua table to the given struct pointer.
func (mapper *Mapper) Map(tbl *lua.LTable, st interface{}) error {
	mp, ok := ToGoValue(tbl).(map[interface{}]interface{})
	if !ok {
		return errors.New("arguments #1 must be a table, but got an array")
	}
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           st,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(mp)
}

// ToGoValue converts the given LValue to a Go object.
func ToGoValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return trimQuotes(string(v))
	case lua.LNumber:
		return float64(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[interface{}]interface{})
			v.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(ToGoValue(key))
				ret[keystr] = ToGoValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, ToGoValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v
	}
}

func trimQuotes(s string) interface{} {
	return regexp.MustCompile(`^"(.*)"$`).ReplaceAllString(s, "$1")
}
