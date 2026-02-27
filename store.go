package main

import "sync"

type CustomerStore struct {
  mu sync.RWMutex
  customers map[string]Customer
}

func NewCustomerStore() *CustomerStore {
  return &CustomerStore{
    customers: make(map[string]Customer),
  }
}

func (s *CustomerStore) GetAll() []Customer {
  s.mu.RLock()
  defer s.mu.RUnlock()

  result := make([]Customer, 0, len(s.customers))
  for _, c := range s.customers {
    result = append(result, c)
  }

  return result
}

func (s *CustomerStore) GetByID(id string) (Customer, bool) {
  s.mu.RLock()
  defer s.mu.RUnlock()

  c, ok := s.customers[id]
  return c, ok
}

func (s *CustomerStore) Create(c Customer) {
  s.mu.Lock()
  defer s.mu.Unlock()

  s.customers[c.ID] = c
}

func (s *CustomerStore) Delete(id string) bool {
  s.mu.Lock()
  defer s.mu.Unlock()

  if _, ok := s.customers[id]; !ok {
    return false
  }

  delete(s.customers, id)
  return true
}
