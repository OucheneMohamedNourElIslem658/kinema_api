package movies

import (
	"encoding/json"
	"net/http"
	"time"

	moviesRepo "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/movies/repositories"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
)

type MoviesController struct {
	moviesRepo *moviesRepo.MoviesRepo
}

func Newcontroller() *MoviesController {
	return &MoviesController{
		moviesRepo: moviesRepo.NewMoviesRepository(),
	}
}

func (moviesController *MoviesController) GetMoviesFromTMDB(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	query := queries.Get("query")
	pageString := queries.Get("page")

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMoviesFromTMDB(query, pageString)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) AddMovie(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)

	tmdbID := body["tmdbID"]
	trailerVideoID := body["trailerVideoID"]
	language := body["language"]
	po := body["po"]

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.AddMovie(tmdbID, trailerVideoID, language, po)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetMovieTrailersFromTMDB(w http.ResponseWriter, r *http.Request) {
	tmdbID := r.URL.Query().Get("tmdbID")

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMovieTrailersFromTMDB(tmdbID)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMovie(id)

	if status == http.StatusOK {
		movie := result["movie"].(models.Movie)
		w.WriteHeader(status)
		reponse, _ := json.MarshalIndent(movie, "", "\t")
		w.Write(reponse)
		return
	}

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetMovies(w http.ResponseWriter, r *http.Request) {
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMovies()

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	json.NewDecoder(r.Body).Decode(&movie)

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.UpdateMovie(movie)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) AddDiffusion(w http.ResponseWriter, r *http.Request) {
	var diffusion models.Diffusion
	json.NewDecoder(r.Body).Decode(&diffusion)

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.AddDiffusion(diffusion)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) AddHall(w http.ResponseWriter, r *http.Request) {
	var hall models.Hall
	json.NewDecoder(r.Body).Decode(&hall)

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.AddHall(hall)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetHalls(w http.ResponseWriter, r *http.Request) {
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetHalls()

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.DeleteMovie(id)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetAllWeeksUntilNextYear(w http.ResponseWriter, r *http.Request) {
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetAllWeeksUntilNextYear()

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetDiffusionsForAdmin(w http.ResponseWriter, r *http.Request) {
	type DiffusionFilter struct {
		FromDate time.Time `json:"fromDate"`
		ToDate   time.Time `json:"toDate"`
		HallID   uint      `json:"hallID"`
	}
	var diffusionFilter DiffusionFilter
	json.NewDecoder(r.Body).Decode(&diffusionFilter)

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetDiffusionsForAdmin(
		diffusionFilter.FromDate,
		diffusionFilter.ToDate,
		diffusionFilter.HallID,
	)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) DeleteDiffusion(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.DeleteDiffusion(id)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetTopDiffusion(w http.ResponseWriter, r *http.Request) {
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetTopDiffusion()

	if status == http.StatusOK {
		diffusion := result["diffusion"].(models.Diffusion)
		w.WriteHeader(status)
		reponse, _ := json.MarshalIndent(diffusion, "", "\t")
		w.Write(reponse)
		return
	}

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetDiffusionsByDay(w http.ResponseWriter, r *http.Request) {
	type DiffusionFilter struct {
		Day time.Time `json:"day"`
	}
	var diffusionFilter DiffusionFilter
	json.NewDecoder(r.Body).Decode(&diffusionFilter)

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetDiffusionsByDay(
		diffusionFilter.Day,
	)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetMostPopularDiffusionsTrailers(w http.ResponseWriter, r *http.Request) {
	trailersCount := r.URL.Query().Get("count")

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMostPopularDiffusionsTrailers(trailersCount)

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetDiffusionsForUsers(w http.ResponseWriter, r *http.Request) {
	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetDiffusionsForUsers()

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}

func (moviesController *MoviesController) GetMoviesDiffusions(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	moviesRepo := moviesController.moviesRepo
	status, result := moviesRepo.GetMoviesDiffusions(id)

	if status == http.StatusOK {
		movie := result["movie"].(models.Movie)
		w.WriteHeader(status)
		reponse, _ := json.MarshalIndent(movie, "", "\t")
		w.Write(reponse)
		return
	}

	w.WriteHeader(status)
	reponse, _ := json.MarshalIndent(result, "", "\t")
	w.Write(reponse)
}
