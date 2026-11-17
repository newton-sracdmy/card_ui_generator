package dto

type Pagination struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

type PaginatedResponse[T any] struct {
	Data        []T   `json:"data"`
	TotalCount  int64 `json:"total_count"`
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
}
