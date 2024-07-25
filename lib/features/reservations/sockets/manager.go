package reservations

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	reservationsRepo "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/reservations/repositories"
	models "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	websocket "github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type SeatChoiceSocketManager struct {
	clients ClientList
	reservationRepo reservationsRepo.ReservationsRepo
	diffusions map[uint][]*models.Seat
	sync.RWMutex
}

func NewSeatChoiceSocketManager() *SeatChoiceSocketManager {
	return &SeatChoiceSocketManager{
		clients: make(ClientList, 0),
		diffusions: make(map[uint][]*models.Seat, 0),
		reservationRepo: *reservationsRepo.NewReservationsRepo(),
	}
}

func (manager *SeatChoiceSocketManager) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New Connection")

	// Get the uid:
	auth, _ := r.Context().Value("auth").(map[string]any)
	uid := uint(auth["id"].(float64))

	// Get diffusionID:
	diffusionIDString := r.URL.Query().Get("diffusionID")
	diffusionID, err := strconv.Atoi(diffusionIDString)
	if err != nil {
		return
	}

	// Upgrade socket:
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	client := NewClient(conn, manager, uint(uid),uint(diffusionID))
	manager.addClient(client)

	// start client processes
	go client.readMessages()
	go client.WriteMessages()
}

func (manager *SeatChoiceSocketManager) addClient(client *Client) {
	manager.Lock()
	defer manager.Unlock()
	manager.clients[client] = true
}

func (manager *SeatChoiceSocketManager) removeClient(client *Client) {
	manager.Lock()
	defer manager.Unlock()
	if _, ok := manager.clients[client]; ok {
		client.connection.Close()
		delete(manager.clients, client)
	}
}
