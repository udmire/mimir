package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func JSONError(logger log.Logger, w http.ResponseWriter, errStr string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{Message: errStr, Code: code}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		level.Error(logger).Log("msg", "failed encoding error message", "message", errStr)
	}
}
