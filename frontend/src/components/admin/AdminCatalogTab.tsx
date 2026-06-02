import React, { useEffect, useState, useRef } from 'react'
import { productAPI } from '../../services/productAPI'
import { uploadAPI } from '../../services/uploadAPI'
import type { Category, Brand, Product } from '../../types'
import ImageUploader from './ImageUploader'

interface AdminCatalogTabProps {
  categories: Category[]
  brands: Brand[]
  reloadLookups: () => Promise<void>
}

// Standard Cartesian Product helper
function cartesianProduct(arrays: string[][]): string[][] {
  return arrays.reduce<string[][]>((a, b) => {
    return a.flatMap(d => b.map(e => [...d, e]))
  }, [[]])
}

export default function AdminCatalogTab({ categories, brands, reloadLookups }: AdminCatalogTabProps) {
  const [subTab, setSubTab] = useState<'product' | 'category' | 'brand'>('product')
  const [loading, setLoading] = useState(false)
  const [saveProgress, setSaveProgress] = useState<string | null>(null)

  // --------------------------------------------------
  // Products List State
  // --------------------------------------------------
  const [products, setProducts] = useState<Product[]>([])
  const [productsLoading, setProductsLoading] = useState(false)
  const [productPage, setProductPage] = useState(1)
  const [productTotalPages, setProductTotalPages] = useState(1)
  const [productTotal, setProductTotal] = useState(0)
  const [productSearch, setProductSearch] = useState('')

  // Toggle full workspace product form
  const [showProductForm, setShowProductForm] = useState(false)
  const [editProductId, setEditProductId] = useState<string | null>(null)

  // Product Form states
  const [productForm, setProductForm] = useState({
    id: '',
    categoryId: '',
    brandId: '',
    name: '',
    metaTitle: '',
    metaDescription: '',
    imgThumb: '',
    images: [] as string[],
    weight: 0,
    lowStockThreshold: 5,
    specs: [] as { group: string; key: string; value: string; sort_order: number }[],
  })

  // Specs helper
  const [newSpec, setNewSpec] = useState({ group: '', key: '', value: '' })

  // --------------------------------------------------
  // Dynamic Variants & Attributes State
  // --------------------------------------------------
  const [hasVariants, setHasVariants] = useState(false)
  const [options, setOptions] = useState<{ name: string; values: string[] }[]>([])
  const [variantRows, setVariantRows] = useState<any[]>([])
  const [deletedVariantIds, setDeletedVariantIds] = useState<number[]>([])

  // Temporary state for adding a new attribute type
  const [newAttributeName, setNewAttributeName] = useState('')
  const [tempAttrValues, setTempAttrValues] = useState<{ [key: number]: string }>({})

  // Bulk Edit states for variants
  const [bulkPrice, setBulkPrice] = useState('')
  const [bulkPriceBase, setBulkPriceBase] = useState('')
  const [bulkWeight, setBulkWeight] = useState('')

  // Variant specific image uploading helper refs
  const variantFileInputRef = useRef<HTMLInputElement>(null)
  const [uploadingVariantIndex, setUploadingVariantIndex] = useState<number | null>(null)

  // Secondary images uploading helper refs
  const secondaryFileInputRef = useRef<HTMLInputElement>(null)
  const [uploadingSecondary, setUploadingSecondary] = useState(false)

  // Category & Brand pagination
  const [editCategoryId, setEditCategoryId] = useState<number | null>(null)
  const [editBrandId, setEditBrandId] = useState<number | null>(null)
  const [categoryPage, setCategoryPage] = useState(1)
  const [brandPage, setBrandPage] = useState(1)

  const categoryPageSize = 10
  const totalCategoryPages = Math.ceil(categories.length / categoryPageSize)
  const safeCategoryPage = Math.min(categoryPage, Math.max(1, totalCategoryPages))
  const paginatedCategories = categories.slice(
    (safeCategoryPage - 1) * categoryPageSize,
    safeCategoryPage * categoryPageSize
  )

  const brandPageSize = 10
  const totalBrandPages = Math.ceil(brands.length / brandPageSize)
  const safeBrandPage = Math.min(brandPage, Math.max(1, totalBrandPages))
  const paginatedBrands = brands.slice(
    (safeBrandPage - 1) * brandPageSize,
    safeBrandPage * brandPageSize
  )

  const [categoryForm, setCategoryForm] = useState({ name: '', parentId: '', sortOrder: 0 })
  const [brandForm, setBrandForm] = useState({ name: '', logoUrl: '', isActive: true })

  // --------------------------------------------------
  // Loader Actions
  // --------------------------------------------------
  const loadProducts = async (pageNumber: number, search: string) => {
    try {
      setProductsLoading(true)
      const res = await productAPI.getProducts({ page: pageNumber, limit: 10, q: search })
      setProducts(res.data || [])
      if (res.pagination) {
        setProductPage(res.pagination.page)
        setProductTotalPages(res.pagination.total_pages)
        setProductTotal(res.pagination.total)
      } else {
        setProductPage(1)
        setProductTotalPages(1)
        setProductTotal(res.data.length || 0)
      }
    } catch (err) {
      console.error('Failed to load products:', err)
    } finally {
      setProductsLoading(false)
    }
  }

  useEffect(() => {
    if (subTab === 'product') {
      void loadProducts(1, '')
    }
  }, [subTab])

  // Automatically update variant names in real-time when the root product name changes
  useEffect(() => {
    if (hasVariants && variantRows.length > 0) {
      setVariantRows(prev => prev.map(row => {
        const comboName = Object.values(row.attributes).join(' - ')
        return {
          ...row,
          name: `${productForm.name} ${comboName}`.trim()
        }
      }))
    }
  }, [productForm.name])

  const handleProductSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    void loadProducts(1, productSearch)
  }

  // Specs helper
  const handleAddSpec = () => {
    if (!newSpec.group || !newSpec.key || !newSpec.value) return
    setProductForm(p => ({
      ...p,
      specs: [...p.specs, { ...newSpec, sort_order: p.specs.length + 1 }],
    }))
    setNewSpec({ group: '', key: '', value: '' })
  }

  const handleRemoveSpec = (index: number) => {
    setProductForm(p => ({
      ...p,
      specs: p.specs.filter((_, i) => i !== index),
    }))
  }

  // --------------------------------------------------
  // Attribute Management Actions
  // --------------------------------------------------
  const handleAddAttributeType = () => {
    const name = newAttributeName.trim()
    if (!name) return
    if (options.some(opt => opt.name.toLowerCase() === name.toLowerCase())) {
      alert('Thuộc tính này đã tồn tại!')
      return
    }
    setOptions([...options, { name, values: [] }])
    setNewAttributeName('')
  }

  const handleRemoveAttributeType = (index: number) => {
    setOptions(options.filter((_, i) => i !== index))
  }

  const handleAddAttributeValue = (index: number, customVal?: string) => {
    const rawVal = customVal !== undefined ? customVal : (tempAttrValues[index] || '')
    if (!rawVal.trim()) return

    // Support comma splitting (e.g. "đỏ, xanh, trắng")
    const newValues = rawVal.split(',')
      .map(v => v.trim())
      .filter(v => v !== '')

    const currentOpt = options[index]
    const currentSet = new Set(currentOpt.values)
    const added: string[] = []

    newValues.forEach(val => {
      if (!currentSet.has(val)) {
        currentSet.add(val)
        added.push(val)
      }
    })

    if (added.length === 0) return

    const updatedOptions = [...options]
    updatedOptions[index] = {
      ...currentOpt,
      values: Array.from(currentSet)
    }
    setOptions(updatedOptions)
    
    // Clear temp input
    if (customVal === undefined) {
      setTempAttrValues(prev => ({ ...prev, [index]: '' }))
    }
  }

  const handleRemoveAttributeValue = (optIndex: number, valIndex: number) => {
    const updatedOptions = [...options]
    updatedOptions[optIndex] = {
      ...updatedOptions[optIndex],
      values: updatedOptions[optIndex].values.filter((_, i) => i !== valIndex)
    }
    setOptions(updatedOptions)
  }

  const handleAutoGenerateVariants = () => {
    // Proactively flush any remaining text in the inputs
    const updatedOptions = options.map((opt, index) => {
      const remainingText = tempAttrValues[index] || ''
      if (remainingText.trim()) {
        const rawSplits = remainingText.split(',').map(v => v.trim()).filter(v => v !== '')
        const combined = Array.from(new Set([...opt.values, ...rawSplits]))
        return { ...opt, values: combined }
      }
      return opt
    })

    // Update state to match
    setOptions(updatedOptions)

    // Clear temp value states
    const clearedTemp: any = {}
    Object.keys(tempAttrValues).forEach(k => {
      clearedTemp[k] = ''
    })
    setTempAttrValues(clearedTemp)

    if (updatedOptions.length === 0 || updatedOptions.some(opt => opt.values.length === 0)) {
      alert('Vui lòng cấu hình ít nhất 1 thuộc tính và điền đầy đủ các lựa chọn giá trị!')
      return
    }

    const attributeNames = updatedOptions.map(opt => opt.name)
    const valueArrays = updatedOptions.map(opt => opt.values)
    const combinations = cartesianProduct(valueArrays)

    const generated = combinations.map(combo => {
      const attrs: { [key: string]: string } = {}
      attributeNames.forEach((name, i) => {
        attrs[name] = combo[i]
      })

      const comboName = combo.join(' - ')
      const variantName = `${productForm.name} ${comboName}`
      const cleanNameForSKU = comboName.replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
      const generatedSKU = `${productForm.id ? productForm.id.toUpperCase() : 'SKU'}-${cleanNameForSKU}`

      return {
        attributes: attrs,
        name: variantName,
        sku: generatedSKU,
        price: 0,
        priceBase: '',
        weight: productForm.weight || 0,
        isActive: true,
        isExisting: false,
        image: ''
      }
    })

    setVariantRows(generated)
  }

  // Bulk edit applies
  const handleBulkApplyPrice = () => {
    const val = Number(bulkPrice)
    if (isNaN(val) || val <= 0) {
      alert('Vui lòng nhập giá trị hợp lệ!')
      return
    }
    setVariantRows(prev => prev.map(row => ({ ...row, price: val })))
  }

  const handleBulkApplyPriceBase = () => {
    setVariantRows(prev => prev.map(row => ({ ...row, priceBase: bulkPriceBase })))
  }

  const handleBulkApplyWeight = () => {
    const val = Number(bulkWeight)
    if (isNaN(val) || val < 0) {
      alert('Vui lòng nhập giá trị hợp lệ!')
      return
    }
    setVariantRows(prev => prev.map(row => ({ ...row, weight: val })))
  }

  // Variant Image trigger
  const triggerVariantImageUpload = (index: number) => {
    setUploadingVariantIndex(index)
    variantFileInputRef.current?.click()
  }

  const handleVariantImageChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (uploadingVariantIndex === null || !e.target.files || !e.target.files[0]) return
    const file = e.target.files[0]
    try {
      setLoading(true)
      const res = await uploadAPI.uploadImage(file)
      setVariantRows(prev => {
        const updated = [...prev]
        updated[uploadingVariantIndex] = {
          ...updated[uploadingVariantIndex],
          image: res.url
        }
        return updated
      })
    } catch (err: any) {
      alert(err.message || 'Lỗi khi tải ảnh lên')
    } finally {
      setLoading(false)
      setUploadingVariantIndex(null)
      if (variantFileInputRef.current) variantFileInputRef.current.value = ''
    }
  }

  // --------------------------------------------------
  // Save Unified Product & Variants
  // --------------------------------------------------
  const handleSaveProductWorkspace = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!productForm.id || !productForm.categoryId || !productForm.brandId || !productForm.name) {
      alert('Vui lòng điền đầy đủ các thông tin bắt buộc (*)')
      return
    }

    try {
      setLoading(true)
      setSaveProgress('Đang lưu thông tin sản phẩm chính...')

      const basePayload = {
        category_id: Number(productForm.categoryId),
        brand_id: Number(productForm.brandId),
        name: productForm.name.trim(),
        meta_title: productForm.metaTitle.trim() || null,
        meta_description: productForm.metaDescription.trim() || null,
        img_thumb: productForm.imgThumb.trim() || null,
        images: productForm.images,
        weight: productForm.weight ? Number(productForm.weight) : null,
        low_stock_threshold: Number(productForm.lowStockThreshold),
      }

      const productId = productForm.id.trim().toLowerCase()

      // Save/Update Root Product
      if (editProductId !== null) {
        await productAPI.adminUpdateProduct(editProductId, {
          ...basePayload,
          specs: productForm.specs.map(s => ({ ...s, unit: '' })),
        })
      } else {
        await productAPI.adminCreateProduct({
          ...basePayload,
          id: productId,
          specs: productForm.specs.map(s => ({ ...s, unit: '' })),
        })
      }

      // Handle Variant Generation if enabled
      if (hasVariants && variantRows.length > 0) {
        setSaveProgress('Đang đăng ký các thuộc tính lựa chọn...')

        const savedValuesMap: { [key: string]: number } = {}
        
        // Loop and add options sequentially
        for (const opt of options) {
          if (opt.values.length === 0) continue
          const res = await productAPI.adminAddOptionValues({
            product_id: productId,
            option_name: opt.name,
            values: opt.values.map((v, index) => ({ value: v, sort_order: index + 1 }))
          })
          res.forEach(item => {
            savedValuesMap[`${opt.name}:${item.value}`] = item.id
          })
        }

        // Loop and create variants that do not already exist
        const newVariants = variantRows.filter(row => !row.isExisting)
        
        for (let i = 0; i < newVariants.length; i++) {
          const row = newVariants[i]
          setSaveProgress(`Đang tạo biến thể (${i + 1}/${newVariants.length}): ${row.name}`)

          const ids = Object.entries(row.attributes)
            .map(([optName, valName]) => savedValuesMap[`${optName}:${valName as string}`])
            .filter(Boolean) as number[]

          if (ids.length === 0) continue

          await productAPI.adminGenerateVariant(productId, {
            name: row.name,
            sku: row.sku,
            price: Number(row.price),
            price_base: row.priceBase ? Number(row.priceBase) : undefined,
            weight: row.weight ? Number(row.weight) : undefined,
            option_value_ids: ids
          })
        }

        // Loop and update existing variants
        const existingVariants = variantRows.filter(row => row.isExisting)
        for (let i = 0; i < existingVariants.length; i++) {
          const row = existingVariants[i]
          setSaveProgress(`Đang cập nhật biến thể (${i + 1}/${existingVariants.length}): ${row.name}`)
          await productAPI.adminUpdateVariant(row.id, {
            name: row.name,
            sku: row.sku,
            price: Number(row.price),
            price_base: row.priceBase ? Number(row.priceBase) : undefined,
            weight: row.weight ? Number(row.weight) : undefined
          })
        }

        // Delete variants that were removed
        if (deletedVariantIds.length > 0) {
          setSaveProgress('Đang xóa các biến thể bị loại bỏ...')
          for (const id of deletedVariantIds) {
            await productAPI.adminDeleteVariant(id)
          }
        }
      }

      alert('Đã lưu thông tin sản phẩm và các biến thể thành công!')
      handleCancelEditProduct()
      void loadProducts(productPage, productSearch)
    } catch (err: any) {
      console.error(err)
      alert(err.message || 'Lỗi khi lưu sản phẩm và biến thể')
    } finally {
      setLoading(false)
      setSaveProgress(null)
    }
  }

  const handleEditProduct = async (p: Product) => {
    try {
      setLoading(true)
      const details = await productAPI.getProductById(p.id)
      setEditProductId(p.id)
      setProductForm({
        id: details.id,
        categoryId: String(details.category_id || ''),
        brandId: String(details.brand_id || ''),
        name: details.name,
        metaTitle: details.meta_title || '',
        metaDescription: details.meta_description || '',
        imgThumb: details.image || '',
        images: details.images || [],
        weight: details.weight || 0,
        lowStockThreshold: details.low_stock_threshold || 5,
        specs: (details.specifications || []).map((s: any) => ({
          group: s.group,
          key: s.key,
          value: s.value,
          sort_order: s.sort_order,
        })),
      })

      // Load existing attributes/options and variants
      if (details.variants && details.variants.length > 0) {
        setHasVariants(true)
        
        // Construct attributes (options) state from variant options
        const extractedOpts: { [name: string]: Set<string> } = {}
        details.variants.forEach((v: any) => {
          if (v.options) {
            v.options.forEach((opt: any) => {
              if (!extractedOpts[opt.option_type_name]) {
                extractedOpts[opt.option_type_name] = new Set()
              }
              extractedOpts[opt.option_type_name].add(opt.value)
            })
          }
        })

        const mappedOptions = Object.entries(extractedOpts).map(([name, valuesSet]) => ({
          name,
          values: Array.from(valuesSet)
        }))
        setOptions(mappedOptions)

        // Map variants
        const mappedRows = details.variants.map((v: any) => {
          const attrs: { [key: string]: string } = {}
          if (v.options) {
            v.options.forEach((opt: any) => {
              attrs[opt.option_type_name] = opt.value
            })
          }
          return {
            id: v.id,
            attributes: attrs,
            name: v.name,
            sku: v.sku,
            price: v.price,
            priceBase: v.price_base || '',
            weight: v.weight || 0,
            isActive: v.is_active,
            isExisting: true,
            image: ''
          }
        })
        setVariantRows(mappedRows)
      } else {
        setHasVariants(false)
        setOptions([])
        setVariantRows([])
      }

      setShowProductForm(true)
    } catch (err: any) {
      alert(err.message || 'Lỗi khi tải chi tiết sản phẩm')
    } finally {
      setLoading(false)
    }
  }

  const handleCancelEditProduct = () => {
    setEditProductId(null)
    setProductForm({
      id: '',
      categoryId: '',
      brandId: '',
      name: '',
      metaTitle: '',
      metaDescription: '',
      imgThumb: '',
      images: [],
      weight: 0,
      lowStockThreshold: 5,
      specs: [],
    })
    setHasVariants(false)
    setOptions([])
    setVariantRows([])
    setShowProductForm(false)
    setBulkPrice('')
    setBulkPriceBase('')
    setBulkWeight('')
    setDeletedVariantIds([])
  }

  const handleDeleteProduct = async (id: string, name: string) => {
    if (!confirm(`Bạn chắc chắn muốn xóa sản phẩm "${name}"? (Xóa mềm)`)) return
    try {
      setLoading(true)
      await productAPI.adminDeleteProduct(id)
      alert('Đã xóa sản phẩm thành công!')
      void loadProducts(productPage, productSearch)
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xóa sản phẩm')
    } finally {
      setLoading(false)
    }
  }

  // --------------------------------------------------
  // Category Actions
  // --------------------------------------------------
  const handleEditCategory = (c: Category) => {
    setEditCategoryId(c.id)
    setCategoryForm({
      name: c.name,
      parentId: c.parent_id ? String(c.parent_id) : '',
      sortOrder: c.sort_order,
    })
  }

  const handleCancelEditCategory = () => {
    setEditCategoryId(null)
    setCategoryForm({ name: '', parentId: '', sortOrder: 0 })
  }

  const handleCreateOrUpdateCategory = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!categoryForm.name) return
    try {
      setLoading(true)
      const payload = {
        name: categoryForm.name.trim(),
        parent_id: categoryForm.parentId ? Number(categoryForm.parentId) : null,
        sort_order: Number(categoryForm.sortOrder),
      }

      if (editCategoryId !== null) {
        await productAPI.adminUpdateCategory(editCategoryId, payload)
        alert('Cập nhật danh mục thành công!')
      } else {
        await productAPI.adminCreateCategory(payload)
        alert('Tạo danh mục thành công!')
      }
      handleCancelEditCategory()
      await reloadLookups()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi lưu danh mục')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteCategory = async (id: number, name: string) => {
    if (!confirm(`Bạn chắc chắn muốn xóa danh mục "${name}"?`)) return
    try {
      setLoading(true)
      await productAPI.adminDeleteCategory(id)
      alert('Đã xóa danh mục thành công!')
      await reloadLookups()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xóa danh mục')
    } finally {
      setLoading(false)
    }
  }

  // --------------------------------------------------
  // Brand Actions
  // --------------------------------------------------
  const handleEditBrand = (b: Brand) => {
    setEditBrandId(b.id)
    setBrandForm({
      name: b.name,
      logoUrl: b.logo || '',
      isActive: b.is_active,
    })
  }

  const handleCancelEditBrand = () => {
    setEditBrandId(null)
    setBrandForm({ name: '', logoUrl: '', isActive: true })
  }

  const handleCreateOrUpdateBrand = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!brandForm.name) return
    try {
      setLoading(true)
      const payload = {
        name: brandForm.name.trim(),
        logo_url: brandForm.logoUrl.trim() ? brandForm.logoUrl.trim() : null,
        is_active: brandForm.isActive,
      }

      if (editBrandId !== null) {
        await productAPI.adminUpdateBrand(editBrandId, payload)
        alert('Cập nhật thương hiệu thành công!')
      } else {
        await productAPI.adminCreateBrand(payload)
        alert('Tạo thương hiệu thành công!')
      }
      handleCancelEditBrand()
      await reloadLookups()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi lưu thương hiệu')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteBrand = async (id: number, name: string) => {
    if (!confirm(`Bạn chắc chắn muốn xóa thương hiệu "${name}"?`)) return
    try {
      setLoading(true)
      await productAPI.adminDeleteBrand(id)
      alert('Đã xóa thương hiệu thành công!')
      await reloadLookups()
    } catch (err: any) {
      alert(err.message || 'Lỗi khi xóa thương hiệu')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Hidden file input for variant image uploads */}
      <input
        type="file"
        ref={variantFileInputRef}
        onChange={handleVariantImageChange}
        className="hidden"
        accept="image/*"
      />

      {/* Hidden file input for secondary image uploads */}
      <input
        type="file"
        ref={secondaryFileInputRef}
        onChange={async (e) => {
          if (!e.target.files || e.target.files.length === 0) return
          const files = Array.from(e.target.files)
          try {
            setUploadingSecondary(true)
            const uploadPromises = files.map(file => uploadAPI.uploadImage(file))
            const results = await Promise.all(uploadPromises)
            const urls = results.map(res => res.url)
            setProductForm(p => ({
              ...p,
              images: [...(p.images || []), ...urls]
            }))
          } catch (err: any) {
            alert(err.message || 'Lỗi khi tải ảnh lên')
          } finally {
            setUploadingSecondary(false)
            if (secondaryFileInputRef.current) secondaryFileInputRef.current.value = ''
          }
        }}
        className="hidden"
        accept="image/*"
        multiple
      />

      {/* Dynamic Save Progress Modal Overlay */}
      {saveProgress && (
        <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white border border-neutral-250 p-6 rounded-lg max-w-sm w-full text-center space-y-4 shadow-xl">
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
            </div>
            <p className="text-xs font-black uppercase tracking-wider text-neutral-800">Đang lưu sản phẩm...</p>
            <p className="text-[11px] text-neutral-500">{saveProgress}</p>
          </div>
        </div>
      )}

      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-xl font-black text-neutral-900 uppercase tracking-tight">Quản lý Danh Mục & Sản Phẩm</h1>
          <p className="text-xs text-neutral-500 mt-1">Quản lý, chỉnh sửa và đăng ký sản phẩm mới, danh mục, thương hiệu hệ thống</p>
        </div>

        {subTab === 'product' && !showProductForm && (
          <button
            onClick={() => {
              handleCancelEditProduct()
              setShowProductForm(true)
            }}
            className="bg-black hover:bg-neutral-850 text-white text-xs font-black uppercase tracking-wider px-5 py-2.5 rounded transition-all shadow-sm flex items-center gap-1.5"
          >
            <span>+ Thêm sản phẩm mới</span>
          </button>
        )}
      </div>

      {/* Sub tabs selectors */}
      <div className="flex border-b border-neutral-200 text-xs font-semibold gap-1">
        {[
          { id: 'product', label: 'Sản phẩm' },
          { id: 'category', label: 'Danh mục' },
          { id: 'brand', label: 'Thương hiệu' },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => {
              setSubTab(tab.id as any)
              setShowProductForm(false)
            }}
            className={`px-4 py-2 border-b-2 transition-all ${
              subTab === tab.id
                ? 'border-black text-black font-black'
                : 'border-transparent text-neutral-500 hover:text-black'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* --------------------------------------------------
          1. PRODUCTS SUBTAB
          -------------------------------------------------- */}
      {subTab === 'product' && (
        <div className="space-y-6">
          {/* A. PRODUCT LIST VIEW */}
          {!showProductForm ? (
            <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4">
              <div className="flex justify-between items-center border-b border-neutral-100 pb-3">
                <span className="text-[10px] text-neutral-450 uppercase font-black tracking-wider">
                  Tổng số: <span className="text-neutral-800 font-bold font-mono">{productTotal}</span> sản phẩm
                </span>
                <form onSubmit={handleProductSearchSubmit} className="flex gap-2 max-w-sm w-full">
                  <input
                    type="text"
                    placeholder="Tìm kiếm tên sản phẩm..."
                    className="flex-1 border border-neutral-250 rounded px-3 py-1.5 text-xs bg-white focus:outline-none"
                    value={productSearch}
                    onChange={(e) => setProductSearch(e.target.value)}
                  />
                  <button
                    type="submit"
                    className="bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase px-4 py-1.5 rounded transition-all"
                  >
                    Tìm
                  </button>
                  {productSearch && (
                    <button
                      type="button"
                      onClick={() => {
                        setProductSearch('')
                        void loadProducts(1, '')
                      }}
                      className="border border-neutral-250 hover:bg-neutral-100 text-[10px] px-3 py-1.5 rounded"
                    >
                      Xóa
                    </button>
                  )}
                </form>
              </div>

              {productsLoading ? (
                <div className="flex justify-center py-12">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-black"></div>
                </div>
              ) : products.length === 0 ? (
                <div className="text-center py-12 border-2 border-dashed border-neutral-200 rounded-lg bg-neutral-50 flex flex-col items-center justify-center">
                  <svg className="w-10 h-10 text-neutral-300 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                  </svg>
                  <p className="text-xs text-neutral-500 font-semibold uppercase tracking-wider">Chưa có sản phẩm nào</p>
                  <button
                    onClick={() => setShowProductForm(true)}
                    className="mt-3 bg-black text-white text-[9px] font-black uppercase tracking-wider px-4 py-2 rounded hover:bg-neutral-850"
                  >
                    + Tạo sản phẩm đầu tiên
                  </button>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs border-collapse">
                    <thead>
                      <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                        <th className="p-3">Sản phẩm</th>
                        <th className="p-3">Danh mục</th>
                        <th className="p-3">Thương hiệu</th>
                        <th className="p-3 text-right">Giá bán</th>
                        <th className="p-3 text-center">Hành động</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-neutral-150">
                      {products.map((p) => (
                        <tr key={p.id} className="hover:bg-neutral-50/50 transition-colors">
                          <td className="p-3 flex gap-3 items-center">
                            <img
                              src={p.image}
                              alt={p.name}
                              className="w-12 h-12 object-cover rounded bg-neutral-100 border border-neutral-250"
                              onError={(e) => {
                                (e.target as HTMLImageElement).src = '/placeholder-product.png'
                              }}
                            />
                            <div>
                              <p className="font-bold text-neutral-850 text-xs leading-tight">{p.name}</p>
                              <p className="text-[9px] text-neutral-400 font-mono mt-1 uppercase tracking-wide">ID: {p.id}</p>
                            </div>
                          </td>
                          <td className="p-3 text-neutral-600 font-medium">
                            {categories.find(c => c.id === p.category_id)?.name || `Category #${p.category_id}`}
                          </td>
                          <td className="p-3 text-neutral-600 font-medium">
                            {p.brand?.name || `Brand #${p.brand_id}`}
                          </td>
                          <td className="p-3 text-right font-mono font-black text-neutral-800">
                            {p.price.toLocaleString('vi-VN')} đ
                          </td>
                          <td className="p-3">
                            <div className="flex gap-2 justify-center">
                              <button
                                onClick={() => void handleEditProduct(p)}
                                className="border border-neutral-350 hover:bg-neutral-100 text-[10px] font-black uppercase tracking-wider px-3 py-1 rounded transition-colors"
                              >
                                Sửa
                              </button>
                              <button
                                onClick={() => void handleDeleteProduct(p.id, p.name)}
                                className="text-red-655 hover:bg-red-50 text-[10px] font-black uppercase tracking-wider px-3 py-1 rounded transition-colors border border-transparent hover:border-red-200"
                              >
                                Xóa
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Product pagination */}
              {productTotalPages > 1 && (
                <div className="flex gap-2 justify-end text-[10px] font-bold pt-2 border-t border-neutral-150">
                  <button
                    disabled={productPage <= 1 || productsLoading}
                    onClick={() => void loadProducts(productPage - 1, productSearch)}
                    className="border border-neutral-350 px-3 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    ← Trước
                  </button>
                  <span className="px-3 py-1 flex items-center uppercase tracking-wider text-neutral-500">Trang {productPage} / {productTotalPages}</span>
                  <button
                    disabled={productPage >= productTotalPages || productsLoading}
                    onClick={() => void loadProducts(productPage + 1, productSearch)}
                    className="border border-neutral-350 px-3 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    Sau →
                  </button>
                </div>
              )}
            </div>
          ) : (
            /* B. PRODUCT & VARIANTS UNIFIED CREATION WORKSPACE (Shopify style) */
            <form onSubmit={handleSaveProductWorkspace} className="max-w-5xl mx-auto space-y-6">
              <div className="flex justify-between items-center bg-white border border-neutral-250 p-4 rounded-lg">
                <div>
                  <h3 className="text-xs font-black uppercase tracking-wide text-neutral-800">
                    {editProductId ? 'Chỉnh Sửa Sản Phẩm & Biến Thể' : 'Tạo Sản Phẩm Mới'}
                  </h3>
                  <p className="text-[10px] text-neutral-450 mt-0.5">Thiết lập các thông tin chi tiết, thuộc tính bán hàng và biến thể</p>
                </div>
                <button
                  type="button"
                  onClick={handleCancelEditProduct}
                  className="bg-neutral-100 hover:bg-neutral-200 text-neutral-800 text-[10px] font-black uppercase tracking-wider px-4 py-2 rounded transition-all"
                >
                  ✕ Hủy bỏ & Quay lại
                </button>
              </div>

              {/* TWO TOP COLUMNS: Balanced layout heights */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch">
                {/* Left Column: Product Information & Specifications */}
                <div className="lg:col-span-2 space-y-6 flex flex-col justify-between">
                  {/* Card: Basic Information */}
                  <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4 flex-1">
                    <h4 className="text-[10px] font-black uppercase tracking-wider text-neutral-800 border-b border-neutral-100 pb-2">Thông tin cơ bản</h4>
                    
                    <div className="grid grid-cols-3 gap-4">
                      <div className="col-span-1">
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Mã sản phẩm *</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: ip15promax"
                          className="border border-neutral-300 rounded px-3 py-2 w-full font-mono font-bold uppercase disabled:bg-neutral-100 disabled:text-neutral-500 focus:border-black focus:outline-none text-xs"
                          value={productForm.id}
                          onChange={(e) => setProductForm(p => ({ ...p, id: e.target.value }))}
                          disabled={editProductId !== null}
                          required
                        />
                      </div>
                      <div className="col-span-2">
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tên sản phẩm gốc *</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: iPhone 15 Pro Max..."
                          className="border border-neutral-300 rounded px-3 py-2 w-full font-bold focus:border-black focus:outline-none text-xs"
                          value={productForm.name}
                          onChange={(e) => setProductForm(p => ({ ...p, name: e.target.value }))}
                          required
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Danh mục *</label>
                        <select
                          className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white font-medium text-xs focus:border-black focus:outline-none"
                          value={productForm.categoryId}
                          onChange={(e) => setProductForm(p => ({ ...p, categoryId: e.target.value }))}
                          required
                        >
                          <option value="">Chọn danh mục</option>
                          {categories.map(c => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                          ))}
                        </select>
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Thương hiệu *</label>
                        <select
                          className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white font-medium text-xs focus:border-black focus:outline-none"
                          value={productForm.brandId}
                          onChange={(e) => setProductForm(p => ({ ...p, brandId: e.target.value }))}
                          required
                        >
                          <option value="">Chọn thương hiệu</option>
                          {brands.map(b => (
                            <option key={b.id} value={b.id}>{b.name}</option>
                          ))}
                        </select>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Trọng lượng (g)</label>
                        <input
                          type="number"
                          placeholder="Trọng lượng sản phẩm..."
                          className="border border-neutral-300 rounded px-3 py-2 w-full text-xs focus:border-black focus:outline-none"
                          value={productForm.weight || ''}
                          onChange={(e) => setProductForm(p => ({ ...p, weight: Number(e.target.value) }))}
                        />
                      </div>
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Cảnh báo tồn kho thấp</label>
                        <input
                          type="number"
                          placeholder="Số lượng cảnh báo hết hàng..."
                          className="border border-neutral-300 rounded px-3 py-2 w-full text-xs focus:border-black focus:outline-none"
                          value={productForm.lowStockThreshold}
                          onChange={(e) => setProductForm(p => ({ ...p, lowStockThreshold: Number(e.target.value) }))}
                        />
                      </div>
                    </div>
                  </div>

                  {/* Card: Dynamic Specifications (Specs) */}
                  <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4 mt-6">
                    <h4 className="text-[10px] font-black uppercase tracking-wider text-neutral-800 border-b border-neutral-100 pb-2">Thông số kỹ thuật (Specs)</h4>
                    
                    {productForm.specs.length > 0 && (
                      <div className="border border-neutral-200 rounded overflow-hidden">
                        <table className="w-full text-left text-[11px] border-collapse">
                          <thead>
                            <tr className="bg-neutral-50 border-b border-neutral-200 font-bold uppercase text-[9px] tracking-wider text-neutral-500">
                              <th className="p-2">Nhóm</th>
                              <th className="p-2">Tên thông số</th>
                              <th className="p-2">Giá trị</th>
                              <th className="p-2 text-center">Xóa</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-neutral-100 font-medium">
                            {productForm.specs.map((spec, i) => (
                              <tr key={i} className="hover:bg-neutral-50">
                                <td className="p-2 text-neutral-800 font-bold">{spec.group}</td>
                                <td className="p-2 text-neutral-600">{spec.key}</td>
                                <td className="p-2 font-semibold text-neutral-800">{spec.value}</td>
                                <td className="p-2 text-center">
                                  <button
                                    type="button"
                                    onClick={() => handleRemoveSpec(i)}
                                    className="text-red-500 hover:text-red-700 font-bold text-xs"
                                  >
                                    ✕
                                  </button>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}

                    <div className="grid grid-cols-3 gap-2 bg-neutral-50 p-3 rounded-lg border border-neutral-200">
                      <div>
                        <label className="block text-[9px] uppercase font-bold text-neutral-450 mb-0.5">Nhóm</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: Màn hình, Pin"
                          className="border border-neutral-300 rounded px-2.5 py-1.5 w-full text-[10px] focus:outline-none focus:border-black"
                          value={newSpec.group}
                          onChange={(e) => setNewSpec(p => ({ ...p, group: e.target.value }))}
                        />
                      </div>
                      <div>
                        <label className="block text-[9px] uppercase font-bold text-neutral-450 mb-0.5">Tên thông số</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: Công nghệ màn hình"
                          className="border border-neutral-300 rounded px-2.5 py-1.5 w-full text-[10px] focus:outline-none focus:border-black"
                          value={newSpec.key}
                          onChange={(e) => setNewSpec(p => ({ ...p, key: e.target.value }))}
                        />
                      </div>
                      <div>
                        <label className="block text-[9px] uppercase font-bold text-neutral-450 mb-0.5">Giá trị</label>
                        <input
                          type="text"
                          placeholder="Ví dụ: OLED, 120Hz"
                          className="border border-neutral-300 rounded px-2.5 py-1.5 w-full text-[10px] focus:outline-none focus:border-black"
                          value={newSpec.value}
                          onChange={(e) => setNewSpec(p => ({ ...p, value: e.target.value }))}
                        />
                      </div>
                      <button
                        type="button"
                        onClick={handleAddSpec}
                        className="col-span-3 bg-neutral-800 text-white font-black uppercase py-2 text-[9px] rounded hover:bg-black mt-2 transition-colors tracking-wider"
                      >
                        + Thêm thông số vào danh sách
                      </button>
                    </div>
                  </div>
                </div>

                {/* Right Column: Media Upload, SEO, Submissions */}
                <div className="space-y-6 lg:col-span-1 flex flex-col justify-between">
                  <div className="space-y-6 flex-1">
                    {/* Card: Images */}
                    <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4">
                      <h4 className="text-[10px] font-black uppercase tracking-wider text-neutral-800 border-b border-neutral-100 pb-2">Hình ảnh sản phẩm</h4>
                      <ImageUploader
                        label="Ảnh đại diện (Thumbnail)"
                        value={productForm.imgThumb}
                        onChange={(url) => setProductForm(p => ({ ...p, imgThumb: url }))}
                        placeholder="Kéo thả ảnh sản phẩm hoặc nhấp để chọn"
                      />

                      {/* Secondary Images Upload Grid */}
                      <div className="space-y-2 pt-2 border-t border-neutral-100">
                        <label className="block text-[10px] uppercase font-bold text-neutral-450">
                          Ảnh chi tiết khác (Thư viện ảnh)
                        </label>
                        <div className="grid grid-cols-4 gap-2">
                          {(productForm.images || []).map((imgUrl, idx) => (
                            <div key={idx} className="relative aspect-square rounded border border-neutral-200 overflow-hidden bg-neutral-50 group">
                              <img
                                src={imgUrl}
                                alt={`sub-${idx}`}
                                className="w-full h-full object-cover"
                              />
                              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                                <button
                                  type="button"
                                  onClick={() => {
                                    setProductForm(p => ({
                                      ...p,
                                      images: (p.images || []).filter((_, i) => i !== idx)
                                    }))
                                  }}
                                  className="bg-red-650 hover:bg-red-700 text-white text-[8px] font-bold uppercase px-2 py-1 rounded transition-colors shadow"
                                >
                                  Xóa
                                </button>
                              </div>
                            </div>
                          ))}

                          {uploadingSecondary ? (
                            <div className="aspect-square border-2 border-dashed border-neutral-250 rounded flex flex-col items-center justify-center bg-neutral-50">
                              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-black"></div>
                            </div>
                          ) : (
                            <button
                              type="button"
                              onClick={() => secondaryFileInputRef.current?.click()}
                              className="aspect-square border-2 border-dashed border-neutral-250 hover:border-neutral-400 rounded flex flex-col items-center justify-center text-neutral-400 hover:text-neutral-600 transition-all bg-white"
                            >
                              <svg className="w-5 h-5 mb-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
                              </svg>
                              <span className="text-[8px] font-bold uppercase tracking-wider">Thêm ảnh</span>
                            </button>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Card: SEO / Meta tags */}
                    <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4">
                      <h4 className="text-[10px] font-black uppercase tracking-wider text-neutral-800 border-b border-neutral-100 pb-2">Thiết lập SEO</h4>
                      
                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Meta Title</label>
                        <input
                          type="text"
                          placeholder="SEO Meta Title..."
                          className="border border-neutral-300 rounded px-3 py-2 w-full text-xs focus:outline-none focus:border-black"
                          value={productForm.metaTitle}
                          onChange={(e) => setProductForm(p => ({ ...p, metaTitle: e.target.value }))}
                        />
                      </div>

                      <div>
                        <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Meta Description</label>
                        <textarea
                          rows={3}
                          placeholder="SEO Meta Description..."
                          className="border border-neutral-300 rounded px-3 py-2 w-full text-xs focus:outline-none focus:border-black resize-none"
                          value={productForm.metaDescription}
                          onChange={(e) => setProductForm(p => ({ ...p, metaDescription: e.target.value }))}
                        />
                      </div>
                    </div>
                  </div>

                  {/* Submission triggers */}
                  <div className="bg-neutral-50 border border-neutral-250 p-4 rounded-lg space-y-3 mt-6">
                    <button
                      type="submit"
                      disabled={loading}
                      className="w-full bg-black hover:bg-neutral-850 text-white text-[11px] font-black uppercase tracking-wider py-3 rounded transition-colors shadow-sm disabled:opacity-50"
                    >
                      {editProductId ? '✓ Lưu tất cả thay đổi' : '✓ Lưu sản phẩm gốc & biến thể'}
                    </button>
                    <button
                      type="button"
                      onClick={handleCancelEditProduct}
                      className="w-full bg-white hover:bg-neutral-100 border border-neutral-300 text-neutral-700 text-[10px] font-black uppercase py-2.5 rounded transition-colors"
                    >
                      Hủy bỏ
                    </button>
                  </div>
                </div>
              </div>

              {/* DYNAMIC VARIANTS PANEL: Full-width layout spanning below upper columns */}
              <div className="bg-white border border-neutral-200 rounded-lg p-5 space-y-4 mt-6">
                <div className="flex justify-between items-center border-b border-neutral-100 pb-2">
                  <h4 className="text-[10px] font-black uppercase tracking-wider text-neutral-800">Biến thể của sản phẩm</h4>
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      className="accent-black h-4 w-4"
                      checked={hasVariants}
                      onChange={(e) => {
                        setHasVariants(e.target.checked)
                        if (e.target.checked && options.length === 0) {
                          setOptions([{ name: 'Màu sắc', values: [] }])
                        }
                      }}
                    />
                    <span className="text-[10px] font-bold uppercase text-neutral-700">Sản phẩm này có nhiều biến thể</span>
                  </label>
                </div>

                {hasVariants && (
                  <div className="space-y-6">
                    {/* Attributes Setup Layout */}
                    <div className="space-y-4">
                      <p className="text-[10px] font-black uppercase tracking-wider text-neutral-450">Thiết lập các thuộc tính & lựa chọn</p>
                      
                      {options.map((opt, optIndex) => (
                        <div key={optIndex} className="bg-neutral-50 p-4 border border-neutral-200 rounded-lg relative space-y-3">
                          <button
                            type="button"
                            onClick={() => handleRemoveAttributeType(optIndex)}
                            className="absolute top-2.5 right-3 text-neutral-400 hover:text-red-650 transition-colors text-xs font-bold"
                          >
                            ✕ Gỡ thuộc tính
                          </button>

                          <div className="max-w-xs">
                            <label className="block text-[9px] uppercase font-bold text-neutral-450 mb-1">Tên thuộc tính</label>
                            <input
                              type="text"
                              className="border border-neutral-300 rounded px-2.5 py-1.5 w-full text-xs font-bold uppercase focus:outline-none focus:border-black bg-white"
                              value={opt.name}
                              onChange={(e) => {
                                const updated = [...options]
                                updated[optIndex] = { ...opt, name: e.target.value }
                                setOptions(updated)
                              }}
                              placeholder="Màu sắc, Dung lượng, RAM..."
                            />
                          </div>

                          <div>
                            <label className="block text-[9px] uppercase font-bold text-neutral-450 mb-1">Các lựa chọn giá trị (Cách nhau bằng dấu phẩy)</label>
                            <div className="flex flex-wrap gap-1.5 items-center bg-white border border-neutral-300 rounded p-2 min-h-[40px]">
                              {opt.values.map((val, valIndex) => (
                                <span key={valIndex} className="text-[10px] bg-neutral-100 border border-neutral-250 px-2 py-0.5 rounded flex items-center gap-1 font-bold">
                                  <span>{val}</span>
                                  <button
                                    type="button"
                                    onClick={() => handleRemoveAttributeValue(optIndex, valIndex)}
                                    className="text-neutral-400 hover:text-black font-bold"
                                  >
                                    ✕
                                  </button>
                                </span>
                              ))}
                              <input
                                type="text"
                                placeholder="đỏ, xanh, trắng..."
                                className="border-none text-xs focus:outline-none flex-1 min-w-[120px] py-0.5"
                                value={tempAttrValues[optIndex] || ''}
                                onChange={(e) => {
                                  const text = e.target.value
                                  if (text.endsWith(',')) {
                                    handleAddAttributeValue(optIndex, text.slice(0, -1))
                                  } else {
                                    setTempAttrValues({ ...tempAttrValues, [optIndex]: text })
                                  }
                                }}
                                onBlur={() => handleAddAttributeValue(optIndex)}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter') {
                                    e.preventDefault()
                                    handleAddAttributeValue(optIndex)
                                  }
                                }}
                              />
                            </div>
                            <p className="text-[9px] text-neutral-400 mt-1">Gõ giá trị và nhấn **Enter** hoặc phẩy (,) để tạo nhãn lựa chọn.</p>
                          </div>
                        </div>
                      ))}

                      <div className="flex gap-2 items-center bg-white p-3 rounded-lg border border-neutral-200">
                        <span className="text-[10px] font-bold uppercase text-neutral-400">Thêm thuộc tính khác:</span>
                        <input
                          type="text"
                          placeholder="Ví dụ: RAM, Kích thước..."
                          className="border border-neutral-300 rounded px-3 py-2 text-xs font-semibold focus:outline-none focus:border-black max-w-xs bg-white"
                          value={newAttributeName}
                          onChange={(e) => setNewAttributeName(e.target.value)}
                        />
                        <button
                          type="button"
                          onClick={handleAddAttributeType}
                          className="bg-black hover:bg-neutral-800 text-white text-[10px] font-black uppercase px-4 py-2.5 rounded transition-all tracking-wider shadow-sm"
                        >
                          + Thêm nhóm thuộc tính
                        </button>
                      </div>
                    </div>

                    {/* Generate Action Button */}
                    <div className="border-t border-neutral-150 pt-4 flex justify-between items-center bg-neutral-50 p-3 rounded-lg">
                      <p className="text-[9px] text-neutral-450 italic">Hệ thống sẽ tự động ghép các lựa chọn để sinh ra các biến thể tương ứng.</p>
                      <button
                        type="button"
                        onClick={handleAutoGenerateVariants}
                        className="bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase tracking-wider px-5 py-2.5 rounded transition-all shadow-sm"
                      >
                        ⚡ Sinh các biến thể tự động
                      </button>
                    </div>

                    {/* Variants List Workspace */}
                    {variantRows.length > 0 && (
                      <div className="space-y-4">
                        <div className="border-t border-neutral-200 pt-4 flex flex-col md:flex-row md:items-center justify-between gap-4">
                          <p className="text-[10px] font-black uppercase tracking-wider text-neutral-800">Danh sách biến thể ({variantRows.length})</p>
                          
                          {/* Premium BULK EDIT Controls */}
                          <div className="flex flex-wrap items-center gap-3 bg-neutral-50 p-2.5 rounded-lg border border-neutral-200 text-xs">
                            <span className="text-[10px] font-black uppercase text-neutral-450">Sửa hàng loạt:</span>
                            <div className="flex gap-1">
                              <input
                                type="number"
                                placeholder="Giá bán..."
                                className="border border-neutral-300 rounded px-2 py-1 w-24 text-[10px] font-mono focus:outline-none focus:border-black"
                                value={bulkPrice}
                                onChange={(e) => setBulkPrice(e.target.value)}
                              />
                              <button
                                type="button"
                                onClick={handleBulkApplyPrice}
                                className="bg-neutral-800 hover:bg-black text-white text-[9px] px-2 py-1 rounded font-bold uppercase transition-colors"
                              >
                                Áp dụng
                              </button>
                            </div>

                            <div className="flex gap-1">
                              <input
                                type="number"
                                placeholder="Giá gốc..."
                                className="border border-neutral-300 rounded px-2 py-1 w-24 text-[10px] font-mono focus:outline-none focus:border-black"
                                value={bulkPriceBase}
                                onChange={(e) => setBulkPriceBase(e.target.value)}
                              />
                              <button
                                type="button"
                                onClick={handleBulkApplyPriceBase}
                                className="bg-neutral-800 hover:bg-black text-white text-[9px] px-2 py-1 rounded font-bold uppercase transition-colors"
                              >
                                Áp dụng
                              </button>
                            </div>

                            <div className="flex gap-1">
                              <input
                                type="number"
                                placeholder="Nặng (g)..."
                                className="border border-neutral-300 rounded px-2 py-1 w-20 text-[10px] font-mono focus:outline-none focus:border-black"
                                value={bulkWeight}
                                onChange={(e) => setBulkWeight(e.target.value)}
                              />
                              <button
                                type="button"
                                onClick={handleBulkApplyWeight}
                                className="bg-neutral-800 hover:bg-black text-white text-[9px] px-2 py-1 rounded font-bold uppercase transition-colors"
                              >
                                Áp dụng
                              </button>
                            </div>
                          </div>
                        </div>

                        {/* Variants matrix table */}
                        <div className="overflow-x-auto border border-neutral-200 rounded-lg">
                          <table className="w-full text-left text-xs border-collapse">
                            <thead>
                              <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                                <th className="p-3 text-center w-14">Ảnh</th>
                                <th className="p-3">Thuộc tính</th>
                                <th className="p-3 w-1/4">Tên biến thể *</th>
                                <th className="p-3 font-mono">SKU *</th>
                                <th className="p-3 text-right">Giá bán *</th>
                                <th className="p-3 text-right">Giá gốc</th>
                                <th className="p-3 text-right">Cân nặng (g)</th>
                                <th className="p-3 text-center w-12">Xóa</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-neutral-150 font-medium">
                              {variantRows.map((row, index) => {
                                const attrString = Object.entries(row.attributes)
                                  .map(([k, v]) => `${k}: ${v as string}`)
                                  .join(' / ')
                                return (
                                  <tr key={index} className="hover:bg-neutral-50/50">
                                    {/* Variant specific inline image uploader */}
                                    <td className="p-3 text-center">
                                      <div
                                        onClick={() => triggerVariantImageUpload(index)}
                                        className="w-10 h-10 border border-dashed border-neutral-300 hover:border-neutral-500 rounded bg-neutral-50 flex items-center justify-center cursor-pointer overflow-hidden transition-all relative group"
                                      >
                                        {row.image ? (
                                          <>
                                            <img src={row.image} alt="v" className="w-full h-full object-cover" />
                                            <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                                              <span className="text-[8px] text-white font-bold uppercase">Sửa</span>
                                            </div>
                                          </>
                                        ) : (
                                          <svg className="w-4 h-4 text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" />
                                          </svg>
                                        )}
                                      </div>
                                    </td>
                                    <td className="p-3">
                                      <span className="bg-neutral-100 text-neutral-800 border border-neutral-200 px-2 py-0.5 rounded text-[10px] font-bold">
                                        {attrString}
                                      </span>
                                    </td>
                                    <td className="p-3 w-1/4">
                                      <input
                                        type="text"
                                        className="border border-neutral-300 rounded px-2 py-1 w-full text-xs font-semibold"
                                        value={row.name}
                                        onChange={(e) => {
                                          const updated = [...variantRows]
                                          updated[index].name = e.target.value
                                          setVariantRows(updated)
                                        }}
                                        required
                                      />
                                    </td>
                                    <td className="p-3">
                                      <input
                                        type="text"
                                        className="border border-neutral-300 rounded px-2 py-1 w-28 font-mono text-[10px] uppercase font-bold"
                                        value={row.sku}
                                        onChange={(e) => {
                                          const updated = [...variantRows]
                                          updated[index].sku = e.target.value
                                          setVariantRows(updated)
                                        }}
                                        required
                                      />
                                    </td>
                                    <td className="p-3">
                                      <input
                                        type="number"
                                        className="border border-neutral-300 rounded px-2 py-1 w-24 text-right text-xs font-bold font-mono"
                                        value={row.price || ''}
                                        onChange={(e) => {
                                          const updated = [...variantRows]
                                          updated[index].price = Number(e.target.value)
                                          setVariantRows(updated)
                                        }}
                                        placeholder="đ"
                                        min={1}
                                        required
                                      />
                                    </td>
                                    <td className="p-3">
                                      <input
                                        type="number"
                                        className="border border-neutral-300 rounded px-2 py-1 w-24 text-right text-xs font-mono text-neutral-500"
                                        value={row.priceBase || ''}
                                        onChange={(e) => {
                                          const updated = [...variantRows]
                                          updated[index].priceBase = e.target.value
                                          setVariantRows(updated)
                                        }}
                                        placeholder="đ"
                                      />
                                    </td>
                                    <td className="p-3">
                                      <input
                                        type="number"
                                        className="border border-neutral-300 rounded px-2 py-1 w-20 text-right text-xs font-mono"
                                        value={row.weight || ''}
                                        onChange={(e) => {
                                          const updated = [...variantRows]
                                          updated[index].weight = Number(e.target.value)
                                          setVariantRows(updated)
                                        }}
                                        placeholder="g"
                                      />
                                    </td>
                                    <td className="p-3 text-center">
                                      <button
                                        type="button"
                                        onClick={() => {
                                          if (row.isExisting && row.id) {
                                            if (confirm(`Bạn chắc chắn muốn xóa biến thể "${row.name}" này? (Sẽ xóa vĩnh viễn khi bấm Lưu)`)) {
                                              setDeletedVariantIds(prev => [...prev, row.id])
                                              setVariantRows(prev => prev.filter((_, i) => i !== index))
                                            }
                                          } else {
                                            setVariantRows(prev => prev.filter((_, i) => i !== index))
                                          }
                                        }}
                                        className="text-red-500 hover:text-red-700 font-bold"
                                      >
                                        ✕
                                      </button>
                                    </td>
                                  </tr>
                                )
                              })}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            </form>
          )}
        </div>
      )}

      {/* --------------------------------------------------
          2. CATEGORIES SUBTAB
          -------------------------------------------------- */}
      {subTab === 'category' && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 items-start">
          <div className="md:col-span-2 bg-white border border-neutral-200 rounded-lg p-5">
            <h3 className="text-xs font-black uppercase tracking-wide mb-3 text-neutral-800">Danh sách danh mục hiện tại</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                    <th className="p-3">ID</th>
                    <th className="p-3">Tên danh mục</th>
                    <th className="p-3">Danh mục cha</th>
                    <th className="p-3 text-center">Sắp xếp</th>
                    <th className="p-3 text-center">Hành động</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-150">
                  {paginatedCategories.map((c) => {
                    const parent = categories.find(p => p.id === c.parent_id)
                    return (
                      <tr key={c.id} className="hover:bg-neutral-50 transition-colors">
                        <td className="p-3 font-mono font-bold text-neutral-500">{c.id}</td>
                        <td className="p-3 font-bold text-neutral-850">{c.name}</td>
                        <td className="p-3 text-neutral-500 font-medium">{parent ? parent.name : 'Không có'}</td>
                        <td className="p-3 text-center font-semibold text-neutral-600">{c.sort_order}</td>
                        <td className="p-3 text-center">
                          <div className="flex gap-2 justify-center">
                            <button
                              onClick={() => handleEditCategory(c)}
                              className="border border-neutral-350 hover:bg-neutral-100 text-[10px] font-bold px-2 py-1 rounded"
                            >
                              Sửa
                            </button>
                            <button
                              onClick={() => void handleDeleteCategory(c.id, c.name)}
                              className="text-red-655 hover:bg-red-50 text-[10px] font-bold px-2 py-1 rounded border border-transparent hover:border-red-200"
                            >
                              Xóa
                            </button>
                          </div>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>

            {totalCategoryPages > 1 && (
              <div className="flex gap-2 justify-between items-center text-[10px] font-semibold pt-3 border-t border-neutral-150 mt-2">
                <span className="text-neutral-500">
                  Hiển thị <span className="font-bold text-neutral-800">{paginatedCategories.length}</span> trên tổng số <span className="font-bold text-neutral-800">{categories.length}</span> danh mục
                </span>
                <div className="flex gap-1.5">
                  <button
                    disabled={safeCategoryPage <= 1}
                    onClick={() => setCategoryPage(p => p - 1)}
                    className="border border-neutral-350 px-2.5 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    ← Trước
                  </button>
                  <span className="px-2 py-1 flex items-center">Trang {safeCategoryPage} / {totalCategoryPages}</span>
                  <button
                    disabled={safeCategoryPage >= totalCategoryPages}
                    onClick={() => setCategoryPage(p => p + 1)}
                    className="border border-neutral-350 px-2.5 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    Sau →
                  </button>
                </div>
              </div>
            )}
          </div>

          <form onSubmit={handleCreateOrUpdateCategory} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs">
            <div className="flex justify-between items-center border-b border-neutral-100 pb-2">
              <h3 className="text-xs font-black uppercase tracking-wide">
                {editCategoryId ? 'Chỉnh sửa Danh mục' : 'Thêm Danh mục mới'}
              </h3>
              {editCategoryId && (
                <button
                  type="button"
                  onClick={handleCancelEditCategory}
                  className="text-neutral-400 hover:text-black font-bold uppercase text-[9px] tracking-wide"
                >
                  ✕ Hủy
                </button>
              )}
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tên danh mục *</label>
                <input
                  type="text"
                  placeholder="Ví dụ: Điện thoại, Laptop..."
                  className="w-full border border-neutral-300 rounded px-3 py-2 font-bold"
                  value={categoryForm.name}
                  onChange={(e) => setCategoryForm(p => ({ ...p, name: e.target.value }))}
                  required
                />
              </div>

              <div>
                <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Danh mục cha (Tùy chọn)</label>
                <select
                  className="w-full border border-neutral-300 rounded px-2.5 py-2 bg-white text-xs"
                  value={categoryForm.parentId}
                  onChange={(e) => setCategoryForm(p => ({ ...p, parentId: e.target.value }))}
                >
                  <option value="">Không có danh mục cha</option>
                  {categories
                    .filter(c => c.id !== editCategoryId)
                    .map(c => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                </select>
              </div>

              <div>
                <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Thứ tự sắp xếp</label>
                <input
                  type="number"
                  placeholder="Thứ tự hiển thị..."
                  className="w-full border border-neutral-300 rounded px-3 py-2 text-xs"
                  value={categoryForm.sortOrder}
                  onChange={(e) => setCategoryForm(p => ({ ...p, sortOrder: Number(e.target.value) }))}
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase tracking-wider py-2.5 rounded transition-colors disabled:opacity-50"
            >
              {editCategoryId ? 'Cập nhật danh mục' : 'Lưu danh mục'}
            </button>
          </form>
        </div>
      )}

      {/* --------------------------------------------------
          3. BRANDS SUBTAB
          -------------------------------------------------- */}
      {subTab === 'brand' && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 items-start">
          <div className="md:col-span-2 bg-white border border-neutral-200 rounded-lg p-5">
            <h3 className="text-xs font-black uppercase tracking-wide mb-3 text-neutral-800">Danh sách thương hiệu hiện tại</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="bg-neutral-50 border-b border-neutral-200 text-neutral-450 uppercase font-black text-[9px] tracking-wider">
                    <th className="p-3">Logo</th>
                    <th className="p-3">ID</th>
                    <th className="p-3">Tên thương hiệu</th>
                    <th className="p-3 text-center">Trạng thái</th>
                    <th className="p-3 text-center">Hành động</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-150">
                  {paginatedBrands.map((b) => (
                    <tr key={b.id} className="hover:bg-neutral-50 transition-colors">
                      <td className="p-3 w-16">
                        {b.logo ? (
                          <img
                            src={b.logo}
                            alt={b.name}
                            className="w-10 h-6 object-contain rounded bg-neutral-50 border border-neutral-100"
                            onError={(e) => {
                              (e.target as HTMLImageElement).style.display = 'none'
                            }}
                          />
                        ) : (
                          <span className="text-[10px] text-neutral-445 italic">No logo</span>
                        )}
                      </td>
                      <td className="p-3 font-mono font-bold text-neutral-500">{b.id}</td>
                      <td className="p-3 font-bold text-neutral-850">{b.name}</td>
                      <td className="p-3 text-center">
                        <span className={`px-2 py-0.5 rounded text-[9px] font-black uppercase ${b.is_active ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-red-50 text-red-700 border border-red-200'}`}>
                          {b.is_active ? 'Hoạt động' : 'Tạm ngưng'}
                        </span>
                      </td>
                      <td className="p-3 text-center">
                        <div className="flex gap-2 justify-center">
                          <button
                            onClick={() => handleEditBrand(b)}
                            className="border border-neutral-350 hover:bg-neutral-100 text-[10px] font-bold px-2 py-1 rounded"
                          >
                            Sửa
                          </button>
                          <button
                            onClick={() => void handleDeleteBrand(b.id, b.name)}
                            className="text-red-655 hover:bg-red-50 text-[10px] font-bold px-2 py-1 rounded border border-transparent hover:border-red-200"
                          >
                            Xóa
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {totalBrandPages > 1 && (
              <div className="flex gap-2 justify-between items-center text-[10px] font-semibold pt-3 border-t border-neutral-150 mt-2">
                <span className="text-neutral-500">
                  Hiển thị <span className="font-bold text-neutral-800">{paginatedBrands.length}</span> trên tổng số <span className="font-bold text-neutral-800">{brands.length}</span> thương hiệu
                </span>
                <div className="flex gap-1.5">
                  <button
                    disabled={safeBrandPage <= 1}
                    onClick={() => setBrandPage(p => p - 1)}
                    className="border border-neutral-350 px-2.5 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    ← Trước
                  </button>
                  <span className="px-2 py-1 flex items-center">Trang {safeBrandPage} / {totalBrandPages}</span>
                  <button
                    disabled={safeBrandPage >= totalBrandPages}
                    onClick={() => setBrandPage(p => p + 1)}
                    className="border border-neutral-350 px-2.5 py-1 bg-white hover:bg-neutral-50 rounded disabled:opacity-40"
                  >
                    Sau →
                  </button>
                </div>
              </div>
            )}
          </div>

          <form onSubmit={handleCreateOrUpdateBrand} className="bg-white border border-neutral-250 rounded-lg p-5 space-y-4 text-xs">
            <div className="flex justify-between items-center border-b border-neutral-100 pb-2">
              <h3 className="text-xs font-black uppercase tracking-wide">
                {editBrandId ? 'Chỉnh sửa Thương hiệu' : 'Thêm Thương hiệu mới'}
              </h3>
              {editBrandId && (
                <button
                  type="button"
                  onClick={handleCancelEditBrand}
                  className="text-neutral-400 hover:text-black font-bold uppercase text-[9px] tracking-wide"
                >
                  ✕ Hủy
                </button>
              )}
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-[10px] uppercase font-bold text-neutral-450 mb-1">Tên thương hiệu *</label>
                <input
                  type="text"
                  placeholder="Ví dụ: Apple, Samsung..."
                  className="w-full border border-neutral-300 rounded px-3 py-2 font-bold"
                  value={brandForm.name}
                  onChange={(e) => setBrandForm(p => ({ ...p, name: e.target.value }))}
                  required
                />
              </div>

              <ImageUploader
                label="Logo thương hiệu"
                value={brandForm.logoUrl}
                onChange={(url) => setBrandForm(p => ({ ...p, logoUrl: url }))}
                placeholder="Kéo thả logo thương hiệu hoặc nhấp để chọn"
              />

              <label className="flex items-center gap-2 cursor-pointer py-1 text-xs">
                <input
                  type="checkbox"
                  className="accent-black h-4 w-4"
                  checked={brandForm.isActive}
                  onChange={(e) => setBrandForm(p => ({ ...p, isActive: e.target.checked }))}
                />
                <span className="font-bold text-neutral-700">Kích hoạt thương hiệu này</span>
              </label>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-black hover:bg-neutral-850 text-white text-[10px] font-black uppercase tracking-wider py-2.5 rounded transition-colors disabled:opacity-50"
            >
              {editBrandId ? 'Cập nhật thương hiệu' : 'Lưu thương hiệu'}
            </button>
          </form>
        </div>
      )}
    </div>
  )
}
