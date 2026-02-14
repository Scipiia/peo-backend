package storage

import "time"

type OrderNormDetails struct {
	OrderNum        string          `json:"order_num"`
	TemplateCode    string          `json:"template_code"`
	Name            string          `json:"name"`
	Count           float64         `json:"count"`
	TotalTime       float64         `json:"total_time"`
	Operations      []NormOperation `json:"operations"`
	Type            string          `json:"type"`
	PartType        string          `json:"part_type"`
	ParentAssembly  string          `json:"parent_assembly"`
	ParentProductID *int64          `json:"parent_product_id"`
	Customer        string          `json:"customer"`
	Position        int             `json:"position"`
	Status          string          `json:"status"`
	Systema         string          `json:"systema"`
	TypeIzd         string          `json:"type_izd"`
	Profile         string          `json:"profile"`
	Sqr             float64         `json:"sqr"`
}

type NormOperation struct {
	Name            string           `json:"operation_name"`
	Label           string           `json:"operation_label"`
	Count           float64          `json:"count"`
	Value           float64          `json:"value"`
	Minutes         float64          `json:"minutes"`
	AssignedWorkers []AssignedWorker `json:"assign_workers,omitempty"`
}

type AssignedWorker struct {
	EmployeeID    int64   `json:"employee_id"`
	ActualMinutes float64 `json:"actual_minutes"`
	ActualValue   float64 `json:"actual_value"`
}

type GetOrderDetails struct {
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
	TemplateCode    string          `json:"template_code"`
	HeadName        string          `json:"head_name"`
	TypeIzd         string          `json:"type_izd"`
	ReadyDate       *string         `json:"ready_date"`
	Position        int             `json:"position"`
	//AssignWorkers   []AssignedWorkers `json:"assign_workers"`
}
