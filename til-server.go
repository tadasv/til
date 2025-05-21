package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"html/template"

	_ "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	"github.com/go-chi/chi/v5"
	"github.com/ncruces/go-sqlite3"
)

const nav = `
<nav>
	<p>
		<a href="/">Tadas Vilkeliskis's TILs</a>
	</p>
</nav>
`

const stylesheet = `
	<style>
		html {
		  padding: 0;
		  margin: 0;
		}

		body {
			font-family: sans-serif;
			margin: 0;
		}

		p.footer-actions {
			font-size: 0.8rem;
		}

		a.topic {
		  font-size: 0.8rem;
		  padding: 0.25rem;
		  background-color: #FDB913;
		  text-decoration: none;
		  color: #000;
		}

		pre {
			white-space: pre-wrap;
		}
		pre[lang=wide] {
			white-space: pre;
			overflow: auto;
		}
		nav {
			text-align: left;
			background: #FDB913;
			color: black;
		}
		nav p {
			display: block;
			margin: 0;
			padding: 4px 0px 4px 2em;
		}
		nav a:link,
		nav a:visited,
		nav a:hover,
		nav a:focus,
		nav a:active {
			color: black;
			text-decoration: none;
		}
		section.body {
			padding: 0.5em 2em;
			max-width: 800px;
		}
		@media (max-width: 600px) {
			section.body {
				padding: 0em 1em;
			}
			nav p {
				padding: 4px 0px 4px 1em;
			}
		}
		a, pre, code {
			overflow-wrap: break-word;
		}
		table {
			border-collapse: collapse;
		}
		th, td {
			padding: 0.3em;
			border: 1px solid #D3D3D3;
			word-wrap: anywhere;
		}
		th {
			background-color: #eee;
		}
		ul.related {
			list-style-type: none;
			margin: 0;
			padding: 0;
		}
		ul.related li {
			margin-bottom: 0.3em;
		}
		blockquote {
			margin: 1em 0;
			border-left: 0.75em solid #9e6bb52e;
			padding-left: 0.75em;
		}
	</style>
`

var indexTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Tadas Vilkeliskis's TILs</title>
	` + stylesheet + `
</head>
<body>
` + nav + `
<section class="body">
  <h1>Tadas Vilkeliskis's TILs</h1>
  <p>A list of things I've learned and collected in <a href="https://github.com/tadasv/til">tadasv/til</a>.</p>

  <p>
  	<strong>Topics:</strong>
	{{range $index, $count := .TopicCounts}}
	<a href="/{{$count.Topic}}">{{$count.Topic}}</a> {{$count.Count}}{{if not (eq $index $.TopicCountsMinusOne) }} · {{end}}
	{{end}}
  </p>

  <h2>Recent TILs</h2>

  <ul>
    {{range .RecentTils}}
	<h3>
		<a class="topic" href="/{{.Topic}}">{{.Topic}}</a> <a href="{{.Topic}}/{{.Slug}}">{{.Title}}</a> - {{.CreatedAt.Format "2006-01-02"}}
	</h3>
    {{end}}
  </ul>
</section>
</body>
</html>
`))

var tilTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	` + stylesheet + `
</head>
<body>
` + nav + `
<section class="body">
{{.Html}}
<hr>
<p class="footer-actions">
	Created at {{.CreatedAt}} · Updated at {{.UpdatedAt}} · <a href="https://github.com/tadasv/til/blob/main/{{.Path}}">Edit</a> · <a href="https://github.com/tadasv/til/commits/main/{{.Path}}">History</a>
</p>
</section>
</body>
</html>
`))

type TopicCount struct {
	Topic string
	Count int
}

func main() {
	db, err := sqlite3.Open("./tils.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		recentTils, err := getTils(db, "", 30)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		topicCounts, err := getTopicCounts(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTemplate.Execute(w, map[string]interface{}{
			"RecentTils":  recentTils,
			"TopicCounts": topicCounts,
			// for template rendering to avoide external functions
			"TopicCountsMinusOne": len(topicCounts) - 1,
		})
	})

	router.Get("/{topic}/{path}", func(w http.ResponseWriter, r *http.Request) {
		topic := chi.URLParam(r, "topic")
		path := chi.URLParam(r, "path")
		til, err := getTil(db, topic+"/"+path+".md")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tilTemplate.Execute(w, til)
	})

	http.ListenAndServe(":3000", router)
}

func getTopicCounts(db *sqlite3.Conn) ([]*TopicCount, error) {
	stmt, _, err := db.Prepare("SELECT topic, COUNT(*) FROM til GROUP BY topic")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var topicCounts []*TopicCount
	for stmt.Step() {
		topicCounts = append(topicCounts, &TopicCount{
			Topic: stmt.ColumnText(0),
			Count: stmt.ColumnInt(1),
		})
	}

	return topicCounts, nil
}

func getTils(db *sqlite3.Conn, topic string, limit int) ([]*Record, error) {
	var stmt *sqlite3.Stmt
	var err error

	if topic == "" {
		stmt, _, err = db.Prepare("SELECT title, html, slug, created_at, updated_at, topic, url FROM til ORDER BY created_at DESC LIMIT ?")
		if err != nil {
			return nil, err
		}
		stmt.BindInt(1, limit)
	} else {
		stmt, _, err = db.Prepare("SELECT title, html, slug, created_at, updated_at, topic, url FROM til WHERE topic = ? ORDER BY created_at DESC LIMIT ?")
		if err != nil {
			return nil, err
		}
		stmt.BindText(1, topic)
		stmt.BindInt(2, limit)
	}
	defer stmt.Close()

	var tils []*Record
	for stmt.Step() {
		til := &Record{}
		til.Title = stmt.ColumnText(0)
		til.Html = template.HTML(stmt.ColumnText(1))
		til.Slug = stmt.ColumnText(2)
		til.CreatedAt, err = time.Parse(time.RFC3339, stmt.ColumnText(3))
		if err != nil {
			return nil, err
		}
		til.UpdatedAt, err = time.Parse(time.RFC3339, stmt.ColumnText(4))
		if err != nil {
			return nil, err
		}
		til.Topic = stmt.ColumnText(5)
		til.Url = stmt.ColumnText(6)
		tils = append(tils, til)
	}

	return tils, nil
}

func getTil(db *sqlite3.Conn, path string) (*Record, error) {
	stmt, _, err := db.Prepare("SELECT title, html, created_at, updated_at, topic, url, path FROM til WHERE path = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var title string
	var content string
	var createdAt time.Time
	var updatedAt time.Time
	var topic string
	var url string
	var tilPath string

	stmt.BindText(1, path)
	if !stmt.Step() {
		return nil, errors.New("not found")
	}

	title = stmt.ColumnText(0)
	content = stmt.ColumnText(1)
	createdAt, err = time.Parse(time.RFC3339, stmt.ColumnText(2))
	if err != nil {
		return nil, err
	}
	updatedAt, err = time.Parse(time.RFC3339, stmt.ColumnText(3))
	if err != nil {
		return nil, err
	}
	topic = stmt.ColumnText(4)
	url = stmt.ColumnText(5)
	tilPath = stmt.ColumnText(6)

	return &Record{
		Title:     title,
		Html:      template.HTML(content),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Topic:     topic,
		Url:       url,
		Path:      tilPath,
	}, nil
}
