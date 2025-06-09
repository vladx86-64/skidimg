CREATE TABLE "image_album" (
    image_id BIGINT NOT NULL,
    album_id BIGINT NOT NULL,
    PRIMARY KEY (image_id, album_id),
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
);
