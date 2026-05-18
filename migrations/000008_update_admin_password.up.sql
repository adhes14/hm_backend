-- Update admin password_hash to represent the plaintext password 'admin123'
UPDATE staff SET password_hash = '$2a$10$pXjnnoyFI3fNm9Ub9.x8O.4OXm6KpAkEEkQnncGtWQGCnmxbbrg5K' WHERE username = 'admin';
