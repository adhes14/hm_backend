package domain

import (
	"time"

	"github.com/google/uuid"
)

type ClinicalLog struct {
	ID           int64      `json:"id"`
	AdmissionID  uuid.UUID  `json:"admission_id"`
	CreatedBy    *uuid.UUID `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	PaSystolic   int16      `json:"pa_systolic"`
	PaDiastolic  int16      `json:"pa_diastolic"`
	HeartRate    int16      `json:"heart_rate"`
	RespRate     int16      `json:"resp_rate"`
	Temperature  float32    `json:"temperature"`
	Spo2         int16      `json:"spo2"`
	PinardStatus bool       `json:"pinard_status"`
	LochiaType   int16      `json:"lochia_type"`   // 1=Rubra, 2=Serosa, 3=Alba
	LochiaAmount int16      `json:"lochia_amount"` // 1=Escaso, 2=Moderado, 3=Abundante
	LochiaOdor   bool       `json:"lochia_odor"`    // true=Normal, false=Fetido
	HasClots     bool       `json:"has_clots"`
	Notes        *string    `json:"notes,omitempty"`
}

// Lochia type constants
const (
	LochiaTypeRubra  = 1
	LochiaTypeSerosa = 2
	LochiaTypeAlba   = 3
)

// Lochia amount constants
const (
	LochiaAmountEscaso   = 1
	LochiaAmountModerado = 2
	LochiaAmountAbundante = 3
)

// VitalSignRange defines acceptable range for a vital sign
type VitalSignRange struct {
	Min float64
	Max float64
}

// VitalSignRanges holds all vital sign ranges
type VitalSignRanges struct {
	PaSystolic  VitalSignRange
	PaDiastolic VitalSignRange
	HeartRate   VitalSignRange
	RespRate    VitalSignRange
	Temperature VitalSignRange
	Spo2        VitalSignRange
}

// DefaultVitalSignRanges returns the default vital sign ranges
func DefaultVitalSignRanges() VitalSignRanges {
	return VitalSignRanges{
		PaSystolic:  VitalSignRange{Min: 50, Max: 300},
		PaDiastolic: VitalSignRange{Min: 30, Max: 200},
		HeartRate:   VitalSignRange{Min: 30, Max: 250},
		RespRate:    VitalSignRange{Min: 5, Max: 60},
		Temperature: VitalSignRange{Min: 30.0, Max: 45.0},
		Spo2:        VitalSignRange{Min: 50, Max: 100},
	}
}