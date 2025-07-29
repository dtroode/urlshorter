package response

// GetStats represents service statistics.
// @Description Response structure for service statistics.
type GetStats struct {
	// URLs is the total number of shortened URLs in the service.
	// @Example 100
	URLs int64 `json:"urls" example:"100"`
	// Users is the total number of unique users in the service.
	// @Example 10
	Users int64 `json:"users" example:"10"`
}
