import { CartItem, Cart } from '@/types';

// 模拟购物车数据
export const mockCartItems: CartItem[] = [
  {
    id: 1,
    product_id: 1,
    sku_id: 101,
    product_name: 'iPhone 15 Pro Max 256GB 深空黑色',
    sku_name: '深空黑 256GB',
    price: '8999.00',
    quantity: 1,
    image: 'https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=400',
    selected: true,
  },
  {
    id: 2,
    product_id: 2,
    sku_id: 201,
    product_name: 'MacBook Pro 14英寸 M3芯片',
    sku_name: '银色 512GB',
    price: '13999.00',
    quantity: 1,
    image: 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400',
    selected: true,
  },
  {
    id: 3,
    product_id: 3,
    sku_id: 301,
    product_name: 'Nike Air Max 270 运动鞋',
    sku_name: '白色 42码',
    price: '699.00',
    quantity: 2,
    image: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400',
    selected: false,
  },
  {
    id: 4,
    product_id: 4,
    sku_id: 401,
    product_name: 'Adidas Ultraboost 22 跑步鞋',
    sku_name: '黑色 43码',
    price: '999.00',
    quantity: 1,
    image: 'https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=400',
    selected: true,
  },
  {
    id: 5,
    product_id: 9,
    sku_id: 901,
    product_name: 'Sony WH-1000XM5 无线降噪耳机',
    sku_name: '黑色',
    price: '1999.00',
    quantity: 1,
    image: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400',
    selected: true,
  },
];

// 计算购物车统计信息
export const calculateCartStats = (items: CartItem[]) => {
  const selectedItems = items.filter(item => item.selected);
  
  const total_quantity = selectedItems.reduce((sum, item) => sum + item.quantity, 0);
  const total_amount = selectedItems.reduce((sum, item) => {
    return sum + (parseFloat(item.price) * item.quantity);
  }, 0);

  return {
    items,
    total_quantity,
    total_amount: total_amount.toFixed(2),
    selected_count: selectedItems.length,
    total_count: items.length,
  };
};

// 模拟购物车API操作
export const mockCartAPI = {
  // 获取购物车
  getCart: async (): Promise<Cart> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    const stats = calculateCartStats(mockCartItems);
    return {
      items: stats.items,
      total_amount: stats.total_amount,
      total_quantity: stats.total_quantity,
    };
  },

  // 添加商品到购物车
  addToCart: async (productId: number, skuId: number, quantity: number): Promise<CartItem> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    
    // 查找是否已存在
    const existingIndex = mockCartItems.findIndex(
      item => item.product_id === productId && item.sku_id === skuId
    );

    if (existingIndex >= 0) {
      // 更新数量
      mockCartItems[existingIndex].quantity += quantity;
      return mockCartItems[existingIndex];
    } else {
      // 创建新商品项（这里简化处理，实际应该从商品数据获取）
      const newItem: CartItem = {
        id: Date.now(),
        product_id: productId,
        sku_id: skuId,
        product_name: `商品 ${productId}`,
        sku_name: `规格 ${skuId}`,
        price: '99.00',
        quantity,
        image: '/images/product-placeholder.svg',
        selected: true,
      };
      
      mockCartItems.push(newItem);
      return newItem;
    }
  },

  // 更新商品数量
  updateQuantity: async (itemId: number, quantity: number): Promise<CartItem> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    
    const item = mockCartItems.find(item => item.id === itemId);
    if (!item) {
      throw new Error('商品不存在');
    }

    if (quantity <= 0) {
      throw new Error('数量必须大于0');
    }

    item.quantity = quantity;
    return item;
  },

  // 删除商品
  removeItem: async (itemId: number): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    
    const index = mockCartItems.findIndex(item => item.id === itemId);
    if (index === -1) {
      throw new Error('商品不存在');
    }

    mockCartItems.splice(index, 1);
  },

  // 批量删除商品
  removeItems: async (itemIds: number[]): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    
    itemIds.forEach(itemId => {
      const index = mockCartItems.findIndex(item => item.id === itemId);
      if (index >= 0) {
        mockCartItems.splice(index, 1);
      }
    });
  },

  // 清空购物车
  clearCart: async (): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    mockCartItems.length = 0;
  },

  // 更新商品选中状态
  updateSelection: async (itemId: number, selected: boolean): Promise<CartItem> => {
    await new Promise(resolve => setTimeout(resolve, 100));
    
    const item = mockCartItems.find(item => item.id === itemId);
    if (!item) {
      throw new Error('商品不存在');
    }

    item.selected = selected;
    return item;
  },

  // 全选/取消全选
  updateAllSelection: async (selected: boolean): Promise<CartItem[]> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    
    mockCartItems.forEach(item => {
      item.selected = selected;
    });
    
    return mockCartItems;
  },

  // 同步购物车（登录后合并游客购物车）
  syncCart: async (guestItems: CartItem[]): Promise<Cart> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // 简化的合并逻辑
    guestItems.forEach(guestItem => {
      const existingIndex = mockCartItems.findIndex(
        item => item.product_id === guestItem.product_id && item.sku_id === guestItem.sku_id
      );

      if (existingIndex >= 0) {
        // 合并数量
        mockCartItems[existingIndex].quantity += guestItem.quantity;
      } else {
        // 添加新商品
        mockCartItems.push({
          ...guestItem,
          id: Date.now() + Math.random(),
        });
      }
    });

    const stats = calculateCartStats(mockCartItems);
    return {
      items: stats.items,
      total_amount: stats.total_amount,
      total_quantity: stats.total_quantity,
    };
  },
};

