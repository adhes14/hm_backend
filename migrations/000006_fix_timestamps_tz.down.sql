-- Alter admissions TIMESTAMPTZ columns back to TIMESTAMP
ALTER TABLE admissions 
    ALTER COLUMN event_at TYPE TIMESTAMP,
    ALTER COLUMN next_control_at TYPE TIMESTAMP,
    ALTER COLUMN estimated_discharge_at TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN discharged_at TYPE TIMESTAMP;

-- Alter clinical_logs TIMESTAMPTZ columns back to TIMESTAMP
ALTER TABLE clinical_logs
    ALTER COLUMN created_at TYPE TIMESTAMP;
