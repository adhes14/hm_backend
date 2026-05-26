ALTER TABLE admissions DROP COLUMN IF EXISTS current_diagnosis_updated_by;
ALTER TABLE admissions RENAME COLUMN current_diagnosis TO discharge_diagnosis;
