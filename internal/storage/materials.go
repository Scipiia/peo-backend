package storage

type KlaesMaterials struct {
	ID         int     `json:"id"`
	OrderID    int     `json:"order_id"`
	Position   int     `json:"position"`
	ArticulMat string  `json:"articul_mat"`
	NameMat    string  `json:"name_mat"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Count      float64 `json:"count"`
	Pole       string  `json:"pole"`
}

type DopInfoDemPrice struct {
	NamePosition string  `json:"name_position"`
	ArticulMat   string  `json:"articul_mat"`
	Position     string  `json:"position"`
	Count        float64 `json:"count"`
}
