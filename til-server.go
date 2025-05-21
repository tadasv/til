package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"html/template"

	_ "github.com/asg017/sqlite-vec-go-bindings/ncruces"
	"github.com/go-chi/chi/v5"
	"github.com/ncruces/go-sqlite3"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
		ul.tils {
			padding-left: 0;
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
  <h1>Tadas Vilkeliskis: TILs</h1>
  <p>A list of things I've learned and collected in <a href="https://github.com/tadasv/til">tadasv/til</a>.</p>

  <p>
  	<strong>Topics:</strong>
	{{range $index, $count := .TopicCounts}}
	<a href="/{{$count.Topic}}">{{$count.Topic}}</a> {{$count.Count}}{{if not (eq $index $.TopicCountsMinusOne) }} · {{end}}
	{{end}}
  </p>

  <h2>Recent TILs</h2>

  <ul class="tils">
    {{range .RecentTils}}
	<h3>
		<a class="topic" href="/{{.Topic}}">{{.Topic}}</a> <a href="{{.Topic}}/{{.Slug}}">{{.Title}}</a> - {{.CreatedAt.Format "2006-01-02"}}
	</h3>
	{{.Html}}
	<a style="font-size: 0.8rem;" href="{{.Topic}}/{{.Slug}}">Continue reading</a>
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

var topicTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Tadas Vilkeliskis: TILs on {{.Topic}}</title>
	` + stylesheet + `
</head>
<body>
` + nav + `
<section class="body">
  <h1>Tadas Vilkeliskis: TILs on {{.Topic}}</h1>
  <ul class="tils">
    {{range .Tils}}
	<h3>
		<a class="topic" href="/{{.Topic}}">{{.Topic}}</a> <a href="{{.Topic}}/{{.Slug}}">{{.Title}}</a> - {{.CreatedAt.Format "2006-01-02"}}
	</h3>
	{{.Html}}
	<a style="font-size: 0.8rem;" href="{{.Topic}}/{{.Slug}}">Continue reading</a>
    {{end}}
  </ul>
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
		recentTils, err := getTils(db, "", 30, true)
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

	router.Get("/{topic}", func(w http.ResponseWriter, r *http.Request) {
		topic := chi.URLParam(r, "topic")
		tils, err := getTils(db, topic, 30, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		topicTemplate.Execute(w, map[string]interface{}{
			"Topic": topic,
			"Tils":  tils,
		})
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

func getTils(db *sqlite3.Conn, topic string, limit int, preview bool) ([]*Record, error) {
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
		if preview {
			til.Html = getPreview(template.HTML(stmt.ColumnText(1)), 2)
		} else {
			til.Html = template.HTML(stmt.ColumnText(1))
		}
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

func getPreview(html template.HTML, numElements int) template.HTML {
	// Define the context for parsing the fragment.
	// Using 'body' as context is common for arbitrary HTML snippets,
	// allowing them to be parsed as if they were direct children of a <body> tag.
	context := &xhtml.Node{
		Type:     xhtml.ElementNode,
		Data:     "body",    // The tag name of the context element
		DataAtom: atom.Body, // The atom for 'body'
	}

	// Parse the snippet as a fragment within the given context.
	// This returns a slice of nodes that are the direct children of the context.
	nodes, err := xhtml.ParseFragment(strings.NewReader(string(html)), context)
	if err != nil {
		return html
	}

	var extractedHTML strings.Builder
	elementsFound := 0

	// Iterate through the parsed top-level nodes from the fragment
	for idx, node := range nodes {
		if idx == 0 {
			if node.Type == xhtml.ElementNode && node.DataAtom == atom.H1 {
				// Skip header
				continue
			}
		}

		if elementsFound >= numElements {
			break // We have found the desired number of elements
		}

		// We are interested in top-level *element* nodes (e.g., <p>, <h1>, <div>)
		// Other node types like TextNode or CommentNode will be skipped.
		if node.Type == xhtml.ElementNode {
			var nodeHTML bytes.Buffer
			// Render the current element node and all its content to HTML
			if err := xhtml.Render(&nodeHTML, node); err != nil {
				// Log a warning and skip this node if rendering fails
				fmt.Fprintf(os.Stderr, "Warning: Error rendering element node %s: %v\n", node.Data, err)
				continue
			}
			extractedHTML.WriteString(nodeHTML.String()) // Append the HTML of the current element
			elementsFound++
		}
	}

	if elementsFound == 0 {
		return html
	}

	return template.HTML(extractedHTML.String())
}
