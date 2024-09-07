package seed

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

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
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s VALUES (NULL, ?, ?, ?)", u.table))
	if err != nil {
		return fmt.Errorf("failed to prepare insert query, %s", err)
	}

	sizeOfData := fmt.Sprint(len(data))
	var batchSize int64
	var timeStart = time.Now()
	var lastBatchStart time.Time = timeStart
	var lastBatchTime time.Duration
	batchSize, err = strconv.ParseInt(sizeOfData[0:len(sizeOfData)-1], 10, 64)
	if err != nil {
		return err
	}
	numRows := int64(len(data))
	numBatchesTotal := numRows / batchSize
	numBatchesLeft := numBatchesTotal

	numInserted := 0
	// rowStartTime := time.Now()
	// var timePerRow time.Duration = 0
	var i int
	var item ParserRow
	print("\x1b[?25l")
	defer func() { print("\x1b[?25h") }()
	for i, item = range data {
		if int64(i)%batchSize == 0 {
			eta := time.Second * time.Duration(float64(numBatchesLeft)*lastBatchTime.Seconds())
			progress_value := float64(i) / float64(len(data))
			numBars := 10
			numProgressBars := int(progress_value * float64(numBars))
			progress := fmt.Sprintf("%s%s", strings.Repeat("▉", numProgressBars), strings.Repeat(" ", numBars-numProgressBars))
			fmt.Fprintf(os.Stderr, "\r\x1b[2K[%s] uploading '\x1b[0;1m%s\x1b[0m' [\x1b[0;34m%s\x1b[0m per batch (%d), %d batches left, ETA: \x1b[0;33m%s\x1b[0m]\r", progress, item.value, lastBatchTime, batchSize, numBatchesLeft, eta)
			now := time.Now()
			lastBatchTime = now.Sub(lastBatchStart)
			lastBatchStart = now
			numBatchesLeft = (numRows - int64(i)) / batchSize
		}
		if !propExists(item.value) {
			_, err := stmt.Exec(item.locale.Id, u.typ, item.value)
			if err != nil {
				return fmt.Errorf("failed to insert item #%d '%#+v': %s", i, item, err)
			}
			numInserted += 1
		}
		// timePerRow = time.Since(rowStartTime)
	}
	println()
	slog.Debug(fmt.Sprintf("Inserted %d/%d fresh values in %s", numInserted, len(data), time.Since(timeStart)))
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
