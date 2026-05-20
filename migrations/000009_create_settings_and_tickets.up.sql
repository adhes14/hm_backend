CREATE TABLE system_settings (
    key VARCHAR(50) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    description VARCHAR(255) NULL
);

INSERT INTO system_settings (key, value, description) VALUES
('sound_alert_control_overdue', 'true', 'Reproducir sonido al vencerse un control clínico de vigilancia estrecha'),
('sound_alert_discharge_ready', 'false', 'Reproducir sonido cuando el alta estimada del paciente esté lista');

CREATE TABLE sse_tickets (
    ticket VARCHAR(50) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
