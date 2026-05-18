ALTER TABLE clinical_logs DROP CONSTRAINT fk_clinical_logs_created_by;

DROP TABLE admission_staff;
DROP TABLE staff;

DROP TYPE staff_role_enum;
