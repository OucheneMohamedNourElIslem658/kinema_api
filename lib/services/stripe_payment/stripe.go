package stripepayment

import (
	stripe "github.com/stripe/stripe-go/v79"
)

var Instance Config

func Init() {
	Instance = stripeConfig
	stripe.Key = stripeConfig.SecretKey
}
