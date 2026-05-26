package domain

import (
	"time"

	"github.com/google/uuid"
)

type SurgicalSchedule struct {
	ID                    uuid.UUID `json:"id"`
	PatientID             uuid.UUID `json:"patient_id"`
	PatientName           string    `json:"patient_name,omitempty"` // populated on read
	ProcedureType         string    `json:"procedure_type"`
	ScheduledAt           time.Time `json:"scheduled_at"`
	PreSurgicalDiagnosis  string    `json:"pre_surgical_diagnosis"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
