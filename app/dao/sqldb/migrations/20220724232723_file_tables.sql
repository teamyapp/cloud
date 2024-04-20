-- +migrate Up
CREATE TABLE file_upload_session
(
    id                         BIGINT PRIMARY KEY,
    status                     VARCHAR(20)  NOT NULL,
    file_id                    BIGINT       NOT NULL DEFAULT 0,
    file_name                  VARCHAR(100) NOT NULL DEFAULT '',
    mime_type                  VARCHAR(255) NOT NULL DEFAULT '',
    chunk_ids                  TEXT         NOT NULL DEFAULT '',
    uploaded_size_in_bytes     BIGINT       NOT NULL DEFAULT 0,
    total_size_in_bytes        BIGINT       NOT NULL DEFAULT 0,
    total_num_of_chunks        INT          NOT NULL DEFAULT 0,
    next_chunk_index_to_upload INT          NOT NULL DEFAULT 0,
    hash_state                 BYTEA        NOT NULL DEFAULT '',
    actual_content_hash        VARCHAR(256)          DEFAULT '',
    expected_content_hash      VARCHAR(256)          DEFAULT '',
    created_at                 TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMP
);

CREATE TABLE file_metadata
(
    id               BIGINT PRIMARY KEY,
    name             VARCHAR(100) NOT NULL,
    size_in_bytes    BIGINT       NOT NULL DEFAULT 0,
    mime_type        VARCHAR(255) NOT NULL DEFAULT '',
    chunk_ids        TEXT         NOT NULL DEFAULT '',
    created_at       TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    last_modified_at TIMESTAMP
);

CREATE TABLE file_chunk_metadata
(
    id            BIGINT PRIMARY KEY,
    size_in_bytes BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMP       DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE file_chunk_metadata;
DROP TABLE file_metadata;
DROP TABLE file_upload_session;