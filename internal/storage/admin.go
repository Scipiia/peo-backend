package storage

type CoefficientPEOAdmin struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	Coefficient float64 `json:"coefficient"`
	IsActive    bool    `json:"is_active"`
}

type EmployeesAdmin struct {
	ID       int64        `json:"id"`
	Name     string       `json:"name"`
	IsActive bool         `json:"is_active"`
	Teams    []*TeamAdmin `json:"teams"`
}

type TeamAdmin struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IsActive bool   `json:"is_active"`
}

type UpdateEmployeeInput struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	TeamIDs  []int  `json:"team_ids"`
}

type CreateEmployeeInput struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	TeamIDs  []int  `json:"team_ids"`
}
