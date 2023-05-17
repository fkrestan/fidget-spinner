package handler

import (
	"net/http"
)

func Liveness(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
