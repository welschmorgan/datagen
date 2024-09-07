package generators

import (
	"fmt"
	"strconv"
	"strings"
)

func ParsePattern(params ...any) (pattern string, err error) {
	if len(params) != 2 {
		return "", fmt.Errorf("invalid arguments, expected ['generator_name', 'pattern'] but got %v", params)
	}
	return params[1].(string), nil
}

func ParseUnion(params ...any) (pattern []string, err error) {
	if len(params) != 2 {
		return nil, fmt.Errorf("invalid arguments, expected ['generator_name', 'pattern'] but got %v", params)
	}
	return strings.Split(params[1].(string), "|"), nil
}

func ParseRange(params ...any) (min, max int64, err error) {
	if len(params) != 2 {
		return -1, -1, fmt.Errorf("invalid arguments, expected ['generator_name', 'min..max'] but got %v", params)
	}
	var expr string
	switch t := params[1].(type) {
	case string:
		expr = params[1].(string)
	case *string:
		expr = *params[1].(*string)
	default:
		return -1, -1, fmt.Errorf("invalid argument 1, expected string but got %T", t)
	}
	parts := strings.Split(expr, "..")
	if len(parts) != 2 {
		return -1, -1, fmt.Errorf("invalid argument 1, expected 'min..max' but got '%s'", expr)
	}
	min, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	max, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return min, max, nil
}

func ParseStrings(minRequired int, params ...any) (args []string, err error) {
	if len(params) != minRequired {
		return nil, fmt.Errorf("invalid arguments: %v", params)
	}
	ret := []string{}
	for i := range len(params) {
		var expr string
		switch t := params[i].(type) {
		case string:
			expr = params[i].(string)
		case *string:
			expr = *params[i].(*string)
		default:
			return nil, fmt.Errorf("invalid argument #%d, expected string but got %T", i, t)
		}
		ret = append(ret, expr)
	}
	return ret, nil
}