// 优惠券模拟数据
export const mockCoupons = [
  {
    id: 1,
    name: '新用户专享',
    description: '满99减10',
    discount_type: 'amount',
    discount_value: 10,
    min_amount: 99,
    valid_until: '2024-12-31',
    status: 'available',
  },
  {
    id: 2,
    name: '限时优惠',
    description: '满199减30',
    discount_type: 'amount',
    discount_value: 30,
    min_amount: 199,
    valid_until: '2024-12-31',
    status: 'available',
  },
  {
    id: 3,
    name: '会员专享',
    description: '9折优惠',
    discount_type: 'percent',
    discount_value: 10,
    min_amount: 0,
    valid_until: '2024-12-31',
    status: 'used',
  },
];

// 计算优惠后价格
export const calculateDiscount = (totalAmount: number, couponId?: number) => {
  if (!couponId) {
    return {
      original_amount: totalAmount,
      discount_amount: 0,
      final_amount: totalAmount,
      coupon: null,
    };
  }

  const coupon = mockCoupons.find(c => c.id === couponId && c.status === 'available');
  if (!coupon) {
    return {
      original_amount: totalAmount,
      discount_amount: 0,
      final_amount: totalAmount,
      coupon: null,
    };
  }

  // 检查最低消费
  if (totalAmount < coupon.min_amount) {
    return {
      original_amount: totalAmount,
      discount_amount: 0,
      final_amount: totalAmount,
      coupon: null,
      error: `需满${coupon.min_amount}元才能使用此优惠券`,
    };
  }

  let discount_amount = 0;
  if (coupon.discount_type === 'amount') {
    discount_amount = coupon.discount_value;
  } else if (coupon.discount_type === 'percent') {
    discount_amount = totalAmount * (coupon.discount_value / 100);
  }

  // 确保折扣不超过总金额
  discount_amount = Math.min(discount_amount, totalAmount);

  return {
    original_amount: totalAmount,
    discount_amount,
    final_amount: totalAmount - discount_amount,
    coupon,
  };
};

// 运费计算
export const calculateShipping = (totalAmount: number, items: CartItem[]) => {
  // 满99免运费
  if (totalAmount >= 99) {
    return {
      shipping_fee: 0,
      free_shipping: true,
      free_shipping_threshold: 99,
      remaining_for_free: 0,
    };
  }

  // 基础运费10元
  const shipping_fee = 10;
  const remaining_for_free = 99 - totalAmount;

  return {
    shipping_fee,
    free_shipping: false,
    free_shipping_threshold: 99,
    remaining_for_free: Math.max(0, remaining_for_free),
  };
};
