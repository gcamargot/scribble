-- User Aggregate Metrics View
-- Provides computed statistics for each user based on their submissions and streaks
-- This view is queried for profile pages and user statistics endpoints
--
-- Note: This view is already defined in schema.sql but extracted here for clarity
-- Run this if the view needs to be recreated or modified

CREATE OR REPLACE VIEW user_aggregate_metrics AS
SELECT
    u.id AS user_id,
    u.username,
    u.avatar_url,

    -- Problem solving metrics
    COUNT(DISTINCT s.problem_id) FILTER (WHERE s.status = 'accepted') AS problems_solved,
    COUNT(DISTINCT s.id) FILTER (WHERE s.status = 'accepted') AS total_accepted_submissions,
    COUNT(s.id) AS total_submissions,

    -- Performance metrics (only for accepted submissions)
    ROUND(AVG(s.execution_time_ms) FILTER (
        WHERE s.status = 'accepted' AND s.execution_time_ms IS NOT NULL
    )::numeric, 2) AS avg_execution_time_ms,

    ROUND(AVG(s.memory_used_kb) FILTER (
        WHERE s.status = 'accepted' AND s.memory_used_kb IS NOT NULL
    )::numeric, 2) AS avg_memory_kb,

    -- Acceptance rate as percentage
    ROUND((
        COUNT(s.id) FILTER (WHERE s.status = 'accepted')::numeric /
        NULLIF(COUNT(s.id), 0) * 100
    )::numeric, 2) AS acceptance_rate,

    -- Streak information (from streaks table)
    COALESCE(st.current_streak, 0) AS current_streak,
    COALESCE(st.longest_streak, 0) AS longest_streak,
    st.last_solved_date,

    -- Language breakdown (most used language)
    (
        SELECT language
        FROM submissions s2
        WHERE s2.user_id = u.id AND s2.status = 'accepted'
        GROUP BY language
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ) AS favorite_language

FROM users u
LEFT JOIN submissions s ON u.id = s.user_id
LEFT JOIN streaks st ON u.id = st.user_id
GROUP BY u.id, u.username, u.avatar_url, st.current_streak, st.longest_streak, st.last_solved_date;

-- Index hint: The view automatically benefits from these indexes:
-- - idx_submissions_user_id
-- - idx_submissions_status
-- - idx_streaks_user_id

-- Usage example:
-- SELECT * FROM user_aggregate_metrics WHERE user_id = 123;
-- SELECT username, problems_solved, current_streak FROM user_aggregate_metrics ORDER BY problems_solved DESC LIMIT 10;
