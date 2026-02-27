package main

import (
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

  customers := h.store.GetAll()

  json.NewEncoder(w).Encode(customers)
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  id := mux.Vars(r)["id"]

  customer, ok := h.store.GetByID(id)

  if !ok {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  json.NewEncoder(w).Encode(customer)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  var c Customer

  err := json.NewDecoder(r.Body).Decode(&c)
  defer  r.Body.Close()
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  if c.ID == "" || c.Name == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  err = h.store.Create(c)

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
  json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
  id := mux.Vars(r)["id"]

  ok := h.store.Delete(id)

  if !ok {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  w.WriteHeader(http.StatusNoContent)
}
