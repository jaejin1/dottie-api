package service

import (
	"math"
	"testing"
)

func TestHaversineMeters(t *testing.T) {
	tests := []struct {
		name      string
		lat1, lng1, lat2, lng2 float64
		wantApprox float64 // 허용 오차 1%
	}{
		{
			name:       "같은 지점",
			lat1: 37.5665, lng1: 126.978,
			lat2: 37.5665, lng2: 126.978,
			wantApprox: 0,
		},
		{
			name:       "서울 광화문 → 강남역 약 9.7km",
			lat1: 37.5759, lng1: 126.9769, // 광화문
			lat2: 37.4981, lng2: 127.0276, // 강남역
			wantApprox: 9737,              // Haversine 기준 ~9.7km
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := haversineMeters(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			if tt.wantApprox == 0 {
				if got != 0 {
					t.Errorf("want 0, got %f", got)
				}
				return
			}
			diff := math.Abs(got-tt.wantApprox) / tt.wantApprox
			if diff > 0.01 {
				t.Errorf("haversineMeters = %f, want ~%f (diff %.2f%%)", got, tt.wantApprox, diff*100)
			}
		})
	}
}

func TestCalcStats_Empty(t *testing.T) {
	dist, places, dur := calcStats(nil)
	if dist != 0 || places != 0 || dur != 0 {
		t.Errorf("empty dots: want 0,0,0 got %f,%d,%d", dist, places, dur)
	}
}
