/**
 * Problem data structures and constants
 *
 * Contains hardcoded problem data for POC phase.
 * In production, this will be fetched from backend API.
 */

export interface ProblemExample {
  input: string;
  output: string;
  explanation: string;
}

export interface TestCase {
  input: string;
  expected: string;
}

export interface Problem {
  id: number;
  title: string;
  difficulty: 'Easy' | 'Medium' | 'Hard';
  description: string;
  examples: ProblemExample[];
  constraints: string[];
  testCases: TestCase[];
}

/**
 * Hardcoded Two Sum problem for POC
 *
 * This demonstrates the problem structure that will eventually
 * come from the backend API's /api/problems/daily endpoint
 */
export const TWO_SUM_PROBLEM: Problem = {
  id: 1,
  title: 'Two Sum',
  difficulty: 'Easy',
  description: `Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target.

You may assume that each input would have exactly one solution, and you may not use the same element twice.

You can return the answer in any order.`,
  examples: [
    {
      input: 'nums = [2,7,11,15], target = 9',
      output: '[0,1]',
      explanation: 'Because nums[0] + nums[1] == 9, we return [0, 1].'
    },
    {
      input: 'nums = [3,2,4], target = 6',
      output: '[1,2]',
      explanation: 'Because nums[1] + nums[2] == 6, we return [1, 2].'
    },
    {
      input: 'nums = [3,3], target = 6',
      output: '[0,1]',
      explanation: 'Because nums[0] + nums[1] == 6, we return [0, 1].'
    }
  ],
  constraints: [
    '2 <= nums.length <= 10⁴',
    '-10⁹ <= nums[i] <= 10⁹',
    '-10⁹ <= target <= 10⁹',
    'Only one valid answer exists.'
  ],
  testCases: [
    { input: '[2,7,11,15]\\n9', expected: '[0,1]' },
    { input: '[3,2,4]\\n6', expected: '[1,2]' },
    { input: '[3,3]\\n6', expected: '[0,1]' }
  ]
};
