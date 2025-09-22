# 第4章：React性能优化技巧 ⚡

> _"性能优化不是过早优化，而是在正确的时机做正确的事情！"_ 🎯

## 📚 本章导览

React应用的性能优化是一个系统性工程，涉及渲染优化、内存管理、网络优化等多个方面。在Mall-Frontend这样的电商应用中，良好的性能直接影响用户体验和转化率。本章将深入探讨React性能优化的各种技巧和最佳实践。

### 🎯 学习目标

通过本章学习，你将掌握：

- **渲染优化** - React.memo、useMemo、useCallback的正确使用
- **组件优化** - 避免不必要的重渲染和组件设计优化
- **列表优化** - 虚拟滚动和大数据渲染技巧
- **代码分割** - 动态导入和懒加载策略
- **内存优化** - 避免内存泄漏和优化内存使用
- **网络优化** - 数据获取和缓存策略
- **性能监控** - 性能指标监控和分析工具
- **实战应用** - 在Mall-Frontend中的性能优化实践

### 🛠️ 技术栈概览

```typescript
{
  "optimization": ["React.memo", "useMemo", "useCallback"],
  "codesplitting": ["React.lazy", "Suspense", "Dynamic Import"],
  "virtualization": ["react-window", "react-virtualized"],
  "monitoring": ["React DevTools", "Web Vitals", "Performance API"]
}
```

### 📖 本章目录

