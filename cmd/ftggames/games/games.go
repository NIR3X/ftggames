package games

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"sync"

	"github.com/NIR3X/ftggames/cmd/ftggames/consts"
	"github.com/NIR3X/logger"
	"github.com/NIR3X/tmplreload"
)

type Games struct {
	Games string
}

type Game struct {
	Name        string
	Description string
	Preview     string
	PlayPath    string
	PlayText    string
}

type GameColl struct {
	mtx   sync.RWMutex
	games map[string]*Game
}

func NewGameColl() *GameColl {
	return &GameColl{
		games: make(map[string]*Game),
	}
}

func (g *GameColl) Update(gameXml string) {
	gameXmlData, err := os.ReadFile(gameXml)
	if err != nil {
		logger.Eprintln(err)
		return
	}

	game := &Game{}
	err = xml.Unmarshal(gameXmlData, game)
	if err != nil {
		logger.Eprintln(err)
		return
	}

	gameDirName := filepath.Base(filepath.Dir(gameXml))
	gamePath := "/" + consts.GameDir + "/" + gameDirName + "/"
	game.Preview = gamePath + game.Preview
	game.PlayPath = gamePath + game.PlayPath

	g.mtx.Lock()
	g.games[gameXml] = game
	g.mtx.Unlock()
}

func (g *GameColl) Remove(gameXml string) {
	g.mtx.Lock()
	delete(g.games, gameXml)
	g.mtx.Unlock()
}

func (g *GameColl) GetGamesTmplData(gameTmpl tmplreload.CollTmpl) *Games {
	g.mtx.RLock()
	defer g.mtx.RUnlock()

	games := &bytes.Buffer{}
	for _, game := range g.games {
		err := gameTmpl.Execute(games, game)
		if err != nil {
			logger.Eprintln(err)
		}
	}
	return &Games{
		Games: games.String(),
	}
}
