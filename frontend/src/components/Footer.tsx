import { Link } from 'react-router-dom'

const Footer = () => {
  return (
    <footer className="bg-[#18181B] text-[#E4E4E7] font-sans border-t border-[#E4E4E7]/25">
      {/* Top Footer Section */}
      <div className="mx-auto max-w-7xl px-4 py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-12">
          
          {/* Logo & Tagline Column */}
          <div className="lg:col-span-2 space-y-6">
            <Link to="/" className="inline-block">
              <span className="font-serif-display text-2xl font-bold tracking-[0.15em] text-[#FAF9F5]">
                BELIBELI.
              </span>
            </Link>
            <p className="text-[10px] uppercase tracking-[0.25em] text-[#8C8273] font-semibold italic">
              "Curated Essentials for the Modern Mind"
            </p>
            <p className="text-xs text-[#A1A1AA] leading-relaxed max-w-sm font-light">
              BeliBeli tuyển lựa những sản phẩm mang ngôn ngữ thiết kế tinh giản, trường tồn cùng thời gian và tối ưu hóa cho công năng sử dụng hàng ngày.
            </p>
            
            {/* Social Icons - Clean tracking style */}
            <div className="flex items-center gap-4 pt-2 text-[10px] tracking-widest uppercase font-semibold text-[#8C8273]">
              <a href="#" className="hover:text-[#FAF9F5] transition-colors">FB</a>
              <span>/</span>
              <a href="#" className="hover:text-[#FAF9F5] transition-colors">IG</a>
              <span>/</span>
              <a href="#" className="hover:text-[#FAF9F5] transition-colors">TW</a>
              <span>/</span>
              <a href="#" className="hover:text-[#FAF9F5] transition-colors">YT</a>
            </div>
          </div>

          {/* Column 1: BeliBeli */}
          <div>
            <h4 className="mb-4 text-[10px] font-bold uppercase tracking-[0.2em] text-[#8C8273]">Thương hiệu</h4>
            <ul className="space-y-3 text-xs text-[#A1A1AA] font-light">
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Câu chuyện BeliBeli</a>
              </li>
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Tuyển dụng</a>
              </li>
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Ấn phẩm truyền thông</a>
              </li>
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Liên hệ</a>
              </li>
            </ul>
          </div>

          {/* Column 2: Mua Sắm (Buy) */}
          <div>
            <h4 className="mb-4 text-[10px] font-bold uppercase tracking-[0.2em] text-[#8C8273]">Mua sắm</h4>
            <ul className="space-y-3 text-xs text-[#A1A1AA] font-light">
              <li>
                <Link to="/browse" className="hover:text-[#FAF9F5] transition-colors">Tất cả sản phẩm</Link>
              </li>
              <li>
                <Link to="/browse?sort=featured" className="hover:text-[#FAF9F5] transition-colors">Tuyển chọn nổi bật</Link>
              </li>
              <li>
                <Link to="/browse?sort=sale" className="hover:text-[#FAF9F5] transition-colors">Ưu đãi thành viên</Link>
              </li>
              <li>
                <Link to="/browse" className="hover:text-[#FAF9F5] transition-colors">Sản phẩm mới về</Link>
              </li>
            </ul>
          </div>

          {/* Column 3: Hướng Dẫn & Trợ Giúp */}
          <div>
            <h4 className="mb-4 text-[10px] font-bold uppercase tracking-[0.2em] text-[#8C8273]">Hỗ trợ</h4>
            <ul className="space-y-3 text-xs text-[#A1A1AA] font-light">
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Chăm sóc khách hàng</a>
              </li>
              <li>
                <Link to="/policies/chinh-sach-van-chuyen" className="hover:text-[#FAF9F5] transition-colors">Giao hàng & Vận chuyển</Link>
              </li>
              <li>
                <Link to="/policies/chinh-sach-doi-tra" className="hover:text-[#FAF9F5] transition-colors">Đổi trả hàng hóa</Link>
              </li>
              <li>
                <a href="#" className="hover:text-[#FAF9F5] transition-colors">Câu hỏi thường gặp</a>
              </li>
            </ul>
          </div>

        </div>

        {/* Newsletter Subscription Row */}
        <div className="mt-16 pt-8 border-t border-[#E4E4E7]/10 flex flex-col lg:flex-row items-center justify-between gap-6">
          <div className="space-y-1 text-center lg:text-left">
            <h4 className="text-sm font-medium text-[#FAF9F5]">Đăng ký nhận bản tin BeliBeli</h4>
            <p className="text-xs text-[#A1A1AA] font-light">Nhận mã ưu đãi đặc quyền và thông tin về các thiết kế giới hạn sớm nhất.</p>
          </div>
          <div className="flex w-full lg:w-auto max-w-md gap-0 border border-[#E4E4E7]/20 bg-[#18181B] focus-within:border-[#FAF9F5] transition-colors duration-300">
            <input
              type="email"
              placeholder="Địa chỉ Email..."
              className="flex-1 bg-transparent px-4 py-2.5 text-xs text-[#FAF9F5] placeholder-[#8C8273]/60 focus:outline-none font-medium"
            />
            <button className="bg-[#FAF9F5] text-[#18181B] px-6 py-2.5 text-xs font-semibold uppercase tracking-wider hover:bg-[#8C8273] hover:text-[#FAF9F5] transition-colors duration-300">
              Đăng ký
            </button>
          </div>
        </div>

        {/* Bottom Section: Divider & Copyright */}
        <div className="mt-12 pt-8 border-t border-[#E4E4E7]/10 flex flex-col md:flex-row items-center justify-between gap-4 text-[10px] text-[#8C8273] tracking-widest uppercase font-semibold">
          <p>© {new Date().getFullYear()} BELIBELI. TOÀN BỘ BẢN QUYỀN ĐƯỢC BẢO LƯU.</p>
          <div className="flex gap-6">
            <Link to="/policies/dieu-khoan-dich-vu" className="hover:text-[#FAF9F5] transition-colors">Điều khoản</Link>
            <span>/</span>
            <Link to="/policies/chinh-sach-bao-mat" className="hover:text-[#FAF9F5] transition-colors">Bảo mật</Link>
          </div>
        </div>
      </div>
    </footer>
  )
}

export default Footer
