DELETE FROM system_settings WHERE key IN ('sound_alert_patient_admitted', 'sound_alert_patient_discharged');

DROP TABLE IF EXISTS auxiliary_orders;
DROP TYPE IF EXISTS order_status_enum;
DROP TYPE IF EXISTS order_category_enum;
