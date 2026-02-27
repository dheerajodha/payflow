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

func (s *CustomerStore) ConfirmPaymentIntent(id string) (PaymentIntent, error) {
  tx, err := s.db.Begin()
  if err != nil {
    return PaymentIntent{}, err
  }

  defer tx.Rollback()

  query := `
  SELECT id, customer_id, amount, currency, status
  FROM payment_intents
  WHERE id=$1
  FOR UPDATE
  `

  var pi PaymentIntent

  err = tx.QueryRow(query, id).Scan(
    &pi.ID,
    &pi.CustomerID,
    &pi.Amount,
    &pi.Currency,
    &pi.Status,
  )

  if err != nil {
    return PaymentIntent{}, err
  }

  if pi.Status != "requires_confirmation" {
    return PaymentIntent{}, errors.New("already confirmed")
  }

  updateQuery := `
  UPDATE payment_intents
  SET status='succeeded'
  WHERE id=$1
  `

  _, err = tx.Exec(updateQuery, id)

  if err != nil {
    return PaymentIntent{}, err
  }

  pi.Status = "succeeded"

  err = tx.Commit()
  if err != nil {
    return PaymentIntent{}, err
  }

  return pi, nil
}

func (s *CustomerStore) CreatePaymentIntentWithIdempotency(key string, pi PaymentIntent) (PaymentIntent, error) {
  tx, err := s.db.Begin()
  if err != nil {
    return PaymentIntent{}, err
  }
  defer tx.Rollback()

  // Step 1: Check if the key exists
  var existingPIID string

  err = tx.QueryRow(
    "SELECT payment_intent_id from idempotency_keys WHERE key=$1",
    key,
  ).Scan(&existingPIID)

  if err == nil {
    // key exists -> return the existing payment intent
    var existingPI PaymentIntent

    err = tx.QueryRow("SELECT id, customer_id, amount, currency, status FROM payment_intents WHERE id=$1", existingPIID).Scan(
      &existingPI.ID,
      &existingPI.CustomerID,
      &existingPI.Amount,
      &existingPI.Currency,
      &existingPI.Status,
    )

    if err != nil {
      return PaymentIntent{}, err
    }

    return existingPI, nil
  }

  // Step 2: create new payment intent

  _, err = tx.Exec(
    `INSERT INTO payment_intents
    (id, customer_id, amount, currency, status)
    VALUES ($1, $2, $3, $4, $5)`,
    pi.ID,
    pi.CustomerID,
    pi.Amount,
    pi.Currency,
    pi.Status,
  )

  if err != nil {
    return PaymentIntent{}, err
  }

  // Store 3: store idempotency key

  _, err = tx.Exec(
    `INSERT INTO idempotency_keys (key, payment_intent_id)
    VALUES ($1, $2)`,
    key,
    pi.ID,
  )

  if err != nil {
    return PaymentIntent{}, err
  }

  err = tx.Commit()
  if err != nil {
    return PaymentIntent{}, err
  }

  return pi, nil
}
