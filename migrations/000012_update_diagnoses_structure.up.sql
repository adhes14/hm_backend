ALTER TABLE admissions RENAME COLUMN discharge_diagnosis TO current_diagnosis;
ALTER TABLE admissions ADD COLUMN current_diagnosis_updated_by UUID REFERENCES staff(id) NULL;
