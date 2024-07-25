package movies

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	models "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	mysql "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/mysql"
	tmdb "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/tmdb"
	youtube "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/youtube"
	gorm "gorm.io/gorm"
)

type MoviesRepo struct {
	database   *gorm.DB
	tmdbAPI    tmdb.Config
	youtubeAPI youtube.Config
}

func NewMoviesRepository() *MoviesRepo {
	return &MoviesRepo{
		database:   mysql.Instance,
		tmdbAPI:    tmdb.Instance,
		youtubeAPI: youtube.Instance,
	}
}

func (moviesRepo *MoviesRepo) GetMoviesFromTMDB(query string, pageString string) (int, map[string]interface{}) {

	// Validate inputs:
	if query == "" {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_QUERY",
		}
	}

	page, err := strconv.Atoi(pageString)
	if err != nil || page <= 0 {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_PAGE",
		}
	}

	if page == 0 {
		page = 1
	}

	tmdbAPI := moviesRepo.tmdbAPI

	// Make tmdb request:
	url := fmt.Sprintf(
		"%v/search/movie?query=%v&api_key=%v&page=%v",
		tmdbAPI.BaseURL,
		query,
		tmdbAPI.ApiKey,
		page,
	)
	response, err := http.Get(url)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "QUERING_FAILED",
		}
	}
	defer response.Body.Close()

	var body map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "DECODING_FAILED",
		}
	}

	// Fill the result map
	result := make(map[string]interface{})
	result["page"] = body["page"]
	result["totalPages"] = body["total_pages"]

	APIMovies, ok := body["results"].([]interface{})
	if !ok {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "RESULTS_CAST_FAILED",
		}
	}

	var movies []map[string]interface{}
	for _, apiMovieInterface := range APIMovies {
		APIMovie, ok := apiMovieInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Initialize the movie map
		movie := make(map[string]interface{})
		movie["tmdbID"] = APIMovie["id"]
		movie["title"] = APIMovie["original_title"]

		posterPath, ok := APIMovie["poster_path"].(string)
		if ok {
			movie["picURL"] = tmdbAPI.ImagesBaseURL + posterPath
		}

		movies = append(movies, movie)
	}

	result["movies"] = movies
	return http.StatusOK, result
}

func (moviesRepo *MoviesRepo) GetMovieTrailersFromTMDB(tmdbID string) (int, map[string]interface{}) {
	if tmdbID == "" {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ID",
		}
	}

	tmdbAPI := moviesRepo.tmdbAPI

	// Build the API URL
	url := fmt.Sprintf(
		"%v/movie/%v/videos?api_key=%v",
		tmdbAPI.BaseURL,
		tmdbID,
		tmdbAPI.ApiKey,
	)

	// Make the API request
	response, err := http.Get(url)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "QUERYING_FAILED",
		}
	}
	defer response.Body.Close()

	// Decode the response body
	var body map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "TRAILERS_FETCHING_FAILED",
		}
	}

	result := make(map[string]interface{})

	APITrailers, ok := body["results"].([]interface{})
	if !ok {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "TRAILERS_CAST_FAILED",
		}
	}

	var trailers []map[string]interface{}
	for _, apiTrailerInterface := range APITrailers {
		APITrailer, ok := apiTrailerInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Initialize the movie map:
		trailer := make(map[string]interface{})

		site := APITrailer["site"].(string)
		if site == "YouTube" {
			youtubeKey := APITrailer["key"].(string)
			trailer["url"] = fmt.Sprintf("https://www.youtube.com/watch?v=%v", youtubeKey)
		} else {
			continue
		}

		trailer["title"] = APITrailer["name"]
		trailer["isOfficial"] = APITrailer["official"]

		trailers = append(trailers, trailer)
	}

	result["trailers"] = trailers
	result["count"] = len(trailers)
	return http.StatusOK, result
}

