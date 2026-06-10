package domain

import (
    "time"

    "github.com/google/uuid"
)

type AdmissionStatus string

const (
    AdmissionStatusActive     AdmissionStatus = "active"
    AdmissionStatusDischarged AdmissionStatus = "discharged"
)

type EventType string

const (
    EventTypeParto    EventType = "parto"
    EventTypeCesarea  EventType = "cesarea"
    EventTypeNinguno  EventType = "ninguno"
)

type Admission struct {
    ID                   uuid.UUID       `json:"id"`
    PatientID            uuid.UUID       `json:"patient_id"`
    BedID                int             `json:"bed_id"`
    Status               AdmissionStatus `json:"status"`
    EventType            EventType       `json:"event_type"`
    EventAt              *time.Time      `json:"event_at"`
    NextControlAt        *time.Time      `json:"next_control_at"`
    EstimatedDischargeAt *time.Time      `json:"estimated_discharge_at"`
    CreatedAt            time.Time       `json:"created_at"`
    DischargedAt         *time.Time      `json:"discharged_at"`
    AdmissionDiagnosis           string     `json:"admission_diagnosis"`
    CurrentDiagnosis             string     `json:"current_diagnosis"`
    Treatment                    string     `json:"treatment"`
    CurrentDiagnosisUpdatedBy    *uuid.UUID `json:"current_diagnosis_updated_by"`
    CurrentDiagnosisUpdatedByName string     `json:"current_diagnosis_updated_by_name"`
}

type AdmissionWithDetails struct {
    Admission
    PatientName      string `json:"patient_name"`
    PatientDNI       string `json:"patient_dni"`
    ClinicalLogCount int    `json:"clinical_log_count"`
}
