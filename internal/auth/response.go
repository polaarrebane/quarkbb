package auth

import (
	"encoding/json"
	"net/http"

	v "codeberg.org/ronia/quarkbb/internal/validator"
)

type response struct {
	data   any
	err    error
	status int
	detail any
}

func (r response) Send(w http.ResponseWriter) {
	if r.err != nil {
		writeError(w, r.status, r.err.Error(), r.detail)
	} else {
		writeJSON(w, r.status, r.data)
	}
}

// func newOkResponse(data any) *response {
// 	return &response{
// 		data:   data,
// 		err:    nil,
// 		status: http.StatusOK,
// 	}
// }

func newCreatedResponse(data any) *response {
	return &response{
		data:   data,
		err:    nil,
		status: http.StatusCreated,
	}
}

func newMalformedBodyResponse() *response {
	return &response{
		data:   nil,
		err:    errMalformedBody,
		status: http.StatusBadRequest,
	}
}

func newInternalErrorResponse() *response {
	return &response{
		data:   nil,
		err:    errInternalError,
		status: http.StatusInternalServerError,
	}
}

func newValidationErrorResponse(ve *v.ValidationErrors) *response {
	return &response{
		data:   nil,
		err:    errValidationError,
		status: http.StatusBadRequest,
		detail: ve,
	}
}

func newRegistrationFailedResponse(err error) *response {
	return &response{
		data:   nil,
		err:    err,
		status: http.StatusInternalServerError,
	}
}

func newUserAlreadyExistsResponse() *response {
	return &response{
		data:   nil,
		err:    errUserAlreadyExists,
		status: http.StatusUnprocessableEntity,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		_ = err // todo: add log
	}
}

func writeError(w http.ResponseWriter, status int, msg string, detail any) {
	r := make(map[string]any)
	r["error"] = msg
	if detail != nil {
		r["detail"] = detail
	}
	writeJSON(w, status, r)
}
