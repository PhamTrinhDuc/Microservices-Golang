import React, { useState, useRef } from 'react'
import { uploadAPI } from '../../services/uploadAPI'

interface ImageUploaderProps {
  label?: string
  value: string
  onChange: (url: string) => void
  placeholder?: string
}

export default function ImageUploader({
  label,
  value,
  onChange,
  placeholder = 'Kéo thả ảnh vào đây hoặc nhấp để chọn'
}: ImageUploaderProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [dragActive, setDragActive] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleUpload = async (file: File) => {
    // Validate file type
    if (!file.type.startsWith('image/')) {
      setError('Vui lòng chọn tệp hình ảnh (.jpg, .png, .webp,...)')
      return
    }
    // Limit to 5MB
    if (file.size > 5 * 1024 * 1024) {
      setError('Dung lượng ảnh tối đa là 5MB')
      return
    }

    try {
      setLoading(true)
      setError(null)
      const res = await uploadAPI.uploadImage(file)
      onChange(res.url)
    } catch (err: any) {
      console.error(err)
      setError(err.response?.data?.message || err.message || 'Lỗi khi tải ảnh lên')
    } finally {
      setLoading(false)
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      void handleUpload(e.target.files[0])
    }
  }

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      void handleUpload(e.dataTransfer.files[0])
    }
  }

  const handleRemove = () => {
    onChange('')
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className="space-y-2">
      {label && (
        <label className="block text-[10px] uppercase font-bold text-neutral-400 mb-1">
          {label}
        </label>
      )}

      {value ? (
        <div className="relative group border border-neutral-200 rounded-lg overflow-hidden bg-neutral-50 h-32 flex items-center justify-center">
          <img
            src={value}
            alt="Uploaded Preview"
            className="h-full w-full object-contain"
            onError={(e) => {
              ;(e.target as HTMLImageElement).src = 'https://placehold.co/600x400?text=Loi+hinh+anh'
            }}
          />
          <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="bg-white/90 hover:bg-white text-neutral-900 font-bold uppercase text-[9px] px-3 py-1.5 rounded transition-all shadow-sm"
            >
              Thay đổi
            </button>
            <button
              type="button"
              onClick={handleRemove}
              className="bg-red-650 hover:bg-red-700 text-white font-bold uppercase text-[9px] px-3 py-1.5 rounded transition-all shadow-sm"
            >
              Gỡ bỏ
            </button>
          </div>
        </div>
      ) : (
        <div
          onDragEnter={handleDrag}
          onDragOver={handleDrag}
          onDragLeave={handleDrag}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={`border-2 border-dashed rounded-lg p-6 flex flex-col items-center justify-center gap-2 cursor-pointer transition-all ${
            dragActive
              ? 'border-black bg-neutral-50 scale-[0.99]'
              : 'border-neutral-250 hover:border-neutral-400 bg-white'
          }`}
        >
          {loading ? (
            <div className="flex flex-col items-center gap-2">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-black"></div>
              <span className="text-[10px] text-neutral-500 font-bold uppercase tracking-wider">Đang tải lên...</span>
            </div>
          ) : (
            <>
              <svg
                className="w-8 h-8 text-neutral-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              <span className="text-[10px] text-neutral-500 font-medium text-center">
                {placeholder}
              </span>
              <span className="text-[9px] text-neutral-400 font-bold uppercase">
                Hỗ trợ: JPG, PNG, WEBP (Tối đa 5MB)
              </span>
            </>
          )}
        </div>
      )}

      {/* Hidden file input */}
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileChange}
        className="hidden"
        accept="image/*"
        disabled={loading}
      />



      {error && (
        <p className="text-[10px] font-bold text-red-600 mt-1">{error}</p>
      )}
    </div>
  )
}
