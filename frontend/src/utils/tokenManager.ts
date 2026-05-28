const TOKEN_KEY = 'auth_token'

const decodePayload = (token: string): { exp?: number } | null => {
  try {
    const payload = token.split('.')[1]
    if (!payload) {
      return null
    }

    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((char) => `%${`00${char.charCodeAt(0).toString(16)}`.slice(-2)}`)
        .join(''),
    )

    return JSON.parse(jsonPayload) as { exp?: number }
  } catch {
    return null
  }
}

export const tokenManager = {
  getToken: (): string | null => localStorage.getItem(TOKEN_KEY),
  setToken: (token: string): void => {
    localStorage.setItem(TOKEN_KEY, token)
  },
  removeToken: (): void => {
    localStorage.removeItem(TOKEN_KEY)
  },
  isTokenExpired: (token: string): boolean => {
    const payload = decodePayload(token)
    if (!payload?.exp) {
      return true
    }

    return payload.exp <= Math.floor(Date.now() / 1000)
  },
  isAuthenticated: (): boolean => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      return false
    }

    if (tokenManager.isTokenExpired(token)) {
      tokenManager.removeToken()
      return false
    }

    return true
  },
}
