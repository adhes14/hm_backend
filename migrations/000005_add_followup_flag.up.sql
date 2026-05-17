-- Add requires_postpartum_followup column to bed_types
ALTER TABLE bed_types ADD COLUMN requires_postpartum_followup BOOLEAN NOT NULL DEFAULT false;

-- Backfill M-prefix bed types (Maternity needs follow-up)
UPDATE bed_types SET requires_postpartum_followup = true WHERE prefix = 'M';

-- Null next_control_at for ARO admissions (they should not have follow-up tracking)
UPDATE admissions
SET next_control_at = NULL
WHERE bed_id IN (
    SELECT b.id FROM beds b
    JOIN bed_types bt ON b.bed_type_id = bt.id
    WHERE bt.prefix = 'ARO'
);