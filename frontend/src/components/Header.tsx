import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useCart } from '../hooks/useCart'

const Header = () => {
  const { isAuthenticated, user, logout } = useAuth()
  const { totalItems } = useCart()
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')

  const handleLogout = async () => {
    await logout()
  }

  const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!searchQuery.trim() && selectedCategory === 'all') {
      return
    }
    let url = '/browse'
    const params = new URLSearchParams()
    
    if (searchQuery.trim()) {
      params.append('q', searchQuery)
    }
    if (selectedCategory && selectedCategory !== 'all') {
      params.append('category', selectedCategory)
    }
    
    const queryString = params.toString()
    if (queryString) {
      url += `?${queryString}`
    }
    navigate(url)
  }

  return (
    <header className="sticky top-0 z-50 w-full bg-white shadow-sm font-sans">
      {/* Top Utility Bar */}
      <div className="w-full bg-neutral-50 border-b border-neutral-200 py-1.5 text-xs text-neutral-500">
        <div className="mx-auto max-w-7xl px-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <span className="cursor-pointer hover:text-neutral-900 transition-colors">Download BeliBeli App</span>
            <span className="h-3 w-px bg-neutral-350"></span>
            <span className="cursor-pointer hover:text-neutral-900 transition-colors">BeliBeli Care</span>
            <span className="h-3 w-px bg-neutral-350"></span>
            <span className="cursor-pointer hover:text-neutral-900 transition-colors text-red-500 font-medium">Promo</span>
          </div>
          <div className="flex items-center gap-4">
            <Link to="/browse?sort=popular" className="hover:text-neutral-900 transition-colors">Bán chạy</Link>
            <span className="h-3 w-px bg-neutral-350"></span>
            {isAuthenticated ? (
              <div className="flex items-center gap-3">
                {user?.role === 'admin' && (
                  <>
                    <Link
                      to="/admin"
                      className="text-neutral-900 hover:text-red-600 font-bold transition-colors uppercase tracking-wider text-[10px]"
                    >
                      Quản trị viên
                    </Link>
                    <span className="h-3 w-px bg-neutral-350"></span>
                  </>
                )}
                <span className="font-semibold text-neutral-800">
                  Hi, {user?.full_name || user?.email}
                </span>
                <span className="h-3 w-px bg-neutral-350"></span>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="hover:text-neutral-900 transition-colors font-medium"
                >
                  Đăng xuất
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <Link to="/login" className="hover:text-neutral-900 transition-colors font-medium">Đăng nhập</Link>
                <span className="h-3 w-px bg-neutral-350"></span>
                <Link to="/register" className="text-neutral-900 hover:underline font-semibold">Đăng ký</Link>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Main Navbar */}
      <div className="w-full bg-white border-b border-neutral-100 py-3.5">
        <div className="mx-auto max-w-7xl px-4 flex items-center justify-between gap-6">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-1.5 group shrink-0">
            <div className="bg-black text-white px-2 py-1 rounded font-black text-lg tracking-wider">
              B
            </div>
            <span className="text-xl font-extrabold tracking-tight text-neutral-900">
              BeliBeli<span className="text-neutral-400 font-semibold text-sm">.com</span>
            </span>
          </Link>

          {/* Search Input Group with Dropdown */}
          <form onSubmit={handleSearch} className="flex-1 max-w-2xl">
            <div className="flex items-center rounded-md border border-neutral-300 bg-white hover:border-neutral-400 focus-within:border-black focus-within:ring-1 focus-within:ring-black transition-all">
              {/* Category Select Dropdown */}
              <div className="relative border-r border-neutral-200">
                <select
                  value={selectedCategory}
                  onChange={(e) => setSelectedCategory(e.target.value)}
                  className="bg-transparent pl-4 pr-8 py-2 text-xs font-semibold text-neutral-700 focus:outline-none appearance-none cursor-pointer"
                >
                  <option value="all">Tất cả danh mục</option>
                  <option value="1">Thời trang</option>
                  <option value="2">Điện tử</option>
                  <option value="3">Gia dụng</option>
                  <option value="4">Sức khỏe & Làm đẹp</option>
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-2.5 flex items-center text-neutral-500">
                  <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </div>

              {/* Main Search Input */}
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Tìm sản phẩm, thương hiệu hoặc cửa hàng mong muốn..."
                className="w-full bg-transparent px-4 py-2 text-xs text-neutral-800 placeholder-neutral-400 focus:outline-none"
              />

              {/* Search Button */}
              <button
                type="submit"
                className="bg-black text-white px-6 py-2.5 rounded-r-[4px] hover:bg-neutral-850 active:bg-black transition-colors"
              >
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>
            </div>
          </form>

          {/* Action Icons */}
          <div className="flex items-center gap-4 shrink-0">
            {/* Notifications */}
            <button className="relative p-2 text-neutral-700 hover:bg-neutral-100 hover:text-black rounded-full transition-colors">
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
              <span className="absolute top-1.5 right-1.5 flex h-2 w-2 rounded-full bg-red-500"></span>
            </button>

            {/* Shopping Bag */}
            <Link
              to="/cart"
              className="relative p-2 text-neutral-700 hover:bg-neutral-100 hover:text-black rounded-full transition-colors"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
              <span className="absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-neutral-900 px-1 text-[9px] font-bold text-white">
                {totalItems}
              </span>
            </Link>

            {/* Profile Avatar if Authenticated */}
            {isAuthenticated && (
              <Link to="/profile" className="flex items-center justify-center h-8 w-8 rounded-full border border-neutral-200 overflow-hidden hover:border-black transition-colors bg-neutral-100">
                <div className="font-bold text-xs text-neutral-700">
                  {(user?.full_name || 'U')[0].toUpperCase()}
                </div>
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Sub-navbar / Category Sub-navigation */}
      <div className="w-full bg-white border-b border-neutral-200 py-2">
        <div className="mx-auto max-w-7xl px-4 flex items-center justify-between">
          <nav className="flex items-center gap-6 text-xs font-semibold text-neutral-600">
            <Link to="/" className="hover:text-black transition-colors relative py-1">Trang chủ</Link>
            <Link to="/browse" className="hover:text-black transition-colors relative py-1">Tất cả sản phẩm</Link>
            <Link to="/browse?sort=featured" className="hover:text-black transition-colors relative py-1 flex items-center gap-1">
              <span>Nổi bật</span>
              <span className="bg-red-100 text-red-650 text-[8px] font-extrabold px-1 rounded-sm uppercase tracking-wide">Hot</span>
            </Link>
            <Link to="/browse?sort=popular" className="hover:text-black transition-colors relative py-1">Phổ biến</Link>
            <Link to="/browse?sort=sale" className="hover:text-black transition-colors relative py-1 text-red-500 hover:text-red-650 font-bold">Khuyến mãi cực sốc</Link>
          </nav>
          <div className="text-xs text-neutral-500 font-medium hidden md:block">
            Miễn phí vận chuyển từ đơn hàng <span className="font-bold text-neutral-850">299kđ</span>
          </div>
        </div>
      </div>
    </header>
  )
}

export default Header

