package generators

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRange(params ...any) (min, max int64, err error) {
	if len(params) != 1 {
		return -1, -1, fmt.Errorf("invalid arguments: %v", params)
	}
	var expr string
	switch t := params[0].(type) {
	case string:
		expr = params[0].(string)
	case *string:
		expr = *params[0].(*string)
	default:
		return -1, -1, fmt.Errorf("invalid argument 0, expected string but got %T", t)
	}
	parts := strings.Split(expr, "..")
	if len(parts) != 2 {
		return -1, -1, fmt.Errorf("invalid argument 0, expected 'min..max' but got '%s'", params[0])
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
