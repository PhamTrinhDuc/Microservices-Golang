import React, { useState, useEffect, useRef } from 'react'
import { agentAPI } from '../services/agentAPI'
import { sessionManager } from '../utils/sessionManager'
import MarkdownRenderer from './MarkdownRenderer'

interface Message {
  id: string
  sender: 'user' | 'bot'
  text: string
  steps?: string[]
  requiresConfirmation?: boolean
  confirmationId?: string
  hint?: string
  payload?: any
}

const getFriendlyErrorMessage = (rawError: string): string => {
  const errLower = rawError.toLowerCase()
  if (
    errLower.includes('limit') ||
    errLower.includes('too large') ||
    errLower.includes('rate limit') ||
    errLower.includes('413') ||
    errLower.includes('tpm') ||
    errLower.includes('rpm')
  ) {
    return '⚠️ Xin lỗi bạn, hệ thống AI đang quá tải giới hạn xử lý (Rate Limit/Context Limit). Vui lòng thử lại sau ít phút hoặc rút ngắn câu hỏi hơn nhé!'
  }
  if (
    errLower.includes('api request failed') ||
    errLower.includes('connection') ||
    errLower.includes('failed to run') ||
    errLower.includes('request failed')
  ) {
    return '⚠️ Jiyuu không thể kết nối tới dịch vụ lúc này. Bạn vui lòng kiểm tra kết nối mạng hoặc thử lại sau nhé!'
  }
  return '⚠️ Jiyuu gặp sự cố ngoài ý muốn khi xử lý yêu cầu. Vui lòng thử lại sau.'
}

