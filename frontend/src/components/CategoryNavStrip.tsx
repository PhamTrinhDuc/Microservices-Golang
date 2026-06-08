import { useLocation, useNavigate } from 'react-router-dom'
import { type Category, type Brand } from '../types'

interface CategoryNavStripProps {
  categories: Category[]
  brands: Brand[]
  activeCategoryId?: string | null
  activeBrandId?: string | null
  onSelectCategory?: (id: string | null) => void
  onSelectBrand?: (id: string | null) => void
  onSelectCategoryAndBrand?: (categoryId: string | null, brandId: string | null) => void
  loading?: boolean
  className?: string
  variant?: 'standalone' | 'header'
}

// Custom vector icons mapping based on category names
const getCategoryIconSvg = (name: string) => {
  const n = name.toLowerCase()
  if (n.includes('điện thoại')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <rect x="5" y="2" width="14" height="20" rx="2.5" />
        <line x1="12" y1="18" x2="12.01" y2="18" strokeLinecap="round" strokeWidth="2.5" />
      </svg>
    )
  }
  if (n.includes('laptop')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <rect x="3" y="4" width="18" height="12" rx="1.5" />
        <path d="M2 20h20M7 16h10" strokeLinecap="round" />
      </svg>
    )
  }
  if (n.includes('phụ kiện')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 18a5 5 0 0110 0v-4m0 0a5 5 0 0110 0v4M13 6a3 3 0 10-6 0v8M17 14V6a3 3 0 10-6 0" />
      </svg>
    )
  }
  if (n.includes('smartwatch') || n.includes('đồng hồ thông minh')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <rect x="6" y="6" width="12" height="12" rx="3" />
        <path d="M9 6V2h6v4M9 18v4h6v-4" strokeLinecap="round" />
        <circle cx="12" cy="12" r="2.5" />
      </svg>
    )
  }
  if (n.includes('đồng hồ')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="9" />
        <polyline points="12 7 12 12 15 15" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    )
  }
  if (n.includes('tablet') || n.includes('máy tính bảng')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <rect x="4" y="3" width="16" height="18" rx="2" />
        <line x1="12" y1="18" x2="12.01" y2="18" strokeLinecap="round" strokeWidth="2.5" />
      </svg>
    )
  }
  if (n.includes('máy cũ') || n.includes('thu cũ')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
      </svg>
    )
  }
  if (n.includes('màn hình')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <rect x="2" y="3" width="20" height="13" rx="2" />
        <path d="M12 16v4M8 20h8" strokeLinecap="round" />
      </svg>
    )
  }
  if (n.includes('sim') || n.includes('thẻ cào')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M7 2h8l5 5v13a2 2 0 01-2 2H7a2 2 0 01-2-2V4a2 2 0 012-2z" />
        <rect x="8" y="11" width="6" height="5" rx="1" />
      </svg>
    )
  }
  if (n.includes('dịch vụ') || n.includes('tiện ích')) {
    return (
      <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
      </svg>
    )
  }
  return (
    <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-14L4 7m8 4v10M4 7v10l8 4" />
    </svg>
  )
}

const renderCategoryIcon = (category: Category, isActive: boolean) => {
  if (!category.icon) {
    return getCategoryIconSvg(category.name)
  }

  // Check if icon is a URL or file path
  const isUrl = category.icon.startsWith('http') || 
                category.icon.startsWith('/') || 
                category.icon.includes('.') || 
                category.icon.includes('/')

  if (isUrl) {
    return (
      <img
        src={category.icon}
        alt={category.name}
        className={`w-5 h-5 object-contain transition-transform group-hover:scale-110 ${isActive ? 'brightness-0 invert' : ''}`}
      />
    )
  }

  // Otherwise treat as emoji or text icon
  return (
    <span className="text-base select-none transition-transform group-hover:scale-110">
      {category.icon}
    </span>
  )
}

