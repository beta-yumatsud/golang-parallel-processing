package main

import (
	"errors"
	"fmt"
	"reflect"
)

func main() {
	sample1()
}

func sample1() {
	type Point struct {
		X int
		Y int
	}
	p := Point{X: 10, Y: 15}

	rv := reflect.ValueOf(p)
	fmt.Printf("rv.Type = %v\n", rv.Type())
	fmt.Printf("rv.Kind = %v\n", rv.Kind())
	fmt.Printf("rv.Interface = %v\n", rv.Interface())

	xv := rv.Field(0)
	fmt.Printf("xv = %d\n", xv.Int())
	if rv.Field(0).CanSet() {
		xv.SetInt(100)
		fmt.Printf("xv(after SetInt) = %d\n", xv.Int())
	}

	if rv.Kind() == reflect.Int {
		// something to do
	}
}

func marshal(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		// For Map
		keyType := rv.Type().Key()
		if keyType.Kind() == reflect.String {
			// key is string
		} else {
			return nil, errors.New("expected map with string key")
		}
	case reflect.Struct:
		// For Struct
	default:
		return nil, errors.New("unsupported type(" + rv.Type().String() + ")")
	}
	return nil, errors.New("do not move")
}