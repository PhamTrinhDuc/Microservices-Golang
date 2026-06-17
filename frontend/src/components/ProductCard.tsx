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

  const idNum = typeof product.id === 'string' 
    ? product.id.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) 
    : product.id
  const soldCount = (product.review_count || 0) * 4 + (idNum % 7) * 3 + 2

  // Extract spec tags
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

    const screen = findSpecValue(['kích thước màn hình', 'màn hình rộng'])
    if (screen) tags.push(screen.replace('inches', '"').replace('inch', '"'))

    const cpu = findSpecValue(['chipset', 'cpu', 'bộ vi xử lý'])
    if (cpu) {
      const parts = cpu.split(' ')
      const shortened = parts[0] + ' ' + (parts[1] || '')
      tags.push(shortened.trim())
    }

    const ram = findSpecValue(['ram', 'dung lượng ram'])
    const storage = findSpecValue(['bộ nhớ trong', 'dung lượng lưu trữ'])
    if (ram && storage) {
      tags.push(`${ram.replace(/\s*gb/i, '')}/${storage.replace(/\s*gb/i, 'GB')}`)
    } else if (storage) {
      tags.push(storage)
    }

    return tags.slice(0, 2)
  }, [product.specs_jsonb])

  return (
    <div className="relative flex flex-col h-full bg-white border border-[#E4E4E7] hover:border-[#18181B] transition-all duration-300 rounded-none group">
      {/* Image Container with Soft Gray Backdrop & Heart Overlay */}
      <div className="relative w-full aspect-square bg-[#FAF9F5] flex items-center justify-center p-4 border-b border-[#E4E4E7] rounded-none">
        <Link to={`/products/${product.id}`} className="w-full h-full flex items-center justify-center overflow-hidden">
          <img
            src={product.image || '/placeholder-product.png'}
            alt={product.name}
            className="object-contain max-h-full max-w-full mix-blend-multiply group-hover:scale-[1.02] transition-transform duration-500 ease-out"
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
          className="absolute top-2 right-2 w-8 h-8 rounded-none border border-[#E4E4E7] bg-[#FAF9F5] hover:bg-[#18181B] hover:text-[#FAF9F5] flex items-center justify-center text-[#8C8273] hover:scale-105 transition-all disabled:opacity-50 cursor-pointer"
        >
          <svg
            className={`w-3.5 h-3.5 transition-colors ${isLiked ? 'fill-[#18181B] text-[#18181B] group-hover:fill-[#FAF9F5]' : 'currentColor'}`}
            fill={isLiked ? 'currentColor' : 'none'}
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.8" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
        </button>

        {/* Minimal Marketing Badges Overlay */}
        <div className="absolute top-2 left-2 flex flex-col gap-1 z-10">
          {hasDiscount && (
            <div className="bg-[#18181B] text-[#FAF9F5] text-[8px] font-bold px-1.5 py-0.5 uppercase tracking-wider rounded-none">
              -{discountPercent}%
            </div>
          )}
        </div>

        {/* Out of Stock Overlay */}
        {product.stock === 0 && (
          <div className="absolute inset-0 bg-[#FAF9F5]/90 backdrop-blur-[1px] flex items-center justify-center">
            <span className="bg-[#18181B] text-[#FAF9F5] text-[9px] font-bold uppercase px-3 py-1 rounded-none tracking-wider">
              Hết hàng
            </span>
          </div>
        )}
      </div>

      {/* Info Details Section */}
      <div className="flex flex-col flex-1 p-4 bg-white">
        {/* Brand */}
        {product.brand && (
          <span className="text-[8px] font-semibold uppercase tracking-[0.2em] text-[#8C8273] mb-1">
            {product.brand.name}
          </span>
        )}
        
        {/* Title */}
        <Link to={`/products/${product.id}`} className="block mb-2">
          <h3 className="line-clamp-2 font-serif-display text-sm text-[#18181B] leading-snug hover:text-[#8C8273] transition-colors min-h-[36px] font-medium">
            {product.name}
          </h3>
        </Link>

        {/* Specification Snippet Pills */}
        {specTags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {specTags.map((tag, idx) => (
              <span key={idx} className="bg-[#FAF9F5] text-[#8C8273] text-[9px] font-semibold px-1.5 py-0.5 rounded-none border border-[#E4E4E7]/60 truncate max-w-[90px]" title={tag}>
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* Rating and Sold inline badge */}
        <div className="flex items-center gap-2 mb-3 text-[10px] text-[#8C8273]">
          <div className="flex items-center text-[#8C8273]">
            <span className="text-[11px] mr-0.5">★</span>
            <span className="font-semibold text-[#18181B]">{product.rating ? product.rating.toFixed(1) : '0.0'}</span>
          </div>
          <span className="text-[#E4E4E7]">/</span>
          <span className="lowercase font-light">đã bán {soldCount}</span>
        </div>

        {/* Price & Cart Actions */}
        <div className="mt-auto pt-3 border-t border-[#E4E4E7]/60 flex items-center justify-between gap-2">
          <div className="flex flex-col">
            {hasDiscount && (
              <span className="text-[9px] text-[#A1A1AA] line-through mb-0.5">
                {originalPrice.toLocaleString('vi-VN')} đ
              </span>
            )}
            <span className="text-xs font-bold text-[#18181B]">
              {displayPrice.toLocaleString('vi-VN')} đ
            </span>
          </div>

          <button
            type="button"
            disabled={product.stock === 0 || loadingCart}
            onClick={handleQuickAdd}
            className="w-8 h-8 rounded-none border border-[#E4E4E7] flex items-center justify-center text-[#18181B] hover:bg-[#18181B] hover:text-[#FAF9F5] hover:border-[#18181B] disabled:opacity-30 transition-all shrink-0 cursor-pointer bg-white"
            title="Thêm vào giỏ hàng"
          >
            {loadingCart ? (
              <svg className="animate-spin h-3 w-3 text-current" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            ) : (
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
