import { Link } from 'react-router-dom'
import { type Brand } from '../types'

interface BrandListProps {
  brands: Brand[]
  loading?: boolean
  className?: string
}

const BrandList = ({ brands, loading = false, className = '' }: BrandListProps) => {
  if (loading) {
    return (
      <div className={`grid gap-4 ${className}`}>
        {[...Array(8)].map((_, i) => (
          <div
            key={i}
            className="aspect-square animate-pulse rounded-lg bg-neutral-200"
          ></div>
        ))}
      </div>
    )
  }

  return (
    <div className={`grid gap-4 ${className}`}>
      {brands.map((brand) => (
        <Link
          key={brand.id}
          to={`/browse?brand=${brand.id}`}
          className="group flex flex-col items-center justify-center gap-3 rounded-lg border border-neutral-200 bg-white p-5 hover:border-black hover:shadow-premium hover:-translate-y-0.5 transition-all duration-300"
        >
          {brand.logo ? (
            <img
              src={brand.logo}
              alt={brand.name}
              className="h-10 w-20 object-contain filter grayscale opacity-60 transition-all duration-300 group-hover:grayscale-0 group-hover:opacity-100 group-hover:scale-105"
            />
          ) : (
            <div className="flex h-10 w-20 items-center justify-center rounded bg-neutral-50 font-black text-neutral-450 group-hover:bg-neutral-900 group-hover:text-white transition-all">
              {brand.name.substring(0, 2).toUpperCase()}
            </div>
          )}
          <span className="text-center text-xs font-bold text-neutral-700 group-hover:text-black transition-colors">
            {brand.name}
          </span>
        </Link>
      ))}
    </div>
  )
}

export default BrandList

