// ABOUTME: HTTP to JSON-RPC error code mapping.
// ABOUTME: Provides JSONRPCCode() method for ErrorX.

package errorsx

// httpToJSONRPC maps HTTP status codes to JSON-RPC error codes.
var httpToJSONRPC = map[int]int{
	400: -32602, // Invalid params
	401: -32001, // Unauthenticated
	403: -32003, // Permission denied
	404: -32004, // Not found
	409: -32009, // Conflict
	429: -32029, // Too many requests
	500: -32603, // Internal error
	503: -32053, // Service unavailable
}

// JSONRPCCode returns the JSON-RPC error code for this error.
func (err *ErrorX) JSONRPCCode() int {
	if code, ok := httpToJSONRPC[err.Code]; ok {
		return code
	}
	return -32603 // Default to Internal error
}
