package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/humbertovnavarro/kismet-topsql/kismet-to-psql/pkg/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	postgresDSN string
	batchSize   int
	copyData    bool
	mu          sync.Mutex
	lastLog     string
)

// --- Core Migration Logic --- //
func MigrateKismet(sqlitePath, postgresDSN string, copyData bool, batchSize int) error {
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open SQLite: %w", err)
	}
	pgDB, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	logMsg("Migrating schema to PostgreSQL...")
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
		return fmt.Errorf("schema migration failed: %w", err)
	}
	logMsg("✅ Schema migration complete")

	if copyData {
		logMsg("Copying data from SQLite to PostgreSQL...")
		totalCopied := 0

		copyTable := func(name string, model interface{}) {
			var rows []map[string]interface{}
			if err := sqliteDB.Model(model).Find(&rows).Error; err != nil {
				logMsg(fmt.Sprintf("⚠️ Failed to query %s: %v", name, err))
				return
			}

			count := len(rows)
			if count == 0 {
				logMsg(fmt.Sprintf("ℹ️ No rows found in %s", name))
				return
			}

			for i := 0; i < count; i += batchSize {
				end := i + batchSize
				if end > count {
					end = count
				}
				batch := rows[i:end]
				if err := pgDB.Model(model).Create(&batch).Error; err != nil {
					logMsg(fmt.Sprintf("⚠️ Failed batch %s (%d–%d): %v", name, i, end, err))
					return
				}
			}
			logMsg(fmt.Sprintf("✅ Copied %d rows from %s", count, name))
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

		logMsg(fmt.Sprintf("🎉 Data copy complete. Total rows copied: %d", totalCopied))
	}
	return nil
}

// --- Web Server --- //
func main() {
	// Load .env if available
	_ = godotenv.Load()

	postgresDSN = os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		log.Fatal("Missing POSTGRES_DSN in environment or .env file")
	}

	batchSizeEnv := os.Getenv("BATCH_SIZE")
	if batchSizeEnv == "" {
		batchSizeEnv = "25"
	}
	var err error
	batchSize, err = strconv.Atoi(batchSizeEnv)
	if err != nil {
		batchSize = 25
	}

	copyDataEnv := os.Getenv("COPY_DATA")
	copyData = copyDataEnv != "false" && copyDataEnv != "0"

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/logs", handleLogs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
	<title>Kismet DB Import</title>
	<style>
		body { font-family: sans-serif; background: #111; color: #eee; display: flex; flex-direction: column; align-items: center; margin-top: 10%; }
		.dropzone { border: 2px dashed #666; padding: 40px; border-radius: 12px; text-align: center; width: 400px; background: #222; cursor: pointer; }
		pre { background: #000; color: #0f0; padding: 10px; width: 400px; height: 300px; overflow-y: scroll; }
	</style>
	</head>
	<body>
		<h2>Kismet SQLite Importer</h2>
		<div class="dropzone" id="dropzone">Drop .kismet file here or click to upload</div>
		<pre id="log"></pre>
		<script>
		const dz = document.getElementById('dropzone');
		const logEl = document.getElementById('log');
		dz.addEventListener('click', () => {
			const inp = document.createElement('input');
			inp.type = 'file';
			inp.accept = '.kismet';
			inp.onchange = () => upload(inp.files[0]);
			inp.click();
		});
		dz.addEventListener('dragover', e => { e.preventDefault(); dz.style.borderColor = '#0f0'; });
		dz.addEventListener('dragleave', e => { e.preventDefault(); dz.style.borderColor = '#666'; });
		dz.addEventListener('drop', e => {
			e.preventDefault(); dz.style.borderColor = '#666';
			const file = e.dataTransfer.files[0];
			upload(file);
		});
		function upload(file) {
			const formData = new FormData();
			formData.append('file', file);
			logEl.textContent = 'Uploading ' + file.name + '...\\n';
			fetch('/upload', { method: 'POST', body: formData })
				.then(r => r.text())
				.then(t => logEl.textContent += t + '\\n');
			setInterval(() => {
				fetch('/logs').then(r => r.text()).then(t => logEl.textContent = t);
			}, 1000);
		}
		</script>
	</body>
	</html>`
	w.Header().Set("Content-Type", "text/html")
	template.Must(template.New("index").Parse(tmpl)).Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File upload error: "+err.Error(), 400)
		return
	}
	defer file.Close()

	tmpPath := filepath.Join(os.TempDir(), header.Filename)
	out, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), 500)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	go func() {
		if err := MigrateKismet(tmpPath, postgresDSN, copyData, batchSize); err != nil {
			logMsg(fmt.Sprintf("❌ Migration failed: %v", err))
		} else {
			logMsg("✅ Migration complete for " + header.Filename)
		}
	}()

	fmt.Fprintf(w, "Uploaded %s, migration started...\n", header.Filename)
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, lastLog)
}

func logMsg(msg string) {
	mu.Lock()
	defer mu.Unlock()
	lastLog += msg + "\n"
	log.Println(msg)
}
