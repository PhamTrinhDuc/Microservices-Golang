# Frontend Design Plan - Ecommerce System

## System Overview
The backend is a full-featured ecommerce platform built with Go (Gin framework) and PostgreSQL with the following core modules:
- **User Management**: Registration, authentication (JWT), user profiles
- **Catalog**: Products, categories, brands, variants with options and specifications
- **Cart**: Session-based and user-based cart management
- **Orders**: Order processing, checkout, multiple payment methods (COD, Bank Transfer, PayOS)
- **Addresses**: User delivery address management
- **Promotions**: Product promotions and voucher system

---

## Frontend Architecture

### Tech Stack
- **Framework**: React 18+ (TypeScript recommended)
- **State Management**: Redux Toolkit or Zustand for global state
- **UI Library**: Material-UI (MUI), TailwindCSS, or custom component library
- **API Communication**: Axios or React Query for REST API calls
- **Routing**: React Router v6+
- **Build Tool**: Vite or Create React App

### Key Pages/Sections

#### 1. **Authentication Module**
- **Login Page**: Email/password authentication with JWT token storage
- **Registration Page**: Full name, email, password validation
- **Password Recovery**: Email-based password reset
- **JWT Token Management**: Secure storage (httpOnly cookies or localStorage), token refresh

#### 2. **Home & Browse**
- **Homepage**: Featured products, banners, promotions carousel
- **Category Listing**: Hierarchical category navigation, subcategories
- **Brand Directory**: Brand logos and dedicated brand pages
- **Product Search**: Full-text search with filters (category, brand, price range, ratings)

#### 3. **Product Details**
- **Product Gallery**: Image slider with thumbnail preview
- **Product Info**: Name, description, brand, category, price, base price
- **Specifications**: Grouped technical specs display (e.g., storage, RAM, display)
- **Variant Selection**: Interactive option selectors (e.g., color, size) with price updates
- **Stock Status**: Real-time availability, low-stock indicators
- **Related Products**: Siblings (same category/brand variants)
- **Reviews/Ratings**: Display and submission (if supported)

#### 4. **Cart Management**
- **Cart Page**: 
  - List cart items with images, variants, prices
  - Quantity adjustment buttons (increment/decrement)
  - Remove item functionality
  - Cart summary: subtotal, tax estimation, shipping calculation
  - "Proceed to Checkout" button
- **Session-based Cart**: Support anonymous shopping via sessionID
- **Merge Cart**: When user logs in after shopping as guest

#### 5. **Checkout Flow**
- **Shipping Address**:
  - List saved addresses with default selection
  - Add new address form (province, district, ward, detail address)
  - Ability to edit/delete addresses
- **Order Review**:
  - Display selected items, quantities, and prices
  - Shipping method selection (if available)
  - Voucher/Promotion code input
- **Payment Method Selection**:
  - COD (Cash on Delivery)
  - Bank Transfer with account details
  - PayOS integration (QR code, payment gateway)
- **Order Confirmation**: Order code, estimated delivery date, order status tracking

#### 6. **User Account Dashboard**
- **Profile Management**:
  - Edit full name, phone, gender, date of birth
  - Avatar upload
  - Email display (read-only or change with verification)
- **Address Book**: CRUD for delivery addresses
- **Order History**:
  - List all orders with filters (status, date range)
  - Order detail view with timeline (order, payment, shipping status)
  - Order tracking (shipping provider, tracking code)
  - Cancel order (if eligible)

#### 7. **Admin/Seller Dashboard** (if applicable)
- **Product Management**: CRUD for products, variants, options
- **Promotion Management**: Create and manage promotions
- **Voucher Management**: Create and manage vouchers
- **Order Management**: View, update order statuses, manage shipping
- **Inventory Management**: Stock levels, low-stock alerts
- **Analytics**: Sales, revenue, popular products

#### 8. **Additional Components**
- **Navigation**: Header with logo, search, categories, user menu, cart icon
- **Footer**: Links, contact info, about, terms & conditions
- **Notifications**: Toast messages for actions (add to cart, errors)
- **Loading States**: Skeletons, spinners for async operations
- **Error Handling**: Error pages (404, 500, network errors)

