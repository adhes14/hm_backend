CREATE TYPE staff_role_enum AS ENUM ('health_staff', 'admin');

CREATE TABLE staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(100) NOT NULL,
    role staff_role_enum NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE admission_staff (
    id BIGSERIAL PRIMARY KEY,
    admission_id UUID REFERENCES admissions(id) ON DELETE CASCADE,
    staff_id UUID REFERENCES staff(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default admin
INSERT INTO staff (full_name, role, username, password_hash, is_active)
VALUES ('Administrador del Sistema', 'admin', 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', true);

-- Add foreign key to clinical_logs.created_by
ALTER TABLE clinical_logs ADD CONSTRAINT fk_clinical_logs_created_by FOREIGN KEY (created_by) REFERENCES staff(id);