func (moviesRepo *MoviesRepo) AddMovie(tmdbID string, trailerVideoID string, language string, po string) (int, map[string]interface{}) {
	// Validate input arguments
	if tmdbID == "" || trailerVideoID == "" || po == ""{
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ARGS",
		}
	}

	tmdbAPI := moviesRepo.tmdbAPI

	// Build the API URL
	url := fmt.Sprintf(
		"%v/movie/%v?api_key=%v",
		tmdbAPI.BaseURL,
		tmdbID,
		tmdbAPI.ApiKey,
	)

	// Make the API request
	response, err := http.Get(url)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "QUERYING_FAILED",
		}
	}
	defer response.Body.Close()

	// Decode the response body
	var body map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "DECODING_FAILED",
		}
	}

	// Get display language:
	displayLanguage := body["original_language"].(string)

	// Get Trailer:
	trailerURL := fmt.Sprintf("https://www.youtube.com/watch?v=%v", trailerVideoID)
	trailerViews, status, err := moviesRepo.getTrailerViews(trailerVideoID)
	if err != nil {
		return status, map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Create the movie object
	movie := models.Movie{
		Title:        body["original_title"].(string),
		Description:  body["overview"].(string),
		Rate:         body["vote_average"].(float64),
		TrailerURL:   trailerURL,
		TrailerViews: trailerViews,
		Duration:     time.Duration(body["runtime"].(float64) * float64(time.Minute)), // Change uint to float64 for runtime
		VoteCount:    uint(body["vote_count"].(float64)),                              // Change uint to float64 for vote_count
		PicURL:       tmdbAPI.ImagesBaseURL + body["poster_path"].(string),
		PO:           po,
		Language:     displayLanguage,
	}

	database := moviesRepo.database

	// Get and store types
	APITypes, ok := body["genres"].([]interface{})
	if !ok {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "TYPES_CAST_FAILED",
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(APITypes))
	for _, apiTypeInterface := range APITypes {
		APIType, ok := apiTypeInterface.(map[string]interface{})
		if !ok {
			wg.Done()
			continue
		}

		// Add the movie type to the slice
		movieType, ok := APIType["name"].(string)
		if ok {
			go func() {
				defer wg.Done()
				mu.Lock()
				moviesRepo.addMovieType(&movie, movieType)
				mu.Unlock()
			}()
		} else {
			wg.Done()
			continue
		}
	}
	wg.Wait()

	// Get and add Cast:
	cast, err := moviesRepo.getMovieCastFromTMDB(tmdbID)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		}
	}

	wg.Add(len(cast))
	for _, actor := range cast {
		// Add the actor to the slice
		go func() {
			defer wg.Done()
			mu.Lock()
			moviesRepo.addActorToMovie(&movie, actor)
			mu.Unlock()
		}()
	}
	wg.Wait()

	err = database.Create(&movie).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "ERROR_ADDING_MOVIE",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"message": "MOVIE_ADDED",
	}
}

