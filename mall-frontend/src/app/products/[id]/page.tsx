'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchProductDetailAsync, selectProduct } from '@/store/slices/productSlice';
import { addToCartAsync } from '@/store/slices/cartSlice';
import MainLayout from '@/components/layout/MainLayout';
import { AppleButton } from '@/components/ui/apple/AppleButton';
import { AppleCard } from '@/components/ui/apple/AppleCard';
import { AppleAmbientBackground } from '@/components/ui/apple/AppleAmbientBackground';
import { ChevronLeft, Share2, Star, Truck, ShieldCheck, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Spin, message, Rate } from 'antd';
import { motion, AnimatePresence } from 'framer-motion';

const ProductDetailPage: React.FC = () => {
  const router = useRouter();
  const params = useParams();
  const dispatch = useAppDispatch();
  const { currentProduct, productLoading } = useAppSelector(selectProduct);
  const [addingToCart, setAddingToCart] = useState(false);

  // Selection State (Mocking options since backend data is generic)
  const [selectedColor, setSelectedColor] = useState('Titanium Black');
  const [selectedStorage, setSelectedStorage] = useState('256GB');

  useEffect(() => {
    if (params?.id) {
      dispatch(fetchProductDetailAsync(Number(params.id)));
    }
  }, [dispatch, params]);

  const handleAddToCart = async () => {
    if (!currentProduct) return;
    setAddingToCart(true);
    try {
      // @ts-ignore
      await dispatch((await import('@/store/slices/cartSlice')).addToCartAsync({
        product_id: currentProduct.id,
        quantity: 1
      })).unwrap();
      message.success("Added to cart");
    } catch (e: any) {
      message.error("Failed to add to cart");
    } finally {
      setAddingToCart(false);
    }
  };

  if (productLoading || !currentProduct) {
    return (
      <MainLayout>
        <div className="h-screen flex items-center justify-center bg-[#f5f5f7]">
          <Spin size="large" />
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="min-h-screen bg-[#f5f5f7] pb-32">
        <AppleAmbientBackground className="fixed inset-0 opacity-50 pointer-events-none" />

        {/* Global Nav with Back Button */}
        <div className="sticky top-0 z-50 bg-[#f5f5f7]/80 backdrop-blur-md border-b border-gray-200">
          <div className="max-w-[1000px] mx-auto px-4 h-14 flex items-center justify-between">
            <div className="flex items-center gap-2 cursor-pointer hover:text-blue-600 transition-colors" onClick={() => router.back()}>
              <ChevronLeft className="w-5 h-5" />
              <span className="font-medium text-sm">Back</span>
            </div>
            <div className="font-semibold text-sm truncate max-w-[200px]">{currentProduct.name}</div>
            <div className="flex gap-4">
              <button className="text-sm font-medium text-blue-600 hover:underline">Overview</button>
              <button className="text-sm font-medium text-gray-500 hover:text-gray-900">Tech Specs</button>
            </div>
          </div>
        </div>

        <div className="max-w-[1000px] mx-auto px-4 pt-10 grid grid-cols-1 md:grid-cols-2 gap-12">

          {/* Left Column - Sticky Image */}
          <div className="relative">
            <div className="sticky top-24 space-y-8">
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.8 }}
                className="relative aspect-square rounded-[40px] overflow-hidden bg-white shadow-xl flex items-center justify-center p-8"
              >
                <img
                  src={currentProduct.images?.[0] || 'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?w=800&h=800&fit=crop'}
                  alt={currentProduct.name}
                  className="max-w-full max-h-full object-contain drop-shadow-2xl"
                />
              </motion.div>

              <div className="grid grid-cols-4 gap-4">
                {(currentProduct.images?.slice(0, 4) || [1, 2, 3]).map((img, i) => (
                  <div key={i} className="aspect-square rounded-2xl bg-white shadow-sm flex items-center justify-center p-2 cursor-pointer border-2 border-transparent hover:border-blue-500 transition-all">
                    <img src={typeof img === 'string' ? img : currentProduct.images?.[0]} className="max-w-full max-h-full object-contain" />
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right Column - Selection content */}
          <div className="pt-4 space-y-10">

            {/* Header Info */}
            <div>
              <h2 className="text-[#bf4800] font-semibold text-sm mb-2 uppercase tracking-wide">New</h2>
              <h1 className="text-4xl md:text-5xl font-bold text-[#1d1d1f] mb-4 leading-tight font-display">{currentProduct.name}</h1>
              <div className="flex items-center gap-2 mb-6">
                <span className="text-2xl font-medium">${currentProduct.price}</span>
                <div className="flex items-center gap-0.5 ml-4">
                  <Star className="w-4 h-4 fill-[#ffb400] text-[#ffb400]" />
                  <span className="text-sm font-medium ml-1">{currentProduct.rating?.toFixed(1) || 5.0}</span>
                </div>
              </div>
              <p className="text-lg text-[#86868b] leading-relaxed">{currentProduct.description}</p>
            </div>

            {/* Config: Color */}
            <div>
              <h3 className="text-xl font-semibold mb-4">Color. <span className="text-gray-500">Pick your favorite.</span></h3>
              <div className="grid grid-cols-2 gap-4">
                {['Titanium Black', 'Natural Desktop', 'White Frost', 'Blue Titanium'].map(color => (
                  <AppleCard
                    key={color}
                    noPadding
                    hoverEffect={selectedColor !== color}
                    className={cn("p-4 border-2 transition-all cursor-pointer flex flex-col items-center gap-3",
                      selectedColor === color ? "border-blue-600 bg-white shadow-md ring-2 ring-blue-100" : "border-transparent bg-white"
                    )}
                    onClick={() => setSelectedColor(color)}
                  >
                    <div className={cn("w-8 h-8 rounded-full border shadow-inner",
                      color.includes('Black') ? 'bg-zinc-800' :
                        color.includes('Natural') ? 'bg-zinc-400' :
                          color.includes('White') ? 'bg-white' : 'bg-blue-900'
                    )} />
                    <span className="font-medium text-sm text-center">{color}</span>
                  </AppleCard>
                ))}
              </div>
            </div>

            {/* Config: Storage (Mock) */}
            <div>
              <h3 className="text-xl font-semibold mb-4">Storage. <span className="text-gray-500">How much space do you need?</span></h3>
              <div className="space-y-4">
                {['128GB', '256GB', '512GB', '1TB'].map(storage => (
                  <AppleCard
                    key={storage}
                    noPadding
                    hoverEffect={selectedStorage !== storage}
                    className={cn("p-6 border-2 transition-all cursor-pointer flex justify-between items-center",
                      selectedStorage === storage ? "border-blue-600 bg-white shadow-md ring-2 ring-blue-100" : "border-transparent bg-white"
                    )}
                    onClick={() => setSelectedStorage(storage)}
                  >
                    <span className="font-semibold">{storage}</span>
                    <span className="text-sm text-gray-500">
                      {storage === '128GB' ? '+$0' : storage === '256GB' ? '+$100' : '+$300'}
                    </span>
                  </AppleCard>
                ))}
              </div>
            </div>

            {/* Delivery Info */}
            <div className="bg-white rounded-2xl p-6 space-y-4 shadow-sm">
              <div className="flex gap-4 items-start">
                <Truck className="w-6 h-6 text-gray-400" />
                <div>
                  <h4 className="font-semibold text-sm">Free Delivery</h4>
                  <p className="text-sm text-gray-500">Get it by tomorrow, Jan 30.</p>
                </div>
              </div>
              <div className="w-full h-[1px] bg-gray-100" />
              <div className="flex gap-4 items-start">
                <RefreshCw className="w-6 h-6 text-gray-400" />
                <div>
                  <h4 className="font-semibold text-sm">Free Returns</h4>
                  <p className="text-sm text-gray-500">You have 14 days to return it.</p>
                </div>
              </div>
            </div>

          </div>
        </div>

        {/* Sticky Buy Bar (Bottom) */}
        <div className="fixed bottom-0 left-0 right-0 bg-[#f5f5f7]/90 backdrop-blur-xl border-t border-gray-200 p-4 pb-8 z-40 transform transition-transform">
          <div className="max-w-[1000px] mx-auto flex items-center justify-between">
            <div className="hidden md:block">
              <div className="font-semibold text-sm">Total</div>
              <div className="font-bold text-xl">${currentProduct.price}</div>
            </div>
            <div className="lex-1 md:flex-none flex gap-4 w-full md:w-auto">
              <AppleButton variant="ghost" size="lg" className="flex-1 md:flex-none">Share</AppleButton>
              <AppleButton
                variant="primary"
                size="lg"
                className="flex-1 md:min-w-[200px]"
                onClick={handleAddToCart}
                isLoading={addingToCart}
              >
                Add to Bag
              </AppleButton>
            </div>
          </div>
        </div>

      </div>
    </MainLayout>
  );
};

export default ProductDetailPage;
