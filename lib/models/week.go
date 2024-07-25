package models

import "time"

type Week struct {
	FromDate time.Time `json:"fromDate"`
	ToDate   time.Time `json:"toDate"`
	Format   string    `json:"fromat"`
}