export const JiyuuChat: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false)
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'welcome',
      sender: 'bot',
      text: 'Xin chào! Mình là Jiyuu 🤖, trợ lý mua sắm của shop Belli. Mình có thể giúp bạn tìm kiếm sản phẩm, kiểm tra thông số kỹ thuật hoặc chính sách mua hàng. Bạn muốn tìm gì hôm nay?',
    },
  ])
  const [inputValue, setInputValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showStepsIndex, setShowStepsIndex] = useState<Record<string, boolean>>({})
  
  // Pending confirmation state
  const [pendingConfirm, setPendingConfirm] = useState<{
    messageId: string
    confirmationId: string
    hint: string
    payload: any
  } | null>(null)

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, isLoading])

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }
    }
  }, [])

  const toggleSteps = (msgId: string) => {
    setShowStepsIndex((prev) => ({
      ...prev,
      [msgId]: !prev[msgId],
    }))
  }

  const handleCancelResponse = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }
    setIsLoading(false)
    
    setMessages((prev) => {
      if (prev.length === 0) return prev
      const lastMsg = prev[prev.length - 1]
      if (lastMsg.sender === 'bot') {
        if (lastMsg.text === '') {
          return prev.slice(0, -1).concat({
            ...lastMsg,
            text: '*(Đã dừng phản hồi)*',
          })
        } else {
          return prev.slice(0, -1).concat({
            ...lastMsg,
            text: lastMsg.text + '\n\n*(Đã dừng phản hồi)*',
          })
        }
      }
      return prev
    })
  }

  const handleSendMessage = async (textToSend: string) => {
    if (!textToSend.trim() || isLoading) return

    const userMsgId = `user-${Date.now()}`
    const userMessage: Message = {
      id: userMsgId,
      sender: 'user',
      text: textToSend,
    }

    setMessages((prev) => [...prev, userMessage])
    setInputValue('')
    setIsLoading(true)
    setPendingConfirm(null) // Reset pending confirm

    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    const controller = new AbortController()
    abortControllerRef.current = controller

    const botMsgId = `bot-${Date.now()}`
    // Thêm placeholder tin nhắn của bot để bắt đầu stream
    setMessages((prev) => [
      ...prev,
      {
        id: botMsgId,
        sender: 'bot',
        text: '',
        steps: [],
      },
    ])

    // Tự động mở danh sách các bước xử lý của bot
    setShowStepsIndex((prev) => ({
      ...prev,
      [botMsgId]: true,
    }))

    try {
      const sessionId = sessionManager.getSessionId()
      await agentAPI.chatStream(
        textToSend,
        {
          onToken: (token) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId ? { ...msg, text: msg.text + token } : msg
              )
            )
          },
          onStep: (step) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? { ...msg, steps: msg.steps ? [...msg.steps, step] : [step] }
                  : msg
              )
            )
          },
          onConfirmation: (confirmationId, hint) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? {
                      ...msg,
                      requiresConfirmation: true,
                      confirmationId: confirmationId,
                      hint: hint,
                    }
                  : msg
              )
            )
            setPendingConfirm({
              messageId: botMsgId,
              confirmationId: confirmationId,
              hint: hint,
              payload: null,
            })
          },
          onError: (err) => {
            if (err === 'AbortError' || err.includes('aborted')) return
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? { ...msg, text: getFriendlyErrorMessage(err) }
                  : msg
              )
            )
            setIsLoading(false)
            abortControllerRef.current = null
          },
          onDone: () => {
            setIsLoading(false)
            abortControllerRef.current = null
          },
        },
        sessionId,
        controller.signal
      )
    } catch (error: any) {
      if (error.name === 'AbortError' || error.message?.includes('aborted')) {
        return
      }
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === botMsgId
            ? {
                ...msg,
                text: getFriendlyErrorMessage(error.message || ''),
              }
            : msg
        )
      )
      setIsLoading(false)
    }
  }

  const handleConfirmation = async (confirmed: boolean) => {
    if (!pendingConfirm || isLoading) return

    const { confirmationId, hint, payload } = pendingConfirm
    setIsLoading(true)
    setPendingConfirm(null) // Ẩn nút ngay lập tức

    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    const controller = new AbortController()
    abortControllerRef.current = controller

    const actionText = confirmed ? '👍 Bạn đã xác nhận đồng ý.' : '👎 Bạn đã từ chối xác nhận.'
    setMessages((prev) => [
      ...prev,
      {
        id: `confirm-action-${Date.now()}`,
        sender: 'user',
        text: actionText,
      },
    ])

    const botMsgId = `bot-${Date.now()}`
    // Thêm tin nhắn bot mới để bắt đầu stream phản hồi từ sau khi xác nhận
    setMessages((prev) => [
      ...prev,
      {
        id: botMsgId,
        sender: 'bot',
        text: '',
        steps: [],
      },
    ])

    setShowStepsIndex((prev) => ({
      ...prev,
      [botMsgId]: true,
    }))

    try {
      const sessionId = sessionManager.getSessionId()
      await agentAPI.confirmStream(
        sessionId,
        confirmationId,
        confirmed,
        {
          onToken: (token) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId ? { ...msg, text: msg.text + token } : msg
              )
            )
          },
          onStep: (step) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? { ...msg, steps: msg.steps ? [...msg.steps, step] : [step] }
                  : msg
              )
            )
          },
          onConfirmation: (nextConfId, nextHint) => {
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? {
                      ...msg,
                      requiresConfirmation: true,
                      confirmationId: nextConfId,
                      hint: nextHint,
                    }
                  : msg
              )
            )
            setPendingConfirm({
              messageId: botMsgId,
              confirmationId: nextConfId,
              hint: nextHint,
              payload: null,
            })
          },
          onError: (err) => {
            if (err === 'AbortError' || err.includes('aborted')) return
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === botMsgId
                  ? { ...msg, text: getFriendlyErrorMessage(err) }
                  : msg
              )
            )
            setIsLoading(false)
            abortControllerRef.current = null
          },
          onDone: () => {
            setIsLoading(false)
            abortControllerRef.current = null
          },
        },
        hint,
        payload,
        controller.signal
      )
    } catch (error: any) {
      if (error.name === 'AbortError' || error.message?.includes('aborted')) {
        return
      }
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === botMsgId
            ? {
                ...msg,
                text: getFriendlyErrorMessage(error.message || ''),
              }
            : msg
        )
      )
      setIsLoading(false)
    }
  }

  const quickReplies = [
    'Đồng hồ thông minh < 3 triệu có GPS chạy bộ',
    'Laptop gaming tầm 20 triệu',
    'Chính sách trả hàng và hoàn tiền',
  ]

  return (
    <div className="fixed bottom-6 right-6 z-50 font-sans">
      {/* Floating Trigger Button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="group relative flex h-14 w-14 items-center justify-center rounded-full bg-slate-900 border border-slate-700/80 text-white shadow-[0_0_30px_rgba(37,99,235,0.3)] hover:shadow-[0_0_40px_rgba(99,102,241,0.5)] transition-all duration-300 transform hover:scale-105"
        >
          {/* Glowing Status Ring */}
          <span className="absolute -top-0.5 -right-0.5 flex h-3.5 w-3.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3.5 w-3.5 bg-emerald-500 shadow-[0_0_8px_#10b981]"></span>
          </span>
          <svg
            className="h-7 w-7 text-indigo-400 group-hover:text-indigo-300 transition-colors"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={1.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z"
            />
          </svg>
          {/* Hover tooltip */}
          <div className="absolute right-16 scale-0 group-hover:scale-100 transition-all origin-right bg-slate-900 border border-slate-800 text-xs text-slate-200 px-3 py-1.5 rounded-lg shadow-xl whitespace-nowrap">
            Chat với Jiyuu 🤖
          </div>
        </button>
      )}

      {/* Chat Window Container */}
      {isOpen && (
        <div className="flex h-[580px] w-[400px] flex-col rounded-2xl bg-[#0b0c10]/95 border border-[#1f2230] shadow-[0_12px_40px_rgba(0,0,0,0.6)] backdrop-blur-xl overflow-hidden transition-all duration-300 animate-in fade-in zoom-in-95 slide-in-from-bottom-5">
          {/* Header */}
          <div className="flex items-center justify-between bg-gradient-to-r from-[#141622] to-[#0c0d15] px-4 py-3 border-b border-[#1f2230]">
            <div className="flex items-center space-x-3">
              <div className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-600 shadow-[0_0_15px_rgba(99,102,241,0.4)]">
                <span className="text-lg font-bold text-white">J</span>
                <span className="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full bg-emerald-500 ring-2 ring-[#0c0d15] shadow-[0_0_6px_#10b981]"></span>
              </div>
              <div>
                <h3 className="text-sm font-semibold text-slate-100">Jiyuu</h3>
                <span className="text-[10px] text-indigo-400 font-medium">Trợ lý mua sắm Belli</span>
              </div>
            </div>
            
            {/* Close Button */}
            <button
              onClick={() => setIsOpen(false)}
              className="rounded-lg p-1.5 text-slate-400 hover:bg-[#1a1c29] hover:text-slate-100 transition-all"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Messages Container */}
          <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4 scrollbar-thin scrollbar-thumb-slate-800 scrollbar-track-transparent">
            {messages.map((msg, index) => {
              const isBot = msg.sender === 'bot'
              return (
                <div key={msg.id || index} className={`flex flex-col ${isBot ? 'items-start' : 'items-end'}`}>
                  
                  {/* Step Journey List (If it has steps, show before text response) */}
                  {isBot && msg.steps && msg.steps.length > 0 && (
                    <div className="w-full max-w-[85%] mb-2 rounded-xl bg-[#141620]/60 border border-[#202334] overflow-hidden text-xs">
                      <button
                        onClick={() => toggleSteps(msg.id)}
                        className="flex w-full items-center justify-between px-3 py-2 text-slate-400 hover:text-slate-200 transition-colors"
                      >
                        <div className="flex items-center space-x-2">
                          <svg className="h-4 w-4 text-indigo-400 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                            <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                          </svg>
                          <span className="font-semibold">Hành trình xử lý của Agent ({msg.steps.length} bước)</span>
                        </div>
                        <svg
                          className={`h-4 w-4 transition-transform duration-200 ${showStepsIndex[msg.id] ? 'rotate-180' : ''}`}
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          strokeWidth={2}
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                        </svg>
                      </button>
                      
                      {showStepsIndex[msg.id] && (
                        <div className="border-t border-[#1e202f] bg-[#0d0e15] px-3 py-2.5 space-y-2 animate-in slide-in-from-top-1">
                          {msg.steps.map((step, idx) => (
                            <div key={idx} className="flex items-start space-x-2 text-slate-300">
                              <span className="text-emerald-400 mt-0.5 font-bold">✓</span>
                              <span className="font-medium text-slate-300">{step}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {/* Message Bubble */}
                  <div
                    className={`rounded-2xl px-4 py-2.5 text-sm max-w-[85%] leading-relaxed ${
                      isBot
                        ? 'bg-[#151722] text-slate-200 border border-[#232738] rounded-tl-none shadow-md'
                        : 'bg-gradient-to-br from-indigo-600 to-violet-600 text-white rounded-tr-none shadow-[0_4px_15px_rgba(99,102,241,0.25)]'
                    }`}
                  >
                    {isBot ? (
                      <MarkdownRenderer content={msg.text} />
                    ) : (
                      <div className="whitespace-pre-line font-normal">
                        {msg.text}
                      </div>
                    )}
                  </div>

                  {/* Confirmation Required Box */}
                  {isBot && msg.requiresConfirmation && pendingConfirm && pendingConfirm.messageId === msg.id && (
                    <div className="mt-3 w-full max-w-[85%] rounded-xl border border-amber-500/30 bg-amber-500/5 p-3 animate-in fade-in slide-in-from-top-2">
                      <div className="flex items-center space-x-2 mb-2">
                        <svg className="h-5 w-5 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                        </svg>
                        <h4 className="text-xs font-bold text-amber-400 uppercase tracking-wider">Yêu cầu xác nhận từ hệ thống</h4>
                      </div>
                      <p className="text-xs text-slate-300 mb-3 bg-[#11121b] p-2 rounded border border-slate-800 font-mono">
                        {msg.hint || 'Hệ thống cần bạn xác nhận để thực hiện tác vụ này.'}
                      </p>
                      <div className="flex space-x-2">
                        <button
                          onClick={() => handleConfirmation(true)}
                          className="flex-1 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-md transition-colors"
                        >
                          Đồng ý
                        </button>
                        <button
                          onClick={() => handleConfirmation(false)}
                          className="flex-1 py-1.5 rounded-lg bg-rose-600/20 border border-rose-500/30 hover:bg-rose-600/40 text-rose-300 text-xs font-semibold transition-colors"
                        >
                          Từ chối
                        </button>
                      </div>
                    </div>
                  )}

                </div>
              )
            })}

            {/* Bouncing Typing Loader */}
            {isLoading && (
              <div className="flex flex-col items-start">
                <div className="rounded-2xl rounded-tl-none bg-[#151722] border border-[#232738] px-4 py-3 shadow-md">
                  <div className="flex items-center space-x-1.5 py-1">
                    <div className="h-2 w-2 animate-bounce rounded-full bg-indigo-500 [animation-delay:-0.3s]"></div>
                    <div className="h-2 w-2 animate-bounce rounded-full bg-indigo-400 [animation-delay:-0.15s]"></div>
                    <div className="h-2 w-2 animate-bounce rounded-full bg-indigo-300"></div>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Quick Replies / suggestions */}
          {messages.length === 1 && !isLoading && (
            <div className="px-4 py-2 space-y-1.5 bg-[#08090f]/50 border-t border-[#131520]">
              <span className="text-[10px] text-slate-400 font-semibold uppercase tracking-wider block">Gợi ý câu hỏi:</span>
              <div className="flex flex-col space-y-1.5">
                {quickReplies.map((reply, i) => (
                  <button
                    key={i}
                    onClick={() => handleSendMessage(reply)}
                    className="w-full text-left text-xs bg-[#12141f] border border-[#202230] hover:border-indigo-500/50 hover:bg-[#181b2a] text-indigo-300 hover:text-indigo-200 px-3 py-1.5 rounded-lg transition-all truncate"
                  >
                    {reply}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Input Panel */}
          <form
            onSubmit={(e) => {
              e.preventDefault()
              handleSendMessage(inputValue)
            }}
            className="flex items-center space-x-2 bg-[#0c0d14] px-3 py-3 border-t border-[#1a1c29]"
          >
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              disabled={isLoading}
              placeholder={isLoading ? 'Agent đang suy nghĩ...' : 'Hỏi Jiyuu điều bạn muốn...'}
              className="flex-1 bg-[#12141f] border border-[#202334] focus:border-indigo-500/80 rounded-xl px-3 py-2 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500/80 disabled:opacity-50 transition-all"
            />
            {isLoading ? (
              <button
                type="button"
                onClick={handleCancelResponse}
                className="flex h-9 w-9 items-center justify-center rounded-xl bg-rose-600 hover:bg-rose-500 text-white transition-colors shadow-md shadow-rose-600/20 focus:outline-none"
                title="Dừng phản hồi"
              >
                <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                  <path fillRule="evenodd" d="M4.5 7.5a3 3 0 013-3h9a3 3 0 013 3v9a3 3 0 01-3 3h-9a3 3 0 01-3-3v-9z" clipRule="evenodd" />
                </svg>
              </button>
            ) : (
              <button
                type="submit"
                disabled={!inputValue.trim() || isLoading}
                className="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-600 hover:bg-indigo-500 disabled:bg-[#1a1c29] text-white disabled:text-slate-600 transition-colors shadow-md shadow-indigo-600/10 focus:outline-none"
              >
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                </svg>
              </button>
            )}
          </form>
        </div>
      )}
    </div>
  )
}

export default JiyuuChat
