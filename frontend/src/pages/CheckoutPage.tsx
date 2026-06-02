import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useCart } from '../hooks/useCart'
import { addressAPI } from '../services/addressAPI'
import { orderAPI } from '../services/orderAPI'
import { voucherAPI } from '../services/voucherAPI'
import { api } from '../services/api'
import VoucherSelector from '../components/VoucherSelector'
import type { Address, Store } from '../types'

export default function CheckoutPage() {
  const navigate = useNavigate()
  const { items, cartSubtotal, fetchCart } = useCart()

  // Addresses
  const [addresses, setAddresses] = useState<Address[]>([])
  const [selectedAddressId, setSelectedAddressId] = useState<number | null>(null)
  
  // New address form
  const [showAddressForm, setShowAddressForm] = useState(false)
  const [newAddress, setNewAddress] = useState({
    fullName: '',
    phone: '',
    province: '',
    district: '',
    ward: '',
    detailAddress: '',
  })

  // Manual receiver info if no addresses
  const [manualReceiver, setManualReceiver] = useState({
    name: '',
    phone: '',
    address: '',
  })

  // Stores
  const [stores, setStores] = useState<Store[]>([])
  const [selectedStoreId, setSelectedStoreId] = useState<number>(1)

  // Vouchers
  const [voucherCode, setVoucherCode] = useState('')
  const [appliedVoucherCode, setAppliedVoucherCode] = useState<string | null>(null)
  const [discountAmount, setDiscountAmount] = useState(0)
  const [voucherError, setVoucherError] = useState<string | null>(null)
  const [showVoucherModal, setShowVoucherModal] = useState(false)

  // Checkout choices
  const [paymentMethod, setPaymentMethod] = useState<'cod' | 'payos' | 'bank_transfer'>('cod')
  const [shippingProvider, setShippingProvider] = useState<'ghn' | 'ghtk'>('ghn')
  const [note, setNote] = useState('')

  // UI state
  const [loading, setLoading] = useState(false)
  const [checkoutError, setCheckoutError] = useState<string | null>(null)

  // Static shipping rules matching backend:
  // - Flat rate: 30,000 VND
  // - Free ship if order >= 500,000 VND
  const FREE_SHIPPING_THRESHOLD = 500000
  const SHIPPING_FEE = 30000
  const shippingCost = items.length === 0 ? 0 : (cartSubtotal >= FREE_SHIPPING_THRESHOLD ? 0 : SHIPPING_FEE)
  
  const finalTotal = Math.max(0, cartSubtotal + shippingCost - discountAmount)

  // Redirect if cart is empty
  useEffect(() => {
    if (items.length === 0) {
      navigate('/cart')
    }
  }, [items, navigate])

  // Load data
  useEffect(() => {
    const loadInitialData = async () => {
      try {
        // Load addresses
        const addrList = await addressAPI.getAddresses()
        setAddresses(addrList.filter(a => !a.is_deleted))
        const defAddr = addrList.find(a => a.is_default && !a.is_deleted)
        if (defAddr) {
          setSelectedAddressId(defAddr.id)
        } else if (addrList.length > 0) {
          setSelectedAddressId(addrList[0].id)
        }

        // Load stores
        const storeRes = await api.get<Store[]>('/stores')
        const activeStores = (storeRes.data || []).filter(s => s.is_active)
        setStores(activeStores)
        if (activeStores.length > 0) {
          setSelectedStoreId(activeStores[0].id)
        }
      } catch (err) {
        console.error('Lỗi khi tải dữ liệu checkout:', err)
      }
    }
    void loadInitialData()
  }, [])

  // Apply voucher function
  const handleApplyVoucher = async (code: string) => {
    if (!code.trim()) return
    try {
      setVoucherError(null)
      const res = await voucherAPI.applyVoucher({
        code: code.trim(),
        order_amount: cartSubtotal,
      })
      if (res.valid) {
        setAppliedVoucherCode(code.trim())
        setVoucherCode(code.trim())
        setDiscountAmount(res.discount_amount)
      } else {
        setVoucherError('Mã giảm giá không hợp lệ')
      }
    } catch (err: any) {
      setVoucherError(err.message || 'Lỗi khi áp dụng mã giảm giá')
      setDiscountAmount(0)
      setAppliedVoucherCode(null)
    }
  }

  // Remove voucher code
  const handleRemoveVoucher = () => {
    setAppliedVoucherCode(null)
    setVoucherCode('')
    setDiscountAmount(0)
    setVoucherError(null)
  }

  // Create address handler
  const handleCreateAddress = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newAddress.fullName || !newAddress.phone || !newAddress.province || !newAddress.district || !newAddress.ward || !newAddress.detailAddress) {
      alert('Vui lòng điền đầy đủ các thông tin địa chỉ')
      return
    }

    try {
      setLoading(true)
      const addr = await addressAPI.createAddress({
        full_name: newAddress.fullName,
        phone: newAddress.phone,
        province: newAddress.province,
        district: newAddress.district,
        ward: newAddress.ward,
        detail_address: newAddress.detailAddress,
        is_default: addresses.length === 0, // default if first address
      })
      setAddresses(prev => [...prev, addr])
      setSelectedAddressId(addr.id)
      setShowAddressForm(false)
      setNewAddress({
        fullName: '',
        phone: '',
        province: '',
        district: '',
        ward: '',
        detailAddress: '',
      })
    } catch (err: any) {
      alert(err.message || 'Không thể tạo địa chỉ mới')
    } finally {
      setLoading(false)
    }
  }

  // Place order handler
  const handlePlaceOrder = async () => {
    // Validate address configuration
    let addressPayload: any = {}
    if (addresses.length > 0) {
      if (!selectedAddressId) {
        setCheckoutError('Vui lòng chọn địa chỉ nhận hàng')
        return
      }
      addressPayload.address_id = selectedAddressId
    } else {
      if (!manualReceiver.name || !manualReceiver.phone || !manualReceiver.address) {
        setCheckoutError('Vui lòng nhập thông tin nhận hàng')
        return
      }
      addressPayload.receiver_name = manualReceiver.name
      addressPayload.receiver_phone = manualReceiver.phone
      addressPayload.receiver_address = manualReceiver.address
    }

    try {
      setLoading(true)
      setCheckoutError(null)

      const payload = {
        store_id: selectedStoreId,
        voucher_code: appliedVoucherCode,
        payment_method: paymentMethod,
        shipping_provider: shippingProvider,
        note: note ? note : undefined,
        ...addressPayload,
      }

      const res = await orderAPI.checkout(payload)

      // Fetch cart again to clear client state
      void fetchCart()

      if (paymentMethod === 'payos' && res.checkout_url) {
        // Redirect to PayOS checkout page
        window.location.href = res.checkout_url
      } else {
        // Go to success screen
        navigate(`/order-success?id=${res.order.id}&code=${res.order.order_code}`)
      }
    } catch (err: any) {
      setCheckoutError(err.message || 'Đã có lỗi xảy ra khi tạo đơn hàng. Vui lòng kiểm tra lại.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex-1 bg-neutral-50 py-8 font-sans">
      <div className="mx-auto max-w-7xl px-4">
        {/* Breadcrumb */}
        <nav className="mb-6 flex items-center gap-1.5 text-xs text-neutral-400 font-medium">
          <Link to="/" className="hover:text-black transition-colors">Trang chủ</Link>
          <span>/</span>
          <Link to="/cart" className="hover:text-black transition-colors">Giỏ hàng</Link>
          <span>/</span>
          <span className="text-neutral-800 font-semibold">Thanh toán</span>
        </nav>

        <h1 className="text-2xl font-black text-neutral-900 tracking-tight uppercase mb-8">
          Thanh Toán Đơn Hàng
        </h1>

        {checkoutError && (
          <div className="mb-6 border border-red-200 bg-red-50 text-red-700 text-xs font-semibold px-4 py-3 rounded">
            {checkoutError}
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Left Column: Checkout Information Form */}
          <div className="lg:col-span-8 space-y-6">
            
            {/* 1. STORE SELECTOR */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <h3 className="text-xs font-black text-neutral-950 uppercase tracking-widest border-b border-neutral-100 pb-2">
                1. Chọn Cửa Hàng Gần Nhất
              </h3>
              {stores.length === 0 ? (
                <p className="text-xs text-neutral-400 font-medium animate-pulse">Đang quét các chi nhánh khả dụng...</p>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {stores.map(store => (
                    <label
                      key={store.id}
                      className={`p-3 border rounded-lg flex items-start gap-3 cursor-pointer transition-all ${
                        selectedStoreId === store.id
                          ? 'border-black ring-1 ring-black bg-neutral-50'
                          : 'border-neutral-250 hover:border-neutral-350'
                      }`}
                    >
                      <input
                        type="radio"
                        name="store"
                        className="mt-1 accent-black"
                        checked={selectedStoreId === store.id}
                        onChange={() => setSelectedStoreId(store.id)}
                      />
                      <div className="text-xs">
                        <p className="font-bold text-neutral-800">{store.name}</p>
                        <p className="text-neutral-450 text-[10px] mt-0.5">
                          {store.road}, {store.ward}, {store.district}, {store.province}
                        </p>
                        {store.hotline && <p className="text-[10px] text-neutral-400 font-mono mt-1">📞 Hotline: {store.hotline}</p>}
                      </div>
                    </label>
                  ))}
                </div>
              )}
            </div>

            {/* 2. SHIPPING ADDRESS */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <div className="flex justify-between items-center border-b border-neutral-100 pb-2">
                <h3 className="text-xs font-black text-neutral-950 uppercase tracking-widest">
                  2. Thông Tin Địa Chỉ Nhận Hàng
                </h3>
                {addresses.length > 0 && !showAddressForm && (
                  <button
                    type="button"
                    onClick={() => setShowAddressForm(true)}
                    className="text-[10px] font-black text-neutral-700 hover:text-black uppercase tracking-wider"
                  >
                    + Thêm địa chỉ mới
                  </button>
                )}
              </div>

              {/* Add New Address Form Modal/Panel */}
              {showAddressForm && (
                <form onSubmit={handleCreateAddress} className="border border-neutral-200 rounded-lg p-4 bg-neutral-50 space-y-3">
                  <h4 className="text-[11px] font-bold text-neutral-800 uppercase tracking-wide">Thêm địa chỉ giao hàng mới</h4>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <input
                      type="text"
                      placeholder="Họ và tên người nhận *"
                      className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                      value={newAddress.fullName}
                      onChange={e => setNewAddress(prev => ({ ...prev, fullName: e.target.value }))}
                      required
                    />
                    <input
                      type="tel"
                      placeholder="Số điện thoại *"
                      className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                      value={newAddress.phone}
                      onChange={e => setNewAddress(prev => ({ ...prev, phone: e.target.value }))}
                      required
                    />
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    <input
                      type="text"
                      placeholder="Tỉnh / Thành phố *"
                      className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                      value={newAddress.province}
                      onChange={e => setNewAddress(prev => ({ ...prev, province: e.target.value }))}
                      required
                    />
                    <input
                      type="text"
                      placeholder="Quận / Huyện *"
                      className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                      value={newAddress.district}
                      onChange={e => setNewAddress(prev => ({ ...prev, district: e.target.value }))}
                      required
                    />
                    <input
                      type="text"
                      placeholder="Phường / Xã *"
                      className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                      value={newAddress.ward}
                      onChange={e => setNewAddress(prev => ({ ...prev, ward: e.target.value }))}
                      required
                    />
                  </div>

                  <input
                    type="text"
                    placeholder="Địa chỉ chi tiết (Số nhà, tên đường, căn hộ...) *"
                    className="border border-neutral-300 rounded px-3 py-2 text-xs w-full bg-white focus:outline-none focus:border-black"
                    value={newAddress.detailAddress}
                    onChange={e => setNewAddress(prev => ({ ...prev, detailAddress: e.target.value }))}
                    required
                  />

                  <div className="flex gap-2 justify-end pt-2">
                    <button
                      type="button"
                      onClick={() => setShowAddressForm(false)}
                      className="text-[10px] border border-neutral-300 px-4 py-2 font-bold uppercase rounded text-neutral-600 hover:text-black"
                    >
                      Hủy bỏ
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="text-[10px] bg-black text-white px-5 py-2 font-black uppercase rounded hover:bg-neutral-800 disabled:opacity-55"
                    >
                      Lưu địa chỉ
                    </button>
                  </div>
                </form>
              )}

              {/* Address Selection List */}
              {addresses.length > 0 ? (
                <div className="space-y-2">
                  {addresses.map(addr => (
                    <label
                      key={addr.id}
                      className={`p-3 border rounded-lg flex items-start gap-3 cursor-pointer transition-all ${
                        selectedAddressId === addr.id
                          ? 'border-black ring-1 ring-black bg-neutral-50'
                          : 'border-neutral-250 hover:border-neutral-350'
                      }`}
                    >
                      <input
                        type="radio"
                        name="address"
                        className="mt-1 accent-black"
                        checked={selectedAddressId === addr.id}
                        onChange={() => {
                          setSelectedAddressId(addr.id)
                          setShowAddressForm(false)
                        }}
                      />
                      <div className="text-xs">
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-neutral-850">{addr.full_name}</span>
                          <span className="text-neutral-400 font-mono">({addr.phone})</span>
                          {addr.is_default && (
                            <span className="text-[9px] bg-neutral-100 text-neutral-600 px-2 py-0.5 rounded font-black uppercase tracking-wider">Mặc định</span>
                          )}
                        </div>
                        <p className="text-neutral-500 text-[10px] mt-1">
                          {addr.detail_address}, Phường {addr.ward}, {addr.district}, {addr.province}
                        </p>
                      </div>
                    </label>
                  ))}
                </div>
              ) : !showAddressForm ? (
                /* Manual Inputs If User Has No Address Records */
                <div className="space-y-3">
                  <p className="text-[10px] text-neutral-400 font-semibold uppercase tracking-wider">Nhập thông tin người nhận hàng</p>
                  
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <input
                      type="text"
                      placeholder="Họ và tên người nhận *"
                      className="border border-neutral-250 rounded px-3 py-2 text-xs w-full focus:outline-none focus:border-black"
                      value={manualReceiver.name}
                      onChange={e => setManualReceiver(prev => ({ ...prev, name: e.target.value }))}
                    />
                    <input
                      type="tel"
                      placeholder="Số điện thoại người nhận *"
                      className="border border-neutral-250 rounded px-3 py-2 text-xs w-full focus:outline-none focus:border-black"
                      value={manualReceiver.phone}
                      onChange={e => setManualReceiver(prev => ({ ...prev, phone: e.target.value }))}
                    />
                  </div>

                  <input
                    type="text"
                    placeholder="Địa chỉ nhận hàng (Ví dụ: 123 Đường Nguyễn Trãi, Quận 1, TP. Hồ Chí Minh) *"
                    className="border border-neutral-250 rounded px-3 py-2 text-xs w-full focus:outline-none focus:border-black"
                    value={manualReceiver.address}
                    onChange={e => setManualReceiver(prev => ({ ...prev, address: e.target.value }))}
                  />
                </div>
              ) : null}
            </div>

            {/* 3. SHIPPING PROVIDER */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <h3 className="text-xs font-black text-neutral-950 uppercase tracking-widest border-b border-neutral-100 pb-2">
                3. Chọn Đơn Vị Vận Chuyển
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <label
                  className={`p-3 border rounded-lg flex items-center justify-between cursor-pointer transition-all ${
                    shippingProvider === 'ghn'
                      ? 'border-black ring-1 ring-black bg-neutral-50'
                      : 'border-neutral-250 hover:border-neutral-350'
                  }`}
                >
                  <div className="flex items-center gap-2 text-xs">
                    <input
                      type="radio"
                      name="shipping"
                      className="accent-black"
                      checked={shippingProvider === 'ghn'}
                      onChange={() => setShippingProvider('ghn')}
                    />
                    <div>
                      <span className="font-bold text-neutral-800">Giao Hàng Nhanh (GHN)</span>
                      <p className="text-[9px] text-neutral-450 mt-0.5">Thời gian nhận hàng dự kiến: 2 - 3 ngày</p>
                    </div>
                  </div>
                </label>

                <label
                  className={`p-3 border rounded-lg flex items-center justify-between cursor-pointer transition-all ${
                    shippingProvider === 'ghtk'
                      ? 'border-black ring-1 ring-black bg-neutral-50'
                      : 'border-neutral-250 hover:border-neutral-350'
                  }`}
                >
                  <div className="flex items-center gap-2 text-xs">
                    <input
                      type="radio"
                      name="shipping"
                      className="accent-black"
                      checked={shippingProvider === 'ghtk'}
                      onChange={() => setShippingProvider('ghtk')}
                    />
                    <div>
                      <span className="font-bold text-neutral-800">Giao Hàng Tiết Kiệm (GHTK)</span>
                      <p className="text-[9px] text-neutral-450 mt-0.5">Thời gian nhận hàng dự kiến: 2 - 4 ngày</p>
                    </div>
                  </div>
                </label>
              </div>
            </div>

            {/* 4. PAYMENT METHOD */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <h3 className="text-xs font-black text-neutral-950 uppercase tracking-widest border-b border-neutral-100 pb-2">
                4. Phương Thức Thanh Toán
              </h3>
              <div className="space-y-2.5">
                {/* PayOS */}
                <label
                  className={`p-4 border rounded-lg flex items-start gap-3.5 cursor-pointer transition-all ${
                    paymentMethod === 'payos'
                      ? 'border-black ring-1 ring-black bg-neutral-50'
                      : 'border-neutral-250 hover:border-neutral-350'
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    className="mt-1 accent-black"
                    checked={paymentMethod === 'payos'}
                    onChange={() => setPaymentMethod('payos')}
                  />
                  <div className="text-xs">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-neutral-800">Thanh toán qua PayOS</span>
                      <span className="text-[9px] bg-sky-50 text-sky-600 px-2 py-0.5 rounded font-black tracking-wider uppercase">Chuyển khoản QR siêu tốc</span>
                    </div>
                    <p className="text-neutral-450 text-[10px] mt-1 leading-relaxed">
                      Bạn sẽ được chuyển đến cổng thanh toán PayOS để quét mã QR ngân hàng.
                    </p>
                  </div>
                </label>

                {/* COD */}
                <label
                  className={`p-4 border rounded-lg flex items-start gap-3.5 cursor-pointer transition-all ${
                    paymentMethod === 'cod'
                      ? 'border-black ring-1 ring-black bg-neutral-50'
                      : 'border-neutral-250 hover:border-neutral-350'
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    className="mt-1 accent-black"
                    checked={paymentMethod === 'cod'}
                    onChange={() => setPaymentMethod('cod')}
                  />
                  <div className="text-xs">
                    <span className="font-bold text-neutral-800">Thanh toán khi nhận hàng (COD)</span>
                    <p className="text-neutral-450 text-[10px] mt-1 leading-relaxed">
                      Thanh toán bằng tiền mặt khi shipper giao hàng tận nơi.
                    </p>
                  </div>
                </label>

                {/* Bank Transfer */}
                <label
                  className={`p-4 border rounded-lg flex items-start gap-3.5 cursor-pointer transition-all ${
                    paymentMethod === 'bank_transfer'
                      ? 'border-black ring-1 ring-black bg-neutral-50'
                      : 'border-neutral-250 hover:border-neutral-350'
                  }`}
                >
                  <input
                    type="radio"
                    name="payment"
                    className="mt-1 accent-black"
                    checked={paymentMethod === 'bank_transfer'}
                    onChange={() => setPaymentMethod('bank_transfer')}
                  />
                  <div className="text-xs">
                    <span className="font-bold text-neutral-800">Chuyển khoản ngân hàng thủ công</span>
                    <p className="text-neutral-450 text-[10px] mt-1 leading-relaxed">
                      Chuyển khoản trực tiếp tới số tài khoản của Jiyuu Store. Nhân viên hỗ trợ sẽ xác nhận đơn hàng sau.
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {/* 5. NOTES */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-3">
              <h3 className="text-xs font-black text-neutral-950 uppercase tracking-widest border-b border-neutral-100 pb-2">
                5. Ghi Chú Đơn Hàng (Tùy chọn)
              </h3>
              <textarea
                rows={3}
                placeholder="Ví dụ: Giao giờ hành chính, gọi điện trước khi giao..."
                className="w-full border border-neutral-250 rounded px-3 py-2 text-xs bg-white focus:outline-none focus:border-black"
                value={note}
                onChange={e => setNote(e.target.value)}
              />
            </div>

          </div>

          {/* Right Column: Checkout Summary & Voucher Applier */}
          <div className="lg:col-span-4 space-y-6">
            
            {/* VOUCHER APPLICATION CARD */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <h3 className="text-xs font-black text-neutral-850 uppercase tracking-wider pb-2 border-b border-neutral-150">
                Mã Giảm Giá
              </h3>

              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Nhập mã ưu đãi..."
                  className="flex-1 border border-neutral-350 rounded px-3 py-2 text-xs bg-white uppercase font-mono font-bold focus:outline-none focus:border-black"
                  value={voucherCode}
                  onChange={e => setVoucherCode(e.target.value)}
                  disabled={!!appliedVoucherCode}
                />
                {appliedVoucherCode ? (
                  <button
                    type="button"
                    onClick={handleRemoveVoucher}
                    className="bg-neutral-100 hover:bg-neutral-200 border border-neutral-350 text-neutral-700 text-xs px-4 py-2 font-black uppercase rounded transition-colors"
                  >
                    Hủy
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => handleApplyVoucher(voucherCode)}
                    disabled={!voucherCode.trim()}
                    className="bg-black text-white text-xs px-4 py-2 font-black uppercase rounded hover:bg-neutral-850 disabled:opacity-35 transition-colors"
                  >
                    Áp dụng
                  </button>
                )}
              </div>

              {voucherError && (
                <p className="text-[10px] text-red-650 font-bold tracking-tight">{voucherError}</p>
              )}

              {appliedVoucherCode && (
                <div className="bg-green-50 border border-green-200 text-green-800 rounded p-3 flex items-start gap-2.5 text-[11px]">
                  <div>
                    <span className="font-extrabold uppercase font-mono bg-green-100 px-1 py-0.5 rounded text-green-900">{appliedVoucherCode}</span>
                    <p className="mt-1 font-semibold">Đã áp dụng giảm giá -{discountAmount.toLocaleString('vi-VN')} đ vào đơn hàng.</p>
                  </div>
                </div>
              )}

              <div className="text-center border-t border-neutral-100 pt-3">
                <button
                  type="button"
                  onClick={() => setShowVoucherModal(true)}
                  className="text-xs font-black text-neutral-850 hover:underline uppercase tracking-wide"
                >
                  Xem tất cả mã giảm giá
                </button>
              </div>
            </div>

            {/* ORDER PRODUCTS SUMMARY LIST */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
              <h3 className="text-xs font-black text-neutral-850 uppercase tracking-wider pb-2 border-b border-neutral-150">
                Tóm Tắt Đơn Hàng
              </h3>

              <div className="divide-y divide-neutral-150 max-h-[220px] overflow-y-auto pr-1">
                {items.map(item => (
                  <div key={item.id} className="py-2.5 flex justify-between gap-3 text-xs">
                    <div className="flex-1 min-w-0">
                      <p className="font-bold text-neutral-800 truncate">{item.product_name}</p>
                      <p className="text-[10px] text-neutral-450 mt-0.5">
                        SL: {item.quantity} {item.variant_name ? `| Phân loại: ${item.variant_name}` : ''}
                      </p>
                    </div>
                    <span className="font-mono font-semibold text-neutral-850 shrink-0 select-none">
                      {(item.price * item.quantity).toLocaleString('vi-VN')} đ
                    </span>
                  </div>
                ))}
              </div>

              {/* Price Breakdown */}
              <div className="border-t border-neutral-150 pt-4 space-y-2 text-xs">
                <div className="flex justify-between text-neutral-550">
                  <span>Tạm tính</span>
                  <span className="font-semibold text-neutral-800">{cartSubtotal.toLocaleString('vi-VN')} đ</span>
                </div>
                
                <div className="flex justify-between text-neutral-550">
                  <span>Phí vận chuyển</span>
                  <span className="font-semibold text-neutral-800">
                    {shippingCost === 0 ? 'Miễn phí' : `${shippingCost.toLocaleString('vi-VN')} đ`}
                  </span>
                </div>

                {appliedVoucherCode && (
                  <div className="flex justify-between text-green-700 font-semibold">
                    <span>Mã giảm giá ({appliedVoucherCode})</span>
                    <span>-{discountAmount.toLocaleString('vi-VN')} đ</span>
                  </div>
                )}
              </div>

              {/* Final total */}
              <div className="border-t border-neutral-150 pt-4 flex justify-between items-baseline">
                <span className="text-xs font-bold text-neutral-850 uppercase tracking-wide">Tổng số tiền</span>
                <span className="text-xl font-black text-neutral-900">
                  {finalTotal.toLocaleString('vi-VN')} đ
                </span>
              </div>

              {/* Place Order CTA */}
              <button
                type="button"
                onClick={handlePlaceOrder}
                disabled={loading || items.length === 0}
                className="w-full text-center bg-black text-white text-xs font-black uppercase tracking-wider py-4 rounded hover:bg-neutral-850 transition-colors shadow-sm disabled:opacity-40 select-none"
              >
                {loading ? 'Đang khởi tạo đơn hàng...' : 'Xác nhận đặt hàng'}
              </button>

              <div className="text-[10px] text-neutral-450 leading-relaxed text-center mt-2.5">
                Cam kết hàng chính hãng 100%. Chính sách bảo hành và đổi trả toàn quốc.
              </div>
            </div>

          </div>
        </div>
      </div>

      {/* Voucher Selector Dialog */}
      {showVoucherModal && (
        <VoucherSelector
          orderAmount={cartSubtotal}
          selectedCode={appliedVoucherCode || undefined}
          onSelect={handleApplyVoucher}
          onClose={() => setShowVoucherModal(false)}
        />
      )}
    </div>
  )
}
