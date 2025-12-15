package models

import "time"

type AnalyticsPoint struct {
	Date   time.Time `json:"date"`
	Total  float64   `json:"total"`
	Count  int       `json:"count"`
}

type AnalyticsPeriod string

const (
	AnalyticsDay   AnalyticsPeriod = "day"
	AnalyticsWeek  AnalyticsPeriod = "week"
	AnalyticsMonth AnalyticsPeriod = "month"
)
