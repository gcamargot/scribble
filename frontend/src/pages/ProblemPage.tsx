import { useState } from 'react';
import Editor from '@monaco-editor/react';

// Hardcoded Two Sum problem for POC
const PROBLEM = {
  id: 1,
  title: 'Two Sum',
  difficulty: 'Easy' as const,
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
    { input: '[2,7,11,15]\n9', expected: '[0,1]' },
    { input: '[3,2,4]\n6', expected: '[1,2]' },
    { input: '[3,3]\n6', expected: '[0,1]' }
  ]
};

const STARTER_CODE = {
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

type Language = keyof typeof STARTER_CODE;

export default function ProblemPage() {
  const [language, setLanguage] = useState<Language>('python');
  const [code, setCode] = useState(STARTER_CODE[language]);

  // Update code when language changes
  const handleLanguageChange = (newLanguage: Language) => {
    setLanguage(newLanguage);
    setCode(STARTER_CODE[newLanguage]);
  };

  return (
    <div className="flex flex-col h-screen bg-dark">
      {/* Problem Section */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel - Problem Description */}
        <div className="w-1/2 border-r border-gray-700 overflow-y-auto">
          <div className="p-6">
            {/* Problem Header */}
            <div className="mb-6">
              <div className="flex items-center gap-3 mb-2">
                <h1 className="text-2xl font-bold text-white">{PROBLEM.title}</h1>
                <span className="px-3 py-1 bg-green-900 text-green-300 text-sm rounded-full">
                  {PROBLEM.difficulty}
                </span>
              </div>
            </div>

            {/* Description */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Description</h2>
              <p className="text-gray-300 whitespace-pre-line leading-relaxed">
                {PROBLEM.description}
              </p>
            </div>

            {/* Examples */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Examples</h2>
              {PROBLEM.examples.map((example, index) => (
                <div key={index} className="mb-4 bg-gray-800 rounded-lg p-4 border border-gray-700">
                  <p className="text-gray-400 text-sm mb-1">Example {index + 1}:</p>
                  <div className="mb-2">
                    <span className="text-white font-mono text-sm">Input: </span>
                    <span className="text-blue-400 font-mono text-sm">{example.input}</span>
                  </div>
                  <div className="mb-2">
                    <span className="text-white font-mono text-sm">Output: </span>
                    <span className="text-green-400 font-mono text-sm">{example.output}</span>
                  </div>
                  <div>
                    <span className="text-white font-mono text-sm">Explanation: </span>
                    <span className="text-gray-300 font-mono text-sm">{example.explanation}</span>
                  </div>
                </div>
              ))}
            </div>

            {/* Constraints */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Constraints</h2>
              <ul className="list-disc list-inside text-gray-300 space-y-1">
                {PROBLEM.constraints.map((constraint, index) => (
                  <li key={index} className="font-mono text-sm">{constraint}</li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* Right Panel - Code Editor */}
        <div className="w-1/2 flex flex-col">
          {/* Language Selector */}
          <div className="bg-gray-800 border-b border-gray-700 p-3 flex items-center gap-2">
            <label className="text-gray-400 text-sm font-medium">Language:</label>
            <select
              value={language}
              onChange={(e) => handleLanguageChange(e.target.value as Language)}
              className="bg-gray-700 text-white rounded px-3 py-1.5 text-sm border border-gray-600 focus:outline-none focus:border-blue-500"
            >
              <option value="python">Python</option>
              <option value="javascript">JavaScript</option>
              <option value="java">Java</option>
              <option value="cpp">C++</option>
              <option value="go">Go</option>
              <option value="rust">Rust</option>
            </select>
          </div>

          {/* Monaco Editor */}
          <div className="flex-1">
            <Editor
              height="100%"
              language={language}
              value={code}
              onChange={(value) => setCode(value || '')}
              theme="vs-dark"
              options={{
                minimap: { enabled: false },
                fontSize: 14,
                lineNumbers: 'on',
                scrollBeyondLastLine: false,
                automaticLayout: true,
                tabSize: 4,
                wordWrap: 'on'
              }}
            />
          </div>

          {/* Test Cases Section */}
          <div className="bg-gray-800 border-t border-gray-700 p-4">
            <h3 className="text-white font-semibold mb-3 text-sm">Sample Test Cases</h3>
            <div className="space-y-2 max-h-32 overflow-y-auto">
              {PROBLEM.testCases.map((testCase, index) => (
                <div key={index} className="bg-gray-900 rounded p-2 text-xs">
                  <div className="mb-1">
                    <span className="text-gray-400">Input: </span>
                    <span className="text-blue-400 font-mono">{testCase.input.replace('\n', ', ')}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Expected: </span>
                    <span className="text-green-400 font-mono">{testCase.expected}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Submit Button */}
          <div className="bg-gray-800 border-t border-gray-700 p-4">
            <button className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 rounded-lg transition-colors">
              Submit Solution
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
