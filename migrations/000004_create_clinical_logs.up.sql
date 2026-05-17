CREATE TABLE clinical_logs (
    id            BIGSERIAL PRIMARY KEY,
    admission_id  UUID NOT NULL REFERENCES admissions(id),
    created_by    UUID NULL,  -- FK deferred to Phase 3
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    pa_systolic   SMALLINT NOT NULL,
    pa_diastolic  SMALLINT NOT NULL,
    heart_rate    SMALLINT NOT NULL,
    resp_rate     SMALLINT NOT NULL,
    temperature   NUMERIC(3,1) NOT NULL,
    spo2          SMALLINT NOT NULL,
    pinard_status BOOLEAN NOT NULL DEFAULT TRUE,
    lochia_type   SMALLINT NOT NULL,  -- 1=Rubra, 2=Serosa, 3=Alba
    lochia_amount SMALLINT NOT NULL,  -- 1=Escaso, 2=Moderado, 3=Abundante
    lochia_odor   BOOLEAN NOT NULL DEFAULT TRUE,  -- true=Normal, false=Fetido
    has_clots     BOOLEAN NOT NULL DEFAULT FALSE,
    notes         VARCHAR(500) NULL
);

CREATE INDEX idx_clinical_logs_admission ON clinical_logs(admission_id);