import { Link } from 'react-router-dom'

const Footer = () => {
  return (
    <footer className="bg-[#0B0F19] text-neutral-400 font-sans border-t border-neutral-800">
      {/* Top Footer Section */}
      <div className="mx-auto max-w-7xl px-4 py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-10">
          
          {/* Logo & Tagline Column */}
          <div className="lg:col-span-2 space-y-5">
            <Link to="/" className="flex items-center gap-1.5 group w-fit">
              <div className="bg-white text-black px-2 py-1 rounded font-black text-lg tracking-wider">
                B
              </div>
              <span className="text-xl font-extrabold tracking-tight text-white">
                BeliBeli<span className="text-neutral-400 font-semibold text-sm">.com</span>
              </span>
            </Link>
            <p className="text-sm font-medium italic text-neutral-300">
              "Let's Shop Beyond Boundaries"
            </p>
            <p className="text-xs text-neutral-400 leading-relaxed max-w-sm">
              Trải nghiệm mua sắm không giới hạn với hàng triệu sản phẩm chất lượng, dịch vụ chăm sóc tận tâm và giao hàng siêu tốc trên toàn quốc.
            </p>
            
            {/* Social Icons */}
            <div className="flex items-center gap-3 pt-2">
              {['facebook', 'instagram', 'twitter', 'youtube'].map((social) => (
                <a
                  key={social}
                  href="#"
                  className="w-8 h-8 rounded-full bg-neutral-800 flex items-center justify-center text-neutral-400 hover:bg-white hover:text-black transition-all duration-200"
                >
                  <span className="sr-only">{social}</span>
                  <span className="text-xs uppercase font-extrabold">{social[0]}</span>
                </a>
              ))}
            </div>
          </div>

          {/* Column 1: BeliBeli */}
          <div>
            <h4 className="mb-4 text-xs font-bold uppercase tracking-wider text-white">BeliBeli</h4>
            <ul className="space-y-3 text-xs">
              <li>
                <a href="#" className="hover:text-white transition-colors">Về chúng tôi</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Tuyển dụng</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Blog tin tức</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Báo chí</a>
              </li>
            </ul>
          </div>

          {/* Column 2: Mua Sắm (Buy) */}
          <div>
            <h4 className="mb-4 text-xs font-bold uppercase tracking-wider text-white">Mua sắm</h4>
            <ul className="space-y-3 text-xs">
              <li>
                <Link to="/browse" className="hover:text-white transition-colors">Tất cả sản phẩm</Link>
              </li>
              <li>
                <Link to="/browse?sort=featured" className="hover:text-white transition-colors">Sản phẩm nổi bật</Link>
              </li>
              <li>
                <Link to="/browse?sort=sale" className="hover:text-white transition-colors text-red-400 font-medium">Săn Sale giá sốc</Link>
              </li>
              <li>
                <Link to="/browse" className="hover:text-white transition-colors">Bộ sưu tập mới</Link>
              </li>
            </ul>
          </div>

          {/* Column 3: Hướng Dẫn & Trợ Giúp */}
          <div>
            <h4 className="mb-4 text-xs font-bold uppercase tracking-wider text-white">Trợ giúp</h4>
            <ul className="space-y-3 text-xs">
              <li>
                <a href="#" className="hover:text-white transition-colors">Liên hệ chúng tôi</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Chính sách vận chuyển</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Đổi trả & hoàn tiền</a>
              </li>
              <li>
                <a href="#" className="hover:text-white transition-colors">Câu hỏi thường gặp</a>
              </li>
            </ul>
          </div>

        </div>

        {/* Newsletter Subscription Row */}
        <div className="mt-12 pt-8 border-t border-neutral-800 flex flex-col lg:flex-row items-center justify-between gap-6">
          <div className="space-y-1 text-center lg:text-left">
            <h4 className="text-sm font-bold text-white">Đăng ký nhận bản tin BeliBeli</h4>
            <p className="text-xs text-neutral-450">Nhận ngay ưu đãi 10% cho đơn hàng đầu tiên của bạn.</p>
          </div>
          <div className="flex w-full lg:w-auto max-w-md gap-2">
            <input
              type="email"
              placeholder="Nhập email của bạn..."
              className="flex-1 bg-neutral-900 border border-neutral-800 rounded px-4 py-2.5 text-xs text-white placeholder-neutral-500 focus:border-white focus:outline-none transition-colors"
            />
            <button className="bg-white text-black px-6 py-2.5 rounded text-xs font-bold hover:bg-neutral-200 transition-colors">
              Đăng ký
            </button>
          </div>
        </div>

        {/* Bottom Section: Divider & Copyright */}
        <div className="mt-12 pt-8 border-t border-neutral-800 flex flex-col md:flex-row items-center justify-between gap-4 text-[11px] text-neutral-550">
          <p>© {new Date().getFullYear()} BeliBeli.com. Tất cả quyền được bảo lưu.</p>
          <div className="flex gap-6 uppercase font-bold tracking-wider">
            <a href="#" className="hover:text-white transition-colors">Điều khoản dịch vụ</a>
            <a href="#" className="hover:text-white transition-colors">Chính sách bảo mật</a>
          </div>
        </div>
      </div>
    </footer>
  )
}

export default Footer

