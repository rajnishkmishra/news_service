package main

import (
	"news_service/services/backends"
	"news_service/utils"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	utils.SetupAndRun(backends.PathHandler)
}
