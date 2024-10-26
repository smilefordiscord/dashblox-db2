package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	// "reflect"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	// "golang.org/x/text/cases"
)

const startupMessage = `Started!`

type Level struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Data        string `json:"data"`
	Owner       string `json:"owner"`
	Difficulty  string `json:"difficulty"`
	Rating      int    `json:"rating"`
	Timestamp   int    `json:"timestamp"`
	Copylock    bool   `json:"copylock"`
	Platformer  bool   `json:"platformer"`
	Featured    bool   `json:"featured"`
	Plays       int    `json:"plays"`
	Newphys     bool   `json:"newphys"`
}

type BasicRequest struct {
	Secret string `json:"secret"`
}

type GetLevelRequest struct {
	Secret string `json:"secret"`
	Id     int    `json:"id"`
}

type RecentTabRequest struct {
	Secret string `json:"secret"`
	MaxId  int    `json:"maxid"`
	Rated  bool   `json:"rated"`
}

type FeaturedTabRequest struct {
	Secret string `json:"secret"`
	MaxId  int    `json:"maxid"`
}

type SearchDiffData struct {
	Unrated      bool `json:"unrated"`
	Auto         bool `json:"auto"`
	Easy         bool `json:"easy"`
	Normal       bool `json:"normal"`
	Hard         bool `json:"hard"`
	Harder       bool `json:"harder"`
	Insane       bool `json:"insane"`
	EasyDemon    bool `json:"easydemon"`
	MediumDemon  bool `json:"mediumdemon"`
	HardDemon    bool `json:"hardemon"`
	InsaneDemon  bool `json:"insanedemon"`
	ExtremeDemon bool `json:"extremedemon"`
}

type SearchRequest struct {
	Secret       string         `json:"secret"`
	Search       string         `json:"search"`
	Offset       int            `json:"offset"`
	SearchType   int            `json:"st"`
	SearchSort   int            `json:"ss"`
	SearchDiffs  SearchDiffData `json:"sd"`
	HideUnrated  bool           `json:"hu"`
	OnlyCopyable bool           `json:"oc"`
	GamemodeLock bool           `json:"gl"`
	Featured     bool           `json:"f"`
}

func logRequest(r *http.Request) {
	uri := r.RequestURI
	method := r.Method
	fmt.Println(method, uri)
}

func main() {
	ctx := context.Background()
	connString := "postgres://" + os.Getenv("db_username") + ":" + os.Getenv("db_password") + "@" + os.Getenv("db_endpoint") + ":5432/postgres"

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		fmt.Fprintf(w, "You should not be here... %s\n", r.URL.Path)
	})

	http.HandleFunc("/get-level", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		logRequest(r)

		var data GetLevelRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Secret != os.Getenv("db_password") {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		query := "SELECT * FROM public.levels WHERE id = " + strconv.Itoa(data.Id)

		var levels []*Level
		err = pgxscan.Select(ctx, db, &levels, query)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			http.Error(w, "Query failed", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(levels)
		if err != nil {
			fmt.Printf("Failed to marshal levels into JSON: %v\n", err)
			http.Error(w, "Failed to marshal levels into JSON", http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})

	http.HandleFunc("/recent-tab", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		logRequest(r)

		var data RecentTabRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Secret != os.Getenv("db_password") {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		query := "SELECT * FROM public.levels WHERE id <= " + strconv.Itoa(data.MaxId)
		if data.Rated {
			query += " AND difficulty > 0"
		}
		query += " ORDER BY id DESC LIMIT 10"

		var levels []*Level
		err = pgxscan.Select(ctx, db, &levels, query)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			http.Error(w, "Query failed", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(levels)
		if err != nil {
			fmt.Printf("Failed to marshal levels into JSON: %v\n", err)
			http.Error(w, "Failed to marshal levels into JSON", http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})

	http.HandleFunc("/last-level", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		logRequest(r)

		var data BasicRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Secret != os.Getenv("db_password") {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		query := "SELECT * FROM public.levels ORDER BY id DESC LIMIT 1"

		var levels []*Level
		err = pgxscan.Select(ctx, db, &levels, query)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			http.Error(w, "Query failed", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(levels)
		if err != nil {
			fmt.Printf("Failed to marshal levels into JSON: %v\n", err)
			http.Error(w, "Failed to marshal levels into JSON", http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})

	http.HandleFunc("/featured-tab", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		logRequest(r)

		var data FeaturedTabRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Secret != os.Getenv("db_password") {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		query := "SELECT * FROM public.levels WHERE id <= " + strconv.Itoa(data.MaxId) + " AND featured IS TRUE ORDER BY id DESC LIMIT 10"

		var levels []*Level
		err = pgxscan.Select(ctx, db, &levels, query)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			http.Error(w, "Query failed", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(levels)
		if err != nil {
			fmt.Printf("Failed to marshal levels into JSON: %v\n", err)
			http.Error(w, "Failed to marshal levels into JSON", http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})
	
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		logRequest(r)

		var data SearchRequest
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Secret != os.Getenv("db_password") {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		query := "SELECT * FROM public.levels WHERE"

		
		switch searchtype := data.SearchType; searchtype {
		case 2:
			query += " owner = " + data.Search
		case 3:
			query += " id = " + data.Search
		case 1:
			query += " title ILIKE '%" + data.Search + "%'"
		}

		diffs := reflect.ValueOf(data.SearchDiffs)
		for i := 0; i < diffs.NumField(); i++ {
			if diffs.Field(i).Bool() {
				query += " AND ABS(difficulty) != " + strconv.Itoa(i)
			}
		}

		if data.HideUnrated {
			query += " AND difficulty > 0"
		}
		if data.OnlyCopyable {
			query += " AND copylock = false"
		}
		if data.GamemodeLock {
			query += " AND platformer = true"
		}
		if data.Featured {
			query += " AND featured = true"
		}
		
		switch sort := data.SearchSort; sort {
		case 1:
			query += " ORDER BY rating ASC, id"
		case 2:
			query += " ORDER BY rating DESC, id"
		case 3:
			query += " ORDER BY id ASC"
		default:
			query += " ORDER BY id DESC"
		}
		
		query += " OFFSET " + strconv.Itoa(data.Offset) + " LIMIT 10;"

		var levels []*Level
		err = pgxscan.Select(ctx, db, &levels, query)
		if err != nil {
			fmt.Printf("Query failed: %v\n", err)
			http.Error(w, "Query failed", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(levels)
		if err != nil {
			fmt.Printf("Failed to marshal levels into JSON: %v\n", err)
			http.Error(w, "Failed to marshal levels into JSON", http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	for _, encodedRoute := range strings.Split(os.Getenv("ROUTES"), ",") {
		if encodedRoute == "" {
			continue
		}
		pathAndBody := strings.SplitN(encodedRoute, "=", 2)
		path, body := pathAndBody[0], pathAndBody[1]
		http.HandleFunc("/"+path, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		})
	}

	bindAddr := fmt.Sprintf(":%s", port)
	lines := strings.Split(startupMessage, "\n")
	fmt.Println()
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Printf("==> Server listening at %s 🚀\n", bindAddr)

	if err := http.ListenAndServe(bindAddr, nil); err != nil {
		panic(err)
	}
}
