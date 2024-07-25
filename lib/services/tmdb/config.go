package tmdb

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiKey     string
	BaseURL string
	ImagesBaseURL string
}

var tmdbAPI = initConfig()

func initConfig() Config {
	godotenv.Load()
	return Config{
		ApiKey:     os.Getenv("TMDB_API_KEY"),
		BaseURL: "https://api.themoviedb.org/3",
		ImagesBaseURL: "https://image.tmdb.org/t/p/w300",
	}
}