import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import ProductCard from '../business/ProductCard';
import { Product } from '@/types';

// Mock store configuration
const createMockStore = (initialState = {}) => {
  return configureStore({
    reducer: {
      cart: (state = { items: [] }) => state,
      auth: (state = { isAuthenticated: true }) => state,
    },
    preloadedState: initialState,
  });
};

// Mock product data
const mockProduct: Product = {
  id: 1,
  name: '测试商品',
  price: '99.99',
  discount_price: '89.99',
  description: '这是一个测试商品',
  images: ['/test-image.jpg'],
  stock: 10,
  sales_count: 500,
  rating: 4.5,
  category_id: 1,
  status: 'active',
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
};

describe('ProductCard组件测试', () => {
  let mockStore: ReturnType<typeof createMockStore>;
  let mockOnAddToCart: jest.Mock;
  let mockOnViewDetail: jest.Mock;

  beforeEach(() => {
    mockStore = createMockStore();
    mockOnAddToCart = jest.fn();
    mockOnViewDetail = jest.fn();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('应该正确渲染商品信息', () => {
    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    expect(screen.getByText('测试商品')).toBeInTheDocument();
    expect(screen.getByText('¥89.99')).toBeInTheDocument(); // 显示折扣价
    expect(screen.getByText('¥99.99')).toBeInTheDocument(); // 显示原价（删除线）
    expect(screen.getByText('这是一个测试商品')).toBeInTheDocument();
  });

  test('应该显示商品图片', () => {
    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    const image = screen.getByAltText('测试商品');
    expect(image).toBeInTheDocument();
    expect(image).toHaveAttribute('src', '/test-image.jpg');
  });

  test('点击商品卡片应该调用onViewDetail回调', async () => {
    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    // 点击卡片主体区域
    const cardElement = screen.getByText('测试商品').closest('.ant-card');
    expect(cardElement).toBeInTheDocument();

    fireEvent.click(cardElement!);

    await waitFor(() => {
      expect(mockOnViewDetail).toHaveBeenCalledWith(1);
    });
  });

  test('点击添加购物车按钮应该调用onAddToCart回调', async () => {
    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    // 通过购物车图标查找按钮
    const addButton = screen.getByRole('button', { name: 'shopping-cart' });
    fireEvent.click(addButton);

    await waitFor(() => {
      expect(mockOnAddToCart).toHaveBeenCalledWith(mockProduct);
    });
  });

  test('应该显示自定义Badge', () => {
    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
          showBadge='热销'
          badgeColor='#ff4d4f'
        />
      </Provider>
    );

    expect(screen.getByText('热销')).toBeInTheDocument();
  });

  test('库存为0时添加购物车按钮应该正常显示', () => {
    const outOfStockProduct = { ...mockProduct, stock: 0 };

    render(
      <Provider store={mockStore}>
        <ProductCard
          product={outOfStockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    // 组件没有内置缺货逻辑，按钮应该正常显示
    const addButton = screen.getByRole('button', { name: 'shopping-cart' });
    expect(addButton).toBeInTheDocument();
    expect(addButton).not.toBeDisabled();
  });

  test('应该正确处理收藏功能', async () => {
    const mockOnToggleFavorite = jest.fn();

    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={mockOnAddToCart}
          onViewDetail={mockOnViewDetail}
          onToggleFavorite={mockOnToggleFavorite}
        />
      </Provider>
    );

    // 获取所有心形按钮，选择底部操作区域的那个（第二个）
    const favoriteButtons = screen.getAllByRole('button', { name: 'heart' });
    expect(favoriteButtons).toHaveLength(2);

    // 点击底部操作区域的收藏按钮
    fireEvent.click(favoriteButtons[1]);

    // 验证回调被调用
    await waitFor(() => {
      expect(mockOnToggleFavorite).toHaveBeenCalledWith(mockProduct);
    });
  });

  test('应该正确处理加载状态', async () => {
    const slowOnAddToCart = jest.fn(
      () => new Promise(resolve => setTimeout(resolve, 50))
    );

    render(
      <Provider store={mockStore}>
        <ProductCard
          product={mockProduct}
          onAddToCart={slowOnAddToCart}
          onViewDetail={mockOnViewDetail}
        />
      </Provider>
    );

    const addButton = screen.getByRole('button', { name: 'shopping-cart' });
    fireEvent.click(addButton);

    // 验证按钮进入加载状态
    await waitFor(() => {
      expect(addButton).toHaveClass('ant-btn-loading');
    });

    // 验证回调被调用
    expect(slowOnAddToCart).toHaveBeenCalledWith(mockProduct);

    // 等待加载完成
    await waitFor(
      () => {
        expect(addButton).not.toHaveClass('ant-btn-loading');
      },
      { timeout: 1000 }
    );
  });
});
