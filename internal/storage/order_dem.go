package storage

type Order struct {
	ID       int    `json:"id"`
	OrderNum string `json:"order_num"`
	Creator  int    `json:"creator"`
	Customer string `json:"customer"`
	DopInfo  string `json:"dop_info"`
	MsNote   string `json:"ms_note"`
}

type OrderDemPrice struct {
	Position     int      `json:"position"`
	Creator      string   `json:"creator"`
	NamePosition string   `json:"name_position"`
	Count        *float64 `json:"count"`
	Image        *string  `json:"image"`
	Color        *string  `json:"color"`
	Sqr          float64  `json:"sqr"`
}

type ResultOrderDetails1 struct {
	Order         *Order           `json:"order_dem_norm"`
	OrderDemPrice []*OrderDemPrice `json:"order_dem_price"`
}

type ResultOrderDetails struct {
	ID           int     `json:"id"`
	NamePosition string  `json:"name_position"`
	Position     string  `json:"position"`
	Sqr          float64 `json:"sqr"`
	OrderNum     string  `json:"order_num"`
	Note         *string `json:"note"`
	Count        int     `json:"count"`
	Image        *string `json:"image"`
	Color        *string `json:"color"`
	Customer     *string `json:"customer"`
}
