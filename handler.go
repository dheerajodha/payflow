package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
  store *CustomerStore
}

func NewHandler(store *CustomerStore) *Handler {
  return &Handler{store: store}
}

func (h *Handler) GetCustomers(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  customers, err := h.store.GetAll()
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
      "error": err.Error(),
    })
    return
  }

  json.NewEncoder(w).Encode(customers)
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  id := mux.Vars(r)["id"]

  customer, err := h.store.GetByID(id)
  if errors.Is(err, sql.ErrNoRows) {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  json.NewEncoder(w).Encode(customer)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
  defer  r.Body.Close()

  w.Header().Set("Content-Type", "application/json")

  type CreateCustomerRequest struct {
    Name string `json:"name"`
  }

  var req CreateCustomerRequest

  err := json.NewDecoder(r.Body).Decode(&req)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "invalid JSON",
    })
    return
  }

  if req.Name == "" {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{
      "error": "name is required",
    })
    return
  }

  customer := Customer {
    ID: generateID("cust"),
    Name: req.Name,
  }

  err = h.store.Create(customer)

  if err != nil {
    if errors.Is(err, ErrCustomerExists) {
      w.WriteHeader(http.StatusConflict)
      json.NewEncoder(w).Encode(map[string]string{
        "error": err.Error(),
      })

      return
    }

    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(customer)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
  id := mux.Vars(r)["id"]

  err := h.store.Delete(id)

  if errors.Is(err, sql.ErrNoRows) {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  w.WriteHeader(http.StatusNoContent)
}
