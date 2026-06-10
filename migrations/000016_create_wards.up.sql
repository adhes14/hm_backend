CREATE TABLE IF NOT EXISTS wards (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT
);

-- Crear una sala por defecto
INSERT INTO wards (name, description) VALUES ('Sala General', 'Sala creada por migración automática');

-- Agregar columna ward_id a camas
ALTER TABLE beds ADD COLUMN ward_id INT REFERENCES wards(id);

-- Asignar todas las camas existentes a la sala por defecto
UPDATE beds SET ward_id = (SELECT id FROM wards WHERE name = 'Sala General' LIMIT 1);

-- Hacer la columna ward_id obligatoria
ALTER TABLE beds ALTER COLUMN ward_id SET NOT NULL;
