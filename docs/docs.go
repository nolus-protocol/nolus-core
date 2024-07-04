package docs

import "embed"

//go:embed swagger/*
var SwaggerFiles embed.FS
