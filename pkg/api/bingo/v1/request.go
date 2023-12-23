package v1

type ListResponse struct {
	Total int64 `json:"total"`
	Data  any   `json:"data"`
}
