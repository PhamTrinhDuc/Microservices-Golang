import { useEffect, useState, useRef } from 'react'
import { useDispatch } from 'react-redux'
import { useAuth } from '../hooks/useAuth'
import { addressAPI } from '../services/addressAPI'
import { orderAPI } from '../services/orderAPI'
import { uploadAPI } from '../services/uploadAPI'
import { authAPI } from '../services/authAPI'
import { locationAPI, type Province, type District, type Ward } from '../services/locationAPI'
import { fetchProfile } from '../store/slices/authSlice'
import type { AppDispatch } from '../store'
import type { Address, OrderResponse } from '../types'

const ProfilePage = () => {
  const dispatch = useDispatch<AppDispatch>()
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState<'profile' | 'addresses' | 'orders'>('profile')

  // Profile Edit State
  const [isEditingProfile, setIsEditingProfile] = useState(false)
  const [editProfileForm, setEditProfileForm] = useState({
    fullName: '',
    phone: '',
    gender: '',
    dob: '',
  })
  const [updatingProfile, setUpdatingProfile] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadingAvatar, setUploadingAvatar] = useState(false)

  // Change Password State
  const [showPasswordForm, setShowPasswordForm] = useState(false)
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [updatingPassword, setUpdatingPassword] = useState(false)

  // Address State
  const [addresses, setAddresses] = useState<Address[]>([])
  const [loadingAddresses, setLoadingAddresses] = useState(false)
  const [showAddressModal, setShowAddressModal] = useState(false)
  const [editingAddress, setEditingAddress] = useState<Address | null>(null)
  const [addressForm, setAddressForm] = useState({
    fullName: '',
    phone: '',
    province: '',
    district: '',
    ward: '',
    detailAddress: '',
    isDefault: false,
  })
  const [submittingAddress, setSubmittingAddress] = useState(false)

  // Location Selector State
  const [provinces, setProvinces] = useState<Province[]>([])
  const [districts, setDistricts] = useState<District[]>([])
  const [wards, setWards] = useState<Ward[]>([])
  const [loadingLocation, setLoadingLocation] = useState(false)
  const [selectedProvinceCode, setSelectedProvinceCode] = useState<number | null>(null)
  const [selectedDistrictCode, setSelectedDistrictCode] = useState<number | null>(null)

  // Orders State
  const [orders, setOrders] = useState<OrderResponse[]>([])
  const [loadingOrders, setLoadingOrders] = useState(false)
  const [ordersPage] = useState(1)
  const [expandedOrderId, setExpandedOrderId] = useState<number | null>(null)
  const [cancellingOrderId, setCancellingOrderId] = useState<number | null>(null)
  const [orderFilter, setOrderFilter] = useState<'all' | 'pending' | 'shipping' | 'completed' | 'cancelled'>('all')

  // Notification Toast State
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null)

  const showToast = (message: string, type: 'success' | 'error') => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 3000)
  }

  // Load addresses
  const loadAddresses = async () => {
    setLoadingAddresses(true)
    try {
      const data = await addressAPI.getAddresses()
      setAddresses(data)
    } catch (err: any) {
      showToast(err?.message || 'Không thể tải danh sách địa chỉ', 'error')
    } finally {
      setLoadingAddresses(false)
    }
  }

  // Load orders
  const loadOrders = async (page = 1) => {
    setLoadingOrders(true)
    try {
      const res = await orderAPI.listMyOrders(page, 20)
      setOrders(res.data || [])
    } catch (err: any) {
      showToast(err?.message || 'Không thể tải lịch sử đơn hàng', 'error')
    } finally {
      setLoadingOrders(false)
    }
  }

  useEffect(() => {
    if (activeTab === 'addresses') {
      void loadAddresses()
    } else if (activeTab === 'orders') {
      void loadOrders()
    }
  }, [activeTab])

  // Initialize Edit Profile Form when user changes or editing is enabled
  useEffect(() => {
    if (user) {
      setEditProfileForm({
        fullName: user.full_name || '',
        phone: user.phone || '',
        gender: user.gender || '',
        dob: user.dob ? new Date(user.dob).toISOString().split('T')[0] : '',
      })
    }
  }, [user, isEditingProfile])

  // Fetch provinces when modal shows up
  useEffect(() => {
    if (showAddressModal) {
      const fetchProvinces = async () => {
        setLoadingLocation(true)
        try {
          const list = await locationAPI.getProvinces()
          setProvinces(list)
        } catch (err: any) {
          showToast('Không thể tải danh mục Tỉnh/Thành phố', 'error')
        } finally {
          setLoadingLocation(false)
        }
      }
      void fetchProvinces()
    }
  }, [showAddressModal])

  // Fetch districts when selectedProvinceCode changes
  useEffect(() => {
    if (selectedProvinceCode) {
      const fetchDistricts = async () => {
        setLoadingLocation(true)
        try {
          const list = await locationAPI.getDistricts(selectedProvinceCode)
          setDistricts(list)
          setWards([])
        } catch (err) {
          showToast('Không thể tải danh mục Quận/Huyện', 'error')
        } finally {
          setLoadingLocation(false)
        }
      }
      void fetchDistricts()
    } else {
      setDistricts([])
      setWards([])
    }
  }, [selectedProvinceCode])

  // Fetch wards when selectedDistrictCode changes
  useEffect(() => {
    if (selectedDistrictCode) {
      const fetchWards = async () => {
        setLoadingLocation(true)
        try {
          const list = await locationAPI.getWards(selectedDistrictCode)
          setWards(list)
        } catch (err) {
          showToast('Không thể tải danh mục Phường/Xã', 'error')
        } finally {
          setLoadingLocation(false)
        }
      }
      void fetchWards()
    } else {
      setWards([])
    }
  }, [selectedDistrictCode])

  // Handle Profile Update
  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editProfileForm.fullName.trim()) {
      showToast('Họ và tên không được để trống', 'error')
      return
    }

    setUpdatingProfile(true)
    try {
      await authAPI.updateProfile({
        full_name: editProfileForm.fullName,
        phone: editProfileForm.phone,
        gender: editProfileForm.gender,
        dob: editProfileForm.dob || undefined,
      })
      showToast('Cập nhật thông tin cá nhân thành công', 'success')
      setIsEditingProfile(false)
      void dispatch(fetchProfile())
    } catch (err: any) {
      showToast(err?.message || 'Không thể cập nhật thông tin', 'error')
    } finally {
      setUpdatingProfile(false)
    }
  }

  // Handle Avatar Change/Upload
  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (file.size > 5 * 1024 * 1024) {
      showToast('Kích thước ảnh đại diện không vượt quá 5MB', 'error')
      return
    }

    setUploadingAvatar(true)
    try {
      const uploadRes = await uploadAPI.uploadImage(file)
      await authAPI.updateProfile({
        full_name: user?.full_name || '',
        avatar: uploadRes.url,
      })
      showToast('Cập nhật ảnh đại diện thành công', 'success')
      void dispatch(fetchProfile())
    } catch (err: any) {
      showToast(err?.message || 'Không thể tải ảnh đại diện lên', 'error')
    } finally {
      setUploadingAvatar(false)
    }
  }

  // Handle Change Password
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!passwordForm.oldPassword || !passwordForm.newPassword || !passwordForm.confirmPassword) {
      showToast('Vui lòng nhập đầy đủ mật khẩu', 'error')
      return
    }

    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      showToast('Mật khẩu mới và xác nhận mật khẩu không khớp', 'error')
      return
    }

    if (passwordForm.newPassword.length < 6) {
      showToast('Mật khẩu mới phải từ 6 ký tự trở lên', 'error')
      return
    }

    setUpdatingPassword(true)
    try {
      await authAPI.updatePassword({
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
      })
      showToast('Đổi mật khẩu thành công', 'success')
      setShowPasswordForm(false)
      setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' })
    } catch (err: any) {
      showToast(err?.message || 'Mật khẩu cũ không chính xác', 'error')
    } finally {
      setUpdatingPassword(false)
    }
  }

  // Handle Open Address Modal (For Create or Update)
  const openAddressModal = async (address: Address | null = null) => {
    setEditingAddress(address)
    setSelectedProvinceCode(null)
    setSelectedDistrictCode(null)
    setDistricts([])
    setWards([])

    if (address) {
      setAddressForm({
        fullName: address.full_name,
        phone: address.phone,
        province: address.province,
        district: address.district,
        ward: address.ward,
        detailAddress: address.detail_address,
        isDefault: address.is_default,
      })

      // Fetch location list first to pre-fill select options
      try {
        const provinceList = await locationAPI.getProvinces()
        setProvinces(provinceList)

        // Find province code
        const pObj = provinceList.find(p => p.name === address.province)
        if (pObj) {
          setSelectedProvinceCode(pObj.code)
          const districtList = await locationAPI.getDistricts(pObj.code)
          setDistricts(districtList)

          // Find district code
          const dObj = districtList.find(d => d.name === address.district)
          if (dObj) {
            setSelectedDistrictCode(dObj.code)
            const wardList = await locationAPI.getWards(dObj.code)
            setWards(wardList)
          }
        }
      } catch (e) {
        console.error('Failed to prefill location selects', e)
      }
    } else {
      setAddressForm({
        fullName: '',
        phone: '',
        province: '',
        district: '',
        ward: '',
        detailAddress: '',
        isDefault: false,
      })
    }
    setShowAddressModal(true)
  }

  // Handle Create or Update Address
  const handleSaveAddress = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!addressForm.fullName || !addressForm.phone || !addressForm.province || !addressForm.district || !addressForm.ward || !addressForm.detailAddress) {
      showToast('Vui lòng nhập đầy đủ thông tin địa chỉ', 'error')
      return
    }

    setSubmittingAddress(true)
    try {
      if (editingAddress) {
        // Edit Mode
        await addressAPI.updateAddress(editingAddress.id, {
          full_name: addressForm.fullName,
          phone: addressForm.phone,
          province: addressForm.province,
          district: addressForm.district,
          ward: addressForm.ward,
          detail_address: addressForm.detailAddress,
          is_default: addressForm.isDefault,
        })
        showToast('Cập nhật địa chỉ thành công', 'success')
      } else {
        // Create Mode
        await addressAPI.createAddress({
          full_name: addressForm.fullName,
          phone: addressForm.phone,
          province: addressForm.province,
          district: addressForm.district,
          ward: addressForm.ward,
          detail_address: addressForm.detailAddress,
          is_default: addressForm.isDefault,
        })
        showToast('Thêm địa chỉ giao hàng thành công', 'success')
      }
      setShowAddressModal(false)
      void loadAddresses()
    } catch (err: any) {
      showToast(err?.message || 'Không thể lưu địa chỉ', 'error')
    } finally {
      setSubmittingAddress(false)
    }
  }

  // Handle Set Default Address
  const handleSetDefaultAddress = async (addressId: number) => {
    try {
      await addressAPI.setDefaultAddress(addressId)
      showToast('Đã đặt địa chỉ mặc định mới', 'success')
      void loadAddresses()
    } catch (err: any) {
      showToast(err?.message || 'Không thể thiết lập địa chỉ mặc định', 'error')
    }
  }

  // Handle Delete Address
  const handleDeleteAddress = async (addressId: number) => {
    if (!confirm('Bạn có chắc chắn muốn xóa địa chỉ này?')) return
    try {
      await addressAPI.deleteAddress(addressId)
      showToast('Xóa địa chỉ thành công', 'success')
      void loadAddresses()
    } catch (err: any) {
      showToast(err?.message || 'Không thể xóa địa chỉ', 'error')
    }
  }

  // Handle Cancel Order
  const handleCancelOrder = async (orderId: number) => {
    if (!confirm('Bạn có chắc chắn muốn hủy đơn hàng này không?')) return
    setCancellingOrderId(orderId)
    try {
      await orderAPI.cancelMyOrder(orderId)
      showToast('Hủy đơn hàng thành công', 'success')
      void loadOrders(ordersPage)
    } catch (err: any) {
      showToast(err?.message || 'Không thể hủy đơn hàng', 'error')
    } finally {
      setCancellingOrderId(null)
    }
  }

  // Format currency
  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(price)
  }

  // Format date
  const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    return d.toLocaleDateString('vi-VN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const getOrderStatusColor = (statusLabel: string) => {
    switch (statusLabel?.toLowerCase()) {
      case 'chờ thanh toán':
      case 'pending':
        return 'bg-amber-50 text-amber-700 border-amber-200'
      case 'chờ xử lý':
        return 'bg-orange-50 text-orange-700 border-orange-200'
      case 'đang giao hàng':
      case 'shipping':
        return 'bg-blue-50 text-blue-700 border-blue-200'
      case 'hoàn thành':
      case 'completed':
      case 'thành công':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200'
      case 'đã hủy':
      case 'cancelled':
        return 'bg-red-50 text-red-700 border-red-200'
      default:
        return 'bg-neutral-50 text-neutral-700 border-neutral-200'
    }
  }

  // Filter orders client-side
  const filteredOrders = orders.filter((orderDetail) => {
    if (orderFilter === 'all') return true
    const label = orderDetail.order_status_label?.toLowerCase()
    if (orderFilter === 'pending') return label === 'chờ xử lý' || label === 'chờ thanh toán' || label === 'pending'
    if (orderFilter === 'shipping') return label === 'đang giao hàng' || label === 'shipping'
    if (orderFilter === 'completed') return label === 'hoàn thành' || label === 'completed' || label === 'thành công'
    if (orderFilter === 'cancelled') return label === 'đã hủy' || label === 'cancelled'
    return true
  })

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 font-sans">
      {/* Toast Alert */}
      {toast && (
        <div className={`fixed bottom-4 right-4 z-50 flex items-center gap-2 rounded-lg px-4 py-3 shadow-lg border text-sm transition-all duration-300 ${
          toast.type === 'success' ? 'bg-emerald-600 border-emerald-500 text-white' : 'bg-red-600 border-red-500 text-white'
        }`}>
          <span>{toast.message}</span>
        </div>
      )}

      {/* Top Banner / Cover */}
      <div className="relative h-40 w-full overflow-hidden rounded-xl bg-gradient-to-r from-neutral-900 via-neutral-800 to-neutral-950 shadow-md">
        <div className="absolute inset-0 opacity-20 bg-[radial-gradient(#fff_1px,transparent_1px)] [background-size:16px_16px]"></div>
      </div>

      {/* Avatar & Simple Details Layout (Completely avoids negative margin username overlapping) */}
      <div className="mb-8 flex flex-col items-center px-6 md:flex-row md:items-end md:gap-6">
        {/* Avatar Container with Pen Hover effect */}
        <div className="relative -mt-14 group cursor-pointer" onClick={() => fileInputRef.current?.click()}>
          <div className="flex h-28 w-28 items-center justify-center rounded-full border-4 border-white bg-neutral-100 text-4xl font-extrabold text-neutral-800 shadow-lg overflow-hidden relative">
            {uploadingAvatar ? (
              <div className="absolute inset-0 bg-black/40 flex items-center justify-center">
                <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
              </div>
            ) : user?.avatar ? (
              <img src={user.avatar} alt={user.full_name} className="h-full w-full object-cover" />
            ) : (
              (user?.full_name || 'U')[0].toUpperCase()
            )}
            
            {/* Hover overlay */}
            <div className="absolute inset-0 bg-black/45 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
              <svg className="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
          </div>
          <input
            type="file"
            ref={fileInputRef}
            onChange={(e) => void handleAvatarChange(e)}
            className="hidden"
            accept="image/*"
          />
        </div>

        {/* User basic info (Safe below the cover, styled cleanly) */}
        <div className="mt-4 flex-1 text-center md:text-left">
          <h1 className="text-2xl font-extrabold text-neutral-900 leading-tight">
            {user?.full_name}
          </h1>
          <p className="text-sm font-semibold text-neutral-500">{user?.email}</p>
          <div className="mt-2.5 flex flex-wrap justify-center gap-2 md:justify-start">
            <span className="rounded-full bg-neutral-150 px-3.5 py-0.5 text-[10px] font-extrabold text-neutral-700 capitalize tracking-wide border border-neutral-200">
              {user?.role === 'admin' ? 'Quản trị viên' : 'Khách hàng'}
            </span>
            {user?.is_verified && (
              <span className="rounded-full bg-emerald-50 border border-emerald-200 px-3.5 py-0.5 text-[10px] font-extrabold text-emerald-700 uppercase tracking-wide">
                Đã xác minh
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-8 lg:grid-cols-4">
        {/* Left Navigation Sidebar */}
        <div className="lg:col-span-1">
          <div className="rounded-xl border border-neutral-200 bg-white p-2.5 shadow-sm">
            <nav className="flex flex-row gap-1 lg:flex-col">
              <button
                onClick={() => setActiveTab('profile')}
                className={`flex flex-1 items-center gap-3 rounded-lg px-4 py-3 text-sm font-bold transition-all ${
                  activeTab === 'profile'
                    ? 'bg-black text-white'
                    : 'text-neutral-600 hover:bg-neutral-100 hover:text-black'
                }`}
              >
                <svg className="h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <span className="hidden sm:inline lg:inline">Thông tin cá nhân</span>
              </button>

              <button
                onClick={() => setActiveTab('addresses')}
                className={`flex flex-1 items-center gap-3 rounded-lg px-4 py-3 text-sm font-bold transition-all ${
                  activeTab === 'addresses'
                    ? 'bg-black text-white'
                    : 'text-neutral-600 hover:bg-neutral-100 hover:text-black'
                }`}
              >
                <svg className="h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span className="hidden sm:inline lg:inline">Sổ địa chỉ</span>
              </button>

              <button
                onClick={() => setActiveTab('orders')}
                className={`flex flex-1 items-center gap-3 rounded-lg px-4 py-3 text-sm font-bold transition-all ${
                  activeTab === 'orders'
                    ? 'bg-black text-white'
                    : 'text-neutral-600 hover:bg-neutral-100 hover:text-black'
                }`}
              >
                <svg className="h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                </svg>
                <span className="hidden sm:inline lg:inline">Đơn hàng của tôi</span>
              </button>
            </nav>
          </div>
        </div>

        {/* Right Content Area */}
        <div className="lg:col-span-3">
          {/* Tab 1: Profile Info & Change Password */}
          {activeTab === 'profile' && (
            <div className="space-y-6">
              {/* Profile Details Block */}
              <div className="rounded-xl border border-neutral-200 bg-white p-6 shadow-sm">
                <div className="mb-6 flex items-center justify-between">
                  <h2 className="text-lg font-extrabold text-neutral-900">Thông tin tài khoản</h2>
                  {!isEditingProfile ? (
                    <button
                      onClick={() => setIsEditingProfile(true)}
                      className="rounded-lg border border-neutral-300 px-4 py-2 text-xs font-bold text-neutral-700 hover:bg-neutral-50 transition-colors"
                    >
                      Chỉnh sửa
                    </button>
                  ) : (
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setIsEditingProfile(false)}
                        className="rounded-lg border border-neutral-200 px-4 py-2 text-xs font-bold text-neutral-500 hover:bg-neutral-50"
                      >
                        Hủy
                      </button>
                      <button
                        onClick={(e) => void handleUpdateProfile(e)}
                        disabled={updatingProfile}
                        className="rounded-lg bg-black px-4 py-2 text-xs font-bold text-white hover:bg-neutral-800 transition-colors"
                      >
                        {updatingProfile ? 'Đang lưu...' : 'Lưu lại'}
                      </button>
                    </div>
                  )}
                </div>

                {!isEditingProfile ? (
                  <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Họ và tên</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800">{user?.full_name || 'Chưa cung cấp'}</p>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Địa chỉ Email</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800">{user?.email}</p>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Số điện thoại</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800">{user?.phone || 'Chưa liên kết'}</p>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Giới tính</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800 capitalize">
                        {user?.gender === 'male' ? 'Nam' : user?.gender === 'female' ? 'Nữ' : user?.gender || 'Chưa thiết lập'}
                      </p>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Ngày sinh</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800">
                        {user?.dob ? new Date(user.dob).toLocaleDateString('vi-VN') : 'Chưa thiết lập'}
                      </p>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400">Ngày gia nhập</label>
                      <p className="mt-1 text-sm font-semibold text-neutral-800">
                        {user?.created_at ? new Date(user.created_at).toLocaleDateString('vi-VN') : 'Mới gia nhập'}
                      </p>
                    </div>
                  </div>
                ) : (
                  <form onSubmit={(e) => void handleUpdateProfile(e)} className="grid grid-cols-1 gap-5 sm:grid-cols-2">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Họ và tên</label>
                      <input
                        type="text"
                        required
                        value={editProfileForm.fullName}
                        onChange={(e) => setEditProfileForm({ ...editProfileForm, fullName: e.target.value })}
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Địa chỉ Email (Không được đổi)</label>
                      <input
                        type="text"
                        disabled
                        value={user?.email}
                        className="w-full rounded-md border border-neutral-200 bg-neutral-100 px-3 py-2 text-xs text-neutral-500 cursor-not-allowed"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Số điện thoại</label>
                      <input
                        type="tel"
                        value={editProfileForm.phone}
                        onChange={(e) => setEditProfileForm({ ...editProfileForm, phone: e.target.value })}
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Giới tính</label>
                      <select
                        value={editProfileForm.gender}
                        onChange={(e) => setEditProfileForm({ ...editProfileForm, gender: e.target.value })}
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black cursor-pointer bg-white"
                      >
                        <option value="">Chọn giới tính</option>
                        <option value="male">Nam</option>
                        <option value="female">Nữ</option>
                        <option value="other">Khác</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Ngày sinh</label>
                      <input
                        type="date"
                        value={editProfileForm.dob}
                        onChange={(e) => setEditProfileForm({ ...editProfileForm, dob: e.target.value })}
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black cursor-pointer"
                      />
                    </div>
                  </form>
                )}
              </div>

              {/* Change Password Block */}
              <div className="rounded-xl border border-neutral-200 bg-white p-6 shadow-sm">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-base font-bold text-neutral-900">Bảo mật tài khoản</h2>
                    <p className="text-xs text-neutral-400 mt-0.5">Thay đổi mật khẩu định kỳ để bảo vệ tài khoản tốt hơn.</p>
                  </div>
                  {!showPasswordForm ? (
                    <button
                      onClick={() => setShowPasswordForm(true)}
                      className="rounded-lg border border-neutral-300 px-4 py-2 text-xs font-bold text-neutral-700 hover:bg-neutral-50 transition-colors"
                    >
                      Đổi mật khẩu
                    </button>
                  ) : (
                    <button
                      onClick={() => {
                        setShowPasswordForm(false)
                        setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' })
                      }}
                      className="text-xs font-bold text-neutral-500 hover:text-black"
                    >
                      Hủy bỏ
                    </button>
                  )}
                </div>

                {showPasswordForm && (
                  <form onSubmit={(e) => void handleChangePassword(e)} className="mt-6 border-t border-neutral-100 pt-5 space-y-4 max-w-md">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Mật khẩu hiện tại</label>
                      <input
                        type="password"
                        required
                        value={passwordForm.oldPassword}
                        onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })}
                        placeholder="••••••••"
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Mật khẩu mới</label>
                      <input
                        type="password"
                        required
                        value={passwordForm.newPassword}
                        onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })}
                        placeholder="Tối thiểu 6 ký tự"
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-neutral-400 mb-1">Xác nhận mật khẩu mới</label>
                      <input
                        type="password"
                        required
                        value={passwordForm.confirmPassword}
                        onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })}
                        placeholder="Nhập lại mật khẩu mới"
                        className="w-full rounded-md border border-neutral-350 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                      />
                    </div>
                    <div className="pt-2">
                      <button
                        type="submit"
                        disabled={updatingPassword}
                        className="rounded-lg bg-black px-5 py-2 text-xs font-bold text-white hover:bg-neutral-800 transition-colors"
                      >
                        {updatingPassword ? 'Đang cập nhật...' : 'Đổi mật khẩu'}
                      </button>
                    </div>
                  </form>
                )}
              </div>
            </div>
          )}

          {/* Tab 2: Addresses (Shopify-style address book with Update support) */}
          {activeTab === 'addresses' && (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-extrabold text-neutral-900">Địa chỉ giao hàng</h2>
                  <p className="text-xs text-neutral-400 mt-0.5">Quản lý và cập nhật các địa chỉ nhận hàng của bạn.</p>
                </div>
                <button
                  onClick={() => openAddressModal(null)}
                  className="flex items-center gap-1.5 rounded-lg bg-black px-4 py-2 text-xs font-bold text-white hover:bg-neutral-800 transition-colors"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M12 4v16m8-8H4" />
                  </svg>
                  Thêm địa chỉ mới
                </button>
              </div>

              {loadingAddresses ? (
                <div className="flex h-32 items-center justify-center rounded-xl border border-neutral-100 bg-white">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-black border-t-transparent"></div>
                </div>
              ) : addresses.length === 0 ? (
                <div className="rounded-xl border border-dashed border-neutral-300 bg-white py-12 text-center">
                  <svg className="mx-auto h-12 w-12 text-neutral-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  </svg>
                  <h3 className="mt-4 text-sm font-bold text-neutral-800">Chưa có địa chỉ nào</h3>
                  <p className="mt-1 text-xs text-neutral-400">Hãy thêm địa chỉ giao hàng để tiện lợi khi thanh toán.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  {addresses.map((address) => (
                    <div
                      key={address.id}
                      className={`relative rounded-xl border p-5 shadow-sm transition-all bg-white flex flex-col justify-between ${
                        address.is_default ? 'border-black ring-1 ring-black' : 'border-neutral-200 hover:border-neutral-300'
                      }`}
                    >
                      <div>
                        <div className="mb-2.5 flex items-center justify-between">
                          <span className="font-extrabold text-sm text-neutral-900">{address.full_name}</span>
                          {address.is_default && (
                            <span className="rounded-md bg-black px-2 py-0.5 text-[9px] font-extrabold text-white uppercase tracking-wider">
                              Mặc định
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-neutral-500 font-semibold mb-2">Số điện thoại: {address.phone}</p>
                        <p className="text-xs leading-relaxed text-neutral-600 font-semibold">
                          {address.detail_address}, {address.ward}, {address.district}, {address.province}
                        </p>
                      </div>

                      <div className="mt-5 flex items-center justify-end gap-3.5 border-t border-neutral-100 pt-3">
                        {!address.is_default && (
                          <button
                            onClick={() => void handleSetDefaultAddress(address.id)}
                            className="text-xs font-bold text-neutral-500 hover:text-black transition-colors mr-auto"
                          >
                            Đặt làm mặc định
                          </button>
                        )}
                        <button
                          onClick={() => openAddressModal(address)}
                          className="text-xs font-bold text-neutral-700 hover:text-black transition-colors"
                        >
                          Chỉnh sửa
                        </button>
                        <button
                          onClick={() => void handleDeleteAddress(address.id)}
                          className="text-xs font-bold text-red-650 hover:text-red-800 transition-colors"
                        >
                          Xóa
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Tab 3: Orders (Shopify-style Order History & Status Filters) */}
          {activeTab === 'orders' && (
            <div className="space-y-6">
              <div>
                <h2 className="text-lg font-extrabold text-neutral-900">Lịch sử đơn hàng</h2>
                <p className="text-xs text-neutral-400 mt-0.5">Theo dõi chi tiết các đơn hàng của bạn qua các trạng thái.</p>
              </div>

              {/* Shopify-style Order Status Tabs */}
              <div className="flex border-b border-neutral-200 overflow-x-auto no-scrollbar scroll-smooth">
                {(['all', 'pending', 'shipping', 'completed', 'cancelled'] as const).map((filter) => (
                  <button
                    key={filter}
                    onClick={() => {
                      setOrderFilter(filter)
                      setExpandedOrderId(null)
                    }}
                    className={`border-b-2 px-4 py-3 text-xs font-bold capitalize whitespace-nowrap transition-all ${
                      orderFilter === filter
                        ? 'border-black text-black'
                        : 'border-transparent text-neutral-400 hover:text-neutral-800'
                    }`}
                  >
                    {filter === 'all'
                      ? 'Tất cả đơn'
                      : filter === 'pending'
                      ? 'Chờ xử lý / Thanh toán'
                      : filter === 'shipping'
                      ? 'Đang giao'
                      : filter === 'completed'
                      ? 'Đã hoàn thành'
                      : 'Đã hủy'}
                  </button>
                ))}
              </div>

              {loadingOrders ? (
                <div className="flex h-32 items-center justify-center rounded-xl border border-neutral-100 bg-white">
                  <div className="h-6 w-6 animate-spin rounded-full border-2 border-black border-t-transparent"></div>
                </div>
              ) : filteredOrders.length === 0 ? (
                <div className="rounded-xl border border-dashed border-neutral-300 bg-white py-12 text-center">
                  <svg className="mx-auto h-12 w-12 text-neutral-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                  </svg>
                  <h3 className="mt-4 text-sm font-bold text-neutral-800">Không tìm thấy đơn hàng</h3>
                  <p className="mt-1 text-xs text-neutral-400">Không có đơn hàng nào khớp với bộ lọc đã chọn.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {filteredOrders.map((orderDetail) => {
                    const order = orderDetail.order
                    const isExpanded = expandedOrderId === order.id
                    const isPending = order.order_status_id === 1 // "Chờ xử lý / Mới"
                    
                    return (
                      <div
                        key={order.id}
                        className="overflow-hidden rounded-xl border border-neutral-200 bg-white shadow-sm transition-all hover:border-neutral-300"
                      >
                        {/* Order Header Card */}
                        <div
                          onClick={() => setExpandedOrderId(isExpanded ? null : order.id)}
                          className="flex cursor-pointer flex-wrap items-center justify-between gap-4 bg-neutral-50 px-5 py-4 transition-colors hover:bg-neutral-100"
                        >
                          <div className="space-y-1">
                            <div className="flex items-center gap-2.5">
                              <span className="font-extrabold text-sm text-neutral-900">
                                Đơn hàng #{order.order_code}
                              </span>
                              <span className={`rounded border px-2 py-0.5 text-[9px] font-extrabold uppercase tracking-wide ${
                                getOrderStatusColor(orderDetail.order_status_label)
                              }`}>
                                {orderDetail.order_status_label}
                              </span>
                            </div>
                            <p className="text-[11px] font-semibold text-neutral-400">
                              Đặt ngày: {formatDate(order.created_at)}
                            </p>
                          </div>
                          
                          <div className="flex items-center gap-4">
                            <div className="text-right">
                              <p className="text-[10px] font-bold uppercase tracking-wider text-neutral-400">Tổng thanh toán</p>
                              <p className="text-sm font-extrabold text-neutral-950">{formatPrice(order.total_amount)}</p>
                            </div>
                            <svg
                              className={`h-5 w-5 text-neutral-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M19 9l-7 7-7-7" />
                            </svg>
                          </div>
                        </div>

                        {/* Order Expanded Details */}
                        {isExpanded && (
                          <div className="border-t border-neutral-200 px-5 py-4 space-y-4">
                            {/* Products List */}
                            <div className="space-y-3">
                              <h4 className="text-xs font-bold uppercase tracking-wider text-neutral-400">Sản phẩm đã chọn</h4>
                              <div className="divide-y divide-neutral-100">
                                {orderDetail.items.map((item) => (
                                  <div key={item.id} className="flex items-center justify-between py-2.5 first:pt-0 last:pb-0">
                                    <div className="space-y-0.5">
                                      <p className="text-xs font-bold text-neutral-900">{item.variant_name}</p>
                                      <p className="text-[10px] font-semibold text-neutral-400">Số lượng: {item.quantity}</p>
                                    </div>
                                    <p className="text-xs font-bold text-neutral-800">{formatPrice(item.unit_price * item.quantity)}</p>
                                  </div>
                                ))}
                              </div>
                            </div>

                            {/* Shipping & Payment Details */}
                            <div className="grid grid-cols-1 gap-6 border-t border-neutral-150 pt-4 sm:grid-cols-2">
                              <div>
                                <h4 className="mb-2 text-xs font-bold uppercase tracking-wider text-neutral-400">Thông tin giao nhận</h4>
                                <p className="text-xs font-extrabold text-neutral-800">{order.receiver_name}</p>
                                <p className="text-[11px] font-semibold text-neutral-500">Số điện thoại: {order.receiver_phone}</p>
                                <p className="text-[11px] leading-relaxed text-neutral-600 font-semibold">{order.receiver_address}</p>
                                {order.shipping_provider && (
                                  <p className="mt-1.5 text-[11px] text-neutral-500 font-extrabold">
                                    Vận chuyển: {order.shipping_provider.toUpperCase()} {order.shipping_code ? `(${order.shipping_code})` : ''}
                                  </p>
                                )}
                              </div>
                              <div className="flex flex-col justify-between">
                                <div>
                                  <h4 className="mb-2 text-xs font-bold uppercase tracking-wider text-neutral-400">Thanh toán</h4>
                                  <p className="text-xs font-bold text-neutral-800 capitalize">
                                    Phương thức: {order.payment_method === 'cod' ? 'Thanh toán COD khi nhận hàng' : order.payment_method === 'payos' ? 'Cổng thanh toán PayOS' : 'Chuyển khoản ngân hàng'}
                                  </p>
                                  <p className="text-[11px] font-semibold text-neutral-500 mt-0.5">
                                    Trạng thái: {orderDetail.payment_status_label}
                                  </p>
                                </div>

                                {/* Order Action Buttons */}
                                {isPending && (
                                  <div className="mt-4 flex justify-end">
                                    <button
                                      disabled={cancellingOrderId === order.id}
                                      onClick={() => void handleCancelOrder(order.id)}
                                      className="rounded-lg border border-red-200 bg-red-50 px-4 py-2 text-xs font-bold text-red-650 hover:bg-red-100 transition-colors disabled:opacity-50"
                                    >
                                      {cancellingOrderId === order.id ? 'Đang hủy...' : 'Hủy đơn hàng'}
                                    </button>
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Address Create or Edit Modal (Shopify-style address modal) */}
      {showAddressModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-lg overflow-hidden rounded-xl bg-white shadow-2xl border border-neutral-100">
            <div className="flex items-center justify-between border-b border-neutral-100 px-6 py-4">
              <h3 className="text-base font-extrabold text-neutral-900">
                {editingAddress ? 'Chỉnh sửa địa chỉ' : 'Thêm địa chỉ giao hàng mới'}
              </h3>
              <button
                onClick={() => setShowAddressModal(false)}
                className="text-neutral-400 hover:text-black transition-colors"
              >
                <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <form onSubmit={(e) => void handleSaveAddress(e)} className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Họ và tên người nhận</label>
                  <input
                    type="text"
                    required
                    value={addressForm.fullName}
                    onChange={(e) => setAddressForm({ ...addressForm, fullName: e.target.value })}
                    placeholder="VD: Nguyễn Văn A"
                    className="w-full rounded-md border border-neutral-300 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                  />
                </div>
                <div>
                  <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Số điện thoại</label>
                  <input
                    type="tel"
                    required
                    value={addressForm.phone}
                    onChange={(e) => setAddressForm({ ...addressForm, phone: e.target.value })}
                    placeholder="VD: 0987654321"
                    className="w-full rounded-md border border-neutral-300 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                  />
                </div>
              </div>

              {/* Shopify-style Dropdown Location Selectors */}
              <div className="grid grid-cols-3 gap-3">
                {/* Provinces Select */}
                <div>
                  <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Tỉnh / Thành phố</label>
                  {loadingLocation && provinces.length === 0 ? (
                    <div className="h-9 rounded-md border border-neutral-200 bg-neutral-50 flex items-center justify-center">
                      <div className="h-3.5 w-3.5 animate-spin rounded-full border border-neutral-400 border-t-transparent"></div>
                    </div>
                  ) : (
                    <select
                      required
                      value={provinces.find(p => p.name === addressForm.province)?.code || ''}
                      onChange={(e) => {
                        const code = Number(e.target.value)
                        const name = provinces.find(p => p.code === code)?.name || ''
                        setSelectedProvinceCode(code)
                        setSelectedDistrictCode(null)
                        setAddressForm({ ...addressForm, province: name, district: '', ward: '' })
                      }}
                      className="w-full rounded-md border border-neutral-300 px-2.5 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black cursor-pointer bg-white"
                    >
                      <option value="">Chọn Tỉnh/TP</option>
                      {provinces.map(p => (
                        <option key={p.code} value={p.code}>{p.name}</option>
                      ))}
                    </select>
                  )}
                </div>

                {/* Districts Select */}
                <div>
                  <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Quận / Huyện</label>
                  <select
                    required
                    disabled={!selectedProvinceCode}
                    value={districts.find(d => d.name === addressForm.district)?.code || ''}
                    onChange={(e) => {
                      const code = Number(e.target.value)
                      const name = districts.find(d => d.code === code)?.name || ''
                      setSelectedDistrictCode(code)
                      setAddressForm({ ...addressForm, district: name, ward: '' })
                    }}
                    className={`w-full rounded-md border border-neutral-300 px-2.5 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black cursor-pointer bg-white ${
                      !selectedProvinceCode ? 'bg-neutral-50 text-neutral-400 cursor-not-allowed border-neutral-200' : ''
                    }`}
                  >
                    <option value="">Chọn Quận/Huyện</option>
                    {districts.map(d => (
                      <option key={d.code} value={d.code}>{d.name}</option>
                    ))}
                  </select>
                </div>

                {/* Wards Select */}
                <div>
                  <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Phường / Xã</label>
                  <select
                    required
                    disabled={!selectedDistrictCode}
                    value={wards.find(w => w.name === addressForm.ward)?.code || ''}
                    onChange={(e) => {
                      const code = Number(e.target.value)
                      const name = wards.find(w => w.code === code)?.name || ''
                      setAddressForm({ ...addressForm, ward: name })
                    }}
                    className={`w-full rounded-md border border-neutral-300 px-2.5 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black cursor-pointer bg-white ${
                      !selectedDistrictCode ? 'bg-neutral-50 text-neutral-400 cursor-not-allowed border-neutral-200' : ''
                    }`}
                  >
                    <option value="">Chọn Phường/Xã</option>
                    {wards.map(w => (
                      <option key={w.code} value={w.code}>{w.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-[10px] font-extrabold uppercase tracking-wider text-neutral-400 mb-1">Địa chỉ chi tiết</label>
                <input
                  type="text"
                  required
                  value={addressForm.detailAddress}
                  onChange={(e) => setAddressForm({ ...addressForm, detailAddress: e.target.value })}
                  placeholder="VD: Số 15, ngõ 100, đường Trần Duy Hưng"
                  className="w-full rounded-md border border-neutral-300 px-3 py-2 text-xs focus:border-black focus:outline-none focus:ring-1 focus:ring-black"
                />
              </div>

              <div className="flex items-center gap-2 pt-2">
                <input
                  type="checkbox"
                  id="defaultAddress"
                  checked={addressForm.isDefault}
                  onChange={(e) => setAddressForm({ ...addressForm, isDefault: e.target.checked })}
                  className="h-4 w-4 rounded border-neutral-355 text-black focus:ring-black cursor-pointer"
                />
                <label htmlFor="defaultAddress" className="text-xs font-bold text-neutral-700 cursor-pointer selection:bg-transparent">
                  Đặt làm địa chỉ mặc định
                </label>
              </div>

              <div className="flex items-center justify-end gap-3 border-t border-neutral-100 pt-4 mt-6">
                <button
                  type="button"
                  onClick={() => setShowAddressModal(false)}
                  className="rounded-lg border border-neutral-200 px-4 py-2 text-xs font-bold text-neutral-600 hover:bg-neutral-50 transition-colors"
                >
                  Hủy
                </button>
                <button
                  type="submit"
                  disabled={submittingAddress}
                  className="rounded-lg bg-black px-5 py-2 text-xs font-bold text-white hover:bg-neutral-800 transition-colors disabled:opacity-50"
                >
                  {submittingAddress ? 'Đang lưu...' : 'Lưu địa chỉ'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default ProfilePage
