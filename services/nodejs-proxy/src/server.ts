import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import morgan from 'morgan';
import helmet from 'helmet';

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
app.use(express.static('public'));

// Health check endpoint
app.get('/api/health', (req: Request, res: Response) => {
  res.json({
    status: 'ok',
    service: 'scribble-nodejs-proxy',
    timestamp: new Date().toISOString()
  });
});

// TODO: Add routes
// - Authentication routes (/api/auth/*)
// - Problems routes (/api/problems/*)
// - Submissions routes (/api/submissions/*)
// - Leaderboards routes (/api/leaderboards/*)
// - Proxy to Go backend

// Error handling middleware
app.use((err: any, req: Request, res: Response, next: any) => {
  console.error('Error:', err);
  res.status(err.status || 500).json({
    error: err.message || 'Internal server error'
  });
});

// 404 handler
app.use((req: Request, res: Response) => {
  res.status(404).json({
    error: 'Route not found'
  });
});

// Start server
app.listen(port, () => {
  console.log(`[${new Date().toISOString()}] Server running on http://localhost:${port}`);
  console.log(`Environment: ${process.env.NODE_ENV || 'development'}`);
});

export default app;
