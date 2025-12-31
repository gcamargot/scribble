import { useState } from 'react';
import Editor from '@monaco-editor/react';
import axios from 'axios';
import { TWO_SUM_PROBLEM } from '../constants/problems';
import { TWO_SUM_STARTER_CODE, LANGUAGE_LABELS, type Language } from '../constants/starterCode';
import ResultPanel, { type SubmissionResult } from '../components/ResultPanel';

export default function ProblemPage() {
  const [language, setLanguage] = useState<Language>('python');
  const [code, setCode] = useState(TWO_SUM_STARTER_CODE[language]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<SubmissionResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Update code when language changes
  const handleLanguageChange = (newLanguage: Language) => {
    setLanguage(newLanguage);
    setCode(TWO_SUM_STARTER_CODE[newLanguage]);
    setResult(null); // Clear previous results
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
        problemId: TWO_SUM_PROBLEM.id
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
                <h1 className="text-2xl font-bold text-white">{TWO_SUM_PROBLEM.title}</h1>
                <span className="px-3 py-1 bg-green-900 text-green-300 text-sm rounded-full">
                  {TWO_SUM_PROBLEM.difficulty}
                </span>
              </div>
            </div>

            {/* Description */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Description</h2>
              <p className="text-gray-300 whitespace-pre-line leading-relaxed">
                {TWO_SUM_PROBLEM.description}
              </p>
            </div>

            {/* Examples */}
            <div className="mb-6">
              <h2 className="text-lg font-semibold text-white mb-3">Examples</h2>
              {TWO_SUM_PROBLEM.examples.map((example, index) => (
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
                {TWO_SUM_PROBLEM.constraints.map((constraint, index) => (
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
              {TWO_SUM_PROBLEM.testCases.map((testCase, index) => (
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
