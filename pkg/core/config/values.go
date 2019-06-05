package config

import (
	"fmt"
	"strconv"
)

type paramValue interface {
	setValue(newVal interface{}) error
}

// StringVal represents a string param value
type StringVal struct {
	val *string
}

// Value returns underlying value of a given param
func (val StringVal) Value() string {
	return *val.val
}

func (val StringVal) setValue(newVal interface{}) error {
	strVal, ok := newVal.(string)
	if !ok {
		return fmt.Errorf("Expected string value but got: %v(%[1]T)", newVal)
	}
	*val.val = strVal
	return nil
}

// IntVal represents an int param value
type IntVal struct {
	val *int
}

// Value returns underlying value of a given param
func (val IntVal) Value() int {
	return *val.val
}

func (val IntVal) setValue(newVal interface{}) error {
	var valPtr *int
	switch newVal.(type) {
	case int:
		intVal := newVal.(int)
		valPtr = &intVal
	case float32:
		intVal := int(newVal.(float32))
		valPtr = &intVal
	case float64:
		intVal := int(newVal.(float64))
		valPtr = &intVal
	case string:
		strVal := newVal.(string)
		if intVal, err := strconv.Atoi(strVal); err == nil {
			valPtr = &intVal
		}
	}
	if valPtr != nil {
		*val.val = *valPtr
		return nil
	}
	return fmt.Errorf("Expected int value but got: %v(%[1]T)", newVal)
}

// BoolVal represents an int param value
type BoolVal struct {
	val *bool
}

// Value returns underlying value of a given param
func (val BoolVal) Value() bool {
	return *val.val
}

func (val BoolVal) setValue(newVal interface{}) error {
	switch newVal.(type) {
	case bool:
		boolVal := newVal.(bool)
		*val.val = boolVal
		return nil
	case string:
		strVal := newVal.(string)
		if boolVal, err := strconv.ParseBool(strVal); err == nil {
			*val.val = boolVal
			return nil
		}
	}
	return fmt.Errorf("Expected bool value but got: %v(%[1]T)", newVal)
}
