package main

import (
	"fmt"

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
	for i := range 3 {
		fmt.Println(i)
	}

}
