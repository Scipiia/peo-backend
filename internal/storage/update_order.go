package storage

import "time"

type UpdateOrderDetails struct {
	ID              int64           `json:"id"`
	OrderNum        string          `json:"order_num"`
	Name            string          `json:"name"`
	Count           float64         `json:"count"`
	TotalTime       float64         `json:"total_time"`
	CreatedAT       time.Time       `json:"created_at"`
	UpdatedAT       time.Time       `json:"updated_at"`
	Operations      []NormOperation `json:"operations"`
	Type            string          `json:"type"`
	PartType        string          `json:"part_type"`
	ParentAssembly  string          `json:"parent_assembly"`
	ParentProductID *int64          `json:"parent_product_id"`
	Status          *string         `json:"status"`
}

type UpdateFinalOrderDetails struct {
	Brigade        *string  `json:"brigade"`
	NormMoney      *float64 `json:"norm_money"`
	ParentAssembly *string  `json:"parent_assembly"`
	Profile        *string  `json:"profile"`
	Sqr            *float64 `json:"sqr"`
	Systema        *string  `json:"systema"`
	TypeIzd        *string  `json:"type_izd"`
	CustomerType   *string  `json:"customer_type"`
	Coefficient    *float64 `json:"coefficient"`
	ID             int64    `json:"id"`
}
