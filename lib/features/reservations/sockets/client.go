package reservations

import (
	"encoding/json"
	"errors"
	"log"

	models "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	websocket "github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection  *websocket.Conn
	manager     *SeatChoiceSocketManager
	uid         uint
	diffusionID uint
	totalPrice  float64
	holdedSeats map[uint]*models.Seat
	egress      chan []byte
}

func NewClient(connection *websocket.Conn, manager *SeatChoiceSocketManager, uid uint, diffusionID uint) *Client {
	return &Client{
		connection:  connection,
		manager:     manager,
		uid:         uid,
		diffusionID: diffusionID,
		holdedSeats: make(map[uint]*models.Seat),
		egress:      make(chan []byte),
	}
}

func (client *Client) readMessages() {
	defer client.cleanSocket()

	for {
		_, payload, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err.Error())
			}
			break
		}

		for wsClient := range client.manager.clients {
			wsClient.egress <- payload
		}
	}
}

type Event struct {
	Event  string                 `json:"event"`
	Body   map[string]any         `json:"body,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
}

func (client *Client) WriteMessages() {
	//Remove onhold seats at the end:
	defer client.cleanSocket()

	// Get seats list:
	reservationsRepo := client.manager.reservationRepo
	var result map[string]interface{}
	var err1 error

	if len(client.manager.diffusions[client.diffusionID]) == 0 {
		result, err1 = reservationsRepo.GetSeats(uint(client.diffusionID))
		if err1 != nil {
			if err2 := client.connection.WriteMessage(websocket.CloseMessage, []byte(err1.Error())); err2 != nil {
				log.Println("connection closed: ", err2.Error())
			}
			return
		}
		client.manager.diffusions[client.diffusionID] = result["seats"].([]*models.Seat)
	} else {
		seatsList := client.manager.diffusions[client.diffusionID]
		result = map[string]interface{}{
			"count":       len(seatsList),
			"seats":       seatsList,
			"totalPrice":  client.totalPrice,
			"holdedSeats": client.holdedSeats,
		}
	}

	var response Event

	// Send initial result:
	response.Event = "data"
	response.Result = result
	initialResponse, _ := json.MarshalIndent(response, "", "\t")
	if err := client.connection.WriteMessage(websocket.TextMessage, initialResponse); err != nil {
		log.Printf("failed to send message %v", err.Error())
	}

	seats := result["seats"].([]*models.Seat)
	seatPrice := result["seatPrice"].(float64)

	for {
		select {
		case message, ok := <-client.egress:
			if !ok {
				if err := client.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err.Error())
				}
				return
			}

			var request Event
			json.Unmarshal(message, &request)

			// Handle events:
			var err1 error
			switch request.Event {
			case "reserve":
				body := request.Body
				reservationIDFloat, ok := body["reservationID"].(float64)
				if !ok || reservationIDFloat == 0 {
					err1 = errors.New("INVALID_RESERVATION_ID")
					break
				}
				reservationID := uint(reservationIDFloat)
				err1 = client.ReserveSeats(reservationID)
			case "unreserve":
				body := request.Body
				reservationIDFloat, ok := body["reservationID"].(float64)
				if !ok || reservationIDFloat == 0 {
					err1 = errors.New("INVALID_RESERVATION_ID")
					break
				}
				reservationID := uint(reservationIDFloat)
				err1 = client.UnreserveSeats(seats, reservationID)
			case "hold":
				body := request.Body
				seatIDFloat, ok := body["seatID"].(float64)
				if !ok || seatIDFloat == 0 {
					err1 = errors.New("INVALID_SEAT_ID")
					break
				}
				seatID := uint(seatIDFloat)

				var requestedSeat *models.Seat
				for index, seat := range seats {
					if seat.ID == seatID {
						requestedSeat = seats[index]
					}
				}
				if requestedSeat == nil {
					err1 = errors.New("INVALID_SEAT_ID")
					break
				}

				err1 = client.HoldSeat(requestedSeat, seatPrice)
			case "unhold":
				body := request.Body
				seatIDFloat, ok := body["seatID"].(float64)
				seatID := uint(seatIDFloat)
				if !ok || seatID == 0 {
					err1 = errors.New("INVALID_SEAT_ID")
					break
				}

				var requestedSeat *models.Seat
				for index, seat := range seats {
					if seat.ID == seatID {
						requestedSeat = seats[index]
					}
				}
				if requestedSeat == nil {
					err1 = errors.New("INVALID_SEAT_ID")
					break
				}

				err1 = client.Unhold(requestedSeat, seatPrice)
			default:
				err1 = errors.New("INVALID_EVENT")
			}

			result["totalPrice"] = client.totalPrice
			result["holdedSeats"] = client.holdedSeats

			// Send new result:
			if err1 != nil {
				response.Event = "error"
				response.Result = map[string]interface{}{
					"error": err1.Error(),
				}
			} else {
				response.Event = "data"
				response.Result = result
			}

			message, _ = json.MarshalIndent(response, "", "\t")
			if err := client.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("failed to send message %v", err.Error())
			}

		}
	}
}

func (client *Client) cleanSocket() {
	reservationsRepo := client.manager.reservationRepo

	reservationsRepo.ResetSeats(client.diffusionID, client.uid)
	client.manager.removeClient(client)

	// Remove diffusion if all clients were deleted
	isDiffusionAbendend := true
	for managerClient := range client.manager.clients {
		if client.diffusionID == managerClient.diffusionID {
			isDiffusionAbendend = false
			break
		}
	}

	if isDiffusionAbendend {
		client.manager.diffusions[client.diffusionID] = make([]*models.Seat, 0)
	}
}

func (client *Client) Unhold(seat *models.Seat, seatPrice float64) error {
	if seat.Status == "onhold" && *seat.UserID == client.uid {
		seat.Status = "availble"
		seat.UserID = nil
		client.totalPrice -= seatPrice
		delete(client.holdedSeats, seat.ID)
		return nil
	}

	return errors.New("SEAT_ALREADY_ONHOLD")
}

func (client *Client) HoldSeat(seat *models.Seat, seatPrice float64) error {
	if seat.Status == "availble" {
		seat.Status = "onhold"
		seat.UserID = &client.uid
		client.totalPrice += seatPrice
		client.holdedSeats[seat.ID] = seat
		return nil
	}

	return errors.New("SEAT_ALREADY_ONHOLD")
}

func (client *Client) ReserveSeats(reservationID uint) error {
	for _, seat := range client.holdedSeats {
		seat.Status = "reserved"
		seat.UserID = &client.uid
		seat.ReservationID = &reservationID
	}
	client.totalPrice = 0
	client.holdedSeats = make(map[uint]*models.Seat)
	return nil
}

func (client *Client) UnreserveSeats(hallSeats []*models.Seat, reservationID uint) error {
	for _, seat := range hallSeats {
		filterUser := seat.UserID != nil && *(seat.UserID) == client.uid
		filterStatus := seat.Status == "reserved"
		filterReservation := seat.ReservationID != nil && reservationID == *seat.ReservationID
		if filterUser && filterStatus && filterReservation {
			seat.Status = "availble"
			seat.UserID = nil
			seat.ReservationID = nil
		}
	}
	return nil
}
