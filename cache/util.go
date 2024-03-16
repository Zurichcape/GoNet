package cache

import (
	"errors"
)

func Incr(val interface{}) (interface{}, error) {
	switch val := val.(type) {
	case int:
		val = val + 1
		return val, nil
	case int32:
		val = val + 1
		return val, nil
	case int64:
		val = val + 1
		return val, nil
	case uint:
		val = val + 1
		return val, nil
	case uint32:
		val = val + 1
		return val, nil
	case uint64:
		val = val + 1
		return val, nil
	default:
		return val, errors.New("item value is not int-type")
	}
}

func Decr(val interface{}) (interface{}, error) {
	switch val := val.(type) {
	case int:
		val = val - 1
		return val, nil
	case int32:
		val = val - 1
		return val, nil
	case int64:
		val = val - 1
		return val, nil
	case uint:
		if val > 0 {
			val = val - 1
			return val, nil
		} else {
			return val, errors.New("item value is less than 0")
		}
	case uint32:
		if val > 0 {
			val = val - 1
			return val, nil
		} else {
			return val, errors.New("item value is less than 0")
		}
	case uint64:
		if val > 0 {
			val = val - 1
			return val, nil
		} else {
			return val, errors.New("item value is less than 0")
		}
	default:
		return val, errors.New("item value is not int-type")
	}
}

