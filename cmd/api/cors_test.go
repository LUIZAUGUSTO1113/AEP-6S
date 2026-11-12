package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllowVite(t *testing.T) {
	handler := allowVite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	for _, origin := range []string{viteOrigin, "https://other.example", ""} {
		for _, method := range []string{http.MethodOptions, http.MethodGet} {
			request := httptest.NewRequest(method, "/samples", nil)
			request.Header.Set("Origin", origin)
			recorder := httptest.NewRecorder()
			if method == http.MethodOptions {
				request.Header.Set("Access-Control-Request-Method", "PUT")
				request.Header.Set("Access-Control-Request-Headers", "content-type")
			}
			handler.ServeHTTP(recorder, request)
			if method == http.MethodOptions {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			} else {
				require.Equal(t, http.StatusTeapot, recorder.Code)
			}
			header := recorder.Header()
			require.Contains(t, header.Values("Vary"), "Origin")
			require.Empty(t, header.Get("Access-Control-Allow-Credentials"))
			if origin == viteOrigin {
				require.Equal(t, origin, header.Get("Access-Control-Allow-Origin"))
				if method == http.MethodOptions {
					require.Equal(t, "GET, POST, PUT, DELETE", header.Get("Access-Control-Allow-Methods"))
					require.Equal(t, "Content-Type", header.Get("Access-Control-Allow-Headers"))
				}
			} else {
				require.Empty(t, header.Get("Access-Control-Allow-Origin"))
			}
		}
	}
}
