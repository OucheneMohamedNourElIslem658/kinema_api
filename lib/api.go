package main

import (
	"fmt"
	"log"
	"net/http"

	authRouters "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/routers"
	moviesRouter "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/movies/routers"
	reservationsRouter "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/reservations/routers"
)

type Server struct {
	address string
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
	}
}

func (server *Server) Serve() {
	mainRouter := http.NewServeMux()

	var subRouter = http.NewServeMux()
	mainRouter.Handle("/api/v1/", http.StripPrefix("/api/v1", subRouter))

	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Write([]byte("Welcome To Your First Golang Big Project!"))
	})

	// Auth router:
	authRouter := authRouters.NewAuthRouter()
	subRouter.Handle("/auth/", http.StripPrefix("/auth", authRouter.Router))
	authRouter.RegisterRouts()

	// Movies router:
	moviesRouter := moviesRouter.NewAuthRouter()
	subRouter.Handle("/movies/", http.StripPrefix("/movies", moviesRouter.Router))
	moviesRouter.RegisterRouts()

	// Reservation Router:
	reservationsRouter := reservationsRouter.NewReservationsRouter()
	subRouter.Handle("/reservations/", http.StripPrefix("/reservations", reservationsRouter.Router))
	reservationsRouter.RegisterRouts()

	// Run server :
	fmt.Println("Server listening on: ", server.address)
	log.Fatal(http.ListenAndServe(server.address, mainRouter).Error())
}
