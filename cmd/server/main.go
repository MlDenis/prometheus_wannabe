package main

import (
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/storage"
	"github.com/MlDenis/prometheus_wannabe/internal/type_converter"
	"net/http"
	"strings"
)

const serverListeningAddress = "localhost:8080"

func main() {

	metricsStorage := storage.NewInMemoryStorage()

	http.HandleFunc("/update/", handleUpdateRequest(metricsStorage))
	http.HandleFunc("/", http.NotFound)

	err := http.ListenAndServe(serverListeningAddress, nil)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func handleUpdateRequest(storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			writeResponse(w, http.StatusUnsupportedMediaType, "Unsupported content type")
			return
		}

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			writeResponse(w, http.StatusNotFound, "404 page not found")
			return
		}

		metricType := parts[2]
		metricName := parts[3]
		stringValue := parts[4]

		if metricName == "" {
			writeResponse(w, http.StatusNotFound, "404 page not found")
			return
		}

		switch metricType {
		case "gauge":
			{
				value, err := type_converter.ToFloat64(stringValue)
				if err != nil {
					writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Value converting fail %v: %v", stringValue, err.Error()))
					return
				}

				storage.AddGaugeMetric(metricName, value)
			}
		case "counter":
			{
				value, err := type_converter.ToInt64(stringValue)
				if err != nil {
					writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Value converting fail %v: %v", stringValue, err.Error()))
					return
				}

				storage.AddCounterMetric(metricName, value)
			}

		default:
			{
				writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Unknown metric type: %v", metricType))
				return
			}
		}

		writeResponse(w, http.StatusOK, "Success")
		fmt.Printf("%v metric was update with value: %v", metricName, stringValue)
	}
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(message))
	if err != nil {
		fmt.Printf("Write response failure: %v", err.Error())
	}
}