func (moviesRepo *MoviesRepo) getTrailerViews(videoID string) (uint, int, error) {
	type YouTubeResponse struct {
		Items []struct {
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	youtubeAPI := moviesRepo.youtubeAPI
	url := fmt.Sprintf("%v/videos?part=statistics&id=%s&key=%s", youtubeAPI.BaseURL, videoID, youtubeAPI.ApiKey)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return 0, http.StatusInternalServerError, errors.New("FETCHING_TRAILER_DATA_FAILED")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, resp.StatusCode, errors.New("FETCHING_TRAILER_DATA_FAILED")
	}

	var data YouTubeResponse
	json.NewDecoder(resp.Body).Decode(&data)
	if len(data.Items) == 0 {
		return 0, http.StatusInternalServerError, errors.New("TRAILER_DATA_NOT_FOUND")
	}
	viewCountStr := data.Items[0].Statistics.ViewCount
	viewCount, err := strconv.Atoi(viewCountStr)
	if err != nil {
		return 0, http.StatusInternalServerError, errors.New("FORMATING_TRAILER_VIEWS_FAILED_FAILED")
	}
	return uint(viewCount), http.StatusOK, nil
}

func (moviesRepo *MoviesRepo) getMovieCastFromTMDB(tmdbID string) ([]models.Actor, error) {
	tmdbAPI := moviesRepo.tmdbAPI

	// Build the API URL
	url := fmt.Sprintf(
		"%v/movie/%v/credits?api_key=%v",
		tmdbAPI.BaseURL,
		tmdbID,
		tmdbAPI.ApiKey,
	)

	fmt.Println(url)

	// Make the API request
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New("CAST_FETCHING_FAILED")
	}
	defer response.Body.Close()

	// Decode the response body
	var body map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return nil, errors.New("CAST_FETCHING_FAILED")
	}

	APICast, ok := body["cast"].([]interface{})
	if !ok {
		return nil, errors.New("CAST_CAST_FAILED")
	}

	var cast []models.Actor
	for _, apiActorInterface := range APICast {
		APIActor, ok := apiActorInterface.(map[string]interface{})
		if !ok {
			continue
		}

		actorPicURL, ok := APIActor["profile_path"].(string)
		if !ok || actorPicURL == "" {
			actorPicURL = ""
		} else {
			actorPicURL = tmdbAPI.ImagesBaseURL + actorPicURL
		}

		actorName, ok := APIActor["name"].(string)
		if !ok {
			continue
		}

		actor := models.Actor{
			Name:   actorName,
			PicURL: actorPicURL,
		}

		cast = append(cast, actor)
	}

	return cast, nil
}

func (moviesRepo *MoviesRepo) addMovieType(movie *models.Movie, typeName string) error {
	var movieType models.Type
	database := moviesRepo.database
	if err := database.Where("name = ?", typeName).First(&movieType).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			movieType = models.Type{
				Name: typeName,
			}
			if err := database.Create(&movieType).Error; err != nil {
				return errors.New("ADDING_TYPE_FAILED")
			}
		}
	}
	movie.Type = append(movie.Type, movieType)
	return nil
}

func (moviesRepo *MoviesRepo) addActorToMovie(movie *models.Movie, actor models.Actor) error {
	database := moviesRepo.database
	if err := database.Where("name = ?", actor.Name).First(&actor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := database.Create(&actor).Error; err != nil {
				return errors.New("ADDING_ACTOR_FAILED")
			}
		}
	}
	movie.Cast = append(movie.Cast, actor)
	return nil
}

func (moviesRepo *MoviesRepo) GetMovie(ID string) (int, map[string]interface{}) {
	if ID == "" {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ID",
		}
	}

	var movie models.Movie
	database := moviesRepo.database

	err := database.Where("id = ?", ID).First(&movie).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FOUNDING_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"movie": movie,
	}
}

func (moviesRepo *MoviesRepo) GetMovies() (int, map[string]interface{}) {
	var movies []models.Movie
	database := moviesRepo.database

	err := database.Preload("Type").Find(&movies).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FOUNDING_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"count":  len(movies),
		"movies": movies,
	}
}

func (moviesRepo *MoviesRepo) UpdateMovie(newMovie models.Movie) (int, map[string]interface{}) {
	// Validating id:
	if newMovie.ID == 0 {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ID",
		}
	}

	var movie models.Movie
	database := moviesRepo.database

	err := database.Where("id = ?", newMovie.ID).Preload("Type").Preload("Cast").First(&movie).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FOUNDING_MOVIE_FAILED",
		}
	}

	// Updating movie
	if newMovie.Title != "" {
		movie.Title = newMovie.Title
	}
	if newMovie.Description != "" {
		movie.Description = newMovie.Description
	}
	if newMovie.TrailerURL != "" {
		movie.TrailerURL = newMovie.TrailerURL
	}
	if newMovie.PicURL != "" {
		movie.PicURL = newMovie.PicURL
	}
	if newMovie.Language != "" {
		movie.Language = newMovie.Language
	}
	if newMovie.PO != "" {
		movie.PO = newMovie.PO
	}

	err = database.Save(&movie).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "UPDATING_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"message": "MOVIE_UPDATED",
	}
}

