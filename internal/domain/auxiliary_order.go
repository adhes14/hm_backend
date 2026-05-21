package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderCategory string
type OrderStatus string

const (
	OrderCategoryLaboratorio OrderCategory = "laboratorio"
	OrderCategoryImagen      OrderCategory = "imagen"
	OrderCategoryProcedimiento OrderCategory = "procedimiento"

	OrderStatusPending  OrderStatus = "pending"
	OrderStatusDone     OrderStatus = "done"
	OrderStatusReported OrderStatus = "reported"
)

type AuxiliaryOrder struct {
	ID            int64         `json:"id"`
	AdmissionID   uuid.UUID     `json:"admission_id"`
	Category      OrderCategory `json:"category"`
	Description   string        `json:"description"`
	Status        OrderStatus   `json:"status"`
	CreatedBy     *uuid.UUID    `json:"created_by"`
	CreatedByName string        `json:"created_by_name,omitempty"` // From JOIN
	UpdatedBy     *uuid.UUID    `json:"updated_by"`
	UpdatedByName string        `json:"updated_by_name,omitempty"` // From JOIN
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	// Optional fields joined for the central view
	PatientName string `json:"patient_name,omitempty"`
	BedNumber   int    `json:"bed_number,omitempty"`
	BedPrefix   string `json:"bed_prefix,omitempty"`
}
