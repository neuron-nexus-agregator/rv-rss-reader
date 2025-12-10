package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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
	fmt.Println(os.Getenv("RSSURL"))
}
