package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	WriteDB *gorm.DB
	ReadDB  *gorm.DB
	Redis   *redis.Client
	AWS     aws.Config
}

func NewDBConfig(ctx context.Context) (*DBConfig, error) {
	writeDB, err := connectWriteDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to write DB: %w", err)
	}

	readDB, err := connectReadDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to read DB: %w", err)
	}

	redisClient, err := connectRedis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	awsConfig, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &DBConfig{
		WriteDB: writeDB,
		ReadDB:  readDB,
		Redis:   redisClient,
		AWS:     awsConfig,
	}, nil
}

func connectWriteDB() (*gorm.DB, error) {
	dsn := os.Getenv("WRITE_DB_URL")
	if dsn == "" {
		host := getEnv("WRITE_DB_HOST", "localhost")
		port := getEnv("WRITE_DB_PORT", "5434")
		user := getEnv("WRITE_DB_USER", "postgres")
		password := getEnv("WRITE_DB_PASSWORD", "postgres")
		dbname := getEnv("WRITE_DB_NAME", "write_model")
		sslmode := getEnv("WRITE_DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to Write DB successfully")
	return db, nil
}

func connectReadDB() (*gorm.DB, error) {
	dsn := os.Getenv("READ_DB_URL")
	if dsn == "" {
		host := getEnv("READ_DB_HOST", "localhost")
		port := getEnv("READ_DB_PORT", "5433")
		user := getEnv("READ_DB_USER", "postgres")
		password := getEnv("READ_DB_PASSWORD", "postgres")
		dbname := getEnv("READ_DB_NAME", "read_model")
		sslmode := getEnv("READ_DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to Read DB successfully")
	return db, nil
}

func connectRedis(ctx context.Context) (*redis.Client, error) {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6380")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis successfully")
	return client, nil
}

func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	localstackURL := getEnv("LOCALSTACK_URL", "http://localhost:4566")
	region := getEnv("AWS_REGION", "us-east-1")

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, r string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           localstackURL,
			SigningRegion: region,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "test"),
				SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "test"),
				SessionToken:    getEnv("AWS_SESSION_TOKEN", "test"),
			}, nil
		})),
	)
	if err != nil {
		return aws.Config{}, err
	}

	log.Println("AWS Config loaded successfully")
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *DBConfig) Close() error {
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			return fmt.Errorf("failed to close Redis: %w", err)
		}
	}

	if c.WriteDB != nil {
		sqlDB, err := c.WriteDB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("failed to close Write DB: %w", err)
			}
		}
	}

	if c.ReadDB != nil {
		sqlDB, err := c.ReadDB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("failed to close Read DB: %w", err)
			}
		}
	}

	log.Println("All database connections closed")
	return nil
}
