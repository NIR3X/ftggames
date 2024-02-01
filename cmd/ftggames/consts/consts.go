package consts

import (
	"time"
)

const (
	FileCacheMaxSize          = 2 * 1024 * 1024
	FileWatcherUpdateInterval = 15 * time.Second
	RootDir                   = "www"
	GameXmlName               = "ftg-game.xml"
	GameDir                   = "games"
)
