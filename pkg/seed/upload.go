package seed

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/welschmorgan/datagen/pkg/models"
	"github.com/welschmorgan/datagen/pkg/utils"
)

type Uploader interface {
	Upload(data []*ParserRow) error
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

func (u *BasicUploader) Upload(data []*ParserRow) error {
	props, err := models.LoadPropsAsMap(u.db, u.table, &u.typ, nil, func(p *models.Prop) string { return strings.ToLower(p.Value) })
	if err != nil {
		return fmt.Errorf("failed to load props, %s", err)
	}
	slog.Debug(fmt.Sprintf("Loaded %d props", len(props)))

	propExists := func(value string) bool {
		_, ok := props[strings.ToLower(value)]
		return ok
	}

	// timeStart := time.Now()
	filteredRows := []*ParserRow{}
	for _, row := range data {
		if !propExists(row.value) {
			filteredRows = append(filteredRows, row)
		}
	}
	// log.Printf("Filtered %d rows in %s -> %d left", len(data), time.Since(timeStart), len(filteredRows))

	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (NULL, ?, ?, ?)", u.table))
	if err != nil {
		return fmt.Errorf("failed to prepare insert query, %s", err)
	}
	handler := func(r *ParserRow) int64 {
		res, err := stmt.Exec(r.locale.Id, u.typ, r.value)
		if err != nil {
			slog.Error("Failed to insert row", "row id", r.id, "err", err)
			return -1
		}
		id, err := res.LastInsertId()
		if err != nil {
			slog.Error("Failed to retrieve lastInsertId", "row id", r.id, "err", err)
		}
		r.id = id
		return id
	}
	// oldConns := u.db.Stats().MaxOpenConnections
	scheduler := utils.NewSchedulerSingleHandler(100, handler, filteredRows...)
	scheduler.Run()
	// u.db.SetMaxOpenConns(oldConns)

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

type QueryUploader struct {
	Uploader

	db    *sql.DB
	query string
}

func NewQueryUploader(db *sql.DB, query string) *QueryUploader {
	return &QueryUploader{db: db, query: query}
}

func (u *QueryUploader) Upload(data []*ParserRow) error {
	// fmt.Printf("Executing query %s", u.query)
	_, err := u.db.Exec(u.query)
	return err
}
