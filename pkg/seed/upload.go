package seed

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/welschmorgan/datagen/pkg/models"
)

type Uploader interface {
	Upload(data []ParserRow) error
}

type BasicUploader struct {
	Uploader

	db    *sql.DB
	table string
	typ   string
}

func NewBasicUploader(db *sql.DB, table, typ string) *BasicUploader {
	return &BasicUploader{db: db, table: table, typ: typ}
}

func (u *BasicUploader) Upload(data []ParserRow) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	props, err := models.LoadProps(u.db, u.table, &u.typ, nil)
	if err != nil {
		return fmt.Errorf("failed to load props, %s", err)
	}
	slog.Debug(fmt.Sprintf("Loaded %d props", len(props)))

	propExists := func(value string) bool {
		for _, prop := range props {
			if strings.EqualFold(prop.Value, value) {
				return true
			}
		}
		return false
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (NULL, ?, ?)", u.table))
	if err != nil {
		return fmt.Errorf("failed to prepare insert query, %s", err)
	}
	numInserted := 0
	for i, item := range data {
		if !propExists(item.value) {
			_, err := stmt.Exec(u.typ, item.value)
			if err != nil {
				return fmt.Errorf("failed to insert item #%d '%s': %s", i, item, err)
			}
			numInserted += 1
		}
	}
	slog.Debug(fmt.Sprintf("Inserted %d/%d fresh values", numInserted, len(data)))
	return tx.Commit()
}

type QueryUploader struct {
	Uploader

	db    *sql.DB
	query string
}

func NewQueryUploader(db *sql.DB, query string) *QueryUploader {
	return &QueryUploader{db: db, query: query}
}

func (u *QueryUploader) Upload(data []ParserRow) error {
	// fmt.Printf("Executing query %s", u.query)
	_, err := u.db.Exec(u.query)
	return err
}
