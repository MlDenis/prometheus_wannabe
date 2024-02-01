package model

type Metrics struct {
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // a parameter that takes the value gauge or counter
	Delta *int64   `json:"delta,omitempty"` // metric value in case of passing counter
	Value *float64 `json:"value,omitempty"` // metric value in case of passing gauge
	Hash  string   `json:"hash,omitempty"`  // hash value
}
