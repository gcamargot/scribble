import { useEffect, useState } from 'react'
import { DiscordSDK } from '@discord/embedded-app-sdk'
import './App.css'

/**
 * Scribble - Discord LeetCode Activity
 *
 * This is the main App component that initializes the Discord SDK
 * and provides authentication context to the entire application.
 */

interface DiscordUser {
  id: string
  username: string
  avatar?: string
}

function App() {
  const [discordUser, setDiscordUser] = useState<DiscordUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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
          scope: ['identify', 'guilds'],
          redirect_uri: window.location.href
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

        const { user } = await response.json()
        setDiscordUser(user)

        // Authenticate with Discord using the access token
        await discordSdk.commands.authenticate({
          access_token: user.accessToken
        })

      } catch (err) {
        console.error('Discord initialization error:', err)
        setError(err instanceof Error ? err.message : 'Unknown error')
      } finally {
        setLoading(false)
      }
    }

    initializeDiscord()
  }, [])

  if (loading) {
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

  if (!discordUser) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-dark">
        <div className="text-center">
          <p className="text-white mb-4">Waiting for authentication...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-dark min-h-screen text-white">
      <header className="bg-darker border-b border-gray-700 p-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold text-primary">Scribble</h1>
          <div className="flex items-center gap-4">
            {discordUser.avatar && (
              <img
                src={discordUser.avatar}
                alt={discordUser.username}
                className="w-8 h-8 rounded-full"
              />
            )}
            <span className="text-gray-300">{discordUser.username}</span>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto p-4">
        {/* TODO: Add main application content here */}
        <div className="mt-8 p-6 bg-darker rounded-lg border border-gray-700">
          <h2 className="text-xl font-bold mb-4">Welcome to Scribble!</h2>
          <p className="text-gray-300 mb-4">
            This is where your daily coding challenges will appear.
          </p>
          <p className="text-sm text-gray-400">
            The main UI components will be added as development progresses.
            Check the issue tracker for tasks: <code>bd ready</code>
          </p>
        </div>
      </main>
    </div>
  )
}

export default App