func (moviesRepo *MoviesRepo) AddHall(hall models.Hall) (int, map[string]interface{}) {
	if err := hall.ValidateHall(); err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		}
	}

	database := moviesRepo.database

	err := database.Create(&hall).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "HALL_CREATION_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"message": "HALL_ADDED",
	}
}

func (moviesRepo *MoviesRepo) AddDiffusion(diffuion models.Diffusion) (int, map[string]interface{}) {
	err := diffuion.Validate()
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		}
	}

	movieID := diffuion.MovieID
	showTime := diffuion.ShowTime
	showDuration := diffuion.ShowDuration
	hallID := diffuion.HallID
	seatPrice := diffuion.SeatPrice

	database := moviesRepo.database

	var movie models.Movie
	err = database.Where("id = ?", movieID).First(&movie).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "MOVIE_NOT_FOUND",
		}
	}

	var hall models.Hall
	err = database.Where("id = ?", hallID).First(&hall).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "HALL_NOT_FOUND",
		}
	}

	var seats []models.Seat
	for i := 0; i < int(hall.RowsCount); i++ {
		rowLetter := rune('A' + i)
		for j := 0; j < int(hall.ColumnsCount); j++ {
			column := j + 1
			seats = append(seats, models.Seat{
				SeatRow:    string(rowLetter),
				SeatColumn: column,
				Status:     "availble",
			})
		}
	}

	diffusion := models.Diffusion{
		MovieID:      movie.ID,
		ShowTime:     showTime,
		ShowDuration: showDuration,
		HallID:       hallID,
		SeatPrice:    seatPrice,
		SeatsStatus:  seats,
	}

	movie.Diffusions = append(movie.Diffusions, diffusion)

	err = database.Save(&movie).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "ADDING_DIFFUSION_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"message": "DIFFUSION_ADDED",
	}
}

func (moviesRepo *MoviesRepo) GetHalls() (int, map[string]interface{}) {
	database := moviesRepo.database

	var halls []models.Hall
	err := database.Find(&halls).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FETCHING_HALLS_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"count": len(halls),
		"halls": halls,
	}
}

func (moviesRepo *MoviesRepo) DeleteMovie(id string) (int, map[string]interface{}) {
	database := moviesRepo.database
	err := database.Unscoped().Where("id = ?", id).Delete(&models.Movie{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusBadRequest, map[string]interface{}{
				"error": "MOVIE_NOT_FOUND",
			}
		}
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "DELETING_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"error": "MOVIE_DELETED",
	}
}

func (moviesRepo *MoviesRepo) GetAllWeeksUntilNextYear() (int, map[string]interface{}) {
	weeks := getAllWeeksUntilNextYear()
	return http.StatusOK, map[string]interface{}{
		"count": len(weeks),
		"weeks": weeks,
	}
}

