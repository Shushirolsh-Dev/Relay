package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Shushirolsh-Dev/Relay/internal/httpapi"
	"github.com/Shushirolsh-Dev/Relay/internal/store"
)

func main() {
	port := env("PORT", "8080")
	dbPath := env("DATABASE_PATH", "./data/relay.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil { log.Fatal(err) }
	db, err := store.Open(dbPath)
	if err != nil { log.Fatalf("database: %v", err) }
	defer db.Close()

	handler := httpapi.New(db, httpapi.Config{AdminToken:env("ADMIN_TOKEN","")})
	server := &http.Server{Addr:":"+port, Handler:handler, ReadHeaderTimeout:5*time.Second, ReadTimeout:10*time.Second, WriteTimeout:20*time.Second, IdleTimeout:60*time.Second}
	go func(){ log.Printf("Relay listening on :%s",port); if err:=server.ListenAndServe(); err!=nil && err!=http.ErrServerClosed { log.Fatal(err) } }()
	stop:=make(chan os.Signal,1); signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM); <-stop
	ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second); defer cancel()
	if err:=server.Shutdown(ctx); err!=nil { log.Printf("shutdown: %v",err) }
}
func env(k, fallback string) string { if v:=os.Getenv(k); v!="" { return v }; return fallback }
