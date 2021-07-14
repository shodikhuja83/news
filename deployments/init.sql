CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscribes
(
    user_id           int NOT NULL,
    tag TEXT NOT NULL,
    readen_articles text[],
    PRIMARY KEY(user_id, tag)
);