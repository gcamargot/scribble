import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import morgan from 'morgan';
import helmet from 'helmet';
import cookieParser from 'cookie-parser';
import authRouter from './routes/auth.js';
import submissionsRouter from './routes/submissions.js';
import problemsRouter from './routes/problems.js';
import { initializeProxy } from './middleware/proxy.js';

// Load environment variables
dotenv.config();

const app: Express = express();
const port = process.env.NODE_PORT || 3000;

// Middleware
app.use(helmet());
app.use(cors({
  origin: process.env.NODE_ENV === 'production'
    ? 'https://discordsays.com'
    : 'http://localhost:5173',
  credentials: true
}));
app.use(morgan('combined'));
app.use(express.json());
app.use(cookieParser());
app.use(express.static('public'));

// Health check endpoint
app.get('/api/health', (_req: Request, res: Response) => {
  res.json({
    status: 'ok',
    service: 'scribble-nodejs-proxy',
    timestamp: new Date().toISOString()
  });
});

// Routes
app.use('/api/auth', authRouter);
app.use('/api/submissions', submissionsRouter);
app.use('/api/problems', problemsRouter);

// Error handling middleware
// eslint-disable-next-line @typescript-eslint/no-explicit-any
app.use((err: any, _req: Request, res: Response, _next: any) => {
  console.error('Error:', err);
  res.status(err.status || 500).json({
    error: err.message || 'Internal server error'
  });
});

// 404 handler
app.use((_req: Request, res: Response) => {
  res.status(404).json({
    error: 'Route not found'
  });
});

// Initialize proxy and start server
async function start() {
  await initializeProxy();
  app.listen(port, () => {
    console.log(`[${new Date().toISOString()}] Server running on http://localhost:${port}`);
    console.log(`Environment: ${process.env.NODE_ENV || 'development'}`);
  });
}

start().catch(err => {
  console.error('Failed to start server:', err);
  process.exit(1);
});

export default app;
