package main

import (
	"context"
	"fmt"
	
	"time"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/berzz26/StreamY/internal/database"
)

func main() {

	cfg := config.LoadDb()
	db := database.New(cfg.DatabaseUrl)
	
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var version string

	err := db.DB.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		panic(err)

	}

	fmt.Println(version)
}
