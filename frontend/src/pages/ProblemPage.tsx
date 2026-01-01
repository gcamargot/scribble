import { useState, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import axios from 'axios';
import { useDailyProblem } from '../hooks/useDailyProblem';
import { TWO_SUM_STARTER_CODE, LANGUAGE_LABELS, type Language } from '../constants/starterCode';
import ResultPanel, { type SubmissionResult } from '../components/ResultPanel';

/**
 * Get difficulty badge color based on difficulty level
 */
function getDifficultyColor(difficulty: string): string {
  switch (difficulty) {
    case 'Easy':
      return 'bg-green-900 text-green-300';
    case 'Medium':
      return 'bg-yellow-900 text-yellow-300';
    case 'Hard':
      return 'bg-red-900 text-red-300';
    default:
      return 'bg-gray-900 text-gray-300';
  }
}

export default function ProblemPage() {
  // Fetch daily problem from API
  const { problem, isLoading: isProblemLoading, isError, isFallback } = useDailyProblem();

  const [language, setLanguage] = useState<Language>('python');
  const [code, setCode] = useState(TWO_SUM_STARTER_CODE[language]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<SubmissionResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Reset code when problem changes
  useEffect(() => {
    // TODO: Fetch starter code for the specific problem from API
    // For now, use hardcoded starter code
    setCode(TWO_SUM_STARTER_CODE[language]);
    setResult(null);
    setError(null);
  }, [problem.id, language]);

  // Update code when language changes
  const handleLanguageChange = (newLanguage: Language) => {
    setLanguage(newLanguage);
    setCode(TWO_SUM_STARTER_CODE[newLanguage]);
    setResult(null);
    setError(null);
  };

  // Handle code submission
  const handleSubmit = async () => {
    setIsSubmitting(true);
    setError(null);
    setResult(null);

    try {
      const response = await axios.post('/api/submissions', {
        code,
        language,
        problemId: problem.id
      });

      setResult(response.data);
    } catch (err) {
      console.error('Submission error:', err);
      setError(
        axios.isAxiosError(err)
          ? err.response?.data?.error || 'Failed to submit solution'
          : 'An unexpected error occurred'
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  // Show loading state while fetching problem
  if (isProblemLoading && !problem) {
    return (
      <div className="flex items-center justify-center h-screen bg-dark">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary mx-auto mb-4"></div>
          <p className="text-gray-300">Loading today's challenge...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-dark">
      {/* API Status Banner */}
      {(isError || isFallback) && (
        <div className="bg-yellow-900/50 border-b border-yellow-700 px-4 py-2 text-center">
          <span className="text-yellow-300 text-sm">
            {isError
              ? 'Unable to fetch daily problem. Showing sample problem.'
              : 'Using cached problem data.'}
          </span>
        </div>
      )}

      {/* Problem Section */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Panel - Problem Description */}
        <div className="w-1/2 border-r border-gray-700 overflow-y-auto">
          <div className="p-6">
            {/* Problem Header */}
            <div className="mb-6">
              <div className="flex items-center gap-3 mb-2">
                <h1 className="text-2xl font-bold text-white">{problem.title}</h1>
                <span className={`px-3 py-1 text-sm rounded-full ${getDifficultyColor(problem.difficulty)}`}>
                  {problem.difficulty}
                </span>
              </div>
            </div>

            {/* Description */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Description</h2>
              <p className="text-gray-300 whitespace-pre-line leading-relaxed">
                {problem.description}
              </p>
            </div>

            {/* Examples */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Examples</h2>
              {problem.examples.map((example, index) => (
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
                {problem.constraints.map((constraint, index) => (
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
              {(Object.keys(LANGUAGE_LABELS) as Language[]).map((lang) => (
                <option key={lang} value={lang}>
                  {LANGUAGE_LABELS[lang]}
                </option>
              ))}
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
              {problem.testCases.map((testCase, index) => (
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

          {/* Results Display */}
          {result && <ResultPanel result={result} />}

          {/* Error Display */}
          {error && (
            <div className="bg-gray-800 border-t border-gray-700 p-4">
              <div className="bg-red-900 border border-red-700 rounded-lg p-4">
                <h3 className="font-bold text-lg text-red-300 mb-2">Submission Error</h3>
                <p className="text-gray-300 text-sm">{error}</p>
              </div>
            </div>
          )}

          {/* Submit Button */}
          <div className="bg-gray-800 border-t border-gray-700 p-4">
            <button
              onClick={handleSubmit}
              disabled={isSubmitting}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-semibold py-3 rounded-lg transition-colors"
            >
              {isSubmitting ? 'Submitting...' : 'Submit Solution'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
