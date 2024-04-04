package models

import "database/sql"

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Metric0 struct {
	ID       int
	MetricID int
	Type     string
	Delta    sql.NullInt64
	Value    sql.NullFloat64
}
