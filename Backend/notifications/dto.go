package notifications

type QueryFilter struct {
	Page   int
	Limit  int
	Offset int
	Unread *bool
}

type ListResponse struct {
	Items []Notification `json:"items"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type CountResponse struct {
	Unread int64 `json:"unread"`
}