// Category static custom dropdown options (subcategories/brands mapping dynamically)
const getCategoryDropdownMenu = (categoryName: string, allBrands: Brand[]) => {
  const name = categoryName.toLowerCase()

  if (name.includes('điện thoại')) {
    const matchedBrands = allBrands.filter(b =>
      ['apple', 'samsung', 'xiaomi', 'oppo', 'realme', 'vivo'].some(t => b.name.toLowerCase().includes(t))
    )
    return {
      type: 'brands',
      title: 'Hãng sản xuất điện thoại',
      items: matchedBrands.map(b => ({ label: b.name, brandId: b.id })),
    }
  }

  if (name.includes('laptop')) {
    const matchedBrands = allBrands.filter(b =>
      ['asus', 'dell', 'hp', 'lenovo', 'acer', 'msi', 'apple'].some(t => b.name.toLowerCase().includes(t))
    )
    return {
      type: 'brands',
      title: 'Hãng sản xuất Laptop',
      items: matchedBrands.map(b => ({ label: b.name === 'Apple' ? 'MacBook (Apple)' : b.name, brandId: b.id })),
    }
  }

  if (name.includes('phụ kiện')) {
    return {
      type: 'links',
      title: 'Nhóm phụ kiện',
      items: [
        { label: 'Tai nghe Bluetooth', query: 'tai nghe' },
        { label: 'Loa không dây', query: 'loa' },
        { label: 'Sạc dự phòng', query: 'sạc dự phòng' },
        { label: 'Cáp sạc, Củ sạc', query: 'cáp sạc' },
        { label: 'Bao da, Ốp lưng', query: 'ốp lưng' },
      ],
    }
  }

  if (name.includes('tablet') || name.includes('máy tính bảng')) {
    const matchedBrands = allBrands.filter(b =>
      ['apple', 'samsung', 'xiaomi'].some(t => b.name.toLowerCase().includes(t))
    )
    return {
      type: 'brands',
      title: 'Hãng sản xuất Tablet',
      items: matchedBrands.map(b => ({ label: b.name === 'Apple' ? 'iPad (Apple)' : b.name, brandId: b.id })),
    }
  }

  if (name.includes('máy cũ') || name.includes('thu cũ')) {
    return {
      type: 'links',
      title: 'Mua bán máy cũ',
      items: [
        { label: 'Điện thoại cũ giá rẻ', query: 'điện thoại' },
        { label: 'Laptop cũ cấu hình cao', query: 'laptop' },
        { label: 'Máy tính bảng cũ', query: 'tablet' },
        { label: 'Thu cũ đổi mới trợ giá 15%', query: '' },
      ],
    }
  }

  if (name.includes('màn hình')) {
    const matchedBrands = allBrands.filter(b =>
      ['asus', 'dell', 'samsung', 'lg'].some(t => b.name.toLowerCase().includes(t))
    )
    return {
      type: 'brands',
      title: 'Thương hiệu màn hình',
      items: matchedBrands.map(b => ({ label: b.name, brandId: b.id })),
    }
  }

  if (name.includes('sim') || name.includes('thẻ cào')) {
    return {
      type: 'links',
      title: 'Sim & Thẻ cào online',
      items: [
        { label: 'Sim Viettel Data siêu khủng', query: 'viettel' },
        { label: 'Sim Mobifone/Vinaphone', query: 'sim' },
        { label: 'Thẻ cào chiết khấu cao', query: 'thẻ' },
      ],
    }
  }

  if (name.includes('dịch vụ') || name.includes('tiện ích')) {
    return {
      type: 'links',
      title: 'Tiện ích thanh toán',
      items: [
        { label: 'Đóng tiền điện, nước 24/7', query: '' },
        { label: 'Mua bảo hiểm thiết bị', query: '' },
        { label: 'Đăng ký gói cước 4G/5G', query: '' },
        { label: 'Mua trả góp lãi suất 0%', query: '' },
      ],
    }
  }

  return null
}

