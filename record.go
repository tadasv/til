package main

import (
	"html/template"
	"time"
)

type Record struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Path      string
	Slug      string
	Title     string
	Topic     string
	Url       string
	Body      string
	Html      template.HTML
}