func getAllWeeksUntilNextYear() []models.Week {
	now := time.Now()
	nextYearSameDay := time.Date(now.Year()+1, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	weeks := []models.Week{}
	for current := now; current.Before(nextYearSameDay); current = current.AddDate(0, 0, 7) {
		weekStart := current.AddDate(0, 0, -int(current.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 6)
		weeks = append(weeks, models.Week{
			FromDate: weekStart,
			ToDate:   weekEnd,
			Format:   fmt.Sprintf("%02d-%02d %s", weekStart.Day(), weekEnd.Day(), weekEnd.Month()),
		})
	}
	return weeks
}

func (moviesRepo *MoviesRepo) GetDiffusionsForAdmin(startDate time.Time, endDate time.Time, hallID uint) (int, map[string]interface{}) {
	if startDate.Weekday() != time.Sunday {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "START_DATE_MUST_BE_SUNDAY",
		}
	}

	expectedEndDate := startDate.AddDate(0, 0, 6)
	if !endDate.Equal(expectedEndDate) {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "DATE_RANGE_MUST_BE_ONE_WEEK",
		}
	}

	database := moviesRepo.database

	var diffusions []models.Diffusion
	endDate = endDate.Add(24 * time.Hour)
	err := database.Where("show_time between ? AND ? and hall_id = ?", startDate, endDate, hallID).
		Preload("Movie", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title")
		}).Find(&diffusions).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "FETCHING_DIFFUSIONS_FAILED",
		}
	}

	// Group diffusions by day of the week
	groupedDiffusions := make(map[string][]map[string]interface{})
	for _, diffusion := range diffusions {
		dayOfWeek := diffusion.ShowTime.Weekday().String()
		if groupedDiffusions[dayOfWeek] == nil {
			groupedDiffusions[dayOfWeek] = []map[string]interface{}{}
		}

		fromHour := diffusion.ShowTime.Format("15:04:05")
		toHour := diffusion.ShowTime.Add(diffusion.ShowDuration).Format("15:04:05")
		groupedDiffusions[dayOfWeek] = append(groupedDiffusions[dayOfWeek], map[string]interface{}{
			"id":        diffusion.ID,
			"movieID":   diffusion.MovieID,
			"title":     diffusion.Movie.Title,
			"startHour": fromHour,
			"endHour":   toHour,
		})
	}

	return http.StatusOK, map[string]interface{}{
		"count":      len(diffusions),
		"diffusions": groupedDiffusions,
	}
}

func (moviesRepo *MoviesRepo) DeleteDiffusion(id string) (int, map[string]interface{}) {
	database := moviesRepo.database

	err := database.Unscoped().Delete(&models.Diffusion{}, id).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "DELETING_DIFFUSION_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"error": "DIFFUSION_DELETED",
	}
}

func (moviesRepo *MoviesRepo) GetTopDiffusion() (int, map[string]interface{}) {
	database := moviesRepo.database

	var diffusion models.Diffusion
	err := database.Joins("JOIN movies ON movies.id = diffusions.movie_id").
		Order("rate desc").
		Preload("Movie").
		First(&diffusion).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FETCHING_TOP_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"diffusion": diffusion,
	}
}

func (moviesRepo *MoviesRepo) GetDiffusionsByDay(day time.Time) (int, map[string]interface{}) {
	database := moviesRepo.database

	startDate := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	endDate := startDate.AddDate(0, 0, 1)

	var diffusions []models.Diffusion
	err := database.Preload("Movie").Where("show_time between ? and ?", startDate, endDate).Find(&diffusions).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "FETCHING_DIFFUSIONS_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"count":      len(diffusions),
		"diffusions": diffusions,
	}
}

func (moviesRepo *MoviesRepo) GetMostPopularDiffusionsTrailers(trailersCount string) (int, map[string]interface{}) {
	count, err := strconv.Atoi(trailersCount)
	if err != nil || count <= 0 {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_COUNT",
		}
	}

	database := moviesRepo.database

	var diffusions []models.Diffusion
	err = database.Joins("join movies on diffusions.movie_id = movies.id").
		Preload("Movie").
		Order("movies.trailer_views DESC").
		Limit(count).
		Find(&diffusions).
		Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FATCHING_DIFFUSIONS_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"count":      len(diffusions),
		"diffusions": diffusions,
	}
}

func (moviesRepo *MoviesRepo) GetDiffusionsForUsers() (int, map[string]interface{}) {
	database := moviesRepo.database

	var diffusions []models.Diffusion
	err := database.Preload("Movie").Find(&diffusions).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FATCHING_DIFFUSIONS_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"count":      len(diffusions),
		"diffusions": diffusions,
	}
}

func (moviesRepo *MoviesRepo) GetMoviesDiffusions(id string) (int, map[string]interface{}) {
	if id == "" {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ID",
		}
	}

	var movie models.Movie
	database := moviesRepo.database

	err := database.Preload("Type").
		Preload("Cast").
		Preload("Diffusions", func(db *gorm.DB) *gorm.DB {
			return db.Order("diffusions.show_time ASC")
		}).
		Where("id = ?", id).
		First(&movie).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FETCHING_MOVIE_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"movie": movie,
	}
}
