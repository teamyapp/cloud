-- +migrate Up
CREATE TABLE resource_relation
(
    child_resource_type  VARCHAR(50) NOT NULL,
    child_resource_id    BIGINT NOT NULL,
    parent_resource_type VARCHAR(50) NOT NULL,
    parent_resource_id   BIGINT NOT NULL,
    PRIMARY KEY (child_resource_type, child_resource_id, parent_resource_type, parent_resource_id)
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
    group_id   BIGINT NOT NULL REFERENCES user_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id    BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE permission
(
    resource_type VARCHAR(50) NOT NULL,
    resource_id   BIGINT      NOT NULL,
    operation     VARCHAR(50) NOT NULL,
    group_id      BIGINT      NOT NULL REFERENCES user_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP,
    PRIMARY KEY (resource_type, resource_id, operation, group_id)
);

CREATE TABLE operation_relation
(
    child_resource_type  VARCHAR(50) NOT NULL,
    child_operation      VARCHAR(50) NOT NULL,
    parent_resource_type VARCHAR(50) NOT NULL,
    parent_operation     VARCHAR(50) NOT NULL,
    PRIMARY KEY (child_resource_type, child_operation, parent_resource_type, parent_operation)
);

-- +migrate Down
DROP TABLE operation_relation;
DROP TABLE permission;
DROP TABLE user_group_member;
DROP TABLE user_group;
DROP TABLE resource_relation;