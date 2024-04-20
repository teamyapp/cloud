-- +migrate Up
ALTER TABLE identity_sign_in_session
    ADD internal_user_id BIGINT;
ALTER TABLE identity_sign_in_session
    ADD type VARCHAR(50) NOT NULL DEFAULT 'UNKNOWN_USER_SIGN_IN';

-- +migrate Down
ALTER TABLE identity_sign_in_session
    DROP COLUMN type;
ALTER TABLE identity_sign_in_session
    DROP COLUMN "internal_user_id";
