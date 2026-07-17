package dto

type APIResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewSuccessResponse(code string, message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(code string, message string) APIResponse {
	return APIResponse{
		Success: false,
		Code:    code,
		Message: message,
		Data:    nil,
	}
}
