package main

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/earlofurl/scenes-of-shakespeare/sqlc"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type result struct {
	Work        string
	WorkID      string
	Act         int
	Scene       int
	Description string
	Snippet     template.HTML
}

func run() error {
	ctx := context.Background()

	// get the port to listen on
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// get the detail of the database to connect to
	dburl := os.Getenv("DATABASE_URL")
	if dburl == "" {
		log.Fatal().Msg("DATABASE_URL env var not set")
	}

	// connect to the database
	db, err := sql.Open("postgres", dburl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	queries := sqlc.New(db)

	// compile the templates
	tplHome := template.Must(template.New(".").Parse(tplStrHome))
	tplResults := template.Must(template.New(".").Parse(tplStrResults))
	tplScene := template.Must(template.New(".").Parse(tplStrScene))

	// handler to render a scene page
	http.HandleFunc("/scene", func(w http.ResponseWriter, r *http.Request) {
		// get parameters and do some basic validation
		a := r.FormValue("a")
		av, err := strconv.Atoi(a)
		if err != nil || av < 0 || av > 5 {
			log.Err(err).Msg("Error in act value")
			http.Error(w, "not found", 404)
			return
		}
		s := r.FormValue("s")
		sv, err := strconv.Atoi(s)
		if err != nil || sv < 0 || sv > 15 {
			log.Err(err).Msg("Error in scene value")
			http.Error(w, "not found", 404)
			return
		}
		wid := r.FormValue("w")
		if len(wid) < 5 || len(wid) > 14 {
			log.Err(err).Msg("Error in work ID value")
			http.Error(w, "not found", 404)
			return
		}
		// fetch the work title
		work, err := queries.GetWork(ctx, wid)
		if err != nil {
			log.Err(err).Msg("Error fetching the work title")
			http.Error(w, "not found", 404)
			return
		}

		// fetch the scene description and body text
		var body string
		scene, err := queries.GetScene(ctx, &sqlc.GetSceneParams{
			Workid: wid,
			Act:    int32(av),
			Scene:  int32(sv),
		})
		if err != nil {
			log.Err(err).Msgf("Error fetching the scene (work: %s, act: %d, scene: %d)", wid, av, sv)
			http.Error(w, "not found", 404)
			return
		} else {
			body = strings.Replace(scene.Body, "\n\n", "<p>", -1)
			body = strings.Replace(body, "\n", "<br>", -1)
			err := tplScene.Execute(w, map[string]interface{}{
				"Work":        work,
				"Act":         av,
				"Scene":       sv,
				"Description": scene.Description,
				"Body":        template.HTML(body),
			})
			if err != nil {
				log.Err(err).Msg("Error executing template for scene page")
				return
			}
		}

	})

	// handler to render home page and query results pages
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.FormValue("q")
		if q == "" {
			err := tplHome.Execute(w, nil)
			if err != nil {
				log.Err(err).Msg("Error executing template for home page")
				return
			}
			return
		}
		if len(q) > 100 {
			q = q[:100]
		}
		rows, err := queries.Search(ctx, q)
		if err != nil {
			log.Err(err).Msg("Error executing search query")
			http.Error(w, "not found", 404)
			return
		}
		var results []result
		for _, row := range rows {
			results = append(results, result{
				Work:        row.Workid,
				WorkID:      row.Workid,
				Act:         int(row.Act),
				Scene:       int(row.Scene),
				Description: row.Description,
				Snippet:     template.HTML(strings.Replace(row.Headline, "\n", "<br>", -1)),
			})
		}
		err = tplResults.Execute(w, map[string]interface{}{
			"Query":   q,
			"Results": results,
		})
	})

	// start the http server
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("failed to run application")
	}
}
