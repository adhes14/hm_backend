-- Seed beds for testing
-- ARO (Alto Riesgo Obstétrico): beds 1-6
-- M (Maternidad): beds 7-12

INSERT INTO beds (bed_type_id, number, is_active) VALUES
    -- Alto Riesgo Obstétrico
    (1, 101, TRUE),
    (1, 102, TRUE),
    (1, 103, TRUE),
    (1, 104, TRUE),
    (1, 105, TRUE),
    (1, 106, TRUE),
    -- Maternidad
    (2, 201, TRUE),
    (2, 202, TRUE),
    (2, 203, TRUE),
    (2, 204, TRUE),
    (2, 205, TRUE),
    (2, 206, TRUE);
