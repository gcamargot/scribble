import { Router, Request, Response } from 'express';
import axios from 'axios';
import jwt from 'jsonwebtoken';

const router = Router();

// Discord OAuth2 configuration from environment
const DISCORD_CLIENT_ID = process.env.DISCORD_CLIENT_ID;
const DISCORD_CLIENT_SECRET = process.env.DISCORD_CLIENT_SECRET;
const DISCORD_REDIRECT_URI = process.env.DISCORD_REDIRECT_URI;
const JWT_SECRET = process.env.JWT_SECRET;

// Discord API endpoints
const DISCORD_TOKEN_URL = 'https://discord.com/api/oauth2/token';
const DISCORD_USER_URL = 'https://discord.com/api/users/@me';

interface DiscordTokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token: string;
  scope: string;
}

interface DiscordUser {
  id: string;
  username: string;
  discriminator: string;
  avatar: string | null;
  email?: string;
}

/**
 * POST /api/auth/discord/callback
 *
 * Exchange Discord authorization code for access token,
 * fetch user information, generate JWT, and set secure cookie.
 */
router.post('/discord/callback', async (req: Request, res: Response) => {
  try {
    const { code } = req.body;

    if (!code) {
      return res.status(400).json({ error: 'Authorization code is required' });
    }

    // Validate environment variables
    if (!DISCORD_CLIENT_ID || !DISCORD_CLIENT_SECRET || !DISCORD_REDIRECT_URI || !JWT_SECRET) {
      console.error('Missing required environment variables for Discord OAuth2');
      return res.status(500).json({ error: 'Server configuration error' });
    }

    // Exchange authorization code for access token
    const tokenParams = new URLSearchParams({
      client_id: DISCORD_CLIENT_ID,
      client_secret: DISCORD_CLIENT_SECRET,
      grant_type: 'authorization_code',
      code: code,
      redirect_uri: DISCORD_REDIRECT_URI
    });

    const tokenResponse = await axios.post<DiscordTokenResponse>(
      DISCORD_TOKEN_URL,
      tokenParams.toString(),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }
    );

    const { access_token } = tokenResponse.data;

    // Fetch user information from Discord
    const userResponse = await axios.get<DiscordUser>(DISCORD_USER_URL, {
      headers: {
        Authorization: `Bearer ${access_token}`
      }
    });

    const discordUser = userResponse.data;

    // Generate JWT token with 7-day expiry
    const jwtPayload = {
      userId: discordUser.id,
      username: discordUser.username,
      discriminator: discordUser.discriminator,
      avatar: discordUser.avatar
    };

    const token = jwt.sign(jwtPayload, JWT_SECRET, {
      expiresIn: '7d'
    });

    // Set HTTP-only cookie for security
    res.cookie('scribble_auth', token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 7 * 24 * 60 * 60 * 1000 // 7 days in milliseconds
    });

    // Return user information to frontend
    // NOTE: Access token is NOT sent to frontend for security
    // The token is only used server-side and stored in HTTP-only cookie
    return res.json({
      success: true,
      user: {
        id: discordUser.id,
        username: discordUser.username,
        discriminator: discordUser.discriminator,
        avatar: discordUser.avatar
          ? `https://cdn.discordapp.com/avatars/${discordUser.id}/${discordUser.avatar}.png`
          : null
      }
    });
  } catch (error) {
    console.error('Discord OAuth2 callback error:', error);

    if (axios.isAxiosError(error)) {
      // Log detailed error information for debugging
      console.error('Discord API error:', {
        status: error.response?.status,
        data: error.response?.data
      });

      return res.status(error.response?.status || 500).json({
        error: 'Failed to authenticate with Discord',
        details: error.response?.data
      });
    }

    return res.status(500).json({
      error: 'Internal server error during authentication'
    });
  }
});

/**
 * POST /api/auth/logout
 *
 * Clear authentication cookie
 */
router.post('/logout', (_req: Request, res: Response) => {
  res.clearCookie('scribble_auth');
  return res.json({ success: true, message: 'Logged out successfully' });
});

/**
 * GET /api/auth/me
 *
 * Get current authenticated user from JWT cookie
 */
router.get('/me', (req: Request, res: Response) => {
  const token = req.cookies?.scribble_auth;

  if (!token) {
    return res.status(401).json({ error: 'Not authenticated' });
  }

  if (!JWT_SECRET) {
    return res.status(500).json({ error: 'Server configuration error' });
  }

  try {
    const decoded = jwt.verify(token, JWT_SECRET);
    return res.json({ user: decoded });
  } catch (error) {
    console.error('JWT verification error:', error);
    return res.status(401).json({ error: 'Invalid or expired token' });
  }
});

export default router;
