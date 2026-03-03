package govee

import "fmt"

// APIError is returned when the Govee API responds with a non-200 status code.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("govee API error %d: %s", e.Code, e.Message)
}
