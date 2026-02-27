package main

type PaymentIntent struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Amount     int64 `json:"amount"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
}
