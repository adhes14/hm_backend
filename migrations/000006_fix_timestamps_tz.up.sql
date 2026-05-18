-- Alter admissions TIMESTAMP columns to TIMESTAMPTZ
ALTER TABLE admissions 
    ALTER COLUMN event_at TYPE TIMESTAMPTZ,
    ALTER COLUMN next_control_at TYPE TIMESTAMPTZ,
    ALTER COLUMN estimated_discharge_at TYPE TIMESTAMPTZ,
    ALTER COLUMN created_at TYPE TIMESTAMPTZ,
    ALTER COLUMN discharged_at TYPE TIMESTAMPTZ;

-- Alter clinical_logs TIMESTAMP columns to TIMESTAMPTZ
ALTER TABLE clinical_logs
    ALTER COLUMN created_at TYPE TIMESTAMPTZ;
