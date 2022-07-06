-- +migrate Up
CREATE TABLE identity_service_account
(
    id BIGINT NOT NULL,
    name VARCHAR(50) NOT NULL,
    secret VARCHAR(10),
    owner_user_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (owner_user_id, id)
);

-- +migrate Down
DROP TABLE identity_service_account;