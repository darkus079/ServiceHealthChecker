package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"servicehealthchecker/internal/handler"
	"servicehealthchecker/internal/pdf"
	"servicehealthchecker/internal/service"
	"servicehealthchecker/internal/storage"
)

func main() {
	store, err := storage.New("data/storage.json", "data/pending.json")
	if err != nil {
		log.Fatal(err)
	}

	checker := service.NewChecker(store)
	pdfGen := pdf.NewGenerator()

	processPendingTasks(checker, store)

	var isShutdown atomic.Bool
	h := handler.New(checker, store, pdfGen, &isShutdown)

	mux := http.NewServeMux()
	mux.HandleFunc("/check", h.CheckLinks)
	mux.HandleFunc("/report", h.GetReport)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	isShutdown.Store(true)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func processPendingTasks(checker *service.Checker, store *storage.Storage) {
	tasks := store.GetPendingTasks()
	if len(tasks) == 0 {
		return
	}

	log.Printf("Processing %d pending tasks", len(tasks))

	for _, task := range tasks {
		if err := checker.ProcessPendingTask(task.ID, task.Links); err != nil {
			log.Printf("Failed to process pending task %d: %v", task.ID, err)
			continue
		}
		store.RemovePendingTask(task.ID)
	}

	log.Println("Pending tasks processed")
}

