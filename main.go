package main

import (
  "log"
  "net/http"

  "github.com/gorilla/mux"
)

func main() {
  db := NewDB()
  store := NewCustomerStore(db)

  handler := NewHandler(store)

  r := mux.NewRouter()

  r.HandleFunc("/customers", handler.GetCustomers).Methods("GET")
  r.HandleFunc("/customers/{id}", handler.GetCustomer).Methods("GET")
  r.HandleFunc("/customers", handler.CreateCustomer).Methods("POST")
  r.HandleFunc("/customers/{id}", handler.DeleteCustomer).Methods("DELETE")
  r.HandleFunc("/payment_intents", handler.CreatePaymentIntent).Methods("POST")
  r.HandleFunc("/payment_intents/{id}/confirm", handler.ConfirmPaymentIntent).Methods("POST")

  log.Println("Server started on :8080")

  log.Fatal(http.ListenAndServe(":8080", r))
}
