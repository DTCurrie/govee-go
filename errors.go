package govee

import "fmt"

// APIError represents an error response from the Govee API.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("govee: API error %d: %s", e.Code, e.Message)
}
