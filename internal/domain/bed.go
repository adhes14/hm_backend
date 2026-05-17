package domain

import "github.com/google/uuid"

type Bed struct {
    ID                 int        `json:"id"`
    BedType            *BedType   `json:"bed_type,omitempty"`
    Number             int        `json:"number"`
    CurrentAdmissionID *uuid.UUID `json:"current_admission_id"`
    IsActive           bool       `json:"is_active"`
}

// IsAvailable returns true if the bed has no active admission
func (b *Bed) IsAvailable() bool {
    return b.CurrentAdmissionID == nil && b.IsActive
}
