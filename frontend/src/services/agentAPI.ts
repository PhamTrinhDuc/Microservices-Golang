import axios from 'axios'
import { sessionManager } from '../utils/sessionManager'
import { tokenManager } from '../utils/tokenManager' // 1. Import tokenManager

// Agent server runs on port 8000
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

export const agentAPI = {
  chat: async (message: string, sessionId?: string): Promise<ChatResponse> => {
    const activeSessionId = sessionId || sessionManager.getSessionId()
    const token = tokenManager.getToken() // 2. Lấy token hiện tại

    const response = await axios.post<ChatResponse>(
      `${AGENT_API_URL}/chat`,
      {
        session_id: activeSessionId,
        message: message,
      },
      {
        headers: token ? { Authorization: `Bearer ${token}` } : {}, // 3. Truyền token vào header
      }
    )

    return response.data
  },

  confirm: async (
    sessionId: string,
    confirmationId: string,
    confirmed: boolean,
    hint?: string,
    payload?: any
  ): Promise<ChatResponse> => {
    const token = tokenManager.getToken() // 2. Lấy token hiện tại

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
        headers: token ? { Authorization: `Bearer ${token}` } : {}, // 3. Truyền token vào header
      }
    )
    return response.data
  }
}
