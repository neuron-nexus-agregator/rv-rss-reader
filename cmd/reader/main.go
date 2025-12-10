package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	rss "gafarov/rss-reader/internal/core/reader/implementation"
)

func init_system() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file")
	}
	return nil
}

func main() {
	if err := init_system(); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(os.Getenv("RSSURL"))

	reader := rss.New()
	defer reader.Stop()

	// out := reader.Output()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// if err := reader.StartParsing(os.Getenv("RSSURL"), 1*time.Minute, ctx); err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Default().Println("Parsing started")
	// }

	items, err := reader.ParseOnce(os.Getenv("RSSURL"), ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Items read:", len(items))

}
