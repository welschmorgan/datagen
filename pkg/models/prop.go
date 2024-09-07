package models

import (
	"database/sql"
	"fmt"
)

type Prop struct {
	Id       int64
	LocaleId int64
	Type     string
	Value    string
}

func NewProp(id int64, locale_id int64, typ, value string) *Prop {
	return &Prop{
		Id:       id,
		LocaleId: locale_id,
		Type:     typ,
		Value:    value,
	}
}

func LoadPropsAsMap(db *sql.DB, table string, typ *string, value *string, key func(*Prop) string) (map[string]*Prop, error) {
	props, err := LoadProps(db, table, typ, value)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]*Prop)
	for _, prop := range props {
		ret[key(prop)] = prop
	}
	return ret, nil
}

func LoadProps(db *sql.DB, table string, typ *string, value *string) ([]*Prop, error) {
	rawQuery := fmt.Sprintf("SELECT id, locale_id, type, value FROM %s WHERE 1=1", table)
	var res *sql.Rows
	var err error
	params := []interface{}{}
	props := []*Prop{}
	if typ != nil {
		rawQuery = fmt.Sprintf("%s AND type = ?", rawQuery)
		params = append(params, *typ)
	}
	if value != nil {
		rawQuery = fmt.Sprintf("%s AND value = ?", rawQuery)
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
		var locale_id int64
		if err := res.Scan(&id, &locale_id, &typ, &value); err != nil {
			return nil, fmt.Errorf("failed to read row #%d of %s, %s", rowId, table, err)
		}
		props = append(props, NewProp(id, locale_id, typ, value))
		rowId += 1
	}
	return props, nil
}
