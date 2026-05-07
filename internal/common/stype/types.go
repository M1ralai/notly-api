package stype

type APIResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      any          `json:"data"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Timestamp string       `json:"timestamp"`
}

type ErrorDetail struct {
	Code    string `json:"code,omitempty"`
	Details string `json:"details"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Ad       string `json:"ad"`
	Soyad    string `json:"soyad"`
}