---

## UI/UX Considerations

### Design Principles
- **Mobile-First Responsive Design**: Optimized for mobile, tablet, and desktop
- **Accessibility**: WCAG 2.1 compliance (keyboard navigation, color contrast, ARIA labels)
- **Performance**: Lazy loading, image optimization, code splitting
- **Consistency**: Unified design system with reusable components

### Color Scheme & Branding
- Primary color: TBD (e.g., #007BFF for blue)
- Secondary color: TBD
- Neutral colors: Grays for backgrounds and text
- Accent colors: Green (success), Red (error/danger), Yellow (warning)

### Typography
- **Headings**: Bold, clear hierarchy (H1, H2, H3, H4)
- **Body Text**: Readable font size (16px+), adequate line height (1.5-1.6)
- **Buttons**: Clear CTAs with hover states

### Layout Components
- **Grid/Card System**: Responsive grid for product listings
- **Modal Dialogs**: For confirmations and forms
- **Sidebar**: For filters on browse pages
- **Breadcrumbs**: Navigation aid on product detail pages

---

## Technical Implementation Todos

### Phase 1: Foundation
- [ ] project-setup: Initialize React project with TypeScript, routing, and API client
- [ ] auth-module: Implement login, registration, JWT token management
- [ ] api-integration: Set up Axios/React Query client with interceptors for API calls

### Phase 2: Core Features
- [ ] homepage: Design and build homepage with featured products
- [ ] product-listing: Build category/search pages with filtering
- [ ] product-details: Implement product detail page with variants and specs
- [ ] cart-feature: Build cart management with local/session storage

### Phase 3: Checkout & Orders
- [ ] address-management: Build address CRUD and selection UI
- [ ] checkout-flow: Implement multi-step checkout (address → payment method → review)
- [ ] payment-integration: Integrate PayOS (if needed), handle COD/Bank transfer options
- [ ] order-confirmation: Design order confirmation and tracking page

### Phase 4: User Account
- [ ] user-profile: Build profile management (edit info, avatar upload)
- [ ] order-history: Implement order history with filtering and detail view
- [ ] address-book: Dedicated address management in user account

### Phase 5: Admin/Seller (Optional)
- [ ] admin-dashboard: Build admin dashboard structure
- [ ] product-admin: CRUD UI for products and variants
- [ ] order-admin: Order management interface
- [ ] promotion-admin: Create and manage promotions/vouchers

### Phase 6: Polish & Optimization
- [ ] ui-refinement: Design system components, theming, accessibility
- [ ] performance: Image optimization, lazy loading, code splitting
- [ ] error-handling: Global error boundary, error pages, retry logic
- [ ] testing: Unit tests for components, integration tests for critical flows
- [ ] deployment: Build configuration, CI/CD setup

---

## Key Integration Points

1. **Authentication**: Use JWT tokens from /api/auth/login, refresh tokens via middleware
2. **Product Search**: Leverage /api/products/search with filters (category, brand, price)
3. **Cart**: Support both session-based (anonymous) and user-based cart storage
4. **Checkout**: Integrate with /api/orders/checkout, handle PayOS redirects
5. **User Account**: Fetch user data from /api/users/{id}, manage addresses
6. **Vouchers**: Validate and apply voucher codes during checkout

---

## Performance & SEO Goals
- **Page Load Time**: < 3 seconds (Lighthouse score > 80)
- **SEO**: Meta tags, open graph, structured data for products
- **Image Optimization**: WebP format with fallbacks, responsive images
- **Accessibility**: WCAG 2.1 AA compliance

---

## Browser & Device Support
- Modern browsers: Chrome, Firefox, Safari, Edge (latest 2 versions)
- Mobile support: iOS Safari, Chrome Mobile, Samsung Internet
- Minimum viewport: 320px (mobile phones)

---

## Future Enhancements
- Product reviews and ratings system
- Wishlist/Favorites feature
- Real-time inventory updates via WebSocket
- Live chat support
- Multi-language support (i18n)
- Dark mode theme
- Progressive Web App (PWA) capabilities