- [渲染优化基础](#渲染优化基础)
- [组件优化策略](#组件优化策略)
- [列表和大数据优化](#列表和大数据优化)
- [代码分割与懒加载](#代码分割与懒加载)
- [内存优化技巧](#内存优化技巧)
- [网络性能优化](#网络性能优化)
- [性能监控与分析](#性能监控与分析)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🎨 渲染优化基础

### React.memo的正确使用

```typescript
import React, { memo, useMemo, useCallback } from 'react';

// 1. 基础memo使用
interface ProductCardProps {
  product: Product;
  onAddToCart: (product: Product) => void;
  onToggleFavorite: (productId: number) => void;
}

// ❌ 错误：每次都会重新渲染
const ProductCard = ({ product, onAddToCart, onToggleFavorite }: ProductCardProps) => {
  return (
    <div className="product-card">
      <img src={product.images[0]} alt={product.name} />
      <h3>{product.name}</h3>
      <p>¥{product.price}</p>
      <button onClick={() => onAddToCart(product)}>
        加入购物车
      </button>
      <button onClick={() => onToggleFavorite(product.id)}>
        {product.isFavorite ? '❤️' : '🤍'}
      </button>
    </div>
  );
};

// ✅ 正确：使用memo优化
const OptimizedProductCard = memo(({ product, onAddToCart, onToggleFavorite }: ProductCardProps) => {
  console.log('ProductCard渲染:', product.name);

  return (
    <div className="product-card">
      <img src={product.images[0]} alt={product.name} />
      <h3>{product.name}</h3>
      <p>¥{product.price}</p>
      <button onClick={() => onAddToCart(product)}>
        加入购物车
      </button>
      <button onClick={() => onToggleFavorite(product.id)}>
        {product.isFavorite ? '❤️' : '🤍'}
      </button>
    </div>
  );
});

// 2. 自定义比较函数
const ProductCardWithCustomCompare = memo(
  ({ product, onAddToCart, onToggleFavorite }: ProductCardProps) => {
    return (
      <div className="product-card">
        {/* 组件内容 */}
      </div>
    );
  },
  (prevProps, nextProps) => {
    // 自定义比较逻辑
    return (
      prevProps.product.id === nextProps.product.id &&
      prevProps.product.name === nextProps.product.name &&
      prevProps.product.price === nextProps.product.price &&
      prevProps.product.isFavorite === nextProps.product.isFavorite
    );
  }
);

// 3. 父组件优化
function ProductList() {
  const [products, setProducts] = useState<Product[]>([]);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);

  // ❌ 错误：每次渲染都创建新函数
  const handleAddToCart = (product: Product) => {
    setCartItems(prev => [...prev, { product, quantity: 1 }]);
  };

  const handleToggleFavorite = (productId: number) => {
    setProducts(prev => prev.map(p =>
      p.id === productId ? { ...p, isFavorite: !p.isFavorite } : p
    ));
  };

  return (
    <div className="product-list">
      {products.map(product => (
        <ProductCard
          key={product.id}
          product={product}
          onAddToCart={handleAddToCart} // 每次都是新函数
          onToggleFavorite={handleToggleFavorite} // 每次都是新函数
        />
      ))}
    </div>
  );
}

// ✅ 正确：使用useCallback优化
function OptimizedProductList() {
  const [products, setProducts] = useState<Product[]>([]);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);

  // 使用useCallback缓存函数
  const handleAddToCart = useCallback((product: Product) => {
    setCartItems(prev => [...prev, { product, quantity: 1 }]);
  }, []);

  const handleToggleFavorite = useCallback((productId: number) => {
    setProducts(prev => prev.map(p =>
      p.id === productId ? { ...p, isFavorite: !p.isFavorite } : p
    ));
  }, []);

  return (
    <div className="product-list">
      {products.map(product => (
        <OptimizedProductCard
          key={product.id}
          product={product}
          onAddToCart={handleAddToCart}
          onToggleFavorite={handleToggleFavorite}
        />
      ))}
    </div>
  );
}
```

### useMemo的高级应用

```typescript
import { useMemo, useState, useCallback } from 'react';

// 1. 复杂计算的memoization
function useProductFiltering(products: Product[], filters: ProductFilters) {
  const filteredProducts = useMemo(() => {
    console.log('执行产品过滤计算...');

    return products.filter(product => {
      // 分类过滤
      if (filters.category && product.category !== filters.category) {
        return false;
      }

      // 价格范围过滤
      const price = parseFloat(product.price);
      if (price < filters.priceRange[0] || price > filters.priceRange[1]) {
        return false;
      }

      // 评分过滤
      if (product.rating < filters.rating) {
        return false;
      }

      // 搜索关键词过滤
      if (filters.search && !product.name.toLowerCase().includes(filters.search.toLowerCase())) {
        return false;
      }

      return true;
    }).sort((a, b) => {
      // 排序逻辑
      switch (filters.sortBy) {
        case 'price':
          const priceA = parseFloat(a.price);
          const priceB = parseFloat(b.price);
          return filters.sortOrder === 'asc' ? priceA - priceB : priceB - priceA;
        case 'rating':
          return filters.sortOrder === 'asc' ? a.rating - b.rating : b.rating - a.rating;
        case 'sales':
          return filters.sortOrder === 'asc' ? a.sales - b.sales : b.sales - a.sales;
        default:
          return 0;
      }
    });
  }, [products, filters]);

  // 统计信息
  const statistics = useMemo(() => {
    return {
      total: filteredProducts.length,
      averagePrice: filteredProducts.reduce((sum, p) => sum + parseFloat(p.price), 0) / filteredProducts.length,
      averageRating: filteredProducts.reduce((sum, p) => sum + p.rating, 0) / filteredProducts.length,
      categories: [...new Set(filteredProducts.map(p => p.category))],
    };
  }, [filteredProducts]);

  return { filteredProducts, statistics };
}

// 2. 对象和数组的memoization
function useCartCalculations(cartItems: CartItem[]) {
  // 计算总价
  const totalPrice = useMemo(() => {
    return cartItems.reduce((total, item) => {
      const price = parseFloat(item.product.discount_price || item.product.price);
      return total + price * item.quantity;
    }, 0);
  }, [cartItems]);

  // 计算总数量
  const totalQuantity = useMemo(() => {
    return cartItems.reduce((total, item) => total + item.quantity, 0);
  }, [cartItems]);

  // 计算优惠信息
  const discountInfo = useMemo(() => {
    const originalTotal = cartItems.reduce((total, item) => {
      const originalPrice = parseFloat(item.product.price);
      return total + originalPrice * item.quantity;
    }, 0);

    const discountAmount = originalTotal - totalPrice;
    const discountPercentage = originalTotal > 0 ? (discountAmount / originalTotal) * 100 : 0;

    return {
      originalTotal,
      discountAmount,
      discountPercentage,
      hasDiscount: discountAmount > 0,
    };
  }, [cartItems, totalPrice]);

  // 分组信息
  const groupedItems = useMemo(() => {
    const groups: Record<string, CartItem[]> = {};

    cartItems.forEach(item => {
      const category = item.product.category;
      if (!groups[category]) {
        groups[category] = [];
      }
      groups[category].push(item);
    });

    return groups;
  }, [cartItems]);

  return {
    totalPrice,
    totalQuantity,
    discountInfo,
    groupedItems,
  };
}

// 3. 复杂组件的优化
interface ProductGridProps {
  products: Product[];
  filters: ProductFilters;
  onProductClick: (product: Product) => void;
  onAddToCart: (product: Product) => void;
}

function ProductGrid({ products, filters, onProductClick, onAddToCart }: ProductGridProps) {
  const { filteredProducts, statistics } = useProductFiltering(products, filters);

  // 分页数据
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 20;

  const paginatedProducts = useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize;
    const endIndex = startIndex + pageSize;
    return filteredProducts.slice(startIndex, endIndex);
  }, [filteredProducts, currentPage, pageSize]);

  // 分页信息
  const paginationInfo = useMemo(() => {
    const totalPages = Math.ceil(filteredProducts.length / pageSize);
    return {
      currentPage,
      totalPages,
      hasNextPage: currentPage < totalPages,
      hasPrevPage: currentPage > 1,
      startIndex: (currentPage - 1) * pageSize + 1,
      endIndex: Math.min(currentPage * pageSize, filteredProducts.length),
      total: filteredProducts.length,
    };
  }, [filteredProducts.length, currentPage, pageSize]);

  // 事件处理函数
  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
    // 滚动到顶部
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, []);

  return (
    <div className="product-grid">
      {/* 统计信息 */}
      <div className="statistics">
        <p>共找到 {statistics.total} 个商品</p>
        <p>平均价格: ¥{statistics.averagePrice.toFixed(2)}</p>
        <p>平均评分: {statistics.averageRating.toFixed(1)}</p>
      </div>

      {/* 产品网格 */}
      <div className="grid">
        {paginatedProducts.map(product => (
          <OptimizedProductCard
            key={product.id}
            product={product}
            onAddToCart={onAddToCart}
            onToggleFavorite={() => {}} // 简化示例
          />
        ))}
      </div>

      {/* 分页组件 */}
      <Pagination
        {...paginationInfo}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
```

---

## 🧩 组件优化策略

### 组件拆分和职责分离

```typescript
// ❌ 错误：单一组件承担过多职责
function ProductPageBad({ productId }: { productId: number }) {
  const [product, setProduct] = useState<Product | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [recommendations, setRecommendations] = useState<Product[]>([]);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [reviewsLoading, setReviewsLoading] = useState(false);

  // 大量的useEffect和业务逻辑...

  return (
    <div className="product-page">
      {/* 大量的JSX... */}
    </div>
  );
}

// ✅ 正确：拆分为多个专职组件
function ProductPage({ productId }: { productId: number }) {
  return (
    <div className="product-page">
      <ProductHeader productId={productId} />
      <ProductDetails productId={productId} />
      <ProductReviews productId={productId} />
      <ProductRecommendations productId={productId} />
    </div>
  );
}

// 产品头部组件
const ProductHeader = memo(({ productId }: { productId: number }) => {
  const { data: product, loading } = useProduct(productId);
  const { addToCart } = useCart();

  if (loading) return <ProductHeaderSkeleton />;
  if (!product) return <ProductNotFound />;

  return (
    <div className="product-header">
      <ProductImageGallery images={product.images} />
      <ProductInfo product={product} onAddToCart={addToCart} />
    </div>
  );
});

// 产品详情组件
const ProductDetails = memo(({ productId }: { productId: number }) => {
  const { data: product } = useProduct(productId);

  if (!product) return null;

  return (
    <div className="product-details">
      <ProductDescription description={product.description} />
      <ProductSpecifications specifications={product.specifications} />
    </div>
  );
});

// 产品评论组件
const ProductReviews = memo(({ productId }: { productId: number }) => {
  const { data: reviews, loading } = useProductReviews(productId);

  return (
    <div className="product-reviews">
      <ReviewsSummary reviews={reviews} />
      <ReviewsList reviews={reviews} loading={loading} />
    </div>
  );
});
```

### 状态提升优化

```typescript
// 状态管理优化
interface ProductPageState {
  selectedVariant: ProductVariant | null;
  quantity: number;
  selectedImageIndex: number;
}

// 使用Context避免prop drilling
const ProductPageContext = createContext<{
  state: ProductPageState;
  actions: {
    setSelectedVariant: (variant: ProductVariant) => void;
    setQuantity: (quantity: number) => void;
    setSelectedImageIndex: (index: number) => void;
  };
} | null>(null);

function ProductPageProvider({ children, productId }: { children: ReactNode; productId: number }) {
  const [state, setState] = useState<ProductPageState>({
    selectedVariant: null,
    quantity: 1,
    selectedImageIndex: 0,
  });

  const actions = useMemo(() => ({
    setSelectedVariant: (variant: ProductVariant) => {
      setState(prev => ({ ...prev, selectedVariant: variant }));
    },
    setQuantity: (quantity: number) => {
      setState(prev => ({ ...prev, quantity }));
    },
    setSelectedImageIndex: (index: number) => {
      setState(prev => ({ ...prev, selectedImageIndex: index }));
    },
  }), []);

  const value = useMemo(() => ({ state, actions }), [state, actions]);

  return (
    <ProductPageContext.Provider value={value}>
      {children}
    </ProductPageContext.Provider>
  );
}

function useProductPage() {
  const context = useContext(ProductPageContext);
  if (!context) {
    throw new Error('useProductPage must be used within ProductPageProvider');
  }
  return context;
}

// 使用优化后的组件
function OptimizedProductPage({ productId }: { productId: number }) {
  return (
    <ProductPageProvider productId={productId}>
      <div className="product-page">
        <ProductImageGallery />
        <ProductInfo />
        <ProductVariants />
        <ProductActions />
      </div>
    </ProductPageProvider>
  );
}

const ProductImageGallery = memo(() => {
  const { state, actions } = useProductPage();
  const { data: product } = useProduct(productId);

  if (!product) return null;

  return (
    <div className="image-gallery">
      <img
        src={product.images[state.selectedImageIndex]}
        alt={product.name}
      />
      <div className="thumbnails">
        {product.images.map((image, index) => (
          <img
            key={index}
            src={image}
            alt=""
            className={index === state.selectedImageIndex ? 'active' : ''}
            onClick={() => actions.setSelectedImageIndex(index)}
          />
        ))}
      </div>
    </div>
  );
});
```

---

## 📋 列表和大数据优化

### 虚拟滚动实现

```typescript
import { FixedSizeList as List, VariableSizeList } from 'react-window';
import { memo, useMemo, useCallback } from 'react';

// 1. 固定高度的虚拟列表
interface VirtualProductListProps {
  products: Product[];
  onProductClick: (product: Product) => void;
}

const VirtualProductList = memo(({ products, onProductClick }: VirtualProductListProps) => {
  const itemHeight = 120; // 每个商品卡片的高度
  const containerHeight = 600; // 容器高度

  // 渲染单个商品项
  const ProductItem = memo(({ index, style }: { index: number; style: React.CSSProperties }) => {
    const product = products[index];

    return (
      <div style={style} className="virtual-product-item">
        <div className="product-card" onClick={() => onProductClick(product)}>
          <img src={product.images[0]} alt={product.name} />
          <div className="product-info">
            <h3>{product.name}</h3>
            <p className="price">¥{product.price}</p>
            <p className="rating">⭐ {product.rating}</p>
          </div>
        </div>
      </div>
    );
  });

  return (
    <List
      height={containerHeight}
      itemCount={products.length}
      itemSize={itemHeight}
      width="100%"
    >
      {ProductItem}
    </List>
  );
});

// 2. 可变高度的虚拟列表
const VariableHeightProductList = memo(({ products, onProductClick }: VirtualProductListProps) => {
  // 计算每个项目的高度
  const getItemSize = useCallback((index: number) => {
    const product = products[index];
    // 根据商品信息计算高度
    const baseHeight = 100;
    const descriptionHeight = product.description ? Math.ceil(product.description.length / 50) * 20 : 0;
    const imageHeight = product.images.length > 1 ? 60 : 0;

    return baseHeight + descriptionHeight + imageHeight;
  }, [products]);

  const VariableProductItem = memo(({ index, style }: { index: number; style: React.CSSProperties }) => {
    const product = products[index];

    return (
      <div style={style} className="variable-product-item">
        <div className="product-card">
          <img src={product.images[0]} alt={product.name} />
          <div className="product-info">
            <h3>{product.name}</h3>
            <p className="price">¥{product.price}</p>
            <p className="description">{product.description}</p>
            {product.images.length > 1 && (
              <div className="additional-images">
                {product.images.slice(1, 4).map((img, idx) => (
                  <img key={idx} src={img} alt="" className="thumb" />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  });

  return (
    <VariableSizeList
      height={600}
      itemCount={products.length}
      itemSize={getItemSize}
      width="100%"
    >
      {VariableProductItem}
    </VariableSizeList>
  );
});

// 3. 无限滚动 + 虚拟列表
function InfiniteVirtualProductList() {
  const [hasNextPage, setHasNextPage] = useState(true);
  const [isNextPageLoading, setIsNextPageLoading] = useState(false);
  const [items, setItems] = useState<Product[]>([]);

  // 加载更多数据
  const loadNextPage = useCallback(async () => {
    if (isNextPageLoading) return;

    setIsNextPageLoading(true);
    try {
      const response = await fetch(`/api/products?page=${Math.ceil(items.length / 20) + 1}`);
      const data = await response.json();

      setItems(prev => [...prev, ...data.products]);
      setHasNextPage(data.hasMore);
    } catch (error) {
      console.error('Failed to load more products:', error);
    } finally {
      setIsNextPageLoading(false);
    }
  }, [items.length, isNextPageLoading]);

  // 检查是否需要加载更多
  const itemCount = hasNextPage ? items.length + 1 : items.length;

  const isItemLoaded = useCallback((index: number) => {
    return !!items[index];
  }, [items]);

  const InfiniteItem = memo(({ index, style }: { index: number; style: React.CSSProperties }) => {
    const isLoading = !isItemLoaded(index);

    if (isLoading) {
      // 触发加载更多
      if (!isNextPageLoading) {
        loadNextPage();
      }

      return (
        <div style={style} className="loading-item">
          <div className="skeleton-card">
            <div className="skeleton-image"></div>
            <div className="skeleton-text"></div>
            <div className="skeleton-text short"></div>
          </div>
        </div>
      );
    }

    const product = items[index];
    return (
      <div style={style} className="product-item">
        <ProductCard product={product} />
      </div>
    );
  });

  return (
    <List
      height={600}
      itemCount={itemCount}
      itemSize={120}
      width="100%"
    >
      {InfiniteItem}
    </List>
  );
}

// 4. 网格虚拟化
import { FixedSizeGrid as Grid } from 'react-window';

interface VirtualProductGridProps {
  products: Product[];
  columnCount: number;
}

const VirtualProductGrid = memo(({ products, columnCount }: VirtualProductGridProps) => {
  const rowCount = Math.ceil(products.length / columnCount);
  const itemWidth = 250;
  const itemHeight = 300;

  const GridItem = memo(({
    columnIndex,
    rowIndex,
    style
  }: {
    columnIndex: number;
    rowIndex: number;
    style: React.CSSProperties;
  }) => {
    const index = rowIndex * columnCount + columnIndex;
    const product = products[index];

    if (!product) {
      return <div style={style} />;
    }

    return (
      <div style={style} className="grid-item">
        <div className="product-card">
          <img src={product.images[0]} alt={product.name} />
          <h3>{product.name}</h3>
          <p>¥{product.price}</p>
        </div>
      </div>
    );
  });

  return (
    <Grid
      columnCount={columnCount}
      columnWidth={itemWidth}
      height={600}
      rowCount={rowCount}
      rowHeight={itemHeight}
      width="100%"
    >
      {GridItem}
    </Grid>
  );
});
```

### 大数据处理优化

```typescript
// 1. 数据分片处理
function useChunkedData<T>(data: T[], chunkSize: number = 1000) {
  const [processedChunks, setProcessedChunks] = useState<T[][]>([]);
  const [currentChunkIndex, setCurrentChunkIndex] = useState(0);
  const [isProcessing, setIsProcessing] = useState(false);

  const processNextChunk = useCallback(() => {
    if (currentChunkIndex >= Math.ceil(data.length / chunkSize)) {
      setIsProcessing(false);
      return;
    }

    setIsProcessing(true);

    // 使用 requestIdleCallback 在浏览器空闲时处理
    const processChunk = (deadline: IdleDeadline) => {
      const startIndex = currentChunkIndex * chunkSize;
      const endIndex = Math.min(startIndex + chunkSize, data.length);
      const chunk = data.slice(startIndex, endIndex);

      // 处理数据块
      const processedChunk = chunk.map(item => {
        // 执行复杂的数据处理逻辑
        return processItem(item);
      });

      setProcessedChunks(prev => [...prev, processedChunk]);
      setCurrentChunkIndex(prev => prev + 1);

      // 如果还有时间且还有数据要处理，继续处理
      if (
        deadline.timeRemaining() > 0 &&
        currentChunkIndex + 1 < Math.ceil(data.length / chunkSize)
      ) {
        processChunk(deadline);
      } else {
        // 安排下一次处理
        if (currentChunkIndex + 1 < Math.ceil(data.length / chunkSize)) {
          requestIdleCallback(processChunk);
        } else {
          setIsProcessing(false);
        }
      }
    };

    requestIdleCallback(processChunk);
  }, [data, chunkSize, currentChunkIndex]);

  useEffect(() => {
    if (data.length > 0 && processedChunks.length === 0) {
      processNextChunk();
    }
  }, [data, processNextChunk]);

  const allProcessedData = useMemo(() => {
    return processedChunks.flat();
  }, [processedChunks]);

  return {
    processedData: allProcessedData,
    isProcessing,
    progress: (currentChunkIndex / Math.ceil(data.length / chunkSize)) * 100,
  };
}

// 2. Web Worker 数据处理
class DataProcessor {
  private worker: Worker | null = null;

  constructor() {
    if (typeof Worker !== 'undefined') {
      this.worker = new Worker('/workers/dataProcessor.js');
    }
  }

  async processLargeDataset(data: any[]): Promise<any[]> {
    if (!this.worker) {
      // 降级到主线程处理
      return this.processInMainThread(data);
    }

    return new Promise((resolve, reject) => {
      const handleMessage = (event: MessageEvent) => {
        const { type, result, error } = event.data;

        if (type === 'PROCESS_COMPLETE') {
          this.worker!.removeEventListener('message', handleMessage);
          resolve(result);
        } else if (type === 'PROCESS_ERROR') {
          this.worker!.removeEventListener('message', handleMessage);
          reject(new Error(error));
        }
      };

      this.worker.addEventListener('message', handleMessage);
      this.worker.postMessage({ type: 'PROCESS_DATA', data });
    });
  }

  private processInMainThread(data: any[]): any[] {
    // 主线程处理逻辑
    return data.map(item => processItem(item));
  }

  destroy() {
    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
    }
  }
}

// 使用 Web Worker 的 Hook
function useDataProcessor() {
  const processorRef = useRef<DataProcessor | null>(null);

  useEffect(() => {
    processorRef.current = new DataProcessor();

    return () => {
      if (processorRef.current) {
        processorRef.current.destroy();
      }
    };
  }, []);

  const processData = useCallback(async (data: any[]) => {
    if (!processorRef.current) {
      throw new Error('Data processor not initialized');
    }

    return processorRef.current.processLargeDataset(data);
  }, []);

  return { processData };
}
```

---

## 🚀 代码分割与懒加载

### React.lazy 和 Suspense

```typescript
import { lazy, Suspense, ComponentType } from 'react';

// 1. 基础懒加载
const LazyProductDetail = lazy(() => import('../components/ProductDetail'));
const LazyCheckout = lazy(() => import('../pages/Checkout'));
const LazyUserProfile = lazy(() => import('../pages/UserProfile'));

// 2. 带错误边界的懒加载
interface LazyComponentProps {
  fallback?: ComponentType;
  errorFallback?: ComponentType<{ error: Error; retry: () => void }>;
}

function LazyWrapper({
  children,
  fallback: Fallback = LoadingSpinner,
  errorFallback: ErrorFallback = ErrorBoundaryFallback
}: LazyComponentProps & { children: React.ReactNode }) {
  return (
    <ErrorBoundary fallback={ErrorFallback}>
      <Suspense fallback={<Fallback />}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
}

// 3. 智能预加载
class ComponentPreloader {
  private static preloadedComponents = new Set<string>();
  private static preloadPromises = new Map<string, Promise<any>>();

  static preload(componentName: string, importFn: () => Promise<any>) {
    if (this.preloadedComponents.has(componentName)) {
      return this.preloadPromises.get(componentName);
    }

    const promise = importFn().then(module => {
      this.preloadedComponents.add(componentName);
      return module;
    });

    this.preloadPromises.set(componentName, promise);
    return promise;
  }

  static preloadOnHover(componentName: string, importFn: () => Promise<any>) {
    return {
      onMouseEnter: () => this.preload(componentName, importFn),
      onFocus: () => this.preload(componentName, importFn),
    };
  }

  static preloadOnIdle(componentName: string, importFn: () => Promise<any>) {
    if ('requestIdleCallback' in window) {
      requestIdleCallback(() => this.preload(componentName, importFn));
    } else {
      setTimeout(() => this.preload(componentName, importFn), 1000);
    }
  }
}

// 4. 路由级别的代码分割
import { Routes, Route } from 'react-router-dom';

// 懒加载页面组件
const HomePage = lazy(() => import('../pages/HomePage'));
const ProductsPage = lazy(() => import('../pages/ProductsPage'));
const ProductDetailPage = lazy(() => import('../pages/ProductDetailPage'));
const CartPage = lazy(() => import('../pages/CartPage'));
const CheckoutPage = lazy(() => import('../pages/CheckoutPage'));
const UserProfilePage = lazy(() => import('../pages/UserProfilePage'));

function AppRoutes() {
  // 预加载关键页面
  useEffect(() => {
    // 预加载购物车页面（用户可能很快访问）
    ComponentPreloader.preloadOnIdle('CartPage', () => import('../pages/CartPage'));

    // 预加载商品页面
    ComponentPreloader.preloadOnIdle('ProductsPage', () => import('../pages/ProductsPage'));
  }, []);

  return (
    <Routes>
      <Route
        path="/"
        element={
          <LazyWrapper>
            <HomePage />
          </LazyWrapper>
        }
      />
      <Route
        path="/products"
        element={
          <LazyWrapper>
            <ProductsPage />
          </LazyWrapper>
        }
      />
      <Route
        path="/products/:id"
        element={
          <LazyWrapper>
            <ProductDetailPage />
          </LazyWrapper>
        }
      />
      <Route
        path="/cart"
        element={
          <LazyWrapper>
            <CartPage />
          </LazyWrapper>
        }
      />
      <Route
        path="/checkout"
        element={
          <LazyWrapper>
            <CheckoutPage />
          </LazyWrapper>
        }
      />
      <Route
        path="/profile"
        element={
          <LazyWrapper>
            <UserProfilePage />
          </LazyWrapper>
        }
      />
    </Routes>
  );
}

// 5. 组件级别的动态导入
function DynamicComponentLoader() {
  const [ComponentToRender, setComponentToRender] = useState<ComponentType | null>(null);
  const [loading, setLoading] = useState(false);

  const loadComponent = useCallback(async (componentName: string) => {
    setLoading(true);
    try {
      let module;

      switch (componentName) {
        case 'ProductComparison':
          module = await import('../components/ProductComparison');
          break;
        case 'WishList':
          module = await import('../components/WishList');
          break;
        case 'RecentlyViewed':
          module = await import('../components/RecentlyViewed');
          break;
        default:
          throw new Error(`Unknown component: ${componentName}`);
      }

      setComponentToRender(() => module.default);
    } catch (error) {
      console.error('Failed to load component:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  return (
    <div>
      <div className="component-selector">
        <button onClick={() => loadComponent('ProductComparison')}>
          加载商品对比
        </button>
        <button onClick={() => loadComponent('WishList')}>
          加载心愿单
        </button>
        <button onClick={() => loadComponent('RecentlyViewed')}>
          加载最近浏览
        </button>
      </div>

      {loading && <LoadingSpinner />}
      {ComponentToRender && <ComponentToRender />}
    </div>
  );
}
```

---

## 🧠 内存优化技巧

### 避免内存泄漏

```typescript
import { useEffect, useRef, useCallback, useState } from 'react';

// 1. 清理事件监听器
function useEventListener<T extends keyof WindowEventMap>(
  eventName: T,
  handler: (event: WindowEventMap[T]) => void,
  element: Window | Element = window
) {
  const savedHandler = useRef(handler);

  useEffect(() => {
    savedHandler.current = handler;
  }, [handler]);

  useEffect(() => {
    const eventListener = (event: WindowEventMap[T]) =>
      savedHandler.current(event);

    element.addEventListener(eventName, eventListener as EventListener);

    return () => {
      element.removeEventListener(eventName, eventListener as EventListener);
    };
  }, [eventName, element]);
}

// 2. 清理定时器
function useInterval(callback: () => void, delay: number | null) {
  const savedCallback = useRef(callback);

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  useEffect(() => {
    if (delay === null) return;

    const id = setInterval(() => savedCallback.current(), delay);

    return () => clearInterval(id);
  }, [delay]);
}

// 3. 清理异步操作
function useAsyncOperation() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const execute = useCallback(async (url: string) => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    abortControllerRef.current = new AbortController();

    try {
      setLoading(true);
      setError(null);

      const response = await fetch(url, {
        signal: abortControllerRef.current.signal,
      });

      if (!response.ok) {
        throw new Error('Request failed');
      }

      const result = await response.json();
      setData(result);
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setError(err);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return { data, loading, error, execute };
}

// 4. 内存监控Hook
function useMemoryMonitor() {
  const [memoryInfo, setMemoryInfo] = useState<any>(null);

  useEffect(() => {
    const updateMemoryInfo = () => {
      if ('memory' in performance) {
        setMemoryInfo({
          usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
          totalJSHeapSize: (performance as any).memory.totalJSHeapSize,
          jsHeapSizeLimit: (performance as any).memory.jsHeapSizeLimit,
        });
      }
    };

    updateMemoryInfo();
    const interval = setInterval(updateMemoryInfo, 5000);

    return () => clearInterval(interval);
  }, []);

  return memoryInfo;
}

// 5. 大对象清理
function useLargeDataCleanup<T>(data: T[], maxSize: number = 1000) {
  const [cleanedData, setCleanedData] = useState<T[]>([]);

  useEffect(() => {
    if (data.length > maxSize) {
      // 只保留最新的数据
      const cleaned = data.slice(-maxSize);
      setCleanedData(cleaned);

      // 强制垃圾回收（仅在开发环境）
      if (process.env.NODE_ENV === 'development' && 'gc' in window) {
        (window as any).gc();
      }
    } else {
      setCleanedData(data);
    }
  }, [data, maxSize]);

  return cleanedData;
}
```

### 图片和资源优化

```typescript
// 1. 图片懒加载Hook
function useImageLazyLoad() {
  const [loadedImages, setLoadedImages] = useState(new Set<string>());
  const observerRef = useRef<IntersectionObserver | null>(null);

  useEffect(() => {
    observerRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const img = entry.target as HTMLImageElement;
            const src = img.dataset.src;

            if (src) {
              img.src = src;
              img.onload = () => {
                setLoadedImages(prev => new Set(prev).add(src));
                observerRef.current?.unobserve(img);
              };
            }
          }
        });
      },
      { threshold: 0.1 }
    );

    return () => {
      observerRef.current?.disconnect();
    };
  }, []);

  const observeImage = useCallback((img: HTMLImageElement | null) => {
    if (img && observerRef.current) {
      observerRef.current.observe(img);
    }
  }, []);

  return { observeImage, loadedImages };
}

// 2. 图片预加载
class ImagePreloader {
  private static cache = new Map<string, HTMLImageElement>();
  private static loadingPromises = new Map<string, Promise<HTMLImageElement>>();

  static preload(src: string): Promise<HTMLImageElement> {
    if (this.cache.has(src)) {
      return Promise.resolve(this.cache.get(src)!);
    }

    if (this.loadingPromises.has(src)) {
      return this.loadingPromises.get(src)!;
    }

    const promise = new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image();

      img.onload = () => {
        this.cache.set(src, img);
        this.loadingPromises.delete(src);
        resolve(img);
      };

      img.onerror = () => {
        this.loadingPromises.delete(src);
        reject(new Error(`Failed to load image: ${src}`));
      };

      img.src = src;
    });

    this.loadingPromises.set(src, promise);
    return promise;
  }

  static preloadBatch(sources: string[]): Promise<HTMLImageElement[]> {
    return Promise.all(sources.map(src => this.preload(src)));
  }

  static clearCache() {
    this.cache.clear();
    this.loadingPromises.clear();
  }
}

// 3. 响应式图片组件
interface ResponsiveImageProps {
  src: string;
  alt: string;
  sizes?: string;
  className?: string;
  lazy?: boolean;
  placeholder?: string;
}

const ResponsiveImage = memo(({
  src,
  alt,
  sizes = '100vw',
  className,
  lazy = true,
  placeholder = '/images/placeholder.jpg'
}: ResponsiveImageProps) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [error, setError] = useState(false);
  const imgRef = useRef<HTMLImageElement>(null);
  const { observeImage } = useImageLazyLoad();

  // 生成不同尺寸的图片URL
  const generateSrcSet = useCallback((baseSrc: string) => {
    const sizes = [320, 640, 768, 1024, 1280, 1920];
    return sizes.map(size => `${baseSrc}?w=${size} ${size}w`).join(', ');
  }, []);

  useEffect(() => {
    if (lazy && imgRef.current) {
      observeImage(imgRef.current);
    }
  }, [lazy, observeImage]);

  const handleLoad = useCallback(() => {
    setIsLoaded(true);
  }, []);

  const handleError = useCallback(() => {
    setError(true);
  }, []);

  return (
    <div className={`responsive-image ${className || ''}`}>
      <img
        ref={imgRef}
        data-src={lazy ? src : undefined}
        src={lazy ? placeholder : src}
        srcSet={lazy ? undefined : generateSrcSet(src)}
        sizes={sizes}
        alt={alt}
        onLoad={handleLoad}
        onError={handleError}
        className={`${isLoaded ? 'loaded' : 'loading'} ${error ? 'error' : ''}`}
      />
      {!isLoaded && !error && (
        <div className="image-skeleton">
          <div className="skeleton-animation"></div>
        </div>
      )}
    </div>
  );
});
```

---

## 🌐 网络性能优化

### 请求优化策略

```typescript
// 1. 请求去重
class RequestDeduplicator {
  private static pendingRequests = new Map<string, Promise<any>>();

  static async request<T>(
    key: string,
    requestFn: () => Promise<T>
  ): Promise<T> {
    if (this.pendingRequests.has(key)) {
      return this.pendingRequests.get(key)!;
    }

    const promise = requestFn().finally(() => {
      this.pendingRequests.delete(key);
    });

    this.pendingRequests.set(key, promise);
    return promise;
  }
}

// 2. 请求缓存
class RequestCache {
  private static cache = new Map<
    string,
    { data: any; timestamp: number; ttl: number }
  >();

  static set(key: string, data: any, ttl: number = 5 * 60 * 1000) {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  static get(key: string): any | null {
    const cached = this.cache.get(key);

    if (!cached) return null;

    if (Date.now() - cached.timestamp > cached.ttl) {
      this.cache.delete(key);
      return null;
    }

    return cached.data;
  }

  static clear() {
    this.cache.clear();
  }
}

// 3. 智能请求Hook
function useSmartRequest<T>(
  key: string,
  requestFn: () => Promise<T>,
  options: {
    cache?: boolean;
    dedupe?: boolean;
    retry?: number;
    retryDelay?: number;
  } = {}
) {
  const { cache = true, dedupe = true, retry = 3, retryDelay = 1000 } = options;
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    // 检查缓存
    if (cache) {
      const cached = RequestCache.get(key);
      if (cached) {
        setData(cached);
        return cached;
      }
    }

    setLoading(true);
    setError(null);

    const executeRequest = async (attempt: number = 1): Promise<T> => {
      try {
        const result = dedupe
          ? await RequestDeduplicator.request(key, requestFn)
          : await requestFn();

        if (cache) {
          RequestCache.set(key, result);
        }

        setData(result);
        setLoading(false);
        return result;
      } catch (err) {
        if (attempt < retry) {
          await new Promise(resolve =>
            setTimeout(resolve, retryDelay * attempt)
          );
          return executeRequest(attempt + 1);
        }

        const error = err as Error;
        setError(error);
        setLoading(false);
        throw error;
      }
    };

    return executeRequest();
  }, [key, requestFn, cache, dedupe, retry, retryDelay]);

  return { data, loading, error, execute };
}

// 4. 批量请求优化
class BatchRequestManager {
  private static batches = new Map<
    string,
    {
      requests: Array<{ resolve: Function; reject: Function; params: any }>;
      timer: NodeJS.Timeout;
    }
  >();

  static addToBatch<T>(
    batchKey: string,
    params: any,
    batchFn: (paramsList: any[]) => Promise<T[]>,
    delay: number = 50
  ): Promise<T> {
    return new Promise((resolve, reject) => {
      let batch = this.batches.get(batchKey);

      if (!batch) {
        batch = {
          requests: [],
          timer: setTimeout(() => this.executeBatch(batchKey, batchFn), delay),
        };
        this.batches.set(batchKey, batch);
      }

      batch.requests.push({ resolve, reject, params });
    });
  }

  private static async executeBatch<T>(
    batchKey: string,
    batchFn: (paramsList: any[]) => Promise<T[]>
  ) {
    const batch = this.batches.get(batchKey);
    if (!batch) return;

    this.batches.delete(batchKey);

    try {
      const paramsList = batch.requests.map(req => req.params);
      const results = await batchFn(paramsList);

      batch.requests.forEach((req, index) => {
        req.resolve(results[index]);
      });
    } catch (error) {
      batch.requests.forEach(req => {
        req.reject(error);
      });
    }
  }
}

// 使用批量请求
function useBatchedProductFetch() {
  const fetchProduct = useCallback(async (productId: number) => {
    return BatchRequestManager.addToBatch(
      'products',
      productId,
      async (productIds: number[]) => {
        const response = await fetch('/api/products/batch', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ids: productIds }),
        });

        const data = await response.json();
        return data.products;
      }
    );
  }, []);

  return { fetchProduct };
}
```

---

## 🎯 面试常考知识点

### 1. React性能优化的核心原则

**Q: React性能优化的主要策略有哪些？**

**A: 核心策略：**

1. **减少不必要的渲染**
   - 使用React.memo、useMemo、useCallback
   - 避免在render中创建新对象/函数
   - 合理使用key属性

2. **代码分割和懒加载**
   - 路由级别的代码分割
   - 组件级别的动态导入
   - 资源的按需加载

3. **虚拟化长列表**
   - 使用react-window或react-virtualized
   - 实现无限滚动
   - 优化大数据渲染

4. **优化网络请求**
   - 请求去重和缓存
   - 批量请求
   - 预加载关键资源

```typescript
// 性能优化检查清单
const PerformanceChecklist = {
  rendering: [
    '使用React.memo包装纯组件',
    '使用useMemo缓存复杂计算',
    '使用useCallback缓存事件处理函数',
    '避免在render中创建新对象',
    '合理拆分组件，避免过大组件',
  ],

  codeSpitting: [
    '实现路由级代码分割',
    '使用React.lazy懒加载组件',
    '预加载关键路由',
    '按需加载第三方库',
  ],

  dataHandling: [
    '使用虚拟滚动处理长列表',
    '实现数据分页',
    '优化图片加载',
    '使用Web Worker处理大数据',
  ],

  network: ['实现请求缓存', '使用请求去重', '批量API调用', '预加载数据'],
};
```

### 2. useMemo vs useCallback

**Q: useMemo和useCallback的区别和使用场景？**

**A: 主要区别：**

```typescript
// useMemo - 缓存计算结果
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(a, b);
}, [a, b]);

// useCallback - 缓存函数引用
const memoizedCallback = useCallback(() => {
  doSomething(a, b);
}, [a, b]);

// 等价于
const memoizedCallback = useMemo(() => {
  return () => doSomething(a, b);
}, [a, b]);
```

**使用场景：**

| Hook            | 使用场景         | 示例               |
| --------------- | ---------------- | ------------------ |
| **useMemo**     | 缓存复杂计算结果 | 过滤/排序大数组    |
| **useCallback** | 缓存事件处理函数 | 传递给子组件的回调 |

### 3. 虚拟滚动的原理

**Q: 虚拟滚动是如何工作的？**

**A: 工作原理：**

1. **只渲染可见区域** - 只渲染视口内的元素
2. **动态计算位置** - 根据滚动位置计算应该渲染哪些元素
3. **复用DOM节点** - 重复使用有限的DOM节点
4. **维护滚动状态** - 保持正确的滚动条高度

```typescript
// 简化的虚拟滚动实现
function VirtualList({ items, itemHeight, containerHeight }) {
  const [scrollTop, setScrollTop] = useState(0);

  const visibleStart = Math.floor(scrollTop / itemHeight);
  const visibleEnd = Math.min(
    visibleStart + Math.ceil(containerHeight / itemHeight),
    items.length
  );

  const visibleItems = items.slice(visibleStart, visibleEnd);
  const totalHeight = items.length * itemHeight;
  const offsetY = visibleStart * itemHeight;

  return (
    <div
      style={{ height: containerHeight, overflow: 'auto' }}
      onScroll={(e) => setScrollTop(e.target.scrollTop)}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        <div style={{ transform: `translateY(${offsetY}px)` }}>
          {visibleItems.map((item, index) => (
            <div key={visibleStart + index} style={{ height: itemHeight }}>
              {item}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

---

## 📚 本章总结

通过本章学习，我们深入掌握了React性能优化的各种技巧：

### 🎯 核心收获

1. **渲染优化** 🎨
   - 掌握了React.memo、useMemo、useCallback的正确使用
   - 学会了避免不必要重渲染的策略
   - 理解了组件优化的设计原则

2. **大数据处理** 📋
   - 掌握了虚拟滚动的实现和应用
   - 学会了处理大数据集的优化技巧
   - 理解了Web Worker的使用场景

3. **代码分割** 🚀
   - 掌握了React.lazy和Suspense的使用
   - 学会了智能预加载策略
   - 理解了代码分割的最佳实践

4. **性能监控** 📊
   - 学会了内存泄漏的预防和检测
   - 掌握了网络请求的优化策略
   - 理解了性能指标的监控方法

### 🚀 技术进阶

- **下一步学习**: 实战项目架构设计
- **实践建议**: 在项目中应用性能优化技巧
- **深入方向**: React并发特性和Fiber架构

性能优化是一个持续的过程，需要在开发中不断实践和改进！ 🎉

---

_下一章我们将进入实战篇，学习《Next.js框架基础与SSR/SSG应用》！_ 🚀

```

```
