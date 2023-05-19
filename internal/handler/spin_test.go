package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestSpinHandler(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cases := []struct {
		desc           string
		query          string
		expectedStatus int
		expectedBody   string
	}{
		{"TestSimple", "", http.StatusOK, "Weeeeeeeeeeeeeeee\n"},
		{"TestParams:1", "?speed=1", http.StatusOK, "Weeeeeeeeeeeeeeee\n"},
		{"TestParams:15", "?speed=15", http.StatusOK, "Weeeeeeeeeeeeeeee\n"},
		{"TestParams:16", "?speed=16", http.StatusOK, "Weeeeeeeeeeeeeeee\n"},
		{"TestBadParams:A", "?speed=A", http.StatusBadRequest, "Invalid spin speed\n"},
		{"TestBadParams:-1", "?speed=-1", http.StatusBadRequest, "Invalid spin speed\n"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/"+tc.query, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			th := SpinHandler{L: logger.Sugar()}
			th.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got '%v', want '%v'", status, tc.expectedStatus)
			}
			if !(rr.Body.String() == tc.expectedBody) {
				t.Errorf("handler returned wrong body: got '%v', want '%v'", rr.Body.String(), tc.expectedBody)
			}
		})
	}
}
