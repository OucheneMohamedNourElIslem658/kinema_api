package tmdb

var Instance Config

func Init() {
	Instance = tmdbAPI
}