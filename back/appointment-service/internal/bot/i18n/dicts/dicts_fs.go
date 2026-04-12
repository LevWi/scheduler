package dicts

import (
	"embed"
)

//go:embed active.*.toml
var BotDictFiles embed.FS
