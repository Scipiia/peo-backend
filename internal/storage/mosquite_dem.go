package storage

import "time"

type MosquitoOrderDTO struct {
	ID        int64
	OrderID   int64
	OrderNum  string
	ClassID   int
	Status    string
	OrderDate time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	// Операции подтягиваются отдельно
	Operations []MosquitoOpDTO
}

type MosquitoOpDTO struct {
	ID          int64
	OpCode      string
	OpName      string
	NormMinutes int
	NormHours   float64
}
