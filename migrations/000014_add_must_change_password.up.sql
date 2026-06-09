ALTER TABLE staff ADD COLUMN must_change_password BOOLEAN NOT NULL DEFAULT true;
UPDATE staff SET must_change_password = false WHERE username = 'admin';
