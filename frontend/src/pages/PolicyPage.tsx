import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { policyAPI } from '../services/policyAPI'
import type { Policy } from '../types'

export default function PolicyPage() {
  const { slug } = useParams<{ slug: string }>()
  const [policy, setPolicy] = useState<Policy | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const categoriesMap: Record<string, string> = {
    refund: 'Chính sách Đổi trả',
    shipping: 'Chính sách Vận chuyển',
    privacy: 'Chính sách Bảo mật',
    terms: 'Điều khoản Dịch vụ',
    payment: 'Chính sách Thanh toán'
  }

  useEffect(() => {
    const fetchPolicy = async () => {
      if (!slug) return
      try {
        setLoading(true)
        setError(null)
        const data = await policyAPI.getPolicyBySlug(slug)
        if (data) {
          setPolicy(data)
        } else {
          setError('Không tìm thấy chính sách này')
        }
      } catch (err: any) {
        setError(err.message || 'Lỗi khi tải thông tin chính sách')
      } finally {
        setLoading(false)
      }
    }

    void fetchPolicy()
  }, [slug])

  if (loading) {
    return (
      <div className="flex-1 flex justify-center items-center py-20 bg-white min-h-[calc(100vh-140px)]">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-black"></div>
      </div>
    )
  }

  if (error || !policy) {
    return (
      <div className="flex-1 flex flex-col justify-center items-center py-20 bg-white text-center px-4 min-h-[calc(100vh-140px)]">
        <h2 className="text-xl font-black text-neutral-800 uppercase tracking-wide">Đã xảy ra lỗi</h2>
        <p className="text-sm text-neutral-500 mt-2">{error || 'Chính sách không tồn tại hoặc đã bị ẩn.'}</p>
        <Link
          to="/"
          className="mt-6 bg-black text-white text-xs font-black uppercase tracking-wider py-3 px-6 rounded hover:bg-neutral-800 transition-colors"
        >
          Quay lại Trang chủ
        </Link>
      </div>
    )
  }

  return (
    <div className="flex-1 bg-white py-12 md:py-16 font-sans min-h-[calc(100vh-140px)]">
      <div className="max-w-3xl mx-auto px-6">
        {/* Navigation Breadcrumb */}
        <nav className="text-[10px] uppercase font-bold tracking-wider text-neutral-400 mb-6 flex items-center gap-1.5">
          <Link to="/" className="hover:text-black transition-colors">Trang chủ</Link>
          <span>/</span>
          <span className="text-neutral-500">Chính sách & Hỗ trợ</span>
          <span>/</span>
          <span className="text-neutral-800 font-extrabold">{categoriesMap[policy.category] || policy.category}</span>
        </nav>

        {/* Policy Header */}
        <header className="border-b border-neutral-100 pb-6 mb-8">
          <span className="bg-neutral-100 text-neutral-600 px-2.5 py-0.5 rounded text-[10px] font-black uppercase tracking-wide">
            {categoriesMap[policy.category] || policy.category}
          </span>
          <h1 className="text-2xl md:text-3xl font-black text-neutral-900 tracking-tight mt-3">
            {policy.title}
          </h1>
          {policy.updated_at && (
            <p className="text-[10px] font-mono text-neutral-400 mt-2">
              CẬP NHẬT LẦN CUỐI: {new Date(policy.updated_at).toLocaleDateString('vi-VN')}
            </p>
          )}
        </header>

        {/* Policy Content */}
        <article className="text-sm text-neutral-700 leading-relaxed space-y-6 whitespace-pre-wrap font-sans">
          {policy.content}
        </article>

        {/* Support Section footer */}
        <section className="bg-neutral-50 border border-neutral-200 rounded-lg p-6 mt-12 text-xs flex flex-col md:flex-row justify-between items-center gap-4">
          <div>
            <h4 className="font-extrabold text-neutral-900 uppercase">Bạn vẫn còn câu hỏi thắc mắc?</h4>
            <p className="text-neutral-500 mt-1">Trò chuyện ngay với Trợ lý AI của chúng tôi ở góc dưới màn hình để được giải đáp tức thì.</p>
          </div>
          <Link
            to="/"
            className="shrink-0 bg-black text-white text-[10px] font-black uppercase tracking-wider py-2.5 px-5 rounded hover:bg-neutral-800 transition-colors"
          >
            Hỏi Trợ lý AI
          </Link>
        </section>
      </div>
    </div>
  )
}
