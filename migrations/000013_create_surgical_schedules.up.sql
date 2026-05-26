CREATE TABLE surgical_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID REFERENCES patients(id) ON DELETE CASCADE NOT NULL,
    procedure_type VARCHAR(100) NOT NULL,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    pre_surgical_diagnosis TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Unique constraint so a patient can have at most one pending surgical schedule
CREATE UNIQUE INDEX idx_surgical_schedules_patient ON surgical_schedules (patient_id);
CREATE INDEX idx_surgical_schedules_date ON surgical_schedules (scheduled_at);

-- Seed surgical procedure types
INSERT INTO system_settings (key, value, description) VALUES
('surgical_procedure_types', '["Cesárea", "Legrado", "Laparoscopía", "Colposcopía", "Histeroscopía", "Cirugía Electiva"]', 'Tipos de procedimiento quirúrgico disponibles para programar (formato JSON array)');
