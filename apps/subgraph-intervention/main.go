package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/lambda/apps/subgraph-intervention/graph"
	"github.com/lambda/apps/subgraph-intervention/graph/generated"
	"github.com/lambda/internal/db"
	"github.com/lambda/internal/events"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

const defaultPort = "8083"

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbConfig, err := db.NewDBConfig(ctx)
	if err != nil {
		log.Fatalf("failed to initialize database connections: %v", err)
	}
	defer dbConfig.Close()

	streamName := getEnv("KINESIS_STREAM_NAME", "intervention-events")
	eventPublisher := events.NewKinesisEventPublisher(dbConfig.AWS, streamName)

	interventionRepo := repository.NewInterventionRepository(dbConfig.WriteDB)

	interventionService := service.NewInterventionService(interventionRepo, eventPublisher)

	resolver := &graph.Resolver{
		InterventionService: interventionService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
