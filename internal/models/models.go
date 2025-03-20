package models

type StationStats struct {
	Min, Max, Sum float64
	Count         int
}

type StationStatsv3 struct {
	Min, Max, Count int32
	Sum             int64
}
