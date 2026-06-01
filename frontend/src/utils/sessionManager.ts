const SESSION_KEY = 'guest_session_id'

export const sessionManager = {
  getSessionId: (): string => {
    let sessionID = localStorage.getItem(SESSION_KEY)
    if (!sessionID) {
      sessionID = 'guest-' + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
      localStorage.setItem(SESSION_KEY, sessionID)
    }
    return sessionID
  },
  clearSessionId: (): void => {
    localStorage.removeItem(SESSION_KEY)
  }
}
