package models

import (
	"database/sql"
	"fmt"
)

type Prop struct {
	Id    int64
	Type  string
	Value string
}

func NewProp(id int64, typ, value string) *Prop {
	return &Prop{
		Id:    id,
		Type:  typ,
		Value: value,
	}
}

func LoadProps(db *sql.DB, table string, typ *string, value *string) ([]*Prop, error) {
	rawQuery := fmt.Sprintf("SELECT * FROM %s", table)
	var res *sql.Rows
	var err error
	params := []interface{}{}
	props := []*Prop{}
	if typ != nil {
		rawQuery = fmt.Sprintf("%s WHERE type = ?", rawQuery)
		params = append(params, *typ)
	}
	if value != nil {
		op := ""
		switch len(params) {
		case 0:
			op = "WHERE"
		default:
			op = "AND"
		}
		rawQuery = fmt.Sprintf("%s %s value = ?", rawQuery, op)
		params = append(params, *value)
	}
	res, err = db.Query(rawQuery, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to list table '%s', %s", table, err)
	}
	rowId := 1
	for res.Next() {
		var id int64 = 0
		typ := ""
		value := ""
		if err := res.Scan(&id, &typ, &value); err != nil {
			return nil, fmt.Errorf("failed to read row #%d of %s, %s", rowId, table, err)
		}
		props = append(props, NewProp(id, typ, value))
		rowId += 1
	}
	return props, nil
}