const CategoryNavStrip = ({
  categories,
  brands,
  activeCategoryId = null,
  activeBrandId = null,
  onSelectCategory,
  onSelectBrand,
  onSelectCategoryAndBrand,
  loading = false,
  className = '',
  variant = 'standalone',
}: CategoryNavStripProps) => {
  const location = useLocation()
  const navigate = useNavigate()
  const isBrowsePage = location.pathname === '/browse'

  const handleCategoryClick = (category: Category) => {
    const idStr = String(category.id)
    if (isBrowsePage) {
      if (onSelectCategoryAndBrand) {
        // Toggle category and clear brand filter for clean category switch
        if (activeCategoryId === idStr) {
          onSelectCategoryAndBrand(null, null)
        } else {
          onSelectCategoryAndBrand(idStr, null)
        }
      } else if (onSelectCategory) {
        onSelectCategory(activeCategoryId === idStr ? null : idStr)
      }
    } else {
      navigate(`/browse?category=${idStr}`)
    }
  }

  const handleBrandClick = (category: Category, brandId: number) => {
    const catIdStr = String(category.id)
    const brandIdStr = String(brandId)
    if (isBrowsePage) {
      if (onSelectCategoryAndBrand) {
        onSelectCategoryAndBrand(catIdStr, brandIdStr)
      } else {
        if (onSelectCategory) onSelectCategory(catIdStr)
        if (onSelectBrand) onSelectBrand(brandIdStr)
      }
    } else {
      navigate(`/browse?category=${catIdStr}&brand=${brandIdStr}`)
    }
  }

  const handleSubLinkClick = (category: Category, query: string) => {
    const catIdStr = String(category.id)
    if (isBrowsePage) {
      navigate(`/browse?category=${catIdStr}${query ? `&q=${encodeURIComponent(query)}` : ''}`)
    } else {
      navigate(`/browse?category=${catIdStr}${query ? `&q=${encodeURIComponent(query)}` : ''}`)
    }
  }

  if (loading) {
    const isHeader = variant === 'header'
    return (
      <div className={`w-full ${isHeader ? 'bg-transparent py-1' : 'bg-white border border-neutral-200/75 rounded-xl px-4 py-2.5 shadow-sm'} animate-pulse ${className}`}>
        <div className="flex gap-3 overflow-x-auto pb-1 scrollbar-none items-center h-8">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="h-5 w-20 bg-neutral-200/60 rounded-full shrink-0"></div>
          ))}
        </div>
      </div>
    )
  }

  const isHeader = variant === 'header'
  return (
    <div className={`w-full relative z-40 ${
      isHeader 
        ? 'bg-transparent py-1' 
        : 'bg-white border border-neutral-200/75 rounded-xl px-3 sm:px-4 py-2 shadow-sm'
    } ${className}`}>
      <div className="flex gap-2 sm:gap-2.5 md:gap-3 overflow-x-auto pb-1.5 pt-0.5 scrollbar-none items-center justify-start">
        {/* "Tất cả" option (only visible on browse page or customizable) */}
        {isBrowsePage && (
          <button
            onClick={() => {
              if (onSelectCategoryAndBrand) {
                onSelectCategoryAndBrand(null, null)
              } else if (onSelectCategory) {
                onSelectCategory(null)
              }
            }}
            className={`flex-shrink-0 text-xs sm:text-[13px] font-bold uppercase whitespace-nowrap flex items-center gap-1 transition-all duration-200 px-2.5 py-1.5 rounded-lg ${
              !activeCategoryId
                ? 'bg-brand-600 text-white shadow-sm font-extrabold'
                : 'text-neutral-600 hover:bg-brand-50 hover:text-brand-600'
            }`}
          >
            🏢 Tất cả
          </button>
        )}

        {categories.map((c) => {
          const idStr = String(c.id)
          const isActive = activeCategoryId === idStr
          const dropdown = getCategoryDropdownMenu(c.name, brands)

          return (
            <div key={c.id} className="relative group shrink-0">
              {/* Category Link Button */}
              <button
                onClick={() => handleCategoryClick(c)}
                className={`text-xs sm:text-[13px] font-bold whitespace-nowrap flex items-center gap-1.5 transition-all duration-200 px-2.5 py-1.5 rounded-lg group ${
                  isActive
                    ? 'bg-brand-600 text-white shadow-sm font-extrabold'
                    : 'text-neutral-600 hover:bg-brand-50 hover:text-brand-600'
                }`}
              >
                {/* Category Icon */}
                <span className={`flex items-center shrink-0 transition-colors ${isActive ? 'text-white' : 'text-neutral-500 group-hover:text-brand-600'}`}>
                  {renderCategoryIcon(c, isActive)}
                </span>
                
                <span>{c.name}</span>

                {/* Down chevron indicator */}
                {dropdown && (
                  <svg
                    className={`w-2.5 h-2.5 mt-0.5 shrink-0 transition-transform duration-200 ${
                      isActive ? 'text-brand-200' : 'text-neutral-400 group-hover:text-brand-500 group-hover:translate-y-0.5'
                    }`}
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.5"
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                )}
              </button>

              {/* Hover Dropdown Menu Card */}
              {dropdown && (
                <div className="absolute left-0 mt-2 w-60 bg-white border border-neutral-200/85 rounded-xl shadow-xl p-4 opacity-0 translate-y-2 pointer-events-none group-hover:opacity-100 group-hover:translate-y-0 group-hover:pointer-events-auto transition-all duration-200 ease-out z-50">
                  <div className="text-[10px] font-black uppercase tracking-wider text-neutral-400 mb-2.5 border-b border-neutral-100 pb-1">
                    {dropdown.title}
                  </div>
                  <div className="flex flex-col gap-1">
                    {dropdown.type === 'brands'
                      ? (dropdown.items as Array<{ label: string; brandId: number }>).map((brand) => (
                          <button
                            key={brand.brandId}
                            onClick={() => handleBrandClick(c, brand.brandId)}
                            className={`w-full text-left text-xs px-2.5 py-1.5 rounded-lg font-semibold transition-colors flex items-center justify-between ${
                              activeBrandId === String(brand.brandId) && isActive
                                ? 'bg-brand-50 text-brand-700 font-bold'
                                : 'text-neutral-600 hover:bg-brand-50 hover:text-brand-600'
                            }`}
                          >
                            <span>{brand.label}</span>
                            <span className="text-neutral-350 group-hover:text-brand-500 text-[10px]">→</span>
                          </button>
                        ))
                      : (dropdown.items as Array<{ label: string; query: string }>).map((item, idx) => (
                          <button
                            key={idx}
                            onClick={() => handleSubLinkClick(c, item.query)}
                            className="w-full text-left text-xs px-2.5 py-1.5 rounded-lg font-semibold text-neutral-600 hover:bg-brand-50 hover:text-brand-600 transition-colors flex items-center justify-between"
                          >
                            <span>{item.label}</span>
                            <span className="text-neutral-350 group-hover:text-brand-500 text-[10px]">→</span>
                          </button>
                        ))}
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default CategoryNavStrip
