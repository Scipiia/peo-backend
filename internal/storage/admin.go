package storage

type CoefficientPEOAdmin struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	Coefficient float64 `json:"coefficient"`
	IsActive    bool    `json:"is_active"`
}

type EmployeesAdmin struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}
