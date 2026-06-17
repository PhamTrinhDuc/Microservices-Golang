import { useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { keycloak } from '../utils/keycloak'

const LoginPage = () => {
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/')
    }
  }, [isAuthenticated, navigate])

  const handleGoogleLogin = () => {
    void keycloak.login({ idpHint: 'google' })
  }

  const handleEmailLogin = () => {
    void keycloak.login()
  }

  return (
    <div className="min-h-screen bg-[#FAF9F5] text-[#18181B] font-sans antialiased flex flex-col justify-between p-4 md:p-6">
      {/* Dynamic Font Loader & Local CSS overrides */}
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&display=swap');
        
        .font-serif-display {
          font-family: 'Plus Jakarta Sans', sans-serif;
        }
        .font-sans-body {
          font-family: 'Plus Jakarta Sans', sans-serif;
        }
        
        @keyframes fadeInSlideUp {
          from {
            opacity: 0;
            transform: translateY(12px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        
        .animate-fade-in-up {
          animation: fadeInSlideUp 0.7s cubic-bezier(0.16, 1, 0.3, 1) forwards;
        }
      `}</style>

      {/* 1. Header */}
      <header className="w-full max-w-5xl mx-auto flex justify-between items-end border-b border-[#E4E4E7] pb-3 animate-fade-in-up">
        <div className="flex items-baseline gap-4">
          <Link to="/" className="font-serif-display text-2xl font-bold tracking-[0.05em] text-[#18181B] hover:opacity-85 transition-opacity">
            BELIBELI.
          </Link>
          <span className="text-[10px] uppercase tracking-[0.2em] text-[#8C8273] font-bold hidden md:inline border-l border-[#E4E4E7] pl-4">
            Không gian tuyển chọn các thiết kế bền vững
          </span>
        </div>
        <Link to="/" className="text-xs font-sans-body text-[#8C8273] hover:text-[#18181B] transition-colors tracking-widest uppercase font-bold">
          Trang chủ ↗
        </Link>
      </header>

      {/* 2. Main Boxed Content Card */}
      <main className="w-full max-w-4xl mx-auto my-6 flex items-center justify-center flex-grow animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
        <div className="w-full grid grid-cols-1 lg:grid-cols-12 border border-[#E4E4E7] bg-white rounded-none overflow-hidden shadow-[0_8px_30px_rgba(24,24,27,0.015)]">
          
          {/* Left Pane - Editorial Content */}
          <div className="lg:col-span-7 bg-[#F6F5F0] p-6 md:p-10 flex flex-col justify-between border-b lg:border-b-0 lg:border-r border-[#E4E4E7]">
            <div className="space-y-4 max-w-md py-2">
              <span className="text-[10px] uppercase tracking-[0.25em] text-[#8C8273] font-extrabold block">TẬP I // ĐỊNH HÌNH TỐI GIẢN</span>
              <h1 className="font-serif-display text-4xl md:text-5xl font-semibold leading-[1.1] text-[#18181B] tracking-tight">
                Vẻ đẹp <br />
                nằm ở sự <br />
                <span className="italic font-normal">tiết chế</span> và chuẩn xác.
              </h1>
              <p className="font-sans-body text-sm text-[#3F3F46] leading-relaxed font-normal">
                BeliBeli tuyển lựa những sản phẩm mang ngôn ngữ thiết kế tinh giản, trường tồn cùng thời gian và tối ưu hóa cho công năng sử dụng hàng ngày.
              </p>
            </div>
            
            <div className="pt-6 border-t border-[#E4E4E7]/60 flex justify-between items-baseline text-[9px] tracking-[0.2em] text-[#8C8273] uppercase font-bold">
              <span>VOLUME 01</span>
              <span>SAIGON, VIETNAM</span>
            </div>
          </div>

          {/* Right Pane - Form Controls */}
          <div className="lg:col-span-5 bg-white p-6 md:p-10 flex flex-col justify-center">
            <div className="mb-6">
              <span className="text-[10px] uppercase tracking-[0.25em] text-[#8C8273] font-extrabold block mb-1">Đăng nhập</span>
              <h2 className="font-serif-display text-3xl font-medium text-[#18181B] mb-2 tracking-tight">Chào mừng quay trở lại</h2>
              <p className="font-sans-body text-xs text-[#52525B] font-normal leading-relaxed">
                Quản lý giỏ hàng của bạn và khám phá những trải nghiệm đặc quyền dành riêng cho thành viên.
              </p>
            </div>

            <div className="space-y-3">
              {/* Google Button */}
              <button
                onClick={handleGoogleLogin}
                className="w-full flex items-center justify-between bg-white border border-[#18181B] hover:bg-[#18181B] hover:text-[#FAF9F5] text-[#18181B] font-sans-body font-bold text-xs py-3 px-4 rounded-none transition-all duration-300 active:scale-[0.99] group cursor-pointer"
              >
                <span className="tracking-widest uppercase text-[10px]">Đăng nhập bằng Google</span>
                <svg className="w-4 h-4 transition-transform group-hover:translate-x-1" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v3.92h6.69c-.29 1.5-.14 3.01-.97 4.29l3.11 2.42c1.82-1.68 2.91-4.17 2.91-6.56z"
                  />
                  <path
                    fill="currentColor"
                    d="M12 24c3.24 0 5.97-1.08 7.96-2.91l-3.11-2.42c-.9.6-2.01.99-3.23.99-3.11 0-5.74-2.11-6.68-4.96L3.74 17.65C5.71 21.6 9.81 24 12 24z"
                  />
                  <path
                    fill="currentColor"
                    d="M5.32 14.7c-.24-.7-.38-1.4-.38-2.2s.14-1.5.38-2.2L1.58 7.38C.57 9.41 0 11.66 0 14s.57 4.59 1.58 6.62l3.74-2.92z"
                  />
                  <path
                    fill="currentColor"
                    d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 9.81 0 5.71 2.4 3.74 6.35L7.48 9.3c.94-2.85 3.57-4.55 6.52-4.55z"
                  />
                </svg>
              </button>

              {/* Divider */}
              <div className="flex items-center py-1.5 w-full">
                <div className="flex-1 h-[0.5px] bg-[#E4E4E7]"></div>
                <span className="text-[9px] text-[#8C8273] font-extrabold uppercase tracking-[0.25em] px-3 select-none">Hoặc</span>
                <div className="flex-1 h-[0.5px] bg-[#E4E4E7]"></div>
              </div>

              {/* Email Button */}
              <button
                onClick={handleEmailLogin}
                className="w-full flex items-center justify-between bg-[#18181B] hover:bg-[#FAF9F5] border border-[#18181B] text-white hover:text-[#18181B] font-sans-body font-bold text-xs py-3 px-4 rounded-none transition-all duration-300 active:scale-[0.99] group cursor-pointer"
              >
                <span className="tracking-widest uppercase text-[10px]">Đăng nhập với Email</span>
                <span className="text-xs transition-transform group-hover:translate-x-1">→</span>
              </button>
            </div>

            <div className="mt-5 text-xs font-sans-body text-[#52525B] font-medium text-center lg:text-left">
              Chưa có tài khoản?{' '}
              <Link to="/register" className="text-[#18181B] font-bold hover:underline underline-offset-4 transition-all">
                Tạo tài khoản mới
              </Link>
            </div>
          </div>

        </div>
      </main>

      {/* 3. Footer */}
      <footer className="w-full max-w-5xl mx-auto border-t border-[#E4E4E7] pt-3 flex flex-col md:flex-row justify-between items-center text-[9px] tracking-[0.2em] text-[#8C8273] uppercase font-bold gap-4 animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
        <span>© 2026 BELIBELI INC. ALL RIGHTS RESERVED.</span>
        <div className="flex gap-6">
          <Link to="/policy" className="hover:text-[#18181B] transition-colors">Điều khoản</Link>
          <span>/</span>
          <Link to="/policy" className="hover:text-[#18181B] transition-colors">Bảo mật</Link>
        </div>
      </footer>
    </div>
  )
}

export default LoginPage
