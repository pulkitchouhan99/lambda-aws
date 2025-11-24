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

	"github.com/lambda/apps/subgraph-auth/graph"
	"github.com/lambda/apps/subgraph-auth/graph/generated"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

const defaultPort = "8081"

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

	// 2. AWS Config for Cognito (LocalStack)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
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
	invitationRepo := repository.NewInvitationRepository(db)
	// userRepo := repository.NewUserRepository(db) // Unused in this main.go, but available

	// 4. Services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "local-secret-key"
	}
	invitationService := service.NewInvitationService(invitationRepo, jwtSecret)

	userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
	if userPoolID == "" {
		log.Println("WARNING: COGNITO_USER_POOL_ID is not set. Auth operations may fail.")
	}

	clientID := os.Getenv("COGNITO_CLIENT_ID")
	if clientID == "" {
		log.Println("WARNING: COGNITO_CLIENT_ID is not set. Auth operations may fail.")
	}

	authService := service.NewAuthService(cfg, userPoolID, clientID)

	// 5. Resolver Injection
	resolver := &graph.Resolver{
		AuthService:       authService,
		InvitationService: invitationService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
