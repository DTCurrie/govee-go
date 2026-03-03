package govee

import "fmt"

// APIError is returned when the Govee API responds with a non-200 status code.
type APIError struct {
	Code    int
	Message string
}

// Error implements the error interface, returning a string with the HTTP status
// code and the message from the Govee API response.
func (e *APIError) Error() string {
	return fmt.Sprintf("govee API error %d: %s", e.Code, e.Message)
}
