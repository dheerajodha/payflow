package main

import (
	"database/sql"
	"errors"

  "github.com/lib/pq"
)

type CustomerStore struct {
  db *sql.DB
}

var ErrCustomerExists = errors.New("customer already exists")

func NewCustomerStore(db *sql.DB) *CustomerStore {
  return &CustomerStore{
    db: db,
  }
}

func (s *CustomerStore) GetAll() ([]Customer, error) {
  query := `SELECT id, name FROM  customers`

  rows, err := s.db.Query(query)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  customers := []Customer{}
  for rows.Next() {
    var c Customer
    if err := rows.Scan(&c.ID, &c.Name); err != nil {
      return nil, err
    }
    customers = append(customers, c)
  }

  return customers, nil
}

func (s *CustomerStore) GetByID(id string) (Customer, error) {
  query := `SELECT id, name FROM customers WHERE id=$1`


  var c Customer
  err := s.db.QueryRow(query, id).Scan(&c.ID, &c.Name)

  if err != nil {
    return Customer{}, err
  }

  return c, nil
}

func (s *CustomerStore) Create(c Customer) error {
  query := `
  INSERT INTO customers (id, name)
  VALUES ($1, $2)
  `

  _, err := s.db.Exec(query, c.ID, c.Name)

  if err != nil {
    if pqErr, ok := err.(*pq.Error); ok {
      if pqErr.Code == "23505" {
        return ErrCustomerExists
      }
    }

    return err
  }

  return nil
}

func (s *CustomerStore) Delete(id string) error {
  query := `DELETE FROM customers WHERE id=$1`

  result, err := s.db.Exec(query, id)
  if err != nil {
    return err
  }

  rowsAffected, err := result.RowsAffected()
  if err != nil {
    return err
  }

  if rowsAffected == 0 {
    return sql.ErrNoRows
  }

  return nil
}

func (s *CustomerStore) CreatePaymentIntent(pi PaymentIntent) error {
  query := `
  INSERT INTO payment_intents
  (id, customer_id, amount, currency, status)
  VALUES ($1, $2, $3, $4, $5)
  `

  _, err := s.db.Exec(
    query,
    pi.ID,
    pi.CustomerID,
    pi.Amount,
    pi.Currency,
    pi.Status,
  )

  return err
}
