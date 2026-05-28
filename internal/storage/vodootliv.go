package storage

type NashchelnikRawData struct {
	LegacyID    int64           `json:"legacy_id"`
	OrderNum    string          `json:"order_num"`
	Customer    string          `json:"customer"`
	Count       float64         `json:"count"`
	Sqr         float64         `json:"sqr"`
	Pgm         float64         `json:"pgm"`
	ExistingOps []NormOperation `json:"existing_ops"` // Только "Разметка", "Резка" и т.д.
}
