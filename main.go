package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/welschmorgan/datagen/internal/models"
)

const DB_FILE string = "resources.db"

func main() {
	db, err := sql.Open("sqlite3", DB_FILE)
	if err != nil {
		log.Fatalf("failed to open resources DB: %s\n", err)
	}
	defer db.Close()
	resources := models.LoadResources(db)
	for _, r := range resources {
		fmt.Println(r)
	}
}
