package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lambda/apps/subgraph-intervention/graph"
	"github.com/lambda/apps/subgraph-intervention/graph/generated"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

const defaultPort = "8082"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// 1. Database Connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=write_model port=5434 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate the intervention table
	// err = db.AutoMigrate(&domain.Intervention{})
	// if err != nil {
	// 	log.Printf("Warning: failed to auto-migrate interventions table: %v", err)
	// }

	// 2. AWS Config (LocalStack)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	})

	_, err = config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
				SessionToken:    "test",
			}, nil
		})),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// 3. Repositories
	interventionRepo := repository.NewInterventionRepository(db)

	// 4. Services
	interventionService := service.NewInterventionService(interventionRepo)

	// 5. Resolver Injection
	resolver := &graph.Resolver{
		InterventionService: interventionService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
