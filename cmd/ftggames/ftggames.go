package main

import (
	"io"
	"net/http"
	"path/filepath"

	"github.com/NIR3X/filecache"
	"github.com/NIR3X/filewatcher"
	"github.com/NIR3X/ftggames/cmd/ftggames/consts"
	"github.com/NIR3X/ftggames/cmd/ftggames/games"
	"github.com/NIR3X/httpcontentwriter"
	"github.com/NIR3X/logger"
	"github.com/NIR3X/multisender"
	"github.com/NIR3X/tmplreload"
)

var listenAddr = ":8000"

func main() {
	gameColl := games.NewGameColl()
	fileCache := filecache.NewFileCache(consts.FileCacheMaxSize)
	multiSender := multisender.NewMultiSender(fileCache)
	// we set minUpdateIntvlSecs to -1 because we manually manage the reloading of the templates
	tmplColl := tmplreload.NewTmplColl(60, -1)
	// we can close it because we manually manage the removal of stale template files
	tmplColl.Close()
	fileWatcher := filewatcher.NewFileWatcher(consts.FileWatcherUpdateInterval, func(path string, isDir bool) { // create
		if isDir {
			return
		}
		switch filepath.Ext(path) {
		case ".gohtml":
			err := tmplColl.ParseFiles(path)
			if err != nil {
				logger.Eprintln(err)
			}
		case ".xml":
			if filepath.Base(path) == consts.GameXmlName {
				gameColl.Update(path)
				break
			}
			fallthrough
		default:
			err := fileCache.Update(path)
			if err != nil {
				logger.Eprintln(err)
			}
		}
	}, func(path string, isDir bool) { // remove
		if isDir {
			return
		}
		switch filepath.Ext(path) {
		case ".gohtml":
			tmplColl.RemoveFiles(path)
		case ".xml":
			if filepath.Base(path) == consts.GameXmlName {
				gameColl.Remove(path)
				break
			}
			fallthrough
		default:
			fileCache.Delete(path)
		}
	}, func(path string, isDir bool) { // modify
		if isDir {
			return
		}
		switch filepath.Ext(path) {
		case ".gohtml":
			err := tmplColl.ReloadFiles(path)
			if err != nil {
				logger.Eprintln(err)
			}
		case ".xml":
			if filepath.Base(path) == consts.GameXmlName {
				gameColl.Update(path)
				break
			}
			fallthrough
		default:
			err := fileCache.Update(path)
			if err != nil {
				logger.Eprintln(err)
			}
		}
	})
	defer fileWatcher.Close()
	err := fileWatcher.Watch(consts.RootDir)
	if err != nil {
		logger.Eprintln(err)
		return
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		ext := filepath.Ext(urlPath)

		w := writer.(io.Writer)
		if ext == ".css" {
			writer.Header().Set("Content-Type", "text/css")
		} else {
			w = httpcontentwriter.NewHttpContentWriter(writer)
		}

		path := filepath.Join(consts.RootDir, urlPath)
		switch urlPath {
		case "/", "/index.html":
			gameTmpl := tmplColl.Lookup(filepath.Join(path, "index.game.gohtml"))
			if gameTmpl == nil {
				logger.Eprintln("index.game.gohtml template not found")
				return
			}
			err := tmplColl.ExecuteTemplate(w, filepath.Join(path, "index.gohtml"), gameColl.GetGamesTmplData(gameTmpl))
			if err != nil {
				logger.Eprintln(err)
			}
		default:
			if ext == ".gohtml" {
				_ = tmplColl.ExecuteTemplate(w, path, nil)
				return
			}

			r, ident := fileCache.GetCached(path)
			switch ident {
			case filecache.Cached:
				_, err = io.Copy(w, r)
				if err != nil {
					logger.Eprintln(err)
				}
			case filecache.Piped:
				multiSenderWriter := multiSender.Add(path, w)
				multiSenderWriter.Wait()
			}
		}
	})
	logger.Println("Listening on " + listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		logger.Eprintln(err)
		return
	}
}
