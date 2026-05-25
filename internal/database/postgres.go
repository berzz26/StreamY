package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(databaseUrl string) *Service {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatalf("unable to parse dburl :%v", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("unable to create connection pool: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	return &Service{
		DB: db,
	}
}

func (s *Service) Close() {
	s.DB.Close()
}
