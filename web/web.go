package web

import (
	"embed"
)

var (
	//go:embed template/*
	Templates embed.FS
	//go:embed static/*
	Static embed.FS
)
