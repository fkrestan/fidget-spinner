package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/crypto/scrypt"
)

type SpinHandler struct {
	L *zap.SugaredLogger
}

func (h *SpinHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	speed := r.URL.Query().Get("speed")
	if speed == "" {
		speed = "15"
	}
	s, err := strconv.Atoi(speed)
	if err != nil || s < 1 {
		http.Error(w, "Invalid spin speed", http.StatusBadRequest)
		return
	}

	// We don't care about security
	salt := []byte{0x4f, 0xe5, 0x3d, 0xa6, 0x5e, 0x97, 0x5c, 0x50}
	// Recommended values. See:
	// https://pkg.go.dev/golang.org/x/crypto@v0.8.0/scrypt#Key
	dk, err := scrypt.Key([]byte("some password"), salt, 1<<s, 8, 1, 32)
	if err != nil {
		h.L.Error(err)
		http.Error(w, fmt.Sprintf("Spin error: %s", err), http.StatusBadRequest)
	}
	// Make sure the hashing doesn't get optimized out
	h.L.Debug("derived scrypt key:", dk)

	w.Write([]byte(fmt.Sprintf("Weeeeeeeeeeeeeeee\n")))
}
