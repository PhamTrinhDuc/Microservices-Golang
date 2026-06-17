import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { useCart } from '../hooks/useCart'
import { useWishlist } from '../hooks/useWishlist'
import { useCatalog } from '../hooks/useCatalog'
import { keycloak } from '../utils/keycloak'
import CategoryNavStrip from './CategoryNavStrip'

const Header = () => {
  const { isAuthenticated, user, logout } = useAuth()
  const { totalItems } = useCart()
  const { items: wishlistItems } = useWishlist()
  const wishlistCount = wishlistItems.length
  const navigate = useNavigate()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('all')

  const { categories, brands, loading: catalogLoading } = useCatalog()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeCategoryId = searchParams.get('category')
  const activeBrandId = searchParams.get('brand')

  const updateCategoryAndBrand = (categoryId: string | null, brandId: string | null) => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (categoryId === null) {
        next.delete('category')
      } else {
        next.set('category', categoryId)
      }
      if (brandId === null) {
        next.delete('brand')
      } else {
        next.set('brand', brandId)
      }
      next.delete('page')
      return next
    })
  }

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
    <header className="sticky top-0 z-50 w-full bg-[#FAF9F5] border-b border-[#E4E4E7] font-sans">
      {/* Top Utility Bar */}
      <div className="w-full bg-[#FAF9F5] border-b border-[#E4E4E7]/60 py-1.5 text-[10px] text-[#8C8273] tracking-wider uppercase font-semibold">
        <div className="mx-auto max-w-7xl px-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <span className="cursor-pointer hover:text-[#18181B] transition-colors">Tải Ứng Dụng</span>
            <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
            <span className="cursor-pointer hover:text-[#18181B] transition-colors">BeliBeli Care</span>
            <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
            <span className="cursor-pointer hover:text-[#18181B] transition-colors text-[#8C8273]">Độc Quyền</span>
          </div>
          <div className="flex items-center gap-4">
            <Link to="/browse?sort=popular" className="hover:text-[#18181B] transition-colors">Bán chạy</Link>
            <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
            {isAuthenticated ? (
              <div className="flex items-center gap-3">
                {user?.role === 'admin' && (
                  <>
                    <Link
                      to="/admin"
                      className="text-[#18181B] hover:text-[#8C8273] font-bold transition-colors uppercase tracking-wider text-[10px]"
                    >
                      Quản trị
                    </Link>
                    <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
                  </>
                )}
                <span className="font-semibold text-[#18181B] lowercase">
                  @{user?.full_name?.replace(/\s+/g, '').toLowerCase() || 'member'}
                </span>
                <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="hover:text-[#18181B] transition-colors font-semibold cursor-pointer uppercase tracking-wider bg-transparent border-0 p-0 text-[10px]"
                >
                  Đăng xuất
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <button
                  onClick={() => void keycloak.login()}
                  className="hover:text-[#18181B] transition-colors font-semibold cursor-pointer bg-transparent border-0 p-0 text-[10px] uppercase tracking-wider"
                >
                  Đăng nhập
                </button>
                <span className="h-2.5 w-[0.5px] bg-[#E4E4E7]"></span>
                <button
                  onClick={() => void keycloak.register()}
                  className="hover:text-[#18181B] transition-colors font-semibold cursor-pointer bg-transparent border-0 p-0 text-[10px] uppercase tracking-wider"
                >
                  Đăng ký
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Main Navbar */}
      <div className="w-full bg-[#FAF9F5] py-4">
        <div className="mx-auto max-w-7xl px-4 flex items-center justify-between gap-6">
          {/* Logo - Print Serif display logo */}
          <Link to="/" className="shrink-0">
            <span className="font-serif-display text-2xl font-bold tracking-[0.15em] text-[#18181B] hover:opacity-85 transition-opacity">
              BELIBELI.
            </span>
          </Link>

          {/* Search Input Group with Dropdown */}
          <form onSubmit={handleSearch} className="flex-1 max-w-xl">
            <div className="flex items-center border border-[#E4E4E7] bg-[#FAF9F5] focus-within:border-[#18181B] transition-colors duration-300 rounded-none overflow-hidden">
              {/* Category Select Dropdown */}
              <div className="relative border-r border-[#E4E4E7] bg-[#FAF9F5]">
                <select
                  value={selectedCategory}
                  onChange={(e) => setSelectedCategory(e.target.value)}
                  className="bg-transparent pl-4 pr-8 py-2 text-[10px] uppercase tracking-widest font-bold text-[#8C8273] focus:outline-none appearance-none cursor-pointer"
                >
                  <option value="all">Tất cả</option>
                  <option value="1">Thời trang</option>
                  <option value="2">Điện tử</option>
                  <option value="3">Gia dụng</option>
                  <option value="4">Làm đẹp</option>
                </select>
                <div className="pointer-events-none absolute inset-y-0 right-2.5 flex items-center text-[#8C8273]">
                  <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </div>

              {/* Main Search Input */}
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Tìm sản phẩm, thiết kế..."
                className="w-full bg-transparent px-4 py-2 text-[11px] text-[#18181B] placeholder-[#8C8273]/60 focus:outline-none font-medium"
              />

              {/* Search Button */}
              <button
                type="submit"
                className="bg-[#18181B] text-[#FAF9F5] px-5 py-2 hover:bg-transparent hover:text-[#18181B] border-l border-[#18181B] transition-all duration-300"
              >
                <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>
            </div>
          </form>

          {/* Action Icons */}
          <div className="flex items-center gap-2 shrink-0">
            {/* Wishlist */}
            <Link
              to="/wishlist"
              className="relative p-2 text-[#8C8273] hover:text-[#18181B] transition-colors"
              title="Yêu thích"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
              {wishlistCount > 0 && (
                <span className="absolute top-1 right-1 flex h-3.5 min-w-3.5 items-center justify-center rounded-none bg-[#18181B] text-[#FAF9F5] px-0.5 text-[8px] font-semibold border border-[#FAF9F5]">
                  {wishlistCount}
                </span>
              )}
            </Link>

            {/* Shopping Bag */}
            <Link
              to="/cart"
              className="relative p-2 text-[#8C8273] hover:text-[#18181B] transition-colors"
              title="Giỏ hàng"
            >
              <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
              {totalItems > 0 && (
                <span className="absolute top-1 right-1 flex h-3.5 min-w-3.5 items-center justify-center rounded-none bg-[#18181B] text-[#FAF9F5] px-0.5 text-[8px] font-semibold border border-[#FAF9F5]">
                  {totalItems}
                </span>
              )}
            </Link>

            {/* Profile Avatar if Authenticated */}
            {isAuthenticated && (
              <Link to="/profile" className="ml-2 flex items-center justify-center h-8 w-8 rounded-none border border-[#E4E4E7] hover:border-[#18181B] transition-colors bg-[#FAF9F5]">
                <div className="font-serif-display font-medium text-sm text-[#18181B]">
                  {(user?.full_name || 'U')[0].toUpperCase()}
                </div>
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Sub-navbar / Category Sub-navigation */}
      <div className="w-full bg-[#FAF9F5] border-t border-[#E4E4E7] relative z-40">
        <div className="mx-auto max-w-7xl px-4 py-1 overflow-visible">
          <CategoryNavStrip
            categories={categories}
            brands={brands}
            activeCategoryId={activeCategoryId}
            activeBrandId={activeBrandId}
            onSelectCategoryAndBrand={updateCategoryAndBrand}
            loading={catalogLoading}
            variant="header"
          />
        </div>
      </div>
    </header>
  )
}

export default Header
