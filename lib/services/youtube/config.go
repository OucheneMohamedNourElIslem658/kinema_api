package youtube

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey     string
	BaseURL string
}

var tmdbAPI = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		ApiKey:     os.Getenv("YOUTUBE_API_KEY"),
		BaseURL: "https://www.googleapis.com/youtube/v3",
	}
}