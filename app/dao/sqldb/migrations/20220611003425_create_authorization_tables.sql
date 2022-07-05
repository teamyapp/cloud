-- +migrate Up
CREATE TABLE resource_relation
(
    child_resource_id    BIGINT PRIMARY KEY,
    child_resource_type  VARCHAR(50) NOT NULL,
    parent_resource_id   BIGINT,
    parent_resource_type VARCHAR(50) NOT NULL,
    CONSTRAINT resource)relation_pk PRIMARY KEY (child_resource_id, child_resource_type, parent_resource_id, parent_resource_type)
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permission
(
    resource_type VARCHAR(50) NOT NULL,
    resource_id   BIGINT      NOT NULL REFERENCES,
    operation     VARCHAR(50) NOT NULL,
    group_id      BIGINT      NOT NULL REFERENCES user_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP
);

CREATE TABLE operation_relation
(
    child_resource_type  VARCHAR(50) NOT NULL,
    child_operation      VARCHAR(50) NOT NULL,
    parent_resource_type VARCHAR(50) NOT NULL,
    parent_operation     VARCHAR(50) NOT NULL
);

-- +migrate Down
DROP TABLE operation_relation;
DROP TABLE permission;
DROP TABLE user_group_member;
DROP TABLE user_group;
DROP TABLE resource_relation;