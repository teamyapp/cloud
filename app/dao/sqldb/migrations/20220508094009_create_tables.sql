-- +migrate Up
CREATE TABLE identity_allocated_range
(
    key       VARCHAR(255) PRIMARY KEY,
    range_end BIGINT
);

CREATE TABLE identity_sign_in_session
(
    id           BIGINT PRIMARY KEY,
    redirect_url VARCHAR(2048)
);

CREATE TABLE identity_user_link
(
    auth_provider    VARCHAR(100),
    external_user_id VARCHAR(50),
    internal_user_id BIGINT,
    PRIMARY KEY (auth_provider, external_user_id)
);

-- +migrate Down
DROP TABLE identity_user_link;
DROP TABLE identity_sign_in_session;
DROP TABLE identity_allocated_range;
