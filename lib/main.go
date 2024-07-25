package main

import (
	mysql "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/mysql"
	stripepayment "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/stripe_payment"
	tmdb "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/tmdb"
	youtube "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/youtube"
)

func init() {
	mysql.Init()
	tmdb.Init()
	youtube.Init()
	stripepayment.Init()
}

func main() {
	server := NewServer("127.0.0.1:8000")
	server.Serve()
}
