package models

import "database/sql"

type Locale struct {
	Id   int64
	Name string
}

func NewLocale(id int64, name string) *Locale {
	return &Locale{
		Id:   id,
		Name: name,
	}
}

func LoadLocales(db *sql.DB) ([]*Locale, error) {
	res, err := db.Query("SELECT * FROM locale")
	if err != nil {
		return nil, err
	}
	ret := []*Locale{}
	for res.Next() {
		var id int64 = 0
		var name string
		if err = res.Scan(&id, &name); err != nil {
			return nil, err
		}
		ret = append(ret, NewLocale(id, name))
	}
	return ret, nil
}
