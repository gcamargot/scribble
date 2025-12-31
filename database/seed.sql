-- Scribble - Sample Data Seed

-- Sample Problems
INSERT INTO problems (title, slug, difficulty, description, constraints, category, tags) VALUES
('Two Sum', 'two-sum', 'easy',
 'Given an array of integers nums and an integer target, return the indices of the two numbers that add up to target. You may assume each input has exactly one solution.',
 'You cannot use the same element twice. The return values must be 1-indexed.',
 'array',
 ARRAY['array', 'hash-table', 'two-pointers']),

('Reverse String', 'reverse-string', 'easy',
 'Write a function that reverses a string. The input string is given as an array of characters s.',
 'You must do this by modifying the input array in-place with O(1) extra memory.',
 'string',
 ARRAY['string', 'two-pointers']),

('Median of Two Sorted Arrays', 'median-two-sorted-arrays', 'hard',
 'Given two sorted arrays nums1 and nums2 of size m and n respectively, return the median of the two sorted arrays.',
 'The overall run time complexity should be O(log (m+n)).',
 'array',
 ARRAY['array', 'binary-search', 'divide-and-conquer']),

('Longest Substring Without Repeating Characters', 'longest-substring-without-repeating', 'medium',
 'Given a string s, find the length of the longest substring without repeating characters.',
 'A substring is a contiguous non-empty sequence of characters within a string.',
 'string',
 ARRAY['string', 'hash-table', 'sliding-window']),

('Binary Tree Level Order Traversal', 'binary-tree-level-order', 'medium',
 'Given the root of a binary tree, return the level order traversal of its nodes values.',
 'BFS or queue-based approach recommended.',
 'tree',
 ARRAY['tree', 'breadth-first-search', 'queue']),

('Valid Parentheses', 'valid-parentheses', 'easy',
 'Given a string s containing just the characters ''('' , '')'' , ''{'' , ''}'' , ''['' and '']'' , determine if the input string is valid.',
 'An input string is valid if all brackets are closed in the correct order.',
 'string',
 ARRAY['string', 'stack']),

('Merge K Sorted Lists', 'merge-k-sorted-lists', 'hard',
 'You are given an array of k linked-lists lists, each linked-list is sorted in ascending order. Merge all the linked-lists into one sorted linked-list and return it.',
 'The number of nodes in all lists is at most 10^4.',
 'linked-list',
 ARRAY['linked-list', 'divide-and-conquer', 'heap']),

('Best Time to Buy and Sell Stock', 'best-time-buy-sell-stock', 'easy',
 'You are given an array prices where prices[i] is the price of a given stock on the ith day. You want to maximize your profit by choosing a single day to buy and a single day to sell. Return the maximum profit you can achieve.',
 'If you cannot achieve any profit, return 0.',
 'array',
 ARRAY['array', 'dynamic-programming']),

('Number of Islands', 'number-of-islands', 'medium',
 'Given an m x n 2D binary grid grid which represents a map of ''1''s (land) and ''0''s (water), return the number of islands.',
 'An island is surrounded by water and is formed by connecting adjacent lands horizontally or vertically.',
 'graph',
 ARRAY['array', 'depth-first-search', 'breadth-first-search']),

('Climbing Stairs', 'climbing-stairs', 'easy',
 'You are climbing a staircase. It takes n steps to reach the top. Each time you can climb 1 or 2 steps. In how many distinct ways can you climb to the top?',
 'The answer could be a very big number, return the answer modulo 10^9 + 7.',
 'dynamic-programming',
 ARRAY['dynamic-programming', 'math', 'memoization']);

-- Sample Test Cases for Two Sum
INSERT INTO test_cases (problem_id, input, expected_output, is_sample) VALUES
((SELECT id FROM problems WHERE slug = 'two-sum'),
 '{"nums": [2, 7, 11, 15], "target": 9}'::jsonb,
 '[0, 1]'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'two-sum'),
 '{"nums": [3, 2, 4], "target": 6}'::jsonb,
 '[1, 2]'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'two-sum'),
 '{"nums": [3, 3], "target": 6}'::jsonb,
 '[0, 1]'::jsonb,
 false),

((SELECT id FROM problems WHERE slug = 'two-sum'),
 '{"nums": [1, 2, 3, 4, 5], "target": 9}'::jsonb,
 '[3, 4]'::jsonb,
 false),

((SELECT id FROM problems WHERE slug = 'two-sum'),
 '{"nums": [-1, 0, 1, 2, -1, -4], "target": 0}'::jsonb,
 '[0, 3]'::jsonb,
 false);

-- Sample Test Cases for Reverse String
INSERT INTO test_cases (problem_id, input, expected_output, is_sample) VALUES
((SELECT id FROM problems WHERE slug = 'reverse-string'),
 '{"s": ["h","e","l","l","o"]}'::jsonb,
 '["o","l","l","e","h"]'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'reverse-string'),
 '{"s": ["H","a","n","n","a","h"]}'::jsonb,
 '["h","a","n","n","a","H"]'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'reverse-string'),
 '{"s": ["a"]}'::jsonb,
 '["a"]'::jsonb,
 false),

((SELECT id FROM problems WHERE slug = 'reverse-string'),
 '{"s": ["a","b","c","d","e","f","g"]}'::jsonb,
 '["g","f","e","d","c","b","a"]'::jsonb,
 false),

((SELECT id FROM problems WHERE slug = 'reverse-string'),
 '{"s": ["hello"]}'::jsonb,
 '["olleh"]'::jsonb,
 false);

-- Sample Test Cases for Valid Parentheses
INSERT INTO test_cases (problem_id, input, expected_output, is_sample) VALUES
((SELECT id FROM problems WHERE slug = 'valid-parentheses'),
 '{"s": "()"}'::jsonb,
 'true'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'valid-parentheses'),
 '{"s": "()[]{}"}'::jsonb,
 'true'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'valid-parentheses'),
 '{"s": "([)]"}'::jsonb,
 'false'::jsonb,
 true),

((SELECT id FROM problems WHERE slug = 'valid-parentheses'),
 '{"s": "{[]}"}'::jsonb,
 'true'::jsonb,
 false),

((SELECT id FROM problems WHERE slug = 'valid-parentheses'),
 '{"s": "("}'::jsonb,
 'false'::jsonb,
 false);

-- Create a sample daily challenge for today
INSERT INTO daily_challenges (problem_id, challenge_date)
VALUES ((SELECT id FROM problems WHERE slug = 'two-sum'), CURRENT_DATE)
ON CONFLICT DO NOTHING;

-- Create a sample user for testing
INSERT INTO users (discord_id, username, avatar_url)
VALUES ('123456789', 'test_user', 'https://example.com/avatar.png')
ON CONFLICT (discord_id) DO NOTHING;

-- Create initial streaks for sample user
INSERT INTO streaks (user_id, current_streak, longest_streak, last_solved_date)
VALUES ((SELECT id FROM users WHERE discord_id = '123456789'), 0, 0, NULL)
ON CONFLICT (user_id) DO NOTHING;
