-- +migrate Up
ALTER TABLE identity_allocated_range
    RENAME TO allocated_range;
ALTER TABLE identity_user_link
    ADD external_user_label VARCHAR(50) NOT NULL DEFAULT 'Unnamed';

-- +migrate Down
ALTER TABLE identity_user_link
    DROP COLUMN external_user_label;
ALTER TABLE allocated_range
    RENAME TO identity_allocated_range;