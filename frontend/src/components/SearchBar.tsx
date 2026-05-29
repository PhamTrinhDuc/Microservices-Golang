import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

interface SearchBarProps {
  onSearch?: (query: string) => void
  initialValue?: string
  className?: string
}

const SearchBar = ({ onSearch, initialValue = '', className = '' }: SearchBarProps) => {
  const [query, setQuery] = useState(initialValue)
  const navigate = useNavigate()

  useEffect(() => {
    setQuery(initialValue)
  }, [initialValue])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (query.trim()) {
      if (onSearch) {
        onSearch(query)
      } else {
        navigate(`/search?q=${encodeURIComponent(query)}`)
      }
    }
  }

  return (
    <form onSubmit={handleSubmit} className={`flex gap-3 ${className}`}>
      <div className="relative flex-1">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Tìm sản phẩm, thương hiệu công nghệ..."
          className="w-full rounded-full border border-slate-200 bg-white/80 px-5 py-3 text-sm transition-all focus:border-brand-500 focus:outline-none focus:ring-4 focus:ring-brand-500/10 shadow-sm"
        />
        {query && (
          <button
            type="button"
            onClick={() => setQuery('')}
            className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors"
          >
            ✕
          </button>
        )}
      </div>
      <button
        type="submit"
        className="rounded-full bg-slate-950 px-6 py-3 text-sm font-bold text-white transition-all hover:bg-brand-600 hover:-translate-y-0.5 active:translate-y-0 shadow-md shadow-slate-950/10 hover:shadow-brand-500/10"
      >
        Tìm kiếm
      </button>
    </form>
  )
}

export default SearchBar
