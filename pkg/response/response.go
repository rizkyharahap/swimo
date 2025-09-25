package response

type Success struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// Pagination represents the pagination metadata.
type Pagination struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	TotalPages int `json:"totalPages" example:"5"`
}

// SuccessPagination is a generic struct for paginated API responses.
// It uses a type parameter 'T' to make the Data field generic.
//
// The 'T' can be any type, and this struct will hold a slice of that type.
type SuccessPagination struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Error struct {
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
}
