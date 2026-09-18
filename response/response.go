package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ResponseStatus(w http.ResponseWriter, status int, resp interface{}, err error) {
	body := Body{}

	if err != nil {
		body.Code = -1
		body.Message = err.Error()
	} else {
		body.Message = "OK"
		body.Data = resp
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Response(w http.ResponseWriter, resp interface{}, err error) {
	ResponseStatus(w, http.StatusOK, resp, err)
}
