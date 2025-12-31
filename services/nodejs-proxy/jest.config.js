/** @type {import('jest').Config} */
export default {
  // Use ts-jest for TypeScript support
  preset: 'ts-jest/presets/default-esm',

  // Test environment - Node.js for backend testing
  testEnvironment: 'node',

  // File extensions to look for
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],

  // Test file patterns
  testMatch: ['**/__tests__/**/*.test.ts', '**/*.test.ts'],

  // Transform TypeScript files with ts-jest
  transform: {
    '^.+\\.tsx?$': [
      'ts-jest',
      {
        useESM: true,
      },
    ],
  },

  // ESM module support
  extensionsToTreatAsEsm: ['.ts'],
  moduleNameMapper: {
    '^(\\.{1,2}/.*)\\.js$': '$1',
  },

  // Inject jest globals for ESM
  injectGlobals: true,

  // Coverage configuration
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.d.ts',
  ],

  // Setup files to run before tests
  setupFilesAfterEnv: [],

  // Clear mocks between tests
  clearMocks: true,

  // Timeout for tests (10 seconds)
  testTimeout: 10000,
};
