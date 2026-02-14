package storage

type SaveWorkers struct {
	Assignments   []OperationWorkers `json:"assignments"`
	UpdateStatus  string             `json:"update_status"`
	ReadyDate     string             `json:"ready_date"`
	RootProductID int64              `json:"root_product_id"`
}

type OperationWorkers struct {
	ProductID     int64   `json:"product_id"`
	OperationName string  `json:"operation_name"`
	EmployeeID    int64   `json:"employee_id"`
	ActualMinutes float64 `json:"actual_minutes"`
	Notes         string  `json:"notes,omitempty"`
	ActualValue   float64 `json:"actual_value"`
}

type GetWorkers struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
