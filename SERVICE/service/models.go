package SERVICE

type UserBalance struct {
	UserID  int     `json:"user_id"`
	Balance float64 `json:"balance"`
}
type ServiceReport struct {
	ServiceName string  `json:"service_name"`
	TotalAmount float64 `json:"total_amount"`
}

type User struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
}
