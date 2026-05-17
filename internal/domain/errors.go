package domain

import "errors"

var (
    ErrBedNotFound       = errors.New("bed not found")
    ErrBedNotAvailable   = errors.New("bed is not available")
    ErrPatientNotFound   = errors.New("patient not found")
    ErrPatientExists     = errors.New("patient with this identity number already exists")
    ErrAdmissionNotFound = errors.New("admission not found")
    ErrAdmissionNotActive = errors.New("admission is not active")
)
