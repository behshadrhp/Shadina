ALTER TABLE comments
ADD CONSTRAINT fk_user_comment
    FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE comments
ADD CONSTRAINT fk_post_comment
    FOREIGN KEY (post_id) REFERENCES posts(id);
