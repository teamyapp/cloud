-- +migrate Up
CREATE TABLE identity_allocated_range
(
    key         VARCHAR(255),
    range_end   BIGINT,
    next_number BIGINT
);

CREATE TABLE identity_sign_in_session
(
    id           BIGINT,
    redirect_url VARCHAR(2048)
);

CREATE TABLE identity_user_link
(
    auth_provider    VARCHAR(100),
    external_user_id VARCHAR(50),
    internal_user_id BIGINT
);

-- +migrate Down
DROP TABLE identity_user_link;
DROP TABLE identity_sign_in_session;
DROP TABLE identity_allocated_range;
