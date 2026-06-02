import React, { useEffect, useState } from 'react'
import { inventoryAPI } from '../../services/inventoryAPI'
import type { Store, Supplier, ProductInventory, LowStockAlertResponse, InventoryLogResponse } from '../../types'

interface AdminInventoryTabProps {
  stores: Store[]
  reloadLookups?: () => Promise<void>
}

export default function AdminInventoryTab({ stores, reloadLookups }: AdminInventoryTabProps) {
  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [selectedStoreId, setSelectedStoreId] = useState<number>(1)
  const [storeInventory, setStoreInventory] = useState<ProductInventory[]>([])
  
  // Tab within inventory: 'adjust' | 'import' | 'alerts' | 'logs' | 'suppliers' | 'stores'
  const [invSubTab, setInvSubTab] = useState<'adjust' | 'import' | 'alerts' | 'logs' | 'suppliers' | 'stores'>('adjust')
  const [loading, setLoading] = useState(false)

  // CRUD Stores state
  const [storeForm, setStoreForm] = useState({
    id: 0, // 0 = create, > 0 = update
    name: '',
    hotline: '',
    province: '',
    district: '',
    ward: '',
    road: '',
    email: '',
    lat: '',
    lng: '',
    is_active: true
  })

  // CRUD Suppliers form
  const [supplierForm, setSupplierForm] = useState({
    id: 0,
    name: '',
    address: '',
    phone: '',
    email: '',
    contactName: '',
    contactPhone: '',
    isDeleted: false,
  })
  const [showSupplierModal, setShowSupplierModal] = useState(false)
  const [showStoreModal, setShowStoreModal] = useState(false)

  // Adjust Stock State
  const [adjustQtyInput, setAdjustQtyInput] = useState<{ [variantId: number]: number }>({})
  const [searchTerm, setSearchTerm] = useState('')

  // Import Goods State
  const [importForm, setImportForm] = useState({
    supplierId: '',
    note: '',
    items: [{ variant_id: 0, quantity: 1, price_import: 0 }],
  })

  // Alerts & Logs
  const [lowStockAlerts, setLowStockAlerts] = useState<LowStockAlertResponse[]>([])
  const [inventoryLogs, setInventoryLogs] = useState<InventoryLogResponse[]>([])

  const loadSuppliers = async () => {
    try {
      const data = await inventoryAPI.adminListSuppliers()
      setSuppliers(data)
      if (data.length > 0) {
        setImportForm(p => ({ ...p, supplierId: String(data[0].id) }))
      }
    } catch (err) {
      console.error(err)
    }
  }

  const loadInventory = async (storeId: number) => {
    try {
      setLoading(true)
      const data = await inventoryAPI.adminListStoreInventory(storeId)
      setStoreInventory(data)
      
      const inputs: any = {}
      data.forEach((i) => {
        inputs[i.variant_id] = i.quantity
      })
      setAdjustQtyInput(inputs)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const loadAlerts = async () => {
    try {
      const data = await inventoryAPI.adminGetLowStockAlerts(selectedStoreId)
      setLowStockAlerts(data)
    } catch (err) {
      console.error(err)
    }
  }

  const loadLogs = async () => {
    try {
      const res = await inventoryAPI.adminGetInventoryLogs({ store_id: selectedStoreId, page: 1, limit: 20 })
      setInventoryLogs(res.logs || [])
    } catch (err) {
      console.error(err)
    }
  }

  useEffect(() => {
    void loadSuppliers()
  }, [])

  useEffect(() => {
    if (selectedStoreId) {
      if (invSubTab === 'adjust') void loadInventory(selectedStoreId)
      if (invSubTab === 'alerts') void loadAlerts()
      if (invSubTab === 'logs') void loadLogs()
    }
  }, [selectedStoreId, invSubTab])

  // Adjust stock
  const handleSaveAdjust = async (variantId: number) => {
    const qty = adjustQtyInput[variantId]
    if (qty === undefined || qty < 0) return
    try {
      setLoading(true)
      await inventoryAPI.adminAdjustInventory(selectedStoreId, {
        adjustments: [{ variant_id: variantId, new_quantity: qty }],
      })
      alert('Điều chỉnh số lượng tồn kho thành công!')
      void loadInventory(selectedStoreId)
    } catch (err: any) {
      alert(err.message || 'Không thể điều chỉnh tồn kho')
    } finally {
      setLoading(false)
    }
  }

  // Create / Update Supplier
  const handleSupplierSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!supplierForm.name) return
    try {
      setLoading(true)
      const payload = {
        name: supplierForm.name,
        address: supplierForm.address ? supplierForm.address : null,
        phone: supplierForm.phone ? supplierForm.phone : null,
        email: supplierForm.email ? supplierForm.email : null,
        contact_name: supplierForm.contactName ? supplierForm.contactName : null,
        contact_phone: supplierForm.contactPhone ? supplierForm.contactPhone : null,
        is_deleted: supplierForm.isDeleted,
      }

      if (supplierForm.id > 0) {
        await inventoryAPI.adminUpdateSupplier(supplierForm.id, payload)
        alert('Cập nhật nhà cung cấp thành công!')
      } else {
        await inventoryAPI.adminCreateSupplier(payload)
        alert('Thêm nhà cung cấp thành công!')
      }
      setSupplierForm({
        id: 0,
        name: '',
        address: '',
        phone: '',
        email: '',
        contactName: '',
        contactPhone: '',
        isDeleted: false,
      })
      setShowSupplierModal(false)
      void loadSuppliers()
    } catch (err: any) {
      alert(err.message || 'Không thể lưu nhà cung cấp')
    } finally {
      setLoading(false)
    }
  }

  const handleEditSupplier = (sup: Supplier) => {
    setSupplierForm({
      id: sup.id,
      name: sup.name,
      address: sup.address || '',
      phone: sup.phone || '',
      email: sup.email || '',
      contactName: sup.contact_name || '',
      contactPhone: sup.contact_phone || '',
      isDeleted: sup.is_deleted,
    })
    setShowSupplierModal(true)
  }

  const handleDeleteSupplier = async (id: number) => {
    if (!window.confirm('Bạn có chắc chắn muốn xóa nhà cung cấp này?')) return
    try {
      setLoading(true)
      await inventoryAPI.adminDeleteSupplier(id)
      alert('Xóa nhà cung cấp thành công!')
      void loadSuppliers()
    } catch (err: any) {
      alert(err.message || 'Không thể xóa nhà cung cấp')
    } finally {
      setLoading(false)
    }
  }

  // Create Import Goods
  const handleAddImportItem = () => {
    setImportForm(p => ({
      ...p,
      items: [...p.items, { variant_id: 0, quantity: 1, price_import: 0 }],
    }))
  }

  const handleRemoveImportItem = (index: number) => {
    setImportForm(p => ({
      ...p,
      items: p.items.filter((_, i) => i !== index),
    }))
  }

  const handleImportSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!importForm.supplierId) {
      alert('Vui lòng chọn nhà cung cấp')
      return
    }

    // Validate items
    const invalidItem = importForm.items.find(i => i.variant_id <= 0 || i.quantity <= 0 || i.price_import <= 0)
    if (invalidItem) {
      alert('Vui lòng điền đúng thông tin Variant ID, Số lượng và giá nhập lớn hơn 0')
      return
    }

    try {
      setLoading(true)
      await inventoryAPI.adminImportGoods({
        supplier_id: Number(importForm.supplierId),
        store_id: selectedStoreId,
        note: importForm.note ? importForm.note : null,
        items: importForm.items,
      })
      alert('Lập hóa đơn nhập hàng thành công!')
      setImportForm({
        supplierId: suppliers.length > 0 ? String(suppliers[0].id) : '',
        note: '',
        items: [{ variant_id: 0, quantity: 1, price_import: 0 }],
      })
      setInvSubTab('adjust')
      void loadInventory(selectedStoreId)
    } catch (err: any) {
      alert(err.message || 'Lỗi khi nhập kho')
    } finally {
      setLoading(false)
    }
  }

  // Create / Update Store
  const handleStoreSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!storeForm.name || !storeForm.province || !storeForm.district || !storeForm.ward) {
      alert('Vui lòng nhập đầy đủ thông tin bắt buộc (Tên, Tỉnh/Thành, Quận/Huyện, Phường/Xã)')
      return
    }

    const payload = {
      name: storeForm.name,
      hotline: storeForm.hotline ? storeForm.hotline : null,
      province: storeForm.province,
      district: storeForm.district,
      ward: storeForm.ward,
      road: storeForm.road ? storeForm.road : null,
      email: storeForm.email ? storeForm.email : null,
      lat: storeForm.lat ? Number(storeForm.lat) : null,
      lng: storeForm.lng ? Number(storeForm.lng) : null,
      is_active: storeForm.is_active
    }

    try {
      setLoading(true)
      if (storeForm.id > 0) {
        await inventoryAPI.adminUpdateStore(storeForm.id, payload)
        alert('Cập nhật cửa hàng thành công!')
      } else {
        await inventoryAPI.adminCreateStore(payload)
        alert('Thêm cửa hàng thành công!')
      }
      setStoreForm({
        id: 0,
        name: '',
        hotline: '',
        province: '',
        district: '',
        ward: '',
        road: '',
        email: '',
        lat: '',
        lng: '',
        is_active: true
      })
      setShowStoreModal(false)
      if (reloadLookups) {
        await reloadLookups()
      }
    } catch (err: any) {
      alert(err.message || 'Không thể lưu thông tin cửa hàng')
    } finally {
      setLoading(false)
    }
  }

  const handleEditStore = (store: Store) => {
    setStoreForm({
      id: store.id,
      name: store.name,
      hotline: store.hotline || '',
      province: store.province,
      district: store.district,
      ward: store.ward,
      road: store.road || '',
      email: store.email || '',
      lat: store.lat ? String(store.lat) : '',
      lng: store.lng ? String(store.lng) : '',
      is_active: store.is_active
    })
    setShowStoreModal(true)
  }

  const handleDeactivateStore = async (id: number) => {
    if (!window.confirm('Bạn có chắc chắn muốn ngưng hoạt động cửa hàng này?')) return
    try {
      setLoading(true)
      await inventoryAPI.adminDeactivateStore(id)
      alert('Đã tạm ngưng hoạt động cửa hàng!')
      if (reloadLookups) {
        await reloadLookups()
      }
    } catch (err: any) {
      alert(err.message || 'Không thể ngưng hoạt động cửa hàng')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Kho & Tồn kho</h1>
          <p className="text-xs text-neutral-555 mt-1">Điều chỉnh lượng hàng, quản lý nhà cung cấp và nhập hàng</p>
        </div>

        {/* Global Store Selector */}
        <select
          className="border border-neutral-350 rounded px-3 py-1.5 text-xs bg-white font-bold"
          value={selectedStoreId}
          onChange={(e) => setSelectedStoreId(Number(e.target.value))}
        >
          {stores.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} ({s.province})
            </option>
          ))}
        </select>
      </div>

      {/* Sub tabs list */}
      <div className="flex border-b border-neutral-200 text-xs font-semibold gap-1">
        {[
          { id: 'adjust', label: 'Tồn kho' },
          { id: 'import', label: 'Nhập hàng' },
          { id: 'suppliers', label: 'Nhà cung cấp' },
          { id: 'stores', label: 'Cửa hàng' },
          { id: 'alerts', label: 'Cảnh báo hết hàng' },
          { id: 'logs', label: 'Lịch sử giao dịch' },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setInvSubTab(tab.id as any)}
            className={`px-4 py-2 border-b-2 transition-all ${
              invSubTab === tab.id
                ? 'border-black text-black font-black'
                : 'border-transparent text-neutral-555 hover:text-black'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* TAB SUBVIEW CONTENT */}
      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
        </div>
      ) : (
        <div className="space-y-4">
          
          {/* 1. ADJUST INVENTORY SUBVIEW */}
          {invSubTab === 'adjust' && (() => {
            const filteredInventory = storeInventory.filter((item) => {
              const term = searchTerm.toLowerCase().trim()
              if (!term) return true
              return (
                (item.variant_name && item.variant_name.toLowerCase().includes(term)) ||
                (item.product_name && item.product_name.toLowerCase().includes(term)) ||
                (item.sku && item.sku.toLowerCase().includes(term))
              )
            })

            return (
              <div className="space-y-4">
                <div className="max-w-md">
                  <input
                    type="text"
                    placeholder="Tìm theo tên sản phẩm, biến thể hoặc SKU..."
                    className="w-full border border-neutral-350 rounded px-3 py-1.5 text-xs bg-white focus:outline-none"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </div>

                <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
                  <table className="w-full text-left text-xs border-collapse">
                    <thead>
                      <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                        <th className="p-4">Sản phẩm / Biến thể</th>
                        <th className="p-4">SKU</th>
                        <th className="p-4 text-center">Đang giữ (Reserved)</th>
                        <th className="p-4 text-center">Tồn khả dụng</th>
                        <th className="p-4 text-center">Tổng tồn kho vật lý</th>
                        <th className="p-4 text-center">Hành động</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-150">
                      {filteredInventory.length === 0 ? (
                        <tr>
                          <td colSpan={6} className="p-8 text-center text-neutral-400">
                            {storeInventory.length === 0 
                              ? "Chưa có hàng trong kho của cửa hàng này. Hãy lập phiếu nhập hàng!"
                              : "Không tìm thấy sản phẩm nào khớp với từ khoá."}
                          </td>
                        </tr>
                      ) : (
                        filteredInventory.map((item) => {
                          const available = Math.max(0, item.quantity - item.reserved)
                          return (
                            <tr key={item.variant_id} className="hover:bg-neutral-50 transition-colors">
                              <td className="p-4">
                                <p className="font-bold text-neutral-900">{item.variant_name || `Biến thể #${item.variant_id}`}</p>
                                {item.product_name && (
                                  <p className="text-[10px] text-neutral-450 mt-0.5">Sản phẩm: {item.product_name}</p>
                                )}
                              </td>
                              <td className="p-4 font-mono font-bold text-neutral-600 text-xs">
                                {item.sku || 'N/A'}
                              </td>
                              <td className="p-4 text-center text-neutral-555 font-semibold">
                                {item.reserved}
                              </td>
                              <td className="p-4 text-center">
                                <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase ${available > 0 ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                                  {available} khả dụng
                                </span>
                              </td>
                              <td className="p-4 text-center">
                                <input
                                  type="number"
                                  min={0}
                                  className="border border-neutral-350 rounded px-2.5 py-1 w-20 text-center font-bold text-xs"
                                  value={adjustQtyInput[item.variant_id] ?? 0}
                                  onChange={(e) =>
                                    setAdjustQtyInput((p) => ({ ...p, [item.variant_id]: Math.max(0, Number(e.target.value)) }))
                                  }
                                />
                              </td>
                              <td className="p-4 text-center">
                                <button
                                  type="button"
                                  onClick={() => handleSaveAdjust(item.variant_id)}
                                  className="bg-black hover:bg-neutral-800 text-white text-[9px] uppercase font-black tracking-wider px-3.5 py-1.5 rounded transition-colors"
                                >
                                  Cập nhật
                                </button>
                              </td>
                            </tr>
                          )
                        })
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )
          })()}

          {/* 2. SUPPLIERS SUBVIEW */}
          {invSubTab === 'suppliers' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center bg-white p-4 border border-neutral-200 rounded-lg shadow-sm">
                <div>
                  <h3 className="text-xs font-black uppercase tracking-wide text-neutral-850">
                    Danh Sách Nhà Cung Cấp
                  </h3>
                  <p className="text-[10px] text-neutral-400 mt-0.5">Quản lý đối tác và theo dõi thống kê giá trị nhập hàng</p>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    setSupplierForm({
                      id: 0,
                      name: '',
                      address: '',
                      phone: '',
                      email: '',
                      contactName: '',
                      contactPhone: '',
                      isDeleted: false,
                    })
                    setShowSupplierModal(true)
                  }}
                  className="bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider px-4 py-2 rounded transition-colors"
                >
                  + Thêm Nhà Cung Cấp
                </button>
              </div>

              {/* Table */}
              <div className="bg-white border border-neutral-200 rounded-lg overflow-x-auto shadow-sm">
                <table className="w-full text-left text-xs border-collapse min-w-[700px]">
                  <thead>
                    <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                      <th className="p-4">Nhà cung cấp</th>
                      <th className="p-4">Thông tin liên hệ</th>
                      <th className="p-4">Người đại diện</th>
                      <th className="p-4">Thống kê nhập hàng</th>
                      <th className="p-4 text-center">Trạng thái</th>
                      <th className="p-4 text-center">Hành động</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-neutral-150">
                    {suppliers.map((s) => (
                      <tr key={s.id} className={`hover:bg-neutral-50 transition-colors ${s.is_deleted ? 'bg-red-50/30' : ''}`}>
                        <td className="p-4">
                          <p className="font-bold text-neutral-855">{s.name}</p>
                          <p className="text-[9px] text-neutral-450 font-mono mt-0.5">ID: {s.id}</p>
                          {s.address && <p className="text-[10px] text-neutral-500 mt-1">Đ/c: {s.address}</p>}
                        </td>
                        <td className="p-4 space-y-0.5">
                          <p className="font-mono text-neutral-600">{s.phone || 'N/A'}</p>
                          {s.email && <p className="text-neutral-500 font-mono">{s.email}</p>}
                        </td>
                        <td className="p-4 space-y-0.5">
                          <p className="font-semibold text-neutral-700">{s.contact_name || 'N/A'}</p>
                          {s.contact_phone && <p className="font-mono text-neutral-500">{s.contact_phone}</p>}
                        </td>
                        <td className="p-4 space-y-1">
                          <div className="text-[10px] text-neutral-600 flex justify-between gap-4">
                            <span>Đơn nhập:</span>
                            <span className="font-bold text-neutral-800">{s.total_imports || 0}</span>
                          </div>
                          <div className="text-[10px] text-neutral-600 flex justify-between gap-4">
                            <span>Tổng giá trị:</span>
                            <span className="font-bold text-neutral-800 font-mono">{(s.total_import_value || 0).toLocaleString('vi-VN')} đ</span>
                          </div>
                          {s.last_imported_at && (
                            <div className="text-[9px] text-neutral-450">
                              Lần cuối: {new Date(s.last_imported_at).toLocaleDateString('vi-VN')}
                            </div>
                          )}
                        </td>
                        <td className="p-4 text-center">
                          <span className={`px-2 py-0.5 rounded-[4px] text-[9px] font-black uppercase tracking-wider ${
                            s.is_deleted 
                              ? 'bg-red-100 text-red-700' 
                              : 'bg-green-100 text-green-700'
                          }`}>
                            {s.is_deleted ? 'Ngừng hoạt động' : 'Hoạt động'}
                          </span>
                        </td>
                        <td className="p-4 text-center">
                          <div className="flex gap-2.5 justify-center">
                            <button
                              type="button"
                              onClick={() => handleEditSupplier(s)}
                              className="text-neutral-800 hover:text-black font-bold uppercase text-[9px] tracking-wider"
                            >
                              Sửa
                            </button>
                            {!s.is_deleted && (
                              <button
                                type="button"
                                onClick={() => handleDeleteSupplier(s.id)}
                                className="text-red-500 hover:text-red-700 font-bold uppercase text-[9px] tracking-wider"
                              >
                                Xóa
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Form Modal */}
              {showSupplierModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
                  <div className="bg-white border border-neutral-200 rounded-lg shadow-2xl w-full max-w-md max-h-[90vh] overflow-y-auto p-6 space-y-5 flex flex-col">
                    <div className="flex justify-between items-center border-b border-neutral-100 pb-3">
                      <h3 className="text-xs font-black uppercase tracking-wider text-neutral-900">
                        {supplierForm.id > 0 ? 'Cập Nhật Nhà Cung Cấp' : 'Thêm Nhà Cung Cấp'}
                      </h3>
                      <button
                        onClick={() => setShowSupplierModal(false)}
                        className="text-neutral-400 hover:text-black"
                      >
                        ✕
                      </button>
                    </div>

                    <form onSubmit={handleSupplierSubmit} className="space-y-4 text-xs flex-1">
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tên nhà cung cấp *</label>
                        <input
                          type="text"
                          placeholder="Tên nhà cc..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={supplierForm.name}
                          onChange={(e) => setSupplierForm(p => ({ ...p, name: e.target.value }))}
                          required
                        />
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Số điện thoại</label>
                        <input
                          type="tel"
                          placeholder="Hotline..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={supplierForm.phone}
                          onChange={(e) => setSupplierForm(p => ({ ...p, phone: e.target.value }))}
                        />
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Email</label>
                        <input
                          type="email"
                          placeholder="Email..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={supplierForm.email}
                          onChange={(e) => setSupplierForm(p => ({ ...p, email: e.target.value }))}
                        />
                      </div>

                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Người đại diện</label>
                          <input
                            type="text"
                            placeholder="Tên..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                            value={supplierForm.contactName}
                            onChange={(e) => setSupplierForm(p => ({ ...p, contactName: e.target.value }))}
                          />
                        </div>

                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">SĐT người đại diện</label>
                          <input
                            type="tel"
                            placeholder="SĐT..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                            value={supplierForm.contactPhone}
                            onChange={(e) => setSupplierForm(p => ({ ...p, contactPhone: e.target.value }))}
                          />
                        </div>
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Địa chỉ văn phòng</label>
                        <input
                          type="text"
                          placeholder="Địa chỉ..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={supplierForm.address}
                          onChange={(e) => setSupplierForm(p => ({ ...p, address: e.target.value }))}
                        />
                      </div>

                      <div className="flex items-center gap-2 pt-1">
                        <input
                          type="checkbox"
                          id="isActiveCheckbox"
                          checked={!supplierForm.isDeleted}
                          onChange={(e) => setSupplierForm(p => ({ ...p, isDeleted: !e.target.checked }))}
                        />
                        <label htmlFor="isActiveCheckbox" className="text-[10px] font-bold text-neutral-600 uppercase cursor-pointer select-none">
                          Đang hoạt động (Active)
                        </label>
                      </div>

                      <div className="flex gap-2 justify-end pt-3">
                        <button
                          type="button"
                          onClick={() => setShowSupplierModal(false)}
                          className="text-[10px] border border-neutral-350 px-5 py-2.5 font-bold uppercase rounded text-neutral-600 hover:text-black"
                        >
                          Hủy
                        </button>
                        <button
                          type="submit"
                          disabled={loading}
                          className="text-[10px] bg-black text-white px-6 py-2.5 font-black uppercase rounded hover:bg-neutral-800"
                        >
                          Lưu
                        </button>
                      </div>
                    </form>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 3. IMPORT GOODS SUBVIEW */}
          {invSubTab === 'import' && (
            <form onSubmit={handleImportSubmit} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs">
              <h3 className="text-xs font-black uppercase tracking-wide">Phiếu Nhập Hàng Kho</h3>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Chọn nhà cung cấp *</label>
                  <select
                    className="w-full border border-neutral-300 rounded px-2.5 py-2"
                    value={importForm.supplierId}
                    onChange={(e) => setImportForm(p => ({ ...p, supplierId: e.target.value }))}
                    required
                  >
                    <option value="">Chọn nhà cung cấp</option>
                    {suppliers.map(s => (
                      <option key={s.id} value={s.id}>
                        {s.name}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Ghi chú phiếu nhập</label>
                  <input
                    type="text"
                    placeholder="Ghi chú đính kèm..."
                    className="w-full border border-neutral-300 rounded px-3 py-2"
                    value={importForm.note}
                    onChange={(e) => setImportForm(p => ({ ...p, note: e.target.value }))}
                  />
                </div>
              </div>

              {/* Items row selectors */}
              <div className="space-y-2.5 border-t border-neutral-100 pt-4">
                <div className="flex justify-between items-center">
                  <p className="text-[10px] uppercase font-bold text-neutral-700 tracking-wider">Danh sách sản phẩm nhập</p>
                  <button
                    type="button"
                    onClick={handleAddImportItem}
                    className="text-[10px] font-black text-neutral-700 hover:text-black uppercase"
                  >
                    + Thêm dòng nhập
                  </button>
                </div>

                {importForm.items.map((item, index) => (
                  <div key={index} className="grid grid-cols-1 md:grid-cols-4 gap-3 items-end border border-neutral-100 p-3 rounded-lg bg-neutral-50 relative">
                    <div>
                      <label className="block text-[9px] uppercase font-bold text-neutral-400 mb-1">Variant ID *</label>
                      <input
                        type="number"
                        placeholder="ID"
                        className="w-full border border-neutral-300 rounded px-2.5 py-1.5 bg-white font-bold"
                        value={item.variant_id || ''}
                        onChange={(e) => {
                          const updated = [...importForm.items]
                          updated[index].variant_id = Number(e.target.value)
                          setImportForm(p => ({ ...p, items: updated }))
                        }}
                        required
                      />
                    </div>

                    <div>
                      <label className="block text-[9px] uppercase font-bold text-neutral-400 mb-1">Số lượng nhập *</label>
                      <input
                        type="number"
                        placeholder="Số lượng"
                        className="w-full border border-neutral-300 rounded px-2.5 py-1.5 bg-white"
                        value={item.quantity || ''}
                        onChange={(e) => {
                          const updated = [...importForm.items]
                          updated[index].quantity = Number(e.target.value)
                          setImportForm(p => ({ ...p, items: updated }))
                        }}
                        required
                      />
                    </div>

                    <div>
                      <label className="block text-[9px] uppercase font-bold text-neutral-400 mb-1">Giá nhập (đ/sp) *</label>
                      <input
                        type="number"
                        placeholder="Giá nhập"
                        className="w-full border border-neutral-300 rounded px-2.5 py-1.5 bg-white font-mono"
                        value={item.price_import || ''}
                        onChange={(e) => {
                          const updated = [...importForm.items]
                          updated[index].price_import = Number(e.target.value)
                          setImportForm(p => ({ ...p, items: updated }))
                        }}
                        required
                      />
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="button"
                        onClick={() => handleRemoveImportItem(index)}
                        disabled={importForm.items.length <= 1}
                        className="text-red-500 hover:text-red-700 disabled:opacity-30 p-2 text-xs"
                      >
                        ✕ Gỡ bỏ
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              <div className="flex justify-end gap-2 border-t border-neutral-100 pt-4">
                <button
                  type="submit"
                  disabled={loading}
                  className="bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider px-6 py-2.5 rounded transition-colors"
                >
                  Xác nhận Nhập hàng
                </button>
              </div>
            </form>
          )}

          {/* 4. LOW STOCK ALERTS SUBVIEW */}
          {invSubTab === 'alerts' && (
            <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                    <th className="p-4">Sản phẩm / Biến thể</th>
                    <th className="p-4">SKU</th>
                    <th className="p-4 text-center">Tồn hiện tại</th>
                    <th className="p-4 text-center">Ngưỡng cảnh báo</th>
                    <th className="p-4">Tình trạng</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-150">
                  {lowStockAlerts.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="p-8 text-center text-neutral-400">
                        Không có cảnh báo tồn kho thấp!
                      </td>
                    </tr>
                  ) : (
                    lowStockAlerts.map((a) => (
                      <tr key={a.variant_id} className="hover:bg-neutral-50">
                        <td className="p-4">
                          <p className="font-bold text-neutral-855">{a.variant_name}</p>
                          <p className="text-[10px] text-neutral-400 mt-0.5">Sản phẩm: {a.product_name}</p>
                        </td>
                        <td className="p-4 font-mono text-neutral-600">{a.sku}</td>
                        <td className="p-4 text-center font-bold text-red-500">{a.quantity}</td>
                        <td className="p-4 text-center text-neutral-555">{a.low_stock_threshold}</td>
                        <td className="p-4">
                          <span className="bg-red-50 border border-red-200 text-red-700 px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-wider">Tồn kho cực thấp</span>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}

          {/* 5. INVENTORY LOGS SUBVIEW */}
          {invSubTab === 'logs' && (
            <div className="bg-white border border-neutral-200 rounded-lg overflow-hidden shadow-sm">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                    <th className="p-4">Thời gian</th>
                    <th className="p-4">Sản phẩm</th>
                    <th className="p-4 text-right">Lượng thay đổi</th>
                    <th className="p-4 text-right">Tồn sau thay đổi</th>
                    <th className="p-4">Lý do</th>
                    <th className="p-4">Người thực hiện</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-150">
                  {inventoryLogs.map((l) => (
                    <tr key={l.id} className="hover:bg-neutral-50">
                      <td className="p-4 text-neutral-400 font-mono">
                        {new Date(l.created_at).toLocaleString('vi-VN')}
                      </td>
                      <td className="p-4 font-bold text-neutral-800">
                        {l.variant_name}
                        <p className="text-[10px] text-neutral-450 font-mono mt-0.5">SKU: {l.sku}</p>
                      </td>
                      <td className={`p-4 text-right font-mono font-bold ${l.change_qty > 0 ? 'text-green-600' : 'text-red-500'}`}>
                        {l.change_qty > 0 ? `+${l.change_qty}` : l.change_qty}
                      </td>
                      <td className="p-4 text-right font-mono font-medium text-neutral-700">{l.qty_after}</td>
                      <td className="p-4 text-neutral-600">{l.reason}</td>
                      <td className="p-4 text-neutral-550 font-medium">{l.creator_name || 'Hệ thống'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* 6. STORES SUBVIEW */}
          {invSubTab === 'stores' && (
            <div className="space-y-4">
              <div className="flex justify-between items-center bg-white p-4 border border-neutral-200 rounded-lg shadow-sm">
                <div>
                  <h3 className="text-xs font-black uppercase tracking-wide text-neutral-850">
                    Danh Sách Cửa Hàng
                  </h3>
                  <p className="text-[10px] text-neutral-400 mt-0.5">Quản lý chi nhánh cửa hàng và thông tin liên hệ, tọa độ</p>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    setStoreForm({
                      id: 0,
                      name: '',
                      hotline: '',
                      province: '',
                      district: '',
                      ward: '',
                      road: '',
                      email: '',
                      lat: '',
                      lng: '',
                      is_active: true
                    })
                    setShowStoreModal(true)
                  }}
                  className="bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase tracking-wider px-4 py-2 rounded transition-colors"
                >
                  + Thêm Cửa Hàng Mới
                </button>
              </div>

              {/* Table */}
              <div className="bg-white border border-neutral-200 rounded-lg overflow-x-auto shadow-sm">
                <table className="w-full text-left text-xs border-collapse min-w-[700px]">
                  <thead>
                    <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                      <th className="p-4">Cửa hàng</th>
                      <th className="p-4">Liên hệ</th>
                      <th className="p-4">Địa chỉ</th>
                      <th className="p-4 text-center">Trạng thái</th>
                      <th className="p-4 text-center">Hành động</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-neutral-150">
                    {stores.map((s) => (
                      <tr key={s.id} className="hover:bg-neutral-50 transition-colors">
                        <td className="p-4 font-bold text-neutral-855">
                          {s.name}
                          <p className="text-[9px] text-neutral-450 font-mono mt-0.5">ID: {s.id}</p>
                        </td>
                        <td className="p-4">
                          {s.hotline && <p className="font-mono text-neutral-600">SĐT: {s.hotline}</p>}
                          {s.email && <p className="text-neutral-500 font-mono">{s.email}</p>}
                          {!s.hotline && !s.email && <p className="text-neutral-400">Chưa cập nhật</p>}
                        </td>
                        <td className="p-4 text-neutral-600">
                          <p className="font-medium">{[s.road, s.ward, s.district, s.province].filter(Boolean).join(', ')}</p>
                          {(s.lat || s.lng) && (
                            <p className="text-[9px] text-neutral-400 font-mono mt-0.5">
                              Tọa độ: {s.lat || '0'}, {s.lng || '0'}
                            </p>
                          )}
                        </td>
                        <td className="p-4 text-center">
                          <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase ${s.is_active ? 'bg-green-50 border border-green-200 text-green-700' : 'bg-red-50 border border-red-200 text-red-700'}`}>
                            {s.is_active ? 'Hoạt động' : 'Tạm ngưng'}
                          </span>
                        </td>
                        <td className="p-4 text-center">
                          <div className="flex gap-2 justify-center">
                            <button
                              type="button"
                              onClick={() => handleEditStore(s)}
                              className="text-neutral-800 hover:text-black font-bold uppercase text-[9px] tracking-wider"
                            >
                              Sửa
                            </button>
                            {s.is_active && (
                              <button
                                type="button"
                                onClick={() => handleDeactivateStore(s.id)}
                                className="text-red-500 hover:text-red-700 font-bold uppercase text-[9px] tracking-wider"
                              >
                                Vô hiệu
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* Form Modal */}
              {showStoreModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
                  <div className="bg-white border border-neutral-200 rounded-lg shadow-2xl w-full max-w-md max-h-[90vh] overflow-y-auto p-6 space-y-5 flex flex-col">
                    <div className="flex justify-between items-center border-b border-neutral-100 pb-3">
                      <h3 className="text-xs font-black uppercase tracking-wider text-neutral-900">
                        {storeForm.id > 0 ? 'Cập Nhật Cửa Hàng' : 'Thêm Cửa Hàng Mới'}
                      </h3>
                      <button
                        onClick={() => setShowStoreModal(false)}
                        className="text-neutral-400 hover:text-black"
                      >
                        ✕
                      </button>
                    </div>

                    <form onSubmit={handleStoreSubmit} className="space-y-4 text-xs flex-1">
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tên cửa hàng *</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: Jiyuu Flagship Store..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={storeForm.name}
                          onChange={(e) => setStoreForm(p => ({ ...p, name: e.target.value }))}
                          required
                        />
                      </div>

                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Hotline</label>
                          <input
                            type="text"
                            placeholder="Số điện thoại..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                            value={storeForm.hotline}
                            onChange={(e) => setStoreForm(p => ({ ...p, hotline: e.target.value }))}
                          />
                        </div>
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Email</label>
                          <input
                            type="email"
                            placeholder="Email cửa hàng..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                            value={storeForm.email}
                            onChange={(e) => setStoreForm(p => ({ ...p, email: e.target.value }))}
                          />
                        </div>
                      </div>

                      <div className="grid grid-cols-3 gap-2">
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tỉnh/Thành *</label>
                          <input
                            type="text"
                            placeholder="Tỉnh..."
                            className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white"
                            value={storeForm.province}
                            onChange={(e) => setStoreForm(p => ({ ...p, province: e.target.value }))}
                            required
                          />
                        </div>
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Quận/Huyện *</label>
                          <input
                            type="text"
                            placeholder="Quận..."
                            className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white"
                            value={storeForm.district}
                            onChange={(e) => setStoreForm(p => ({ ...p, district: e.target.value }))}
                            required
                          />
                        </div>
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Phường/Xã *</label>
                          <input
                            type="text"
                            placeholder="Phường..."
                            className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white"
                            value={storeForm.ward}
                            onChange={(e) => setStoreForm(p => ({ ...p, ward: e.target.value }))}
                            required
                          />
                        </div>
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Số nhà / Đường phố</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: 123 Nguyễn Trãi..."
                          className="w-full border border-neutral-300 rounded px-3 py-2 bg-white"
                          value={storeForm.road}
                          onChange={(e) => setStoreForm(p => ({ ...p, road: e.target.value }))}
                        />
                      </div>

                      <div className="grid grid-cols-2 gap-3">
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Vĩ độ (Latitude)</label>
                          <input
                            type="number"
                            step="any"
                            placeholder="10.7626..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white font-mono"
                            value={storeForm.lat}
                            onChange={(e) => setStoreForm(p => ({ ...p, lat: e.target.value }))}
                          />
                        </div>
                        <div>
                          <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Kinh độ (Longitude)</label>
                          <input
                            type="number"
                            step="any"
                            placeholder="106.6602..."
                            className="w-full border border-neutral-300 rounded px-3 py-2 bg-white font-mono"
                            value={storeForm.lng}
                            onChange={(e) => setStoreForm(p => ({ ...p, lng: e.target.value }))}
                          />
                        </div>
                      </div>

                      {storeForm.id > 0 && (
                        <div className="flex items-center gap-2 pt-1">
                          <input
                            type="checkbox"
                            id="store_is_active"
                            checked={storeForm.is_active}
                            onChange={(e) => setStoreForm(p => ({ ...p, is_active: e.target.checked }))}
                          />
                          <label htmlFor="store_is_active" className="text-[10px] uppercase font-bold text-neutral-700 cursor-pointer select-none">
                            Hoạt động (Active)
                          </label>
                        </div>
                      )}

                      <div className="flex gap-2 justify-end pt-3">
                        <button
                          type="button"
                          onClick={() => setShowStoreModal(false)}
                          className="text-[10px] border border-neutral-350 px-5 py-2.5 font-bold uppercase rounded text-neutral-600 hover:text-black"
                        >
                          Hủy
                        </button>
                        <button
                          type="submit"
                          disabled={loading}
                          className="text-[10px] bg-black text-white px-6 py-2.5 font-black uppercase rounded hover:bg-neutral-800"
                        >
                          Lưu
                        </button>
                      </div>
                    </form>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
