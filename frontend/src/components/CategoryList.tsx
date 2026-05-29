import { Link } from 'react-router-dom'
import { type Category } from '../types'

interface CategoryListProps {
  categories: Category[]
  loading?: boolean
  className?: string
}

const CategoryList = ({ categories, loading = false, className = '' }: CategoryListProps) => {
  if (loading) {
    return (
      <div className={`grid gap-3 ${className}`}>
        {[...Array(6)].map((_, i) => (
          <div
            key={i}
            className="h-14 animate-pulse rounded-lg bg-neutral-200"
          ></div>
        ))}
      </div>
    )
  }

  return (
    <div className={`grid gap-4 ${className}`}>
      {categories.map((category) => (
        <Link
          key={category.id}
          to={`/browse?category=${category.id}`}
          className="flex items-center gap-4 rounded-lg border border-neutral-200 bg-white p-4 hover:border-black hover:shadow-premium hover:-translate-y-0.5 transition-all duration-300 group"
        >
          {category.icon && (
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded bg-neutral-50 group-hover:bg-neutral-100 transition-colors">
              <img
                src={category.icon}
                alt={category.name}
                className="h-8 w-8 object-contain transition-transform duration-300 group-hover:scale-110"
              />
            </div>
          )}
          <span className="flex-1 text-sm font-bold text-neutral-850 group-hover:text-black transition-colors">
            {category.name}
          </span>
          <span className="text-neutral-400 group-hover:text-black transition-colors font-bold group-hover:translate-x-0.5 transform duration-150">
            →
          </span>
        </Link>
      ))}
    </div>
  )
}

export default CategoryList

