package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	app := pocketbase.New()

	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			if !e.Router.HasRoute(http.MethodGet, "/{path...}") {
				e.Router.GET("/{path...}", apis.Static(os.DirFS("pb_public"), true))
			}
			return e.Next()
		},
		Priority: 999, // execute as latest as possible to allow users to provide their own route
	})
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			path := e.Request.URL.Path
			if strings.HasPrefix(path, "/_/") { // Skip Gzip for Admin UI
				return e.Next()
			}
			return apis.Gzip().Func(e)
		})

		slog.Info("Importing collections")
		collections, err := os.ReadFile("collections.json")
		if err == nil {
			err = app.ImportCollectionsByMarshaledJSON(collections, true)
			if err != nil {
				return err
			}
		} else {
			slog.Warn("Failed to import collections", "err", err)
		}

		importCategoriesFromJson(app)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

type CollectionImport struct {
	Name        string `json:"name"`
	MustHint    bool   `json:"mustHint"`
	MustPresent bool   `json:"mustPresent"`
	Words       []struct {
		Word     string `json:"word"`
		Hint     string `json:"hint"`
		HintLong string `json:"hintLong"`
	} `json:"words"`
}

func importCategoriesFromJson(app *pocketbase.PocketBase) {
	imports, err := os.ReadDir("imports")
	if err != nil {
		slog.Warn("Failed to read json imports", "err", err)
	}
	for _, f := range imports {
		b, err := os.ReadFile(filepath.Join("imports", f.Name()))
		if err != nil {
			slog.Warn("Failed to read json import", "err", err)
			continue
		}
		var imprt CollectionImport
		if err := json.Unmarshal(b, &imprt); err != nil {
			slog.Warn("Failed to unmarshal json import", "err", err)
			continue
		}

		if r, _ := app.FindFirstRecordByData("categories", "name",
			imprt.Name); r == nil {
			slog.Info("Importing category", "file", f.Name())
			coll, _ := app.FindCollectionByNameOrId("categories")
			recordCat := core.NewRecord(coll)
			recordCat.Set("name", imprt.Name)
			recordCat.Set("mustHint", imprt.MustHint)
			recordCat.Set("mustPresent", imprt.MustPresent)
			if err := app.Save(recordCat); err != nil {
				slog.Warn("Failed to save category", "err", err)
				continue
			}

			for _, w := range imprt.Words {
				coll, _ := app.FindCollectionByNameOrId("words")
				record := core.NewRecord(coll)
				record.Set("category", recordCat.Id)
				record.Set("word", w.Word)
				record.Set("hint", w.Hint)
				record.Set("hintLong", w.HintLong)
				if err := app.Save(record); err != nil {
					slog.Warn("Failed to save word", "err", err)
					continue
				}
			}
		}
	}
}
