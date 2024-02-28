package encrypt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if MetricsDecryptor != nil {
			buf, _ := io.ReadAll(r.Body)

			message, err := MetricsDecryptor.Decrypt(buf)
			if err != nil {
				fmt.Printf("DECRYPT REQUEST BODY ERROR: %s", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(message))
		}

		h.ServeHTTP(w, r)
	})
}
