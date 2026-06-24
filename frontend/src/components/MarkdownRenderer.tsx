import React, { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { Link } from 'react-router-dom'
import { useCart } from '../hooks/useCart'
import { productAPI } from '../services/productAPI'

interface MarkdownRendererProps {
  content: string
}

interface InlineProductCardProps {
  id: string
}

const InlineProductCard: React.FC<InlineProductCardProps> = ({ id }) => {
  const { addToCart } = useCart()
  const [product, setProduct] = useState<any | null>(null)
  const [loading, setLoading] = useState(true)
  const [adding, setAdding] = useState(false)
  const [success, setSuccess] = useState(false)

  React.useEffect(() => {
    let active = true
    const fetchDetail = async () => {
      try {
        setLoading(true)
        const data = await productAPI.getProductById(id)
        if (active) {
          setProduct(data)
        }
      } catch (err) {
        console.error('Failed to fetch product details for inline card:', err)
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    fetchDetail()
    return () => {
      active = false
    }
  }, [id])

  const handleAddToCart = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (!product) return
    try {
      setAdding(true)
      const targetVariant = product.variants?.[0]
      if (!targetVariant) {
        alert('Sản phẩm này chưa cấu hình phiên bản bán hàng.')
        return
      }
      await addToCart(targetVariant.id, 1).unwrap()
      setSuccess(true)
      setTimeout(() => setSuccess(false), 2000)
    } catch (err: any) {
      alert(err || 'Không thể thêm sản phẩm vào giỏ hàng')
    } finally {
      setAdding(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center gap-3 bg-[#11131c]/90 border border-slate-800/85 rounded-xl p-2.5 shadow-sm w-full sm:w-[280px] animate-pulse">
        <div className="w-12 h-12 bg-slate-800 rounded-lg shrink-0" />
        <div className="flex-1 space-y-1.5">
          <div className="h-2.5 bg-slate-800 rounded w-3/4" />
          <div className="h-2.5 bg-slate-800 rounded w-1/4" />
        </div>
      </div>
    )
  }

  if (!product) {
    return (
      <Link to={`/products/${id}`} className="text-indigo-400 hover:text-indigo-300 underline font-semibold text-xs transition-colors">
        Xem sản phẩm (#{id})
      </Link>
    )
  }

  const displayPrice = product.discount_price || product.price || 0
  const originalPrice = product.price || 0
  const hasDiscount = !!(product.discount_price && product.price && product.discount_price < product.price)

  return (
    <div className="flex items-center gap-3 bg-[#131522] border border-[#1e2235] hover:border-indigo-500/40 rounded-xl p-2.5 shadow-[0_4px_12px_rgba(0,0,0,0.15)] hover:shadow-[0_4px_18px_rgba(99,102,241,0.1)] transition-all duration-300 w-full sm:w-[280px]">
      {/* Product Image */}
      <div className="w-12 h-12 bg-white rounded-lg flex items-center justify-center p-1 shrink-0 overflow-hidden shadow-inner border border-slate-800/20">
        <img src={product.image || '/placeholder-product.png'} alt={product.name} className="max-h-full max-w-full object-contain mix-blend-multiply" />
      </div>
      
      {/* Product Details */}
      <div className="flex-1 min-w-0 flex flex-col gap-1 text-left">
        <h4 className="text-xs font-bold text-slate-100 truncate" title={product.name}>{product.name}</h4>
        <div className="flex items-baseline gap-1">
          <span className="text-[10px] font-bold text-indigo-400">
            {displayPrice.toLocaleString('vi-VN')} đ
          </span>
          {hasDiscount && (
            <span className="text-[8px] text-slate-500 line-through">
              {originalPrice.toLocaleString('vi-VN')} đ
            </span>
          )}
        </div>
        
        {/* Actions Row */}
        <div className="flex items-center gap-1.5 mt-0.5">
          <Link to={`/products/${id}`} className="px-2 py-0.5 rounded bg-[#1e2030] hover:bg-[#282a3f] text-slate-300 hover:text-white text-[9px] font-bold transition-all uppercase tracking-wide border border-[#2d314a]">
            Chi tiết
          </Link>
          <button
            onClick={handleAddToCart}
            disabled={adding || product.stock === 0}
            className={`px-2 py-0.5 rounded text-[9px] font-bold uppercase tracking-wide transition-all ${
              product.stock === 0
                ? 'bg-slate-800 text-slate-500 cursor-not-allowed border border-slate-700'
                : success 
                  ? 'bg-emerald-600 text-white'
                  : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm'
            }`}
          >
            {product.stock === 0 ? 'Hết' : adding ? 'Đang thêm' : success ? '✓ Thêm' : 'Thêm'}
          </button>
        </div>
      </div>
    </div>
  )
}

export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ content }) => {
  if (!content) return null

  // Extract all unique product IDs from standard relative routes in the text
  const productRegex = /\/products\/([^/?#\s"'()]+)/g
  const productIds: string[] = []
  let match
  while ((match = productRegex.exec(content)) !== null) {
    if (match[1] && !productIds.includes(match[1])) {
      productIds.push(match[1])
    }
  }

  return (
    <div className="space-y-2.5 text-slate-300 leading-relaxed text-sm w-full overflow-hidden flex flex-col">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
        components={{
          p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
          strong: ({ children }) => <strong className="font-semibold text-white">{children}</strong>,
          ul: ({ children }) => <ul className="list-disc pl-4 space-y-1 my-1 text-slate-300">{children}</ul>,
          ol: ({ children }) => <ol className="list-decimal pl-4 space-y-1 my-1 text-slate-300">{children}</ol>,
          li: ({ children }) => <li className="leading-relaxed">{children}</li>,
          h1: ({ children }) => <h1 className="text-lg font-bold text-white mt-4 mb-2">{children}</h1>,
          h2: ({ children }) => <h2 className="text-base font-bold text-indigo-300 mt-4 mb-2 border-b border-slate-800 pb-1">{children}</h2>,
          h3: ({ children }) => <h3 className="text-sm font-bold text-slate-100 mt-3 mb-1">{children}</h3>,
          h4: ({ children }) => <h4 className="text-xs font-bold text-slate-200 mt-2 mb-1">{children}</h4>,
          hr: () => <hr className="border-slate-800 my-3" />,
          code: ({ children }) => <code className="bg-[#1a1c29] text-indigo-300 px-1.5 py-0.5 rounded font-mono text-xs">{children}</code>,
          img: ({ src, alt }) => (
            <img 
              src={src} 
              alt={alt} 
              className="max-w-[140px] md:max-w-[180px] h-auto rounded-lg border border-slate-800/80 my-2 shadow-md block bg-white p-1.5 mix-blend-multiply" 
            />
          ),
          a: ({ href, children }) => {
            const targetHref = href || ''
            if (targetHref.startsWith('/')) {
              return (
                <Link to={targetHref} className="text-indigo-400 hover:text-indigo-300 underline font-semibold transition-colors">
                  {children}
                </Link>
              )
            }
            return (
              <a href={targetHref} target="_blank" rel="noopener noreferrer" className="text-indigo-400 hover:text-indigo-300 underline font-semibold transition-colors">
                {children}
              </a>
            )
          },
          table: ({ children }) => (
            <div className="overflow-x-auto my-3 rounded-lg border border-slate-700/60 shadow-sm w-full">
              <table className="min-w-full divide-y divide-slate-700/85 text-xs text-left text-slate-300 bg-[#0f111a]/40 border-collapse">
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-[#161925] text-slate-100 font-semibold border-b border-slate-700/80">
              {children}
            </thead>
          ),
          th: ({ children }) => (
            <th className="px-4 py-2.5 text-left font-semibold border-r border-slate-700/50 last:border-r-0">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="px-4 py-2.5 border-b border-slate-700/50 border-r border-slate-700/50 last:border-r-0 leading-relaxed align-middle">
              {children}
            </td>
          ),
          tr: ({ children }) => (
            <tr className="hover:bg-slate-800/40 even:bg-slate-800/15 transition-colors duration-100 border-b border-slate-700/30 last:border-b-0">
              {children}
            </tr>
          ),
        }}
      >
        {content}
      </ReactMarkdown>

      {/* Dedicated Related Products Section at the Bottom */}
      {productIds.length > 0 && (
        <div className="mt-3 pt-3 border-t border-[#1e2235] w-full flex flex-col gap-2">
          <span className="text-[10px] text-slate-450 font-bold uppercase tracking-wider block text-left">
            🛍️ Sản phẩm được đề xuất:
          </span>
          <div className="flex flex-wrap gap-2.5 justify-start">
            {productIds.map(id => (
              <InlineProductCard key={id} id={id} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
export default MarkdownRenderer
