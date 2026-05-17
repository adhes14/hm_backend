-- Enable trigram extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Enums
CREATE TYPE admission_status_enum AS ENUM ('active', 'discharged');
CREATE TYPE event_type_enum AS ENUM ('parto', 'cesarea', 'ninguno');

-- Bed types catalog
CREATE TABLE bed_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    prefix VARCHAR(10) NOT NULL
);

-- Physical beds
CREATE TABLE beds (
    id SERIAL PRIMARY KEY,
    bed_type_id INT REFERENCES bed_types(id),
    number INT NOT NULL,
    current_admission_id UUID NULL,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT uq_bed_number UNIQUE (number)
);

-- Patients
CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_number VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    birth_date DATE NOT NULL,
    obstetric_history JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- Admissions
CREATE TABLE admissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID REFERENCES patients(id),
    bed_id INT REFERENCES beds(id),
    status admission_status_enum DEFAULT 'active',
    event_type event_type_enum DEFAULT 'ninguno',
    event_at TIMESTAMP NULL,
    next_control_at TIMESTAMP NULL,
    estimated_discharge_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    discharged_at TIMESTAMP NULL
);

-- Indexes
CREATE INDEX idx_beds_current_admission ON beds (current_admission_id);
CREATE INDEX idx_patients_full_name ON patients USING gin (full_name gin_trgm_ops);
CREATE INDEX idx_admissions_bed_id ON admissions (bed_id);
CREATE INDEX idx_admissions_status ON admissions (status);