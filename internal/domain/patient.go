package domain

import (
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type Patient struct {
    ID                 uuid.UUID       `json:"id"`
    IdentityNumber     string          `json:"identity_number"`
    FullName           string         `json:"full_name"`
    BirthDate          time.Time       `json:"birth_date"`
    ObstetricHistory   json.RawMessage `json:"obstetric_history"`
    IsAdmitted         bool            `json:"is_admitted"`
    CurrentAdmissionID *uuid.UUID      `json:"current_admission_id"`
}
