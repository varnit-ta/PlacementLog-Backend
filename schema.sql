-- Placement Log Database Schema

-- Users table
CREATE TABLE IF NOT EXISTS placement_log_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regno VARCHAR(20) UNIQUE NOT NULL, -- Registration number (e.g., 22bcs1234)
    username VARCHAR(255) NOT NULL,    -- User's name (not unique)
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Admins table
CREATE TABLE IF NOT EXISTS placement_log_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Posts table with reviewed field
CREATE TABLE IF NOT EXISTS placement_log_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES placement_log_users(id) ON DELETE CASCADE,
    post_body JSONB NOT NULL,
    reviewed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Placement Companies Table
CREATE TABLE IF NOT EXISTS placement_companies (
    id SERIAL PRIMARY KEY,
    company VARCHAR(100) NOT NULL,
    ctc DECIMAL(10,2) NULL, -- Allow NULL for cases where CTC is not available (NA)
    placement_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Placement Branchwise Record Table
CREATE TABLE IF NOT EXISTS placement_branchwise_record (
    id SERIAL PRIMARY KEY,
    placement_id INT NOT NULL REFERENCES placement_companies(id) ON DELETE CASCADE,
    branch VARCHAR(10) NOT NULL,
    count INT NOT NULL
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON placement_log_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_reviewed ON placement_log_posts(reviewed);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON placement_log_posts(created_at);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON placement_log_users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON placement_log_admins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON placement_log_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 