import React, { useEffect, useState } from 'react'
import { useAuth } from '../../hooks/useAuth'
import { userAPI } from '../../services/userAPI'
import type { User } from '../../types'

const getPageNumbers = (current: number, total: number) => {
  const pages: (number | string)[] = []
  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    pages.push(1)
    if (current > 3) {
      pages.push('...')
    }
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }
    if (current < total - 2) {
      pages.push('...')
    }
    pages.push(total)
  }
  return pages
}

export default function AdminUsersTab() {
  const { user: currentUser } = useAuth()
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [totalUsers, setTotalUsers] = useState(0)

  // Confirmation Modal State
  const [modalState, setModalState] = useState<{
    show: boolean
    userId: number
    userName: string
    isLock: boolean
  } | null>(null)

  const loadUsers = async (pageNumber: number, query: string) => {
    try {
      setLoading(true)
      const res = await userAPI.adminListUsers(pageNumber, 10, query)
      setUsers(res.data || [])
      setPage(res.pagination.page)
      setTotalPages(res.pagination.total_pages)
      setTotalUsers(res.pagination.total)
    } catch (err: any) {
      console.error(err)
      alert(err.message || 'Lỗi khi tải danh sách thành viên')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadUsers(1, '')
  }, [])

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    void loadUsers(1, searchQuery)
  }

  const triggerLockUnlock = (userId: number, userName: string, currentLock: boolean) => {
    setModalState({
      show: true,
      userId,
      userName,
      isLock: !currentLock,
    })
  }

  const executeLockUnlock = async () => {
    if (!modalState) return
    try {
      setLoading(true)
      await userAPI.adminLockUser(modalState.userId, modalState.isLock)
      alert(`Đã ${modalState.isLock ? 'khóa' : 'mở khóa'} tài khoản ${modalState.userName} thành công!`)
      setModalState(null)
      void loadUsers(page, searchQuery)
    } catch (err: any) {
      alert(err.message || 'Lỗi khi thay đổi trạng thái tài khoản')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Thành Viên</h1>
          <p className="text-xs text-neutral-500 mt-1">Xem danh sách, tìm kiếm và thay đổi trạng thái khóa tài khoản thành viên</p>
        </div>
      </div>

      {/* Search Input Bar */}
      <form onSubmit={handleSearchSubmit} className="flex gap-2 max-w-md">
        <input
          type="text"
          placeholder="Tìm kiếm theo email, họ tên..."
          className="flex-1 border border-neutral-300 rounded px-3 py-1.5 text-xs bg-white focus:outline-none focus:ring-1 focus:ring-black"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
        <button
          type="submit"
          className="bg-black hover:bg-neutral-850 text-white text-xs font-bold px-4 py-1.5 rounded transition-all"
        >
          Tìm kiếm
        </button>
        {searchQuery && (
          <button
            type="button"
            onClick={() => {
              setSearchQuery('')
              void loadUsers(1, '')
            }}
            className="border border-neutral-300 hover:bg-neutral-100 text-xs px-3 py-1.5 rounded transition-all"
          >
            Xóa
          </button>
        )}
      </form>

      {/* Users Table Card */}
      <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-600 font-bold uppercase tracking-wider text-[10px]">
                <th className="py-3.5 px-4 font-black">ID</th>
                <th className="py-3.5 px-4 font-black">Thành viên</th>
                <th className="py-3.5 px-4 font-black">Email</th>
                <th className="py-3.5 px-4 font-black">Vai trò</th>
                <th className="py-3.5 px-4 font-black">Xác thực</th>
                <th className="py-3.5 px-4 font-black">Trạng thái</th>
                <th className="py-3.5 px-4 font-black">Ngày tạo</th>
                <th className="py-3.5 px-4 font-black text-right">Thao tác</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-neutral-150">
              {loading && users.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-8 text-center text-neutral-400 font-medium">
                    Đang tải danh sách thành viên...
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-8 text-center text-neutral-400 font-medium">
                    Không tìm thấy thành viên nào.
                  </td>
                </tr>
              ) : (
                users.map((u) => (
                  <tr key={u.id} className="hover:bg-neutral-50/50 transition-colors">
                    <td className="py-3 px-4 font-mono font-bold text-neutral-500">{u.id}</td>
                    <td className="py-3 px-4 font-bold text-neutral-800">{u.full_name}</td>
                    <td className="py-3 px-4 text-neutral-600 font-medium">{u.email}</td>
                    <td className="py-3 px-4">
                      <span className={`inline-block px-2 py-0.5 rounded-[3px] text-[9px] font-bold uppercase tracking-wider ${
                        u.role === 'admin'
                          ? 'bg-neutral-900 text-white'
                          : 'bg-neutral-100 text-neutral-700 border border-neutral-200'
                      }`}>
                        {u.role}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={`inline-block px-1.5 py-0.5 rounded-[3px] text-[9px] font-bold uppercase tracking-wider ${
                        u.is_verified
                          ? 'bg-blue-50 text-blue-600 border border-blue-100'
                          : 'bg-neutral-50 text-neutral-500 border border-neutral-150'
                      }`}>
                        {u.is_verified ? 'Đã xác thực' : 'Chưa xác thực'}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <span className={`inline-block px-1.5 py-0.5 rounded-[3px] text-[9px] font-bold uppercase tracking-wider ${
                        u.is_lock
                          ? 'bg-red-50 text-red-650 border border-red-100'
                          : 'bg-green-50 text-green-700 border border-green-100'
                      }`}>
                        {u.is_lock ? 'Đã khóa' : 'Hoạt động'}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-neutral-500 font-medium">
                      {new Date(u.created_at).toLocaleDateString('vi-VN')}
                    </td>
                    <td className="py-3 px-4 text-right">
                      {currentUser?.id === u.id ? (
                        <span className="text-[10px] text-neutral-400 font-semibold italic">Tài khoản của bạn</span>
                      ) : (
                        <button
                          onClick={() => triggerLockUnlock(u.id, u.full_name || u.email, u.is_lock)}
                          className={`text-[10px] font-black uppercase tracking-wider px-2.5 py-1 rounded transition-colors ${
                            u.is_lock
                              ? 'bg-green-600 hover:bg-green-700 text-white'
                              : 'bg-red-600 hover:bg-red-700 text-white'
                          }`}
                        >
                          {u.is_lock ? 'Mở khóa' : 'Khóa'}
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Controls */}
        {totalPages > 1 && (
          <div className="bg-neutral-50 border-t border-neutral-200 px-4 py-3.5 flex flex-col sm:flex-row gap-4 items-center sm:justify-between">
            <div className="text-[11px] text-neutral-555 font-medium shrink-0">
              Hiển thị <span className="font-bold text-neutral-800">{users.length}</span> trên tổng số{' '}
              <span className="font-bold text-neutral-800">{totalUsers}</span> thành viên
            </div>
            <div className="flex items-center gap-1 flex-wrap justify-center">
              <button
                onClick={() => void loadUsers(page - 1, searchQuery)}
                disabled={page === 1 || loading}
                className="px-2.5 py-1 text-[10px] font-black uppercase tracking-wider border border-neutral-300 rounded hover:bg-neutral-100 bg-white disabled:opacity-40 disabled:hover:bg-transparent transition-all"
              >
                Trước
              </button>
              
              {getPageNumbers(page, totalPages).map((p, idx) => {
                if (p === '...') {
                  return (
                    <span key={`dots-${idx}`} className="px-2 py-1 text-[10px] font-semibold text-neutral-400">
                      ...
                    </span>
                  )
                }
                return (
                  <button
                    key={p}
                    onClick={() => void loadUsers(p as number, searchQuery)}
                    disabled={loading}
                    className={`px-2.5 py-1 text-[10px] font-bold rounded transition-all ${
                      page === p
                        ? 'bg-black text-white'
                        : 'border border-neutral-300 hover:bg-neutral-100 bg-white text-neutral-700'
                    }`}
                  >
                    {p}
                  </button>
                )
              })}

              <button
                onClick={() => void loadUsers(page + 1, searchQuery)}
                disabled={page === totalPages || loading}
                className="px-2.5 py-1 text-[10px] font-black uppercase tracking-wider border border-neutral-300 rounded hover:bg-neutral-100 bg-white disabled:opacity-40 disabled:hover:bg-transparent transition-all"
              >
                Sau
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Confirmation Modal */}
      {modalState?.show && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="bg-white border border-neutral-200 rounded-2xl shadow-2xl w-full max-w-sm p-6 space-y-4 text-center">
            <div className={`w-12 h-12 rounded-full flex items-center justify-center mx-auto border ${
              modalState.isLock
                ? 'bg-red-50 text-red-650 border-red-100'
                : 'bg-green-50 text-green-700 border-green-100'
            }`}>
              {modalState.isLock ? (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              ) : (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                </svg>
              )}
            </div>

            <div className="space-y-1.5">
              <h3 className="text-sm font-black uppercase tracking-wide text-neutral-900">
                Xác nhận {modalState.isLock ? 'Khóa' : 'Mở khóa'} tài khoản
              </h3>
              <p className="text-xs text-neutral-550 leading-relaxed font-semibold">
                Bạn có chắc chắn muốn {modalState.isLock ? 'đình chỉ hoạt động (khóa)' : 'mở khóa hoạt động'} của tài khoản{' '}
                <span className="font-bold text-neutral-800">{modalState.userName}</span> không?
              </p>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={() => setModalState(null)}
                className="flex-1 border border-neutral-300 hover:bg-neutral-100 text-neutral-750 text-xs font-black uppercase tracking-wider py-2.5 rounded-xl transition-all"
              >
                Hủy
              </button>
              <button
                type="button"
                onClick={executeLockUnlock}
                className={`flex-1 text-white text-xs font-black uppercase tracking-wider py-2.5 rounded-xl transition-all ${
                  modalState.isLock
                    ? 'bg-red-650 hover:bg-red-750'
                    : 'bg-green-600 hover:bg-green-700'
                }`}
              >
                Đồng ý
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
