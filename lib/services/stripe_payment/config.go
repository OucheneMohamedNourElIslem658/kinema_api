package stripepayment

import (
	"os"
)

type Config struct {
	SecretKey     string
	PublishableKey string
}

var stripeConfig = initConfig()

func initConfig() Config {
	return Config{
		SecretKey: os.Getenv("STRIPE_SECRET_KEY"),
		PublishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
	}
}