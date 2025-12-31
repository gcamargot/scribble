/**
 * Starter code templates for all supported languages
 *
 * Contains boilerplate code for each problem in all 6 supported languages.
 * In production, these will be fetched from backend along with problem data.
 */

export type Language = 'python' | 'javascript' | 'java' | 'cpp' | 'go' | 'rust';

/**
 * Language display names for UI
 */
export const LANGUAGE_LABELS: Record<Language, string> = {
  python: 'Python',
  javascript: 'JavaScript',
  java: 'Java',
  cpp: 'C++',
  go: 'Go',
  rust: 'Rust'
};

/**
 * Monaco Editor language identifiers
 */
export const MONACO_LANGUAGES: Record<Language, string> = {
  python: 'python',
  javascript: 'javascript',
  java: 'java',
  cpp: 'cpp',
  go: 'go',
  rust: 'rust'
};

/**
 * Starter code templates for Two Sum problem
 *
 * Each template includes the function signature and a comment
 * prompting the user to write their solution
 */
export const TWO_SUM_STARTER_CODE: Record<Language, string> = {
  python: `def twoSum(nums: List[int], target: int) -> List[int]:
    # Write your solution here
    pass`,

  javascript: `/**
 * @param {number[]} nums
 * @param {number} target
 * @return {number[]}
 */
var twoSum = function(nums, target) {
    // Write your solution here
};`,

  java: `class Solution {
    public int[] twoSum(int[] nums, int target) {
        // Write your solution here
    }
}`,

  cpp: `class Solution {
public:
    vector<int> twoSum(vector<int>& nums, int target) {
        // Write your solution here
    }
};`,

  go: `func twoSum(nums []int, target int) []int {
    // Write your solution here
}`,

  rust: `impl Solution {
    pub fn two_sum(nums: Vec<i32>, target: i32) -> Vec<i32> {
        // Write your solution here
    }
}`
};
