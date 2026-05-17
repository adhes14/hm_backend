package service

import (
	"testing"
	"time"
)

// Note: CalculateNextControlAt does NOT check bed type - it only calculates intervals.
// The bed type check (requires_postpartum_followup) happens in CreateClinicalLog BEFORE
// calling this function. So ARO beds never call this function - they skip it entirely.

func TestCalculateNextControlAt_M_Bed_Controls1to4_15Min(t *testing.T) {
	// M bed controls 1-4 should return now + 15 minutes
	now := time.Now()

	testCases := []struct {
		controlCount int
		desc         string
	}{
		{1, "Control #1"},
		{2, "Control #2"},
		{3, "Control #3"},
		{4, "Control #4"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := CalculateNextControlAt(tc.controlCount, now)
			if result == nil {
				t.Fatalf("Expected non-nil for control %d", tc.controlCount)
			}

			expected := now.Add(15 * time.Minute)
			diff := result.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			// Allow 1 second tolerance
			if diff > time.Second {
				t.Errorf("Control %d: expected ~%v, got %v", tc.controlCount, expected, *result)
			}
		})
	}
}

func TestCalculateNextControlAt_M_Bed_Controls5to8_30Min(t *testing.T) {
	// M bed controls 5-8 should return now + 30 minutes
	// Note: control 8 returns nil (process complete after 8 controls)
	now := time.Now()

	testCases := []struct {
		controlCount     int
		desc             string
		expectedDuration time.Duration
		expectedNil      bool
	}{
		{5, "Control #5", 30 * time.Minute, false},
		{6, "Control #6", 30 * time.Minute, false},
		{7, "Control #7", 30 * time.Minute, false},
		{8, "Control #8 (complete)", 0, true}, // After 8 controls, process is complete
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := CalculateNextControlAt(tc.controlCount, now)

			if tc.expectedNil {
				if result != nil {
					t.Errorf("Control %d: expected nil (complete), got %v", tc.controlCount, *result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Control %d: expected non-nil", tc.controlCount)
			}

			expected := now.Add(tc.expectedDuration)
			diff := result.Sub(expected)
			if diff < 0 {
				diff = -diff
			}

			if diff > time.Second {
				t.Errorf("Control %d: expected ~%v, got %v", tc.controlCount, expected, *result)
			}
		})
	}
}

func TestCalculateNextControlAt_M_Bed_Control9OrMore_ReturnsNil(t *testing.T) {
	// Control 9+ returns nil (process complete)
	testCases := []struct {
		controlCount int
		desc         string
	}{
		{9, "Control #9"},
		{10, "Control #10"},
		{15, "Control #15"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := CalculateNextControlAt(tc.controlCount, time.Now())
			if result != nil {
				t.Errorf("Expected nil for control %d (complete), got %v", tc.controlCount, *result)
			}
		})
	}
}