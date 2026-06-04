import { useEffect, useState, useMemo } from 'react'
import { analyticsAPI, type GetAnalyticsParams } from '../../services/analyticsAPI'
import type { AnalyticsSummary, Store } from '../../types'

interface AdminAnalyticsTabProps {
  stores: Store[]
}

type DatePreset = 'today' | 'yesterday' | '7days' | '30days' | 'custom'

export default function AdminAnalyticsTab({ stores }: AdminAnalyticsTabProps) {
  const [preset, setPreset] = useState<DatePreset>('30days')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [selectedStore, setSelectedStore] = useState<number | undefined>(undefined)
  
  const [data, setData] = useState<AnalyticsSummary | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Date range picker helper
  const getDatesForPreset = (presetType: DatePreset) => {
    const today = new Date()
    const formatDate = (d: Date) => d.toISOString().split('T')[0]

    switch (presetType) {
      case 'today':
        return { start: formatDate(today), end: formatDate(today) }
      case 'yesterday': {
        const yesterday = new Date(today)
        yesterday.setDate(today.getDate() - 1)
        return { start: formatDate(yesterday), end: formatDate(yesterday) }
      }
      case '7days': {
        const last7 = new Date(today)
        last7.setDate(today.getDate() - 7)
        return { start: formatDate(last7), end: formatDate(today) }
      }
      case '30days': {
        const last30 = new Date(today)
        last30.setDate(today.getDate() - 30)
        return { start: formatDate(last30), end: formatDate(today) }
      }
      default:
        return { start: startDate, end: endDate }
    }
  }

  // Effect to handle preset change
  useEffect(() => {
    if (preset !== 'custom') {
      const dates = getDatesForPreset(preset)
      setStartDate(dates.start)
      setEndDate(dates.end)
    }
  }, [preset])

  // Load analytics summary
  const loadAnalytics = async () => {
    try {
      setLoading(true)
      setError(null)
      const params: GetAnalyticsParams = {
        start_date: startDate || undefined,
        end_date: endDate || undefined,
        store_id: selectedStore,
      }
      const summary = await analyticsAPI.getAnalytics(params)
      setData(summary)
    } catch (err: any) {
      setError(err?.message || 'Không thể tải báo cáo thống kê')
    } finally {
      setLoading(false)
    }
  }

  // Reload when date or store filter changes
  useEffect(() => {
    if (startDate && endDate) {
      void loadAnalytics()
    }
  }, [startDate, endDate, selectedStore])

  // SVG Chart Computations
  const lineChartData = useMemo(() => {
    if (!data || !data.sales_over_time || data.sales_over_time.length === 0) return null
    const points = data.sales_over_time
    const maxVal = Math.max(...points.map((p) => p.revenue), 1000000)
    
    const width = 600
    const height = 200
    const paddingLeft = 60
    const paddingRight = 20
    const paddingTop = 20
    const paddingBottom = 30

    const graphWidth = width - paddingLeft - paddingRight
    const graphHeight = height - paddingTop - paddingBottom

    const svgPoints = points.map((p, i) => {
      const x = paddingLeft + (points.length > 1 ? (i / (points.length - 1)) * graphWidth : graphWidth / 2)
      const y = paddingTop + graphHeight - (p.revenue / maxVal) * graphHeight
      return { x, y, raw: p }
    })

    const pathD = svgPoints.reduce((path, p, i) => {
      return path + `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y} `
    }, '')

    const fillD = pathD 
      ? pathD + `L ${svgPoints[svgPoints.length - 1].x} ${height - paddingBottom} L ${svgPoints[0].x} ${height - paddingBottom} Z`
      : ''

    return { svgPoints, pathD, fillD, maxVal, width, height, paddingLeft, graphHeight, paddingTop, paddingBottom }
  }, [data])

  if (loading && !data) {
    return (
      <div className="py-12 flex items-center justify-center min-h-[400px]">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black mx-auto"></div>
          <p className="text-xs font-semibold text-neutral-500 uppercase tracking-wider">Đang tính toán dữ liệu báo cáo...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Top Filter Bar */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-white border border-neutral-200 rounded-lg p-4 shadow-sm">
        <div className="space-y-1">
          <h1 className="text-lg font-black uppercase text-neutral-900 tracking-tight">Phân Tích & Thống Kê</h1>
          <p className="text-xs text-neutral-400">Xem doanh thu, hiệu suất bán hàng và sức khỏe kho hàng</p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          {/* Preset Buttons */}
          <div className="flex items-center rounded border border-neutral-300 bg-neutral-50 p-0.5">
            {(['today', 'yesterday', '7days', '30days', 'custom'] as DatePreset[]).map((pType) => (
              <button
                key={pType}
                type="button"
                onClick={() => setPreset(pType)}
                className={`px-3 py-1.5 rounded text-[11px] font-bold uppercase transition-all ${
                  preset === pType
                    ? 'bg-white text-black shadow-sm'
                    : 'text-neutral-500 hover:text-black'
                }`}
              >
                {pType === 'today' && 'Hôm nay'}
                {pType === 'yesterday' && 'Hôm qua'}
                {pType === '7days' && '7 ngày'}
                {pType === '30days' && '30 ngày'}
                {pType === 'custom' && 'Tùy chọn'}
              </button>
            ))}
          </div>

          {/* Custom Date Picker inputs */}
          {preset === 'custom' && (
            <div className="flex items-center gap-2">
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="border border-neutral-300 text-xs rounded px-2 py-1 focus:outline-none focus:border-black"
              />
              <span className="text-neutral-400 text-xs">đến</span>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="border border-neutral-300 text-xs rounded px-2 py-1 focus:outline-none focus:border-black"
              />
            </div>
          )}

          {/* Store Selector */}
          <select
            value={selectedStore === undefined ? '' : selectedStore}
            onChange={(e) => setSelectedStore(e.target.value === '' ? undefined : Number(e.target.value))}
            className="border border-neutral-300 bg-white text-xs font-semibold rounded px-3 py-1.5 focus:outline-none focus:border-black cursor-pointer"
          >
            <option value="">Tất cả cửa hàng</option>
            {stores.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {error && (
        <div className="border border-red-200 bg-red-50 text-red-650 text-xs font-semibold px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {data && (
        <>
          {/* Summary KPIs Row */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
            {/* KPI: Sales */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm hover:shadow transition-shadow">
              <span className="text-[10px] font-black uppercase text-neutral-400 tracking-wider">Tổng Doanh Thu</span>
              <div className="flex items-baseline justify-between mt-2.5">
                <span className="text-2xl font-black text-neutral-900">
                  {data.total_sales.toLocaleString('vi-VN')} đ
                </span>
                {/* Growth indicator */}
                <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${
                  data.sales_growth >= 0 ? 'text-emerald-700 bg-emerald-50' : 'text-red-650 bg-red-50'
                }`}>
                  {data.sales_growth >= 0 ? '↑' : '↓'} {Math.abs(data.sales_growth).toFixed(1)}%
                </span>
              </div>
              <div className="text-[10px] text-neutral-400 mt-2 font-medium">
                So với chu kỳ trước: {data.prev_total_sales.toLocaleString('vi-VN')} đ
              </div>
            </div>

            {/* KPI: Orders */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm hover:shadow transition-shadow">
              <span className="text-[10px] font-black uppercase text-neutral-400 tracking-wider">Đơn Hàng</span>
              <div className="flex items-baseline justify-between mt-2.5">
                <span className="text-2xl font-black text-neutral-900">
                  {data.total_orders} đơn
                </span>
                <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${
                  data.orders_growth >= 0 ? 'text-emerald-700 bg-emerald-50' : 'text-red-650 bg-red-50'
                }`}>
                  {data.orders_growth >= 0 ? '↑' : '↓'} {Math.abs(data.orders_growth).toFixed(1)}%
                </span>
              </div>
              <div className="text-[10px] text-neutral-400 mt-2 font-medium">
                So với chu kỳ trước: {data.prev_total_orders} đơn
              </div>
            </div>

            {/* KPI: AOV */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm hover:shadow transition-shadow">
              <span className="text-[10px] font-black uppercase text-neutral-400 tracking-wider">Giá Trị Đơn Trung Bình</span>
              <div className="flex items-baseline justify-between mt-2.5">
                <span className="text-2xl font-black text-neutral-900">
                  {data.average_order_value.toLocaleString('vi-VN')} đ
                </span>
                <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${
                  data.aov_growth >= 0 ? 'text-emerald-700 bg-emerald-50' : 'text-red-650 bg-red-50'
                }`}>
                  {data.aov_growth >= 0 ? '↑' : '↓'} {Math.abs(data.aov_growth).toFixed(1)}%
                </span>
              </div>
              <div className="text-[10px] text-neutral-400 mt-2 font-medium">
                So với chu kỳ trước: {data.prev_average_order_value.toLocaleString('vi-VN')} đ
              </div>
            </div>

            {/* KPI: Total Items Sold */}
            <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm hover:shadow transition-shadow">
              <span className="text-[10px] font-black uppercase text-neutral-400 tracking-wider">Sản Phẩm Đã Bán</span>
              <div className="flex items-baseline mt-2.5">
                <span className="text-2xl font-black text-neutral-900">
                  {data.items_sold} món
                </span>
              </div>
              <div className="text-[10px] text-neutral-400 mt-2 font-medium">
                Tổng sản phẩm bán được theo khoảng thời gian đã chọn
              </div>
            </div>
          </div>

          {/* Inventory Summary Banner Widgets */}
          <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm space-y-4">
            <h3 className="text-xs font-black uppercase tracking-wider text-neutral-850">
              📊 Tổng Quan Kho Hàng & Tồn Kho (Hiện Tại)
            </h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6 text-center">
              <div className="p-3 bg-neutral-50 rounded border border-neutral-150">
                <span className="text-[10px] font-bold text-neutral-450 uppercase tracking-wide">Tổng số SKUs</span>
                <div className="text-xl font-black text-neutral-800 mt-1">{data.inventory.total_sku}</div>
              </div>
              <div className="p-3 bg-neutral-50 rounded border border-neutral-150">
                <span className="text-[10px] font-bold text-neutral-450 uppercase tracking-wide">Tổng sản phẩm hiện có</span>
                <div className="text-xl font-black text-neutral-800 mt-1">{data.inventory.total_quantity}</div>
              </div>
              <div className="p-3 bg-amber-50/50 rounded border border-amber-200 text-amber-900">
                <span className="text-[10px] font-bold text-amber-700 uppercase tracking-wide">Sắp hết hàng</span>
                <div className="text-xl font-black mt-1">{data.inventory.low_stock_count} SKUs</div>
              </div>
              <div className="p-3 bg-red-50/50 rounded border border-red-200 text-red-950">
                <span className="text-[10px] font-bold text-red-700 uppercase tracking-wide">Đã cháy hàng</span>
                <div className="text-xl font-black mt-1">{data.inventory.out_of_stock_count} SKUs</div>
              </div>
            </div>
          </div>

          {/* Sales charts and store sales */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            {/* Sales Daily Trend chart */}
            <div className="lg:col-span-8 bg-white border border-neutral-200 rounded-lg p-5 shadow-sm flex flex-col justify-between">
              <div className="pb-3 border-b border-neutral-150 flex items-center justify-between">
                <h3 className="text-xs font-black uppercase text-neutral-850 tracking-wider">📈 Xu Hướng Doanh Thu Ngày</h3>
                <span className="text-[10px] text-neutral-400">Doanh thu qua từng ngày (đ)</span>
              </div>

              {/* Line chart SVG */}
              {lineChartData ? (
                <div className="relative mt-4">
                  <svg
                    viewBox={`0 0 ${lineChartData.width} ${lineChartData.height}`}
                    className="w-full h-auto overflow-visible"
                  >
                    <defs>
                      <linearGradient id="sales-gradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor="#000000" stopOpacity="0.1" />
                        <stop offset="100%" stopColor="#000000" stopOpacity="0" />
                      </linearGradient>
                    </defs>

                    {/* Y-axis helper grids */}
                    {[0, 0.25, 0.5, 0.75, 1].map((ratio, index) => {
                      const y = lineChartData.paddingTop + lineChartData.graphHeight * (1 - ratio)
                      return (
                        <g key={index}>
                          <line
                            x1={lineChartData.paddingLeft}
                            y1={y}
                            x2={lineChartData.width - 20}
                            y2={y}
                            stroke="#f0f0f0"
                            strokeWidth="1"
                          />
                          <text
                            x={lineChartData.paddingLeft - 10}
                            y={y + 4}
                            textAnchor="end"
                            className="text-[9px] font-semibold text-neutral-450 font-mono"
                          >
                            {Math.round((lineChartData.maxVal * ratio) / 1000).toLocaleString('vi-VN')}k
                          </text>
                        </g>
                      )
                    })}

                    {/* Gradient Shading */}
                    {lineChartData.fillD && (
                      <path d={lineChartData.fillD} fill="url(#sales-gradient)" />
                    )}

                    {/* Main Line path */}
                    {lineChartData.pathD && (
                      <path
                        d={lineChartData.pathD}
                        fill="none"
                        stroke="#000000"
                        strokeWidth="2.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      />
                    )}

                    {/* Grid dots points on hover */}
                    {lineChartData.svgPoints.map((p, idx) => (
                      <circle
                        key={idx}
                        cx={p.x}
                        cy={p.y}
                        r="3.5"
                        fill="#ffffff"
                        stroke="#000000"
                        strokeWidth="2"
                        className="cursor-pointer hover:r-5 transition-all"
                      >
                        <title>{`${p.raw.date}: ${p.raw.revenue.toLocaleString('vi-VN')}đ`}</title>
                      </circle>
                    ))}
                  </svg>
                </div>
              ) : (
                <div className="h-48 flex items-center justify-center text-xs text-neutral-400 italic">
                  Không có dữ liệu doanh thu trong khoảng thời gian này
                </div>
              )}
            </div>

            {/* Store performance sales */}
            <div className="lg:col-span-4 bg-white border border-neutral-200 rounded-lg p-5 shadow-sm flex flex-col justify-between">
              <div className="pb-3 border-b border-neutral-150">
                <h3 className="text-xs font-black uppercase text-neutral-850 tracking-wider">🏢 Bán Hàng Theo Cửa Hàng</h3>
              </div>

              <div className="space-y-4 mt-4 flex-1 flex flex-col justify-center">
                {(data.store_sales || []).length > 0 ? (
                  (data.store_sales || []).map((store) => {
                    const maxRev = Math.max(...(data.store_sales || []).map((s) => s.revenue), 1)
                    const percent = (store.revenue / maxRev) * 100

                    return (
                      <div key={store.store_id} className="space-y-1.5">
                        <div className="flex items-center justify-between text-xs font-semibold">
                          <span className="text-neutral-800 line-clamp-1">{store.store_name}</span>
                          <span className="text-neutral-900 font-bold font-mono">
                            {store.revenue.toLocaleString('vi-VN')}đ
                          </span>
                        </div>
                        {/* Horizontal Bar */}
                        <div className="w-full bg-neutral-100 rounded-full h-2 overflow-hidden">
                          <div
                            className="bg-black h-full rounded-full transition-all duration-500"
                            style={{ width: `${percent}%` }}
                          ></div>
                        </div>
                        <div className="text-[9px] text-neutral-400 font-mono">
                          {store.orders_count} đơn hàng hoàn thành
                        </div>
                      </div>
                    )
                  })
                ) : (
                  <div className="text-center text-xs text-neutral-400 italic py-12">
                    Không có doanh thu theo cửa hàng
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Insights block: Top products and Top categories */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
            
            {/* Top categories rank */}
            <div className="lg:col-span-4 bg-white border border-neutral-200 rounded-lg p-5 shadow-sm">
              <h3 className="text-xs font-black uppercase text-neutral-850 tracking-wider pb-3 border-b border-neutral-150 mb-4">
                🏷️ Phân Loại Danh Mục Bán Chạy
              </h3>

              {(data.top_categories || []).length > 0 ? (
                <div className="space-y-4">
                  {(data.top_categories || []).map((cat, idx) => {
                    const totalCatRevenue = (data.top_categories || []).reduce((acc, c) => acc + c.revenue, 0)
                    const sharePercent = totalCatRevenue > 0 ? (cat.revenue / totalCatRevenue) * 100 : 0
                    return (
                      <div key={idx} className="space-y-1">
                        <div className="flex justify-between items-baseline text-xs font-semibold">
                          <span className="text-neutral-700">{cat.category_name}</span>
                          <span className="text-neutral-900 font-bold">{cat.revenue.toLocaleString('vi-VN')}đ</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="flex-1 bg-neutral-100 h-1.5 rounded-full overflow-hidden">
                            <div className="bg-black h-full" style={{ width: `${sharePercent}%` }}></div>
                          </div>
                          <span className="text-[10px] text-neutral-450 font-bold font-mono shrink-0">
                            {sharePercent.toFixed(0)}%
                          </span>
                        </div>
                        <div className="text-[9px] text-neutral-400">Đã bán {cat.sold_qty} sản phẩm</div>
                      </div>
                    )
                  })}
                </div>
              ) : (
                <div className="text-center text-xs text-neutral-400 italic py-12">
                  Chưa có dữ liệu danh mục bán chạy
                </div>
              )}
            </div>

            {/* Top products table */}
            <div className="lg:col-span-8 bg-white border border-neutral-200 rounded-lg p-5 shadow-sm">
              <h3 className="text-xs font-black uppercase text-neutral-850 tracking-wider pb-3 border-b border-neutral-150 mb-4">
                👑 Top 5 Sản Phẩm Bán Chạy Nhất
              </h3>

              {(data.top_products || []).length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs divide-y divide-neutral-200">
                    <thead>
                      <tr className="text-neutral-450 uppercase font-black tracking-wider text-[9px]">
                        <th className="pb-3">Sản phẩm</th>
                        <th className="pb-3 text-right">Số lượng bán</th>
                        <th className="pb-3 text-right">Doanh thu đạt được</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-150">
                      {(data.top_products || []).map((prod, idx) => (
                        <tr key={idx} className="text-neutral-800 font-medium">
                          <td className="py-3 font-semibold text-neutral-900">{prod.product_name}</td>
                          <td className="py-3 text-right font-mono">{prod.sold_qty}</td>
                          <td className="py-3 text-right font-bold text-neutral-900 font-mono">
                            {prod.revenue.toLocaleString('vi-VN')} đ
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center text-xs text-neutral-400 italic py-12">
                  Chưa có sản phẩm bán chạy trong khoảng thời gian này
                </div>
              )}
            </div>
          </div>

          {/* Bottom Row: Order Status Distributions */}
          <div className="bg-white border border-neutral-200 rounded-lg p-5 shadow-sm">
            <h3 className="text-xs font-black uppercase text-neutral-850 tracking-wider pb-3 border-b border-neutral-150 mb-4">
              📦 Phân Bố Trạng Thái Đơn Hàng
            </h3>
            
            {(data.status_distribution || []).length > 0 ? (
              <div className="flex flex-wrap gap-4 items-center justify-around">
                {(data.status_distribution || []).map((status, index) => {
                  const totalCount = (data.status_distribution || []).reduce((acc, s) => acc + s.count, 0)
                  const percent = totalCount > 0 ? (status.count / totalCount) * 100 : 0
                  return (
                    <div key={index} className="flex items-center gap-3 bg-neutral-50 border border-neutral-200 rounded px-4 py-2.5 min-w-[140px]">
                      <div className="space-y-1">
                        <span className="text-[10px] font-bold text-neutral-400 uppercase tracking-wide">
                          {status.status_label}
                        </span>
                        <div className="flex items-baseline gap-1.5">
                          <span className="text-sm font-black text-neutral-800">{status.count} đơn</span>
                          <span className="text-[10px] font-bold text-neutral-500 font-mono">
                            ({percent.toFixed(0)}%)
                          </span>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="text-center text-xs text-neutral-400 italic py-4">
                Chưa có đơn hàng nào
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
