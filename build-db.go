package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Use sqlite-vec for semanting search in the future.
	_ "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	"github.com/ncruces/go-sqlite3"
	"github.com/yuin/goldmark"
)

func main() {
	db, err := sqlite3.Open("./tils.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS til (
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			path TEXT PRIMARY KEY NOT NULL,
			slug TEXT NOT NULL,
			topic TEXT NOT NULL,
			title TEXT NOT NULL,
			url TEXT NOT NULL,
			body TEXT NOT NULL,
			html TEXT NOT NULL
		)
	`); err != nil {
		log.Fatal(err)
	}

	// Read all markdown files
	files, err := filepath.Glob("*/*.md")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		bodyBytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		body := strings.TrimSpace(string(bodyBytes))
		title := strings.ReplaceAll(strings.TrimSpace(strings.Split(body, "\n")[0]), "# ", "")

		// Extract topic from directory name
		topic := filepath.Base(filepath.Dir(file))

		// Create slug from filename without extension
		slug := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		path := filepath.Join(topic, slug+".md")

		var htmlBuffer bytes.Buffer
		if err := goldmark.Convert(bodyBytes, &htmlBuffer); err != nil {
			log.Fatal(err)
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			log.Fatal(err)
		}

		record := Record{
			CreatedAt: fileInfo.ModTime(),
			UpdatedAt: time.Now(),
			Path:      path,
			Slug:      slug,
			Topic:     topic,
			Title:     title,
			Url:       fmt.Sprintf("https://github.com/tadasv/til/blob/main/%s", path),
			Body:      body,
			Html:      template.HTML(htmlBuffer.String()),
		}

		oldRecord := Record{}
		stmt, _, err := db.Prepare(`
			SELECT path, slug, topic, title, url, body, html
			FROM til
			WHERE path = ?
		`)
		if err != nil {
			log.Fatal(err)
		}

		stmt.BindText(1, path)

		if stmt.Step() {
			oldRecord.Path = stmt.ColumnText(0)
			oldRecord.Slug = stmt.ColumnText(1)
			oldRecord.Topic = stmt.ColumnText(2)
			oldRecord.Title = stmt.ColumnText(3)
			oldRecord.Url = stmt.ColumnText(4)
			oldRecord.Body = stmt.ColumnText(5)
			oldRecord.Html = template.HTML(stmt.ColumnText(6))
		}
		stmt.Close()

		if oldRecord.Body != record.Body {
			if oldRecord.Body == "" {
				println("inserting record", path)
				insertStmt, _, err := db.Prepare(`	
					INSERT INTO til (created_at, updated_at, path, slug, topic, title, url, body, html)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				`)
				if err != nil {
					log.Fatal(err)
				}
				insertStmt.BindText(1, time.Now().Format(time.RFC3339))
				insertStmt.BindText(2, time.Now().Format(time.RFC3339))
				insertStmt.BindText(3, record.Path)
				insertStmt.BindText(4, record.Slug)
				insertStmt.BindText(5, record.Topic)
				insertStmt.BindText(6, record.Title)
				insertStmt.BindText(7, record.Url)
				insertStmt.BindText(8, record.Body)
				insertStmt.BindText(9, string(record.Html))
				if err := insertStmt.Exec(); err != nil {
					log.Fatal(err)
				}
				insertStmt.Close()
			} else {
				println("updating record", path)
				updateStmt, _, err := db.Prepare(`
					UPDATE til
					SET updated_at = ?, body = ?, html = ?, title = ?, url = ?, slug = ?, topic = ?
					WHERE path = ?
				`)
				if err != nil {
					log.Fatal(err)
				}
				updateStmt.BindText(1, time.Now().Format(time.RFC3339))
				updateStmt.BindText(2, record.Body)
				updateStmt.BindText(3, string(record.Html))
				updateStmt.BindText(4, record.Title)
				updateStmt.BindText(5, record.Url)
				updateStmt.BindText(6, record.Slug)
				updateStmt.BindText(7, record.Topic)
				updateStmt.BindText(8, record.Path)
				if err := updateStmt.Exec(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
