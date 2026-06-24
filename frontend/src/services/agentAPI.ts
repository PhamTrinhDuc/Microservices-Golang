import axios from 'axios'
import { sessionManager } from '../utils/sessionManager'
import { tokenManager } from '../utils/tokenManager'

// Agent server runs on port 8080
const AGENT_API_URL = 'http://localhost:8080/api'

export interface ChatResponse {
  session_id: string
  message: string
  steps?: string[]
  requires_confirmation?: boolean
  confirmation_id?: string
  hint?: string
  payload?: any
}

export interface StreamCallbacks {
  onSession?: (sessionId: string) => void
  onToken?: (token: string) => void
  onStep?: (message: string, tool: string) => void
  onConfirmation?: (confirmationId: string, hint: string) => void
  onError?: (error: string) => void
  onDone?: () => void
}

export const agentAPI = {
  chat: async (message: string, sessionId?: string): Promise<ChatResponse> => {
    const activeSessionId = sessionId || sessionManager.getSessionId()
    const token = tokenManager.getToken()

    const response = await axios.post<ChatResponse>(
      `${AGENT_API_URL}/chat`,
      {
        session_id: activeSessionId,
        message: message,
      },
      {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      }
    )
    console.log(response.data.message)

    return response.data
  },

  confirm: async (
    sessionId: string,
    confirmationId: string,
    confirmed: boolean,
    hint?: string,
    payload?: any
  ): Promise<ChatResponse> => {
    const token = tokenManager.getToken()

    const response = await axios.post<ChatResponse>(
      `${AGENT_API_URL}/chat/confirm`,
      {
        session_id: sessionId,
        confirmation_id: confirmationId,
        confirmed: confirmed,
        hint: hint,
        payload: payload,
      },
      {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      }
    )
    return response.data
  },

  chatStream: async (
    message: string,
    callbacks: StreamCallbacks,
    sessionId?: string,
    signal?: AbortSignal
  ): Promise<void> => {
    const activeSessionId = sessionId || sessionManager.getSessionId()
    const token = tokenManager.getToken()

    const response = await fetch(`${AGENT_API_URL}/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        session_id: activeSessionId,
        message: message,
      }),
      signal: signal,
    })

    if (!response.body) {
      throw new Error('Readable stream not supported.')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        const messages = buffer.split('\n\n')
        buffer = messages.pop() || ''

        for (const rawMsg of messages) {
          if (!rawMsg.trim()) continue

          let eventName = 'message'
          let dataStr = ''

          const lines = rawMsg.split('\n')
          for (const line of lines) {
            if (line.startsWith('event:')) {
              eventName = line.substring(6).trim()
            } else if (line.startsWith('data:')) {
              dataStr = line.substring(5).trim()
            }
          }

          if (!dataStr) continue
          try {
            const data = JSON.parse(dataStr)
            console.log("agent streaming: ", data)
            switch (eventName) {
              case 'session':
                if (callbacks.onSession && data.session_id) callbacks.onSession(data.session_id)
                break
              case 'token':
                if (callbacks.onToken && data.text) callbacks.onToken(data.text)
                break
              case 'step':
                if (callbacks.onStep) callbacks.onStep(data.message, data.tool)
                break
              case 'confirmation':
                if (callbacks.onConfirmation) callbacks.onConfirmation(data.confirmation_id, data.hint)
                break
              case 'error':
                if (callbacks.onError) callbacks.onError(data.error)
                break
              case 'done':
                if (callbacks.onDone) callbacks.onDone()
                break
            }
          } catch (e) {
            console.error('Failed to parse SSE data:', dataStr, e)
          }
        }
      }
    } finally {
      reader.releaseLock()
    }
  },

  confirmStream: async (
    sessionId: string,
    confirmationId: string,
    confirmed: boolean,
    callbacks: StreamCallbacks,
    hint?: string,
    payload?: any,
    signal?: AbortSignal
  ): Promise<void> => {
    const token = tokenManager.getToken()

    const response = await fetch(`${AGENT_API_URL}/chat/confirm/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        session_id: sessionId,
        confirmation_id: confirmationId,
        confirmed: confirmed,
        hint: hint,
        payload: payload,
      }),
      signal: signal,
    })

    if (!response.body) {
      throw new Error('Readable stream not supported.')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        const messages = buffer.split('\n\n')
        buffer = messages.pop() || ''

        for (const rawMsg of messages) {
          if (!rawMsg.trim()) continue

          let eventName = 'message'
          let dataStr = ''

          const lines = rawMsg.split('\n')
          for (const line of lines) {
            if (line.startsWith('event:')) {
              eventName = line.substring(6).trim()
            } else if (line.startsWith('data:')) {
              dataStr = line.substring(5).trim()
            }
          }

          if (!dataStr) continue
          try {
            const data = JSON.parse(dataStr)
            switch (eventName) {
              case 'token':
                if (callbacks.onToken && data.text) callbacks.onToken(data.text)
                break
              case 'step':
                if (callbacks.onStep) callbacks.onStep(data.message, data.tool)
                break
              case 'confirmation':
                if (callbacks.onConfirmation) callbacks.onConfirmation(data.confirmation_id, data.hint)
                break
              case 'error':
                if (callbacks.onError) callbacks.onError(data.error)
                break
              case 'done':
                if (callbacks.onDone) callbacks.onDone()
                break
            }
          } catch (e) {
            console.error('Failed to parse SSE data:', dataStr, e)
          }
        }
      }
    } finally {
      reader.releaseLock()
    }
  }
}
