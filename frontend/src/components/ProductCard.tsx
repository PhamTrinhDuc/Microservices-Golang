import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { type Product } from '../types'
import { useCart } from '../hooks/useCart'
import { useWishlist } from '../hooks/useWishlist'
import { productAPI } from '../services/productAPI'

interface ProductCardProps {
  product: Product
}

const ProductCard = ({ product }: ProductCardProps) => {
  const { addToCart } = useCart()
  const { items: wishlistItems, addToWishlist, removeFromWishlist } = useWishlist()
  const [loadingCart, setLoadingCart] = useState(false)
  const [loadingWishlist, setLoadingWishlist] = useState(false)

  const wishlistItem = wishlistItems.find(item => item.product_id === product.id)
  const isLiked = !!wishlistItem


  const handleQuickAdd = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()

    if (product.stock === 0) return

    try {
      setLoadingCart(true)
      const detailedProduct = await productAPI.getProductById(product.id)
      const targetVariant = detailedProduct.variants?.[0]
      
      if (!targetVariant) {
        alert('Sản phẩm này chưa cấu hình phiên bản bán hàng.')
        return
      }

      await addToCart(targetVariant.id, 1).unwrap()
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setLoadingCart(false)
    }
  }
  
  const displayPrice = product.discount_price || product.price || 0
  const originalPrice = product.price || 0
  const hasDiscount = !!(product.discount_price && product.price && product.discount_price < product.price)
  const discountPercent = hasDiscount 
    ? Math.round(((originalPrice - displayPrice) / originalPrice) * 100)
    : 0

  // Calculate a realistic sold count for premium aesthetics
  const idNum = typeof product.id === 'string' 
    ? product.id.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) 
    : product.id
  const soldCount = (product.review_count || 0) * 4 + (idNum % 7) * 3 + 2

  // Extract spec tags (e.g. Screen size, RAM, Storage) for TGDĐ spec pills
  const specTags = useMemo(() => {
    let specs: any = {}
    if (product.specs_jsonb) {
      try {
        specs = typeof product.specs_jsonb === 'string'
          ? JSON.parse(product.specs_jsonb)
          : product.specs_jsonb
      } catch (e) {
        // ignore
      }
    }

    const tags: string[] = []
    
    // Look for screen size, CPU/Chipset, RAM, Storage in specs
    const findSpecValue = (targetKeys: string[]) => {
      for (const groupName of Object.keys(specs)) {
        const group = specs[groupName]
        if (group && typeof group === 'object') {
          for (const key of Object.keys(group)) {
            if (targetKeys.some(tk => key.toLowerCase().includes(tk.toLowerCase()))) {
              const info = group[key]
              const val = info?.raw_value || info?.value
              if (val && val !== '-' && val !== 'Không') {
                return String(val).trim()
              }
            }
          }
        }
      }
      return null
    }

    // 1. Screen size
    const screen = findSpecValue(['kích thước màn hình', 'màn hình rộng'])
    if (screen) tags.push(screen.replace('inches', '"').replace('inch', '"'))

    // 2. CPU/Chip
    const cpu = findSpecValue(['chipset', 'cpu', 'bộ vi xử lý'])
    if (cpu) {
      const parts = cpu.split(' ')
      const shortened = parts[0] + ' ' + (parts[1] || '')
      tags.push(shortened.trim())
    }

    // 3. RAM/ROM (e.g. 8 GB / 256 GB)
    const ram = findSpecValue(['ram', 'dung lượng ram'])
    const storage = findSpecValue(['bộ nhớ trong', 'dung lượng lưu trữ'])
    if (ram && storage) {
      tags.push(`${ram.replace(/\s*gb/i, '')}/${storage.replace(/\s*gb/i, 'GB')}`)
    } else if (storage) {
      tags.push(storage)
    }

    return tags.slice(0, 3)
  }, [product.specs_jsonb])

  const marketingBadges = useMemo(() => {
    const list = []
    if (product.stock > 0 && product.stock <= 5) {
      list.push({ text: 'Sắp hết hàng', css: 'bg-amber-500 text-white' })
    }
    if (discountPercent >= 20) {
      list.push({ text: 'Giá rẻ quá', css: 'bg-orange-600 text-white' })
    }
    if (displayPrice >= 10000000) {
      list.push({ text: 'Trả góp 0%', css: 'bg-emerald-600 text-white' })
    }
    if (product.review_count >= 15) {
      list.push({ text: 'Bán chạy', css: 'bg-red-650 text-white' })
    }
    return list.slice(0, 2)
  }, [product.stock, discountPercent, displayPrice, product.review_count])

  return (
    <div className="relative flex flex-col h-full bg-white rounded-lg border border-neutral-200 overflow-hidden hover:shadow-premium hover:-translate-y-0.5 transition-all duration-300 group">
      
      {/* Image Container with Soft Gray Backdrop & Heart Overlay */}
      <div className="relative w-full aspect-square bg-neutral-100/70 flex items-center justify-center p-3">
        <Link to={`/products/${product.id}`} className="w-full h-full flex items-center justify-center overflow-hidden rounded">
          <img
            src={product.image || '/placeholder-product.png'}
            alt={product.name}
            className="object-contain max-h-full max-w-full mix-blend-multiply group-hover:scale-105 transition-transform duration-550 ease-out"
          />
        </Link>

        {/* Heart Icon Overlay */}
        <button
          type="button"
          onClick={async (e) => {
            e.preventDefault()
            e.stopPropagation()
            if (loadingWishlist) return
            try {
              setLoadingWishlist(true)
              if (isLiked && wishlistItem) {
                await removeFromWishlist(wishlistItem.variant_id).unwrap()
              } else {
                let variantId = product.variants?.[0]?.id
                if (!variantId) {
                  const detailedProduct = await productAPI.getProductById(product.id)
                  variantId = detailedProduct.variants?.[0]?.id
                }
                if (!variantId) {
                  alert('Không tìm thấy phiên bản sản phẩm để thêm vào yêu thích.')
                  return
                }
                await addToWishlist(variantId).unwrap()
              }
            } catch (err: any) {
              alert(err || 'Không thể cập nhật danh sách yêu thích')
            } finally {
              setLoadingWishlist(false)
            }
          }}
          disabled={loadingWishlist}
          className="absolute top-2.5 right-2.5 w-8 h-8 rounded-full bg-white shadow-sm flex items-center justify-center text-neutral-450 hover:text-red-500 hover:scale-105 transition-all disabled:opacity-50"
        >
          <svg
            className={`w-4 h-4 transition-colors ${isLiked ? 'fill-red-500 text-red-500' : 'text-neutral-400'}`}
            fill={isLiked ? 'currentColor' : 'none'}
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
        </button>

        {/* TGDĐ Marketing Badges Overlay */}
        <div className="absolute top-2.5 left-2.5 flex flex-col gap-1 z-10">
          {hasDiscount && (
            <div className="bg-red-505 text-white text-[9px] font-extrabold px-1.5 py-0.5 rounded-sm uppercase tracking-wide w-fit shadow-sm">
              -{discountPercent}%
            </div>
          )}
          {marketingBadges.map((badge, idx) => (
            <div key={idx} className={`${badge.css} text-[9px] font-extrabold px-1.5 py-0.5 rounded-sm uppercase tracking-wide w-fit shadow-sm`}>
              {badge.text}
            </div>
          ))}
        </div>

        {/* Out of Stock Overlay */}
        {product.stock === 0 && (
          <div className="absolute inset-0 bg-white/80 backdrop-blur-[1px] flex items-center justify-center">
            <span className="bg-neutral-900 text-white text-[10px] font-extrabold uppercase px-3 py-1 rounded-sm tracking-wider">
              Hết hàng
            </span>
          </div>
        )}
      </div>

      {/* Info Details Section */}
      <div className="flex flex-col flex-1 p-3.5">
        
        {/* Brand */}
        {product.brand && (
          <span className="text-[9px] font-bold uppercase tracking-wider text-neutral-400 mb-1">
            {product.brand.name}
          </span>
        )}
        
        {/* Title */}
        <Link to={`/products/${product.id}`} className="block mb-2">
          <h3 className="line-clamp-2 text-[13px] font-semibold text-neutral-800 leading-snug hover:text-black transition-colors min-h-[36px]">
            {product.name}
          </h3>
        </Link>

        {/* Specification Snippet Pills (TGDĐ style) */}
        {specTags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {specTags.map((tag, idx) => (
              <span key={idx} className="bg-neutral-100 text-neutral-600 text-[9px] font-bold px-1.5 py-0.5 rounded-sm border border-neutral-200/50 truncate max-w-[80px]" title={tag}>
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* Rating and Sold inline badge */}
        <div className="flex items-center gap-2 mb-3 text-[11px] text-neutral-500">
          <div className="flex items-center text-amber-500">
            <span className="text-xs mr-0.5">★</span>
            <span className="font-bold text-neutral-700">{product.rating ? product.rating.toFixed(1) : '0.0'}</span>
          </div>
          <span className="text-neutral-300">|</span>
          <span>Đã bán {soldCount}</span>
        </div>

        {/* Price & Cart Actions */}
        <div className="mt-auto pt-3 border-t border-neutral-100 flex items-center justify-between gap-2">
          <div className="flex flex-col">
            {hasDiscount && (
              <div className="flex items-center gap-1.5 mb-0.5">
                <span className="text-[10px] text-neutral-400 line-through">
                  {originalPrice.toLocaleString('vi-VN')} đ
                </span>
                <span className="text-[9px] font-black text-red-500 bg-red-50 px-1 py-0.2 rounded">
                  Tiết kiệm {(originalPrice - displayPrice).toLocaleString('vi-VN')}đ
                </span>
              </div>
            )}
            <span className="text-sm font-black text-neutral-900">
              {displayPrice.toLocaleString('vi-VN')} đ
            </span>
          </div>

          <button
            type="button"
            disabled={product.stock === 0 || loadingCart}
            onClick={handleQuickAdd}
            className="w-8 h-8 rounded-full border border-neutral-250 flex items-center justify-center text-neutral-750 hover:bg-black hover:text-white hover:border-black active:scale-95 disabled:opacity-40 disabled:hover:bg-transparent disabled:hover:text-neutral-750 disabled:hover:border-neutral-250 transition-all shrink-0"
            title="Thêm vào giỏ hàng"
          >
            {loadingCart ? (
              <svg className="animate-spin h-4 w-4 text-black" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            ) : (
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
              </svg>
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

export default ProductCard
