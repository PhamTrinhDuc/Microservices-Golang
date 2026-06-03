import React from 'react'

interface SearchSkeletonProps {
  viewMode: 'grid' | 'list'
  count?: number
}

const SearchSkeleton: React.FC<SearchSkeletonProps> = ({ viewMode, count = 8 }) => {
  return (
    <div
      className={
        viewMode === 'grid'
          ? 'grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6'
          : 'space-y-4'
      }
    >
      {[...Array(count)].map((_, idx) => (
        <div
          key={idx}
          className={`bg-white rounded-xl border border-neutral-100 overflow-hidden shadow-sm animate-pulse ${
            viewMode === 'list' ? 'flex gap-6 p-4' : 'flex flex-col'
          }`}
        >
          {/* Product Thumbnail Shimmer */}
          <div
            className={`bg-gradient-to-r from-neutral-200 to-neutral-100 relative ${
              viewMode === 'list'
                ? 'w-32 h-32 md:w-40 md:h-40 rounded-lg flex-shrink-0'
                : 'aspect-square w-full'
            }`}
          >
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent animate-shimmer" style={{ backgroundSize: '200% 100%' }}></div>
          </div>

          {/* Product Info Shimmer */}
          <div className="flex-1 p-4 flex flex-col justify-between space-y-3">
            <div className="space-y-2">
              {/* Category/Brand Tag */}
              <div className="h-3 w-16 bg-neutral-200 rounded"></div>
              {/* Product Title */}
              <div className="h-4 w-3/4 bg-neutral-200 rounded"></div>
              <div className="h-4 w-1/2 bg-neutral-200 rounded"></div>
            </div>

            {/* Price & Rating Row */}
            <div className="flex items-center justify-between pt-2">
              <div className="space-y-1">
                <div className="h-5 w-24 bg-neutral-200 rounded"></div>
                <div className="h-3.5 w-16 bg-neutral-200 rounded"></div>
              </div>
              <div className="h-4 w-12 bg-neutral-200 rounded"></div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

export default SearchSkeleton
