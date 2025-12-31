-- Scribble Discord LeetCode Activity - Database Schema

-- Users table (synced from Discord)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    discord_id VARCHAR(64) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    discriminator VARCHAR(4),
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_discord_id ON users(discord_id);
CREATE INDEX idx_users_username ON users(username);

-- Problems table (800 problems pre-seeded)
CREATE TABLE IF NOT EXISTS problems (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    difficulty VARCHAR(20) NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    description TEXT NOT NULL,
    constraints TEXT,
    hints TEXT[],
    category VARCHAR(50),
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_problems_difficulty ON problems(difficulty);
CREATE INDEX idx_problems_category ON problems(category);
CREATE INDEX idx_problems_slug ON problems(slug);

-- Test cases for each problem
CREATE TABLE IF NOT EXISTS test_cases (
    id BIGSERIAL PRIMARY KEY,
    problem_id INTEGER NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    input JSONB NOT NULL,
    expected_output JSONB NOT NULL,
    is_sample BOOLEAN DEFAULT FALSE,
    weight DECIMAL(3,2) DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_test_cases_problem_id ON test_cases(problem_id);
CREATE INDEX idx_test_cases_is_sample ON test_cases(is_sample);

-- Daily challenges (one problem per day at UTC midnight)
CREATE TABLE IF NOT EXISTS daily_challenges (
    id SERIAL PRIMARY KEY,
    problem_id INTEGER NOT NULL REFERENCES problems(id),
    challenge_date DATE UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_daily_challenges_date ON daily_challenges(challenge_date DESC);
CREATE INDEX idx_daily_challenges_problem_id ON daily_challenges(problem_id);

-- Submissions (ALL attempts stored)
CREATE TABLE IF NOT EXISTS submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id INTEGER NOT NULL REFERENCES problems(id),
    daily_challenge_id INTEGER REFERENCES daily_challenges(id),
    language VARCHAR(20) NOT NULL CHECK (language IN ('python', 'javascript', 'java', 'cpp', 'rust', 'go')),
    code TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('accepted', 'wrong_answer', 'runtime_error', 'timeout', 'memory_limit', 'compilation_error')),

    -- Execution metrics
    execution_time_ms INTEGER,
    memory_used_kb INTEGER,
    tests_passed INTEGER DEFAULT 0,
    tests_total INTEGER DEFAULT 0,

    -- Metadata
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    error_message TEXT
);

CREATE INDEX idx_submissions_user_id ON submissions(user_id);
CREATE INDEX idx_submissions_problem_id ON submissions(problem_id);
CREATE INDEX idx_submissions_daily_challenge ON submissions(daily_challenge_id);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_submitted_at ON submissions(submitted_at DESC);
CREATE INDEX idx_submissions_user_problem ON submissions(user_id, problem_id);
CREATE INDEX idx_submissions_language ON submissions(language);

-- User streaks (consecutive days solving daily challenge)
CREATE TABLE IF NOT EXISTS streaks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    last_solved_date DATE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_streaks_user_id ON streaks(user_id);

-- Leaderboard cache (computed periodically for performance)
CREATE TABLE IF NOT EXISTS leaderboard_cache (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL CHECK (metric_type IN ('fastest_avg', 'lowest_memory_avg', 'problems_solved', 'longest_streak')),
    metric_value DECIMAL(12,2) NOT NULL,
    rank INTEGER NOT NULL,
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, metric_type)
);

CREATE INDEX idx_leaderboard_metric_rank ON leaderboard_cache(metric_type, rank);
CREATE INDEX idx_leaderboard_user_id ON leaderboard_cache(user_id);

-- Execution jobs tracking (for Kubernetes job management)
CREATE TABLE IF NOT EXISTS execution_jobs (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    k8s_job_name VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'timeout')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_execution_jobs_status ON execution_jobs(status);
CREATE INDEX idx_execution_jobs_submission ON execution_jobs(submission_id);
CREATE INDEX idx_execution_jobs_created_at ON execution_jobs(created_at DESC);

-- Views for common queries
CREATE OR REPLACE VIEW user_aggregate_metrics AS
SELECT
    u.id AS user_id,
    u.username,
    u.avatar_url,
    COUNT(DISTINCT s.problem_id) FILTER (WHERE s.status = 'accepted') AS problems_solved,
    COUNT(DISTINCT s.id) FILTER (WHERE s.status = 'accepted') AS total_accepted_submissions,
    ROUND(AVG(s.execution_time_ms) FILTER (WHERE s.status = 'accepted' AND s.execution_time_ms IS NOT NULL)::numeric, 2) AS avg_execution_time_ms,
    ROUND(AVG(s.memory_used_kb) FILTER (WHERE s.status = 'accepted' AND s.memory_used_kb IS NOT NULL)::numeric, 2) AS avg_memory_kb,
    ROUND((COUNT(s.id) FILTER (WHERE s.status = 'accepted')::numeric / NULLIF(COUNT(s.id), 0) * 100)::numeric, 2) AS acceptance_rate,
    st.current_streak,
    st.longest_streak,
    st.last_solved_date
FROM users u
LEFT JOIN submissions s ON u.id = s.user_id
LEFT JOIN streaks st ON u.id = st.user_id
GROUP BY u.id, u.username, u.avatar_url, st.current_streak, st.longest_streak, st.last_solved_date;

-- Enable UUID extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add any initial settings or configuration
CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO settings (key, value) VALUES
    ('daily_challenge_hour_utc', '0'),
    ('daily_challenge_minute_utc', '0'),
    ('leaderboard_update_interval_minutes', '5'),
    ('execution_timeout_seconds', '10'),
    ('execution_memory_limit_mb', '512'),
    ('problems_total', '800')
ON CONFLICT (key) DO NOTHING;
