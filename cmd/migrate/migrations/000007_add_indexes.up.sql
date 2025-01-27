CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Indeks untuk tabel comments
CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);

-- Indeks untuk tabel posts
CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

-- Indeks untuk tabel users
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- Indeks untuk foreign keys di tabel posts dan comments
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);
