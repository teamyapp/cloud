-- +migrate Up
CREATE TABLE resource_user_group_relation
(
    resource_type VARCHAR(50) NOT NULL REFERENCES resource_type (resource_type) ON UPDATE CASCADE ON DELETE CASCADE,
    resource_id   BIGINT      NOT NULL,
    user_group_id BIGINT      NOT NULL REFERENCES user_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
    key           VARCHAR(50),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    creator_user_id BIGINT NOT NULL,
    PRIMARY KEY (resource_type, resource_id, user_group_id)
);

CREATE INDEX idx_resource_user_group_relation_user_group_id
    ON resource_user_group_relation(user_group_id);

-- +migrate Down
DROP INDEX idx_resource_user_group_relation_user_group_id
DROP TABLE resource_user_group_relation;