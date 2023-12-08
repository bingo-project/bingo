package v1

type ListRequest struct {
	// Offset
	Offset int `form:"offset"`

	// Limit
	Limit int `form:"limit"`

	// Order by field.
	Order string `form:"order"`

	// Sort: asc or desc.
	Sort string `form:"sort"`
}

type ListResponse struct {
	Total int64 `json:"total"`
	Data  any   `json:"data"`
}
