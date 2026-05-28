# Frontend - Ứng dụng E-Commerce

Dự án frontend xây dựng bằng **React**, **TypeScript**, **Vite**, **Tailwind CSS** và **Redux**.

## Yêu cầu

- **Node.js** >= 16.0.0
- **npm** hoặc **yarn**

## Cài đặt

### 1. Cài đặt dependencies

```bash
# Sử dụng npm
npm install

# Hoặc sử dụng yarn
yarn install
```

### 2. Cấu hình biến môi trường

Tạo file `.env.local` dựa trên file `.env.example`:

```bash
cp .env.example .env.local
```

Nội dung file `.env.local`:

```
VITE_API_BASE_URL=http://localhost:8080/api
VITE_GOOGLE_CLIENT_ID=your_google_client_id_here
```

**Lưu ý quan trọng:**
- Đảm bảo backend API đang chạy trên `http://localhost:8080` trước khi khởi động frontend
- Để sử dụng Google Login, bạn cần:
  1. Tạo Google OAuth 2.0 credentials tại [Google Cloud Console](https://console.cloud.google.com)
  2. Lấy **Client ID** và cập nhật vào `VITE_GOOGLE_CLIENT_ID` trong `.env.local`
  3. Thêm `http://localhost:5173` vào **Authorized JavaScript origins** trong Google Console

## Chạy dự án

### Chế độ Development (Phát triển)

Khởi động development server với Hot Module Replacement (HMR):

```bash
npm run dev
```

Ứng dụng sẽ tự động mở tại `http://localhost:5173` trong trình duyệt.

### Build Production

Biên dịch TypeScript và tạo bản build tối ưu hóa cho production:

```bash
npm run build
```

Output sẽ được lưu trong thư mục `dist/`.

### Xem Preview Build

Xem trước bản build production trên local:

```bash
npm run preview
```

### Kiểm tra Type Errors

Kiểm tra lỗi TypeScript mà không cần biên dịch:

```bash
npm run lint
```

## Cấu trúc dự án

```
frontend/
├── src/
│   ├── App.tsx              # Component chính
│   ├── main.tsx             # Entry point
│   ├── index.css            # Global styles
│   ├── components/          # React components
│   ├── pages/               # Page components
│   ├── store/               # Redux store
│   ├── hooks/               # Custom React hooks
│   ├── services/            # API services
│   ├── types/               # TypeScript types
│   └── utils/               # Utility functions
├── public/                  # Static files
├── index.html              # HTML template
├── vite.config.ts          # Vite configuration
├── tailwind.config.js      # Tailwind CSS config
├── postcss.config.js       # PostCSS config
└── tsconfig.json           # TypeScript config
```

## Công nghệ sử dụng

| Công nghệ | Phiên bản | Mục đích |
|-----------|----------|---------|
| React | 18.3.1 | UI Library |
| TypeScript | 5.5.4 | Static typing |
| Vite | 5.4.2 | Build tool & Dev server |
| Tailwind CSS | 3.4.13 | Styling |
| Redux Toolkit | 2.2.7 | State management |
| React Router | 6.26.2 | Routing |
| Axios | 1.7.7 | HTTP client |
| @react-oauth/google | 0.12.1 | Google OAuth login |

## Tính năng xác thực (Authentication)

### Form Validation (Client-side)

Tất cả các form đều có validation **ngay trên frontend** trước khi gửi request:

#### Login Form
- **Email**: Bắt buộc, định dạng email hợp lệ
- **Password**: Bắt buộc, tối thiểu 6 ký tự

#### Register Form
- **Full Name**: Bắt buộc, tối thiểu 2 ký tự
- **Email**: Bắt buộc, định dạng email hợp lệ
- **Password**: Bắt buộc, tối thiểu 6 ký tự, có thể xem strength indicator

### Đăng nhập / Đăng ký

Ứng dụng hỗ trợ 2 phương thức xác thực:

#### 1. Email & Password
- Đăng nhập: `/login`
- Đăng ký: `/register`

#### 2. Google OAuth
- Click nút "Sign in with Google"
- Backend sẽ tự động tạo tài khoản nếu chưa tồn tại
- **Lưu ý:** Email của Google account không được trùng lặp

### Flow xác thực

```
User điền form → Frontend validate → Gửi request → Backend validate lại → 
Backend trả về token + user info → Lưu token → Redirect đến home
```

**Backend luôn là chốt chặn cuối cùng**, frontend chỉ validate để cải thiện UX.

## Lưu ý

- **Backend API:** Đảm bảo backend chạy trên `http://localhost:8080` trước khi khởi động frontend
- **Node version:** Khuyên dùng Node.js 18+ để tránh các vấn đề tương thích
- **Port mặc định:** Development server chạy trên port `5173` (tự động mở browser)
- **Token storage:** JWT token được lưu trữ an toàn trong localStorage
