-- +migrate Up
CREATE TABLE resource_relation
(
    id                   BIGINT PRIMARY KEY,
    resource_type        VARCHAR(50) NOT NULL,
    parent_resource_id   BIGINT,
    parent_resource_type VARCHAR(50) NOT NULL
);

CREATE TABLE user_group
(
    id          BIGINT PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    description VARCHAR(240),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE TABLE user_group_member
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

CREATE TABLE operation_relation
(
    resource_type        VARCHAR(50) NOT NULL,
    operation            VARCHAR(50) NOT NULL,
    parent_resource_type VARCHAR(50) NOT NULL,
    parent_operation     VARCHAR(50) NOT NULL
);

-- +migrate Down
DROP TABLE operation_relation;
DROP TABLE permission;
DROP TABLE user_group_member;
DROP TABLE user_group;
DROP TABLE resource_relation;