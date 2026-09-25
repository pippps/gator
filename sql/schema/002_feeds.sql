-- +goose Up
CREATE TABLE feeds(
                      id UUID PRIMARY KEY UNIQUE NOT NULL ,
                      created_at TIMESTAMP,
                      updated_at TIMESTAMP,
                      name TEXT UNIQUE NOT NULL,
                      url TEXT UNIQUE ,
                      user_id UUID NOT NULL,
                      CONSTRAINT fk_user_id
                          FOREIGN KEY (user_id)
                          REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;