-- +migrate Up
CREATE TABLE resource
(
    id                   BIGINT PRIMARY KEY,
    resource_type        VARCHAR(50) NOT NULL,
    parent_resource_id   BIGINT,
    parent_resource_type VARCHAR(50) NOT NULL
);

CREATE TABLE security_group
(
    id          BIGINT PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    description VARCHAR(240),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE TABLE security_group_user
(
    group_id   BIGINT NOT NULL REFERENCES resource (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id    BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permission
(
    resource_type VARCHAR(50) NOT NULL,
    resource_id   BIGINT      NOT NULL REFERENCES resource (id) ON UPDATE CASCADE ON DELETE CASCADE,
    operation     VARCHAR(50) NOT NULL,
    group_id      BIGINT      NOT NULL REFERENCES security_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP
);

CREATE TABLE resource_operation
(
    resource_type        VARCHAR(50) NOT NULL,
    operation            VARCHAR(50) NOT NULL,
    parent_resource_type VARCHAR(50) NOT NULL,
    parent_operation     VARCHAR(50) NOT NULL
);

-- +migrate Down
DROP TABLE resource_operation;
DROP TABLE permission;
DROP TABLE security_group_user;
DROP TABLE security_group;
DROP TABLE resource;

