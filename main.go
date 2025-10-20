package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/humbertovnavarro/kismet-topsql/kismet-to-psql/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	sqlitePath := flag.String("sqlite", "", "Path to source SQLite database (e.g. ./kismet.db)")
	postgresDSN := flag.String("dsn", "", "PostgreSQL DSN (e.g. postgres://user:pass@localhost:5432/dbname?sslmode=disable)")
	copyData := flag.Bool("copy", true, "Copy data from SQLite to PostgreSQL after migrating schema")
	batchSize := flag.Int("batch", 25, "Batch size for inserts to avoid PostgreSQL parameter limit")
	flag.Parse()

	if *sqlitePath == "" || *postgresDSN == "" {
		fmt.Println("Usage: migrate --sqlite ./kismet.db --dsn postgres://user:pass@localhost/dbname?sslmode=disable [--copy] [--batch N]")
		os.Exit(1)
	}

	sqliteDB, err := gorm.Open(sqlite.Open(*sqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open SQLite: %v", err)
	}
	pgDB, err := gorm.Open(postgres.Open(*postgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	fmt.Println("Migrating schema to PostgreSQL...")

	err = pgDB.AutoMigrate(
		&models.Kismet{},
		&models.Device{},
		&models.Packet{},
		&models.Data{},
		&models.DataSource{},
		&models.Alert{},
		&models.Message{},
		&models.Snapshot{},
	)
	if err != nil {
		log.Fatalf("Schema migration failed: %v", err)
	}

	fmt.Println("✅ Schema migration complete")

	if *copyData {
		fmt.Println("Copying data from SQLite to PostgreSQL...")
		totalCopied := 0

		copyTable := func(name string, model interface{}) {
			var rows []map[string]interface{}
			if err := sqliteDB.Model(model).Find(&rows).Error; err != nil {
				log.Printf("⚠️  Failed to query %s: %v", name, err)
				return
			}

			count := len(rows)
			if count == 0 {
				fmt.Printf("ℹ️  No rows found in %s\n", name)
				return
			}

			for i := 0; i < count; i += *batchSize {
				end := i + *batchSize
				if end > count {
					end = count
				}
				batch := rows[i:end]

				if err := pgDB.Model(model).Create(&batch).Error; err != nil {
					log.Printf("⚠️  Failed to copy batch for %s (%d–%d): %v", name, i, end, err)
					return
				}
			}

			fmt.Printf("✅ Copied %d rows from %s\n", count, name)
			totalCopied += count
		}

		copyTable("Kismet", &models.Kismet{})
		copyTable("Device", &models.Device{})
		copyTable("Packet", &models.Packet{})
		copyTable("Data", &models.Data{})
		copyTable("DataSource", &models.DataSource{})
		copyTable("Alert", &models.Alert{})
		copyTable("Message", &models.Message{})
		copyTable("Snapshot", &models.Snapshot{})

		fmt.Printf("\n🎉 Data copy complete. Total rows copied: %d\n", totalCopied)
	}
}
