package mangadexapi

import "fmt"

type ErrorDetail struct {
	ID      string `json:"id"`
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Context string `json:"context"`
}

type ErrorResponse struct {
	Result string        `json:"result"`
	Errors []ErrorDetail `json:"errors"`
}

func (e *ErrorResponse) Error() string {
	errorMsg := fmt.Sprintf("result: %s ; errors: [", e.Result)
	for i, err := range e.Errors {
		errorMsg += fmt.Sprintf("{id: %s, status: %d, title: %s, detail: %s, context: %s}",
			err.ID, err.Status, err.Title, err.Detail, err.Context)
		if i < len(e.Errors)-1 {
			errorMsg += ", "
		}
	}
	errorMsg += "]"

	return errorMsg
}
