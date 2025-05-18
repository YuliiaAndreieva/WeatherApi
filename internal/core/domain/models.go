package domain

type Weather struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

type Subscription struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	City        string `json:"city"`
	Frequency   string `json:"frequency"`
	Token       string `json:"token"`
	IsConfirmed bool   `json:"is_confirmed"`
}
