import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { productAPI } from '../services/productAPI'
import { inventoryAPI } from '../services/inventoryAPI'
import type { Store, Category, Brand } from '../types'

import AdminOrdersTab from '../components/admin/AdminOrdersTab'
import AdminVouchersTab from '../components/admin/AdminVouchersTab'
import AdminInventoryTab from '../components/admin/AdminInventoryTab'
import AdminCatalogTab from '../components/admin/AdminCatalogTab'
import AdminUsersTab from '../components/admin/AdminUsersTab'
import AdminBannersTab from '../components/admin/AdminBannersTab'
import AdminFlashSaleTab from '../components/admin/AdminFlashSaleTab'

type ActiveTab = 'orders' | 'vouchers' | 'inventory' | 'catalog' | 'users' | 'banners' | 'flashsales'

export default function AdminDashboardPage() {
  const [activeTab, setActiveTab] = useState<ActiveTab>('orders')

  // Global Lookups
  const [stores, setStores] = useState<Store[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [brands, setBrands] = useState<Brand[]>([])

  const loadLookups = async () => {
    try {
      const sList = await inventoryAPI.listStores()
      setStores(sList)
      const cList = await productAPI.getCategories()
      setCategories(cList)
      const bList = await productAPI.getBrands()
      setBrands(bList)
    } catch (err) {
      console.error('Failed to load lookup data:', err)
    }
  }

  useEffect(() => {
    void loadLookups()
  }, [])

  return (
    <div className="flex-1 min-h-[calc(100vh-140px)] bg-neutral-50 flex flex-col md:flex-row font-sans">
      {/* Sidebar Navigation */}
      <aside className="w-full md:w-64 bg-black text-white shrink-0 flex flex-col justify-between py-6 border-r border-neutral-850">
        <div className="space-y-6">
          <div className="px-6 pb-4 border-b border-neutral-850">
            <h2 className="text-xs font-black uppercase tracking-widest text-neutral-400">Jiyuu Control</h2>
            <p className="text-[10px] text-neutral-500 font-mono mt-1">Hệ thống quản trị viên</p>
          </div>

          <nav className="space-y-1.5 px-3">
            {[
              { id: 'orders', label: 'Đơn hàng', desc: 'Quản lý đơn đặt hàng' },
              { id: 'vouchers', label: 'Mã giảm giá (Vouchers)', desc: 'Mã giảm giá đơn hàng' },
              { id: 'flashsales', label: 'Quản lý Flash Sale', desc: 'Giảm giá sốc, đếm ngược' },
              { id: 'inventory', label: 'Kho & Cửa hàng', desc: 'Kho hàng, nhập kho, nhà cc' },
              { id: 'catalog', label: 'Danh mục & Sản phẩm', desc: 'Thêm mới catalog sản phẩm' },
              { id: 'banners', label: 'Banners trang chủ', desc: 'Quản lý ảnh slider trang chủ' },
              { id: 'users', label: 'Thành viên', desc: 'Khóa / mở khóa tài khoản' },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as ActiveTab)}
                className={`w-full text-left px-4 py-3 rounded transition-all flex flex-col ${
                  activeTab === tab.id
                    ? 'bg-neutral-850 text-white border-l-4 border-white'
                    : 'text-neutral-450 hover:bg-neutral-900 hover:text-white'
                }`}
              >
                <span className="text-xs font-bold">{tab.label}</span>
                <span className="text-[9px] text-neutral-555 font-medium mt-0.5">{tab.desc}</span>
              </button>
            ))}
          </nav>
        </div>

        <div className="px-6 pt-6 border-t border-neutral-850">
          <Link
            to="/"
            className="text-[10px] uppercase font-black tracking-widest text-neutral-400 hover:text-white flex items-center gap-1.5 transition-colors"
          >
            ← Về cửa hàng
          </Link>
        </div>
      </aside>

      {/* Main Administrative Views Area */}
      <main className="flex-1 p-6 md:p-8 max-w-7xl overflow-hidden">
        {activeTab === 'orders' && <AdminOrdersTab stores={stores} />}
        {activeTab === 'vouchers' && <AdminVouchersTab />}
        {activeTab === 'flashsales' && <AdminFlashSaleTab />}
        {activeTab === 'inventory' && <AdminInventoryTab stores={stores} reloadLookups={loadLookups} setActiveTab={(tab) => setActiveTab(tab as any)} />}
        {activeTab === 'catalog' && (
          <AdminCatalogTab categories={categories} brands={brands} reloadLookups={loadLookups} />
        )}
        {activeTab === 'banners' && <AdminBannersTab />}
        {activeTab === 'users' && <AdminUsersTab />}
      </main>
    </div>
  )
}
