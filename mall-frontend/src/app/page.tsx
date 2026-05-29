'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import MainLayout from '@/components/layout/MainLayout';
import { AppleAmbientBackground } from '@/components/ui/apple/AppleAmbientBackground';
import { AppleCard } from '@/components/ui/apple/AppleCard';
import { AppleButton } from '@/components/ui/apple/AppleButton';
import { motion } from 'framer-motion';
import { ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchProductsAsync, fetchCategoriesAsync, selectProducts, selectCategories, selectProductLoading } from '@/store/slices/productSlice';
import { getCategoryVisual, getBentoSize } from '@/lib/visuals';
import { Spin } from 'antd';

export default function Home() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const products = useAppSelector(selectProducts);
  const categories = useAppSelector(selectCategories);
  const loading = useAppSelector(selectProductLoading);

  useEffect(() => {
    dispatch(fetchProductsAsync({ page: 1, page_size: 10 }));
    dispatch(fetchCategoriesAsync());
  }, [dispatch]);

  // Derived state for display
  const heroProducts = products.slice(0, 2);
  const newArrivals = products.slice(2, 7);
  const displayCategories = categories.length > 0 ? categories.slice(0, 4) : [];

  return (
    <MainLayout>
      <div className="relative min-h-screen pb-20 overflow-hidden">

        {/* Ambient Background for the whole page */}
        <AppleAmbientBackground className="fixed inset-0" fallbackColor="#007AFF" />

        <div className="relative z-10 max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 pt-6 space-y-16">

          {/* SECTION 1: HERO */}
          <section className="relative">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
              className="text-center mb-12 pt-10"
            >
              <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-[#1d1d1f] mb-4 font-display">
                Store. <span className="text-[#6e6e73]">The best way to buy the products you love.</span>
              </h1>
            </motion.div>

            {/* Hero Carousel / Featured */}
            <Spin spinning={loading && products.length === 0}>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 min-h-[500px]">
                {heroProducts.map((product, idx) => (
                  <AppleCard
                    key={product.id}
                    className="group relative h-[500px] flex flex-col justify-between overflow-hidden border-0 bg-white"
                    noPadding
                  >
                    <div className="absolute inset-0 z-0">
                      <img
                        src={product.images?.[0] || 'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?w=800&q=80'}
                        alt={product.name}
                        className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
                      />
                      {/* Subtle overlay to ensure text readability */}
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-80" />
                    </div>

                    <div className="relative z-10 p-8 flex flex-col h-full justify-between items-start text-white w-full">
                      <div>
                        <span className="inline-block px-3 py-1 rounded-full bg-white/20 backdrop-blur-md text-xs font-medium mb-3 border border-white/10 shadow-sm">
                          Featured
                        </span>
                        <h2 className="text-4xl font-semibold mb-2 tracking-tight">{product.name}</h2>
                        <p className="text-lg text-white/90 max-w-sm font-medium line-clamp-2">{product.description}</p>
                      </div>

                      <div className="w-full flex justify-end transform translate-y-4 opacity-0 group-hover:translate-y-0 group-hover:opacity-100 transition-all duration-300">
                        <AppleButton
                          variant="glass"
                          onClick={() => router.push(`/products/${product.id}`)}
                          className="text-white border-white/40 hover:bg-white/20 backdrop-blur-md"
                        >
                          Buy
                        </AppleButton>
                      </div>
                    </div>
                  </AppleCard>
                ))}
                {/* Skeleton Loading State if needed, or handled by Spin */}
              </div>
            </Spin>
          </section>

          {/* SECTION 2: BENTO CATEGORIES */}
          <section>
            <div className="flex justify-between items-end mb-8">
              <h2 className="text-3xl font-semibold text-[#1d1d1f]">Categories</h2>
              <AppleButton variant="ghost" className="text-sm font-normal text-[#007AFF] hover:underline" onClick={() => router.push('/products')}>
                View all <ChevronRight className="w-4 h-4 ml-1" />
              </AppleButton>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 auto-rows-[250px]">
              {displayCategories.map((cat, idx) => {
                const visual = getCategoryVisual(cat.id);
                const size = getBentoSize(idx);

                return (
                  <AppleCard
                    key={cat.id}
                    className={cn(
                      "relative group flex flex-col justify-end overflow-hidden border-0 cursor-pointer",
                      size === 'large' ? 'col-span-2 row-span-2 md:row-span-2' :
                        size === 'medium' ? 'col-span-2 row-span-1 md:col-span-1 md:row-span-1' : 'col-span-1 row-span-1'
                    )}
                    noPadding
                    onClick={() => router.push(`/products?category=${cat.id}`)}
                  >
                    {/* Background Image - Object Cover! */}
                    <div className="absolute inset-0 z-0">
                      <img
                        src={visual.image}
                        alt={cat.name}
                        className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
                      />
                      <div className="absolute inset-0 bg-black/20 group-hover:bg-black/10 transition-colors duration-500" />
                      <div className={cn("absolute inset-0 bg-gradient-to-br opacity-50 mix-blend-overlay", visual.color)} />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent" />
                    </div>

                    <div className="relative z-10 p-6 w-full">
                      <h3 className={cn("font-semibold text-white tracking-tight drop-shadow-md", size === 'large' ? 'text-4xl' : 'text-2xl')}>
                        {cat.name}
                      </h3>
                    </div>
                  </AppleCard>
                );
              })}
            </div>
          </section>

          {/* SECTION 3: NEW ARRIVALS SHELF */}
          <section className="py-8">
            <h2 className="text-3xl font-semibold text-[#1d1d1f] mb-8">The Latest. <span className="text-[#6e6e73]">Take a look at what's new.</span></h2>

            <div className="relative -mx-4 px-4 sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8 overflow-x-auto no-scrollbar pb-8">
              <div className="flex space-x-6 w-max">
                {newArrivals.map((item) => (
                  <AppleCard
                    key={item.id}
                    className="w-[280px] md:w-[320px] h-[420px] flex-shrink-0 flex flex-col justify-between group bg-white border border-gray-100/50"
                    onClick={() => router.push(`/products/${item.id}`)}
                    hoverEffect
                  >
                    {/* Reduced padding from p-6 to p-4, better containment */}
                    <div className="h-[260px] w-full flex items-center justify-center p-2 bg-[#f5f5f7] rounded-2xl mb-4 group-hover:bg-gray-50 transition-colors duration-300">
                      <img
                        src={item.images?.[0] || 'https://placehold.co/400x400?text=Product'}
                        className="h-full w-full object-contain mix-blend-multiply drop-shadow-md group-hover:scale-105 transition-transform duration-500"
                        loading="lazy"
                      />
                    </div>

                    <div className="flex flex-col items-start px-2 space-y-1">
                      <p className="text-xs font-semibold text-[#bf4800] uppercase tracking-wide">New</p>
                      <h3 className="text-lg font-bold text-[#1d1d1f] group-hover:text-[#007AFF] transition-colors">{item.name}</h3>
                      <p className="text-sm text-[#1d1d1f] font-medium">From ${item.price}</p>
                    </div>
                    <div className="mt-4 opacity-0 group-hover:opacity-100 transition-opacity duration-300 w-full">
                      <AppleButton size="sm" className="w-full">Buy</AppleButton>
                    </div>
                  </AppleCard>
                ))}
              </div>
            </div>
          </section>

          {/* SECTION 4: RICH FOOTER */}
          <section className="py-16 border-t border-[#d2d2d7] mt-12">
            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-8 text-xs text-[#1d1d1f]">
              <div className="col-span-2 md:col-span-1">
                <h3 className="font-semibold mb-3 text-[#1d1d1f]">Shop and Learn</h3>
                <ul className="space-y-2 text-[#424245]">
                  <li>Store</li>
                  <li>Mac</li>
                  <li>iPad</li>
                  <li>iPhone</li>
                  <li>Watch</li>
                  <li>Accessories</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold mb-3 text-[#1d1d1f]">Account</h3>
                <ul className="space-y-2 text-[#424245]">
                  <li>Manage Your ID</li>
                  <li>Apple Store Account</li>
                  <li>iCloud.com</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold mb-3 text-[#1d1d1f]">Apple Store</h3>
                <ul className="space-y-2 text-[#424245]">
                  <li>Find a Store</li>
                  <li>Genius Bar</li>
                  <li>Today at Apple</li>
                  <li>Apple Camp</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold mb-3 text-[#1d1d1f]">About</h3>
                <ul className="space-y-2 text-[#424245]">
                  <li>Newsroom</li>
                  <li>Career</li>
                  <li>Investors</li>
                  <li>Ethics & Compliance</li>
                </ul>
              </div>
            </div>

            <div className="mt-12 pt-8 border-t border-[#d2d2d7] flex flex-col md:flex-row justify-between items-start md:items-center text-xs text-[#86868b]">
              <p>Copyright © 2026 Mall Inc. All rights reserved.</p>
              <div className="flex flex-wrap gap-4 mt-4 md:mt-0">
                <span className="hover:underline cursor-pointer">Privacy Policy</span>
                <span className="border-l border-gray-300 pl-4 hover:underline cursor-pointer">Terms of Use</span>
                <span className="border-l border-gray-300 pl-4 hover:underline cursor-pointer">Sales and Refunds</span>
                <span className="border-l border-gray-300 pl-4 hover:underline cursor-pointer">Legal</span>
                <span className="border-l border-gray-300 pl-4 hover:underline cursor-pointer">Site Map</span>
              </div>
              <div className="mt-4 md:mt-0 text-[#1d1d1f]">
                United States
              </div>
            </div>
          </section>

        </div>
      </div>
    </MainLayout>
  );
}
