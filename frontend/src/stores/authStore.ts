import { create } from 'zustand';

interface DiscordUser {
  id: string;
  username: string;
  discriminator: string;
  avatar: string | null;
  accessToken?: string;
}

interface AuthState {
  // State
  user: DiscordUser | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  setUser: (user: DiscordUser | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  logout: () => void;
  reset: () => void;
}

/**
 * Zustand store for authentication state
 *
 * Manages Discord user authentication, loading states, and errors.
 * Used throughout the app to access current user information.
 */
export const useAuthStore = create<AuthState>((set) => ({
  // Initial state
  user: null,
  isLoading: false,
  error: null,

  // Set authenticated user
  setUser: (user) => set({ user, error: null }),

  // Set loading state (used during OAuth flow)
  setLoading: (isLoading) => set({ isLoading }),

  // Set error message (used when authentication fails)
  setError: (error) => set({ error, isLoading: false }),

  // Logout user and clear state
  logout: () => set({ user: null, error: null, isLoading: false }),

  // Reset all state to initial values
  reset: () => set({ user: null, isLoading: false, error: null })
}));
