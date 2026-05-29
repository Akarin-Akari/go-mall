'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchProductsAsync, fetchCategoriesAsync, selectProduct, selectCategories } from '@/store/slices/productSlice';
import MainLayout from '@/components/layout/MainLayout';
import { Product } from '@/types';
import { AppleCard } from '@/components/ui/apple/AppleCard';
import { AppleButton } from '@/components/ui/apple/AppleButton';
import { AppleAmbientBackground } from '@/components/ui/apple/AppleAmbientBackground';
import { Search, Filter, Star } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Spin, Drawer, Slider, Checkbox } from 'antd';
import { addToCartAsync } from '@/store/slices/cartSlice';

const ProductsPage: React.FC = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const dispatch = useAppDispatch();

  const { products, loading } = useAppSelector(selectProduct);
  const categories = useAppSelector(selectCategories);

  const [activeCategory, setActiveCategory] = useState<number>(0);
  const [filterOpen, setFilterOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // Fetch Categories once
  useEffect(() => {
    dispatch(fetchCategoriesAsync());
  }, [dispatch]);

  // Fetch Products when params change
  useEffect(() => {
    const category = searchParams?.get('category');
    const catId = category ? Number(category) : 0;
    setActiveCategory(catId);

    dispatch(fetchProductsAsync({
      page: 1,
      page_size: 24,
      category_id: catId === 0 ? undefined : catId,
      keyword: searchQuery || undefined
    }));
  }, [dispatch, searchParams, searchQuery]); // re-fetch when search changes (debounce ideally needed but keeping simple for logic completeness)

  const handleCategoryClick = (id: number) => {
    setActiveCategory(id);
    const params = new URLSearchParams(searchParams?.toString());
    if (id === 0) {
      params.delete('category');
    } else {
      params.set('category', id.toString());
    }
    router.push(`/products?${params.toString()}`);
  };

  const handleAddToCart = async (product: Product) => {
    try {
      // @ts-ignore
      await dispatch((await import('@/store/slices/cartSlice')).addToCartAsync({
        product_id: product.id,
        quantity: 1
      })).unwrap();
    } catch (e) {
      console.error("Cart error", e);
    }
  };

  return (
    <MainLayout>
      <div className="relative min-h-screen pb-20 bg-[#f5f5f7]">
        {/* Subtle Ambient BG */}
        <AppleAmbientBackground className="fixed inset-0 opacity-30" fallbackColor="#F5F5F7" />

        <div className="relative z-10 max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 pt-8">

          {/* Header & Search */}
          <div className="flex flex-col md:flex-row justify-between items-center mb-8 gap-4">
            <div>
              <h1 className="text-3xl font-bold text-[#1d1d1f]">Store. <span className="text-[#86868b]">All Products.</span></h1>
            </div>

            <div className="flex items-center gap-3 w-full md:w-auto">
              <div className="relative flex-1 md:w-[300px]">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search"
                  className="w-full h-10 pl-10 pr-4 rounded-full bg-white border-0 shadow-sm focus:ring-2 focus:ring-[#007AFF] focus:outline-none transition-all placeholder:text-gray-400 text-sm"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />
              </div>
              <AppleButton variant="secondary" size="icon" onClick={() => setFilterOpen(true)}>
                <Filter className="w-4 h-4" />
              </AppleButton>
            </div>
          </div>

          {/* Dynamic Pill Filters */}
          <div className="mb-8 overflow-x-auto no-scrollbar pb-2">
            <div className="flex gap-3">
              <button
                onClick={() => handleCategoryClick(0)}
                className={cn(
                  "px-5 py-2 rounded-full text-sm font-medium transition-all duration-300 flex-shrink-0",
                  activeCategory === 0
                    ? "bg-[#1d1d1f] text-white shadow-md transform scale-105"
                    : "bg-white text-[#1d1d1f] hover:bg-white/80"
                )}
              >
                All
              </button>
              {categories.map((cat) => (
                <button
                  key={cat.id}
                  onClick={() => handleCategoryClick(cat.id)}
                  className={cn(
                    "px-5 py-2 rounded-full text-sm font-medium transition-all duration-300 flex-shrink-0",
                    activeCategory === cat.id
                      ? "bg-[#1d1d1f] text-white shadow-md transform scale-105"
                      : "bg-white text-[#1d1d1f] hover:bg-white/80"
                  )}
                >
                  {cat.name}
                </button>
              ))}
            </div>
          </div>

          {/* Product Grid */}
          <Spin spinning={loading}>
            {products.length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                {products.map((product) => (
                  <AppleCard
                    key={product.id}
                    className="group flex flex-col justify-between h-[420px]"
                    noPadding
                  >
                    <div className="relative h-[280px] p-6 flex items-center justify-center bg-[#fbfbfd]">
                      <img
                        src={product.images?.[0] || 'https://placehold.co/400x400?text=Product'}
                        alt={product.name}
                        className="max-h-full max-w-full object-contain mix-blend-multiply transition-transform duration-500 group-hover:scale-105"
                      />
                      {/* Quick Add Button Overlay */}
                      <div className="absolute bottom-4 right-4 translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-300">
                        <AppleButton size="sm" onClick={(e) => { e.stopPropagation(); handleAddToCart(product); }}>
                          Add
                        </AppleButton>
                      </div>
                    </div>

                    <div className="p-6 flex flex-col flex-1 justify-between">
                      <div>
                        <p className="text-[#bf4800] text-xs font-semibold mb-1 uppercase">New</p>
                        <h3 className="text-lg font-semibold text-[#1d1d1f] leading-tight mb-2 group-hover:text-[#007AFF] transition-colors cursor-pointer" onClick={() => router.push(`/products/${product.id}`)}>
                          {product.name}
                        </h3>
                        <p className="text-sm text-[#86868b] line-clamp-2">{product.description}</p>
                      </div>
                      <div className="pt-4 flex justify-between items-end">
                        <span className="text-lg font-medium text-[#1d1d1f]">${product.price}</span>
                        {/* Star Rating Visualization */}
                        <div className="flex gap-0.5">
                          {[...Array(5)].map((_, i) => (
                            <Star key={i} className={cn("w-3 h-3", i < 4 ? "fill-[#ffb400] text-[#ffb400]" : "text-gray-300")} />
                          ))}
                        </div>
                      </div>
                    </div>
                  </AppleCard>
                ))}
              </div>
            ) : (
              <div className="h-[400px] flex flex-col items-center justify-center text-[#86868b]">
                <p>No products found.</p>
                <AppleButton variant="ghost" className="mt-4" onClick={() => handleCategoryClick(0)}>Clear Filters</AppleButton>
              </div>
            )}
          </Spin>

          {/* Filter Drawer (Keeping it functional but minimal) */}
          <Drawer
            title="Filter"
            placement="right"
            onClose={() => setFilterOpen(false)}
            open={filterOpen}
            width={320}
          >
            <div className="space-y-6">
              <div>
                <h3 className="font-semibold mb-3">Price Range</h3>
                <Slider range min={0} max={2000} defaultValue={[0, 2000]} />
              </div>
              <div>
                <h3 className="font-semibold mb-3">Features</h3>
                {[1, 2, 3].map(i => (
                  <div key={i} className="flex items-center gap-2 mb-2">
                    <Checkbox /> <span>Feature {i}</span>
                  </div>
                ))}
              </div>
            </div>
          </Drawer>

        </div>
      </div>
    </MainLayout>
  );
};

export default ProductsPage;
