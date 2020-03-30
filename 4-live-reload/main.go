package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func calcUpdateDuration() time.Duration {
	if StartTime.IsZero() {
		return 0
	}
	return time.Since(StartTime)
}

type ExampleRouter struct {
	*mux.Router
	tmpl           *template.Template
	updateDuration time.Duration
	shuttingDown   bool
	buildID        string
}

func NewExampleRouter() (*ExampleRouter, error) {
	r := mux.NewRouter()

	tmpl, err := template.ParseGlob("./web/templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	updateDuration := calcUpdateDuration()
	router := &ExampleRouter{
		Router:         r,
		tmpl:           tmpl,
		updateDuration: updateDuration,
		shuttingDown:   false,
		buildID:        RandStringRunes(16),
	}

	fs := http.FileServer(http.Dir("./web"))
	r.HandleFunc("/", router.index)
	if os.Getenv("ENVIRONMENT") == "development" {
		r.HandleFunc("/live-reload", router.livereload)
	}
	r.PathPrefix("/").Handler(fs)

	return router, nil
}

func (r *ExampleRouter) updateTimeDisplay() string {
	if r.updateDuration != 0 {
		return r.updateDuration.Truncate(100 * time.Millisecond).String()
	}
	return "N/A"
}

func (r *ExampleRouter) index(w http.ResponseWriter, req *http.Request) {
	congrats := r.updateDuration != 0 && r.updateDuration < 2*time.Second
	err := r.tmpl.ExecuteTemplate(w, "index.tmpl", map[string]interface{}{
		"Duration":     r.updateTimeDisplay(),
		"Congrats":     congrats,
		"LiveReload":   os.Getenv("ENVIRONMENT") == "development",
		"CurrentBuild": r.buildID,
	})
	if err != nil {
		log.Printf("index: %v")
	}
}

func hashFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String := hex.EncodeToString(hashInBytes)

	return returnMD5String, nil
}

func (r *ExampleRouter) livereload(w http.ResponseWriter, req *http.Request) {
	// build := strconv.Itoa(os.Getpid())
	// build, err := hashFileMD5(os.Args[0])
	// if err != nil {
	// 	log.Fatal("calculate build hash for live reload: %v", err)
	// }
	build := r.buildID

	deployedBuild := req.URL.Query().Get("currentBuild")

	//	if deployedBuild == "" {
	//		w.Write([]byte(fmt.Sprintf("{\"reload\":false, \"build\":\"%s\"}", build)))
	//		return
	//	}

	if deployedBuild != build {
		w.Write([]byte(fmt.Sprintf("{\"reload\":true, \"build\":\"%s\"}", build)))
		return
	}

	for i := 0; i < 120; i++ {
		if r.shuttingDown {
			w.Write([]byte(fmt.Sprintf("{\"reload\":true, \"build\":\"\"}", build)))
		}
		time.Sleep(500 * time.Millisecond)
	}

	w.Write([]byte(fmt.Sprintf("{\"reload\":false, \"build\":\"%s\"}", build)))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	router, err := NewExampleRouter()
	if err != nil {
		log.Fatalf("Router creation failed: %v", err)
	}

	srv := &http.Server{Addr: ":8000", Handler: router.Router}

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	go func() {
		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown: %v", err)
		}
	}()

	srv.RegisterOnShutdown(func() {
		router.shuttingDown = true
	})

	log.Println("Serving on port: 8000")
	log.Printf("Deploy time: %s\n", router.updateTimeDisplay())
	err = srv.ListenAndServe()
	if err != nil {
		log.Printf("Server exited with: %v", err)
	}
}
