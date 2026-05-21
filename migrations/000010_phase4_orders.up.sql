CREATE TYPE order_category_enum AS ENUM ('laboratorio', 'imagen', 'procedimiento');
CREATE TYPE order_status_enum AS ENUM ('pending', 'done', 'reported');

CREATE TABLE auxiliary_orders (
    id BIGSERIAL PRIMARY KEY,
    admission_id UUID REFERENCES admissions(id) ON DELETE CASCADE,
    category order_category_enum NOT NULL,
    description VARCHAR(150) NOT NULL,
    status order_status_enum DEFAULT 'pending',
    created_by UUID REFERENCES staff(id),
    updated_by UUID REFERENCES staff(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_pending ON auxiliary_orders (status) WHERE status = 'pending';
CREATE INDEX idx_orders_admission_id ON auxiliary_orders(admission_id);

INSERT INTO system_settings (key, value, description) VALUES
('sound_alert_patient_admitted', 'false', 'Reproducir sonido al ingresar a un nuevo paciente a una cama'),
('sound_alert_patient_discharged', 'false', 'Reproducir sonido al dar de alta a un paciente y liberar la cama')
ON CONFLICT (key) DO NOTHING;
