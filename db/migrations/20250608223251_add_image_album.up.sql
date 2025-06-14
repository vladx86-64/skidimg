CREATE TABLE "image_album" (
    image_id INT NOT NULL,
    album_id INT NOT NULL,
    PRIMARY KEY (image_id, album_id),
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
);
