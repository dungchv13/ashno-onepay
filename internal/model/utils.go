package model

type TimeDuration struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

type PaginationReq struct {
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	OrderBy OrderBy `json:"order,omitempty"`
}

type PaginationRes struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type OrderBy struct {
	Field string `json:"field"`
	Order string `json:"order"`
}
