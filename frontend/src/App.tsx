import { useEffect } from 'react'
import { DiscordSDK } from '@discord/embedded-app-sdk'
import { useAuthStore } from './stores/authStore'
import ProblemPage from './pages/ProblemPage'
import './App.css'

/**
 * Scribble - Discord LeetCode Activity
 *
 * This is the main App component that initializes the Discord SDK
 * and provides authentication context to the entire application.
 */

function App() {
  const { user, isLoading, error, setUser, setLoading, setError } = useAuthStore()

  useEffect(() => {
    /**
     * Initialize Discord SDK and authenticate user
     *
     * This follows the pattern from Discord's getting-started-activity:
     * 1. Create DiscordSDK instance with client ID
     * 2. Wait for Discord client to be ready
     * 3. Authorize with scopes
     * 4. Exchange code for token
     * 5. Authenticate with token
     */
    const initializeDiscord = async () => {
      try {
        setLoading(true)

        // Get Discord client ID from environment
        const clientId = import.meta.env.VITE_DISCORD_CLIENT_ID
        if (!clientId) {
          throw new Error('Discord Client ID not configured')
        }

        // Initialize Discord SDK
        const discordSdk = new DiscordSDK(clientId)

        // Wait for Discord client
        await discordSdk.ready()

        // Get authorization code from Discord
        const { code } = await discordSdk.commands.authorize({
          client_id: clientId,
          response_type: 'code',
          state: '',
          prompt: 'none',
          scope: ['identify', 'guilds']
        })

        // Send code to our backend to exchange for access token
        const response = await fetch('/api/auth/discord/callback', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ code })
        })

        if (!response.ok) {
          throw new Error('Failed to authenticate with backend')
        }

        const { user: discordUser } = await response.json()
        setUser(discordUser)

        // Authenticate with Discord using the access token
        await discordSdk.commands.authenticate({
          access_token: discordUser.accessToken
        })

      } catch (err) {
        console.error('Discord initialization error:', err)
        setError(err instanceof Error ? err.message : 'Unknown error')
      } finally {
        setLoading(false)
      }
    }

    initializeDiscord()
  }, [setUser, setLoading, setError])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-dark">
        <div className="text-white text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary mx-auto mb-4"></div>
          <p>Loading Scribble...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-dark">
        <div className="text-center bg-error bg-opacity-10 border border-error border-opacity-30 rounded-lg p-6 max-w-md">
          <h1 className="text-xl font-bold text-error mb-2">Authentication Error</h1>
          <p className="text-gray-300">{error}</p>
          <p className="text-sm text-gray-400 mt-4">
            Please make sure Discord is configured correctly and you have the necessary permissions.
          </p>
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-dark">
        <div className="text-center">
          <p className="text-white mb-4">Waiting for authentication...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-dark h-screen flex flex-col text-white">
      <header className="bg-darker border-b border-gray-700 p-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold text-primary">Scribble</h1>
          <div className="flex items-center gap-4">
            {user.avatar && (
              <img
                src={user.avatar}
                alt={user.username}
                className="w-8 h-8 rounded-full"
              />
            )}
            <span className="text-gray-300">{user.username}</span>
          </div>
        </div>
      </header>

      <main className="flex-1 overflow-hidden">
        <ProblemPage />
      </main>
    </div>
  )
}

export default App
