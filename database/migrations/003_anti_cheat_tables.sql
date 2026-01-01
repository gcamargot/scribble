-- Anti-Cheat Tables Migration
-- Creates tables for tracking flagged submissions and rate limiting

-- Flagged submissions table
-- Stores submissions flagged for potential cheating
CREATE TABLE IF NOT EXISTS flagged_submissions (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_id INTEGER NOT NULL REFERENCES problems(id),
    reason VARCHAR(50) NOT NULL CHECK (reason IN (
        'suspicious_time',
        'zero_memory',
        'rate_limit_abuse',
        'identical_code',
        'pattern_match'
    )),
    details TEXT, -- JSON details about the flag
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',
        'reviewed',
        'cleared',
        'banned'
    )),
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for flagged submissions
CREATE INDEX idx_flagged_submissions_submission ON flagged_submissions(submission_id);
CREATE INDEX idx_flagged_submissions_user ON flagged_submissions(user_id);
CREATE INDEX idx_flagged_submissions_status ON flagged_submissions(status);
CREATE INDEX idx_flagged_submissions_reason ON flagged_submissions(reason);
CREATE INDEX idx_flagged_submissions_created ON flagged_submissions(created_at DESC);

-- Rate limit entries table
-- Tracks submission rate per user for brute force prevention
CREATE TABLE IF NOT EXISTS rate_limit_entries (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    submissions INTEGER NOT NULL DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    last_submit TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(user_id)
);

CREATE INDEX idx_rate_limit_user ON rate_limit_entries(user_id);
CREATE INDEX idx_rate_limit_last_submit ON rate_limit_entries(last_submit);

-- View for admin dashboard
CREATE OR REPLACE VIEW flag_summary AS
SELECT
    f.status,
    f.reason,
    COUNT(*) as count,
    COUNT(DISTINCT f.user_id) as unique_users,
    MIN(f.created_at) as oldest,
    MAX(f.created_at) as newest
FROM flagged_submissions f
GROUP BY f.status, f.reason
ORDER BY f.status, count DESC;

-- Comments for documentation
COMMENT ON TABLE flagged_submissions IS 'Stores submissions flagged for potential cheating';
COMMENT ON TABLE rate_limit_entries IS 'Tracks submission rate per user for brute force prevention';
COMMENT ON COLUMN flagged_submissions.reason IS 'Why the submission was flagged (suspicious_time, zero_memory, etc.)';
COMMENT ON COLUMN flagged_submissions.status IS 'Review status (pending, reviewed, cleared, banned)';
COMMENT ON COLUMN rate_limit_entries.submissions IS 'Number of submissions in current rate limit window';
