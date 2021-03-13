package maptostruct

import (
	"errors"
	"fmt"
	"reflect"
)

// Do converts a map to a structure
func Do(m map[string]interface{}, s interface{}) error {
	tagInfo := make(map[string]string)
	rt := reflect.TypeOf(s).Elem()
	for i := 0; i < rt.NumField(); i++ {
		if mts := rt.Field(i).Tag.Get("mts"); mts != "" {
			tagInfo[mts] = rt.Field(i).Name
		}
	}

	for k, v := range m {
		var err error
		// If the struct has the key, set value
		if name, ok := tagInfo[k]; ok {
			err = setField(s, name, v)
		} else {
			err = setField(s, k, v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func setField(obj interface{}, name string, value interface{}) error {
	rv := reflect.ValueOf(obj).Elem().FieldByName(name)

	if !rv.IsValid() {
		// No such field => SKIP
		return nil
	}

	if !rv.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)
	if val.Type() != rv.Type() {
		return errors.New("Provided value type didn't match obj field type")
	}
	rv.Set(val)

	return nil
}
