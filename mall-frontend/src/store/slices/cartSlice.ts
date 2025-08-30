import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { CartItem, Cart } from '@/types';
import { cartAPI } from '@/services/api';
import { message } from 'antd';

// 购物车状态接口
interface CartState {
  items: CartItem[];
  total_amount: string;
  total_quantity: number;
  loading: boolean;
  syncing: boolean;
}

// 初始状态
const initialState: CartState = {
  items: [],
  total_amount: '0.00',
  total_quantity: 0,
  loading: false,
  syncing: false,
};

// 异步actions
export const fetchCartAsync = createAsyncThunk(
  'cart/fetchCart',
  async (_, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockCartAPI } = await import('@/data/mockCart');
        return await mockCartAPI.getCart();
      }

      // 生产环境使用真实API
      const response = await cartAPI.getCart();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取购物车失败');
    }
  }
);

export const addToCartAsync = createAsyncThunk(
  'cart/addToCart',
  async (data: {
    product_id: number;
    sku_id?: number;
    quantity: number;
  }, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockCartAPI } = await import('@/data/mockCart');
        return await mockCartAPI.addToCart(params.product_id, params.sku_id || 0, params.quantity);
      }

      // 生产环境使用真实API
      const response = await cartAPI.addToCart(data);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '添加到购物车失败');
    }
  }
);

export const updateCartItemAsync = createAsyncThunk(
  'cart/updateCartItem',
  async (data: {
    id: number;
    quantity: number;
    selected?: boolean;
  }, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockCartAPI } = await import('@/data/mockCart');
        if (data.quantity !== undefined) {
          return await mockCartAPI.updateQuantity(data.id, data.quantity);
        }
        if (data.selected !== undefined) {
          return await mockCartAPI.updateSelection(data.id, data.selected);
        }
      }

      // 生产环境使用真实API
      const response = await cartAPI.updateCartItem(data);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新购物车失败');
    }
  }
);

export const removeCartItemAsync = createAsyncThunk(
  'cart/removeCartItem',
  async (id: number, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockCartAPI } = await import('@/data/mockCart');
        await mockCartAPI.removeItem(id);
        return id;
      }

      // 生产环境使用真实API
      await cartAPI.removeCartItem(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除商品失败');
    }
  }
);

export const clearCartAsync = createAsyncThunk(
  'cart/clearCart',
  async (_, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockCartAPI } = await import('@/data/mockCart');
        await mockCartAPI.clearCart();
        return null;
      }

      // 生产环境使用真实API
      await cartAPI.clearCart();
      return null;
    } catch (error: any) {
      return rejectWithValue(error.message || '清空购物车失败');
    }
  }
);

export const syncCartAsync = createAsyncThunk(
  'cart/syncCart',
  async (items: CartItem[], { rejectWithValue }) => {
    try {
      const response = await cartAPI.syncCart(items);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '同步购物车失败');
    }
  }
);

// 创建slice
const cartSlice = createSlice({
  name: 'cart',
  initialState,
  reducers: {
    // 本地添加商品到购物车
    addItemLocal: (state, action: PayloadAction<CartItem>) => {
      const existingItem = state.items.find(
        item => item.product_id === action.payload.product_id && 
                item.sku_id === action.payload.sku_id
      );
      
      if (existingItem) {
        existingItem.quantity += action.payload.quantity;
      } else {
        state.items.push(action.payload);
      }
      
      cartSlice.caseReducers.calculateTotals(state);
    },
    
    // 本地更新购物车商品
    updateItemLocal: (state, action: PayloadAction<{
      id: number;
      quantity?: number;
      selected?: boolean;
    }>) => {
      const item = state.items.find(item => item.id === action.payload.id);
      if (item) {
        if (action.payload.quantity !== undefined) {
          item.quantity = action.payload.quantity;
        }
        if (action.payload.selected !== undefined) {
          item.selected = action.payload.selected;
        }
      }
      
      cartSlice.caseReducers.calculateTotals(state);
    },
    
    // 本地删除购物车商品
    removeItemLocal: (state, action: PayloadAction<number>) => {
      state.items = state.items.filter(item => item.id !== action.payload);
      cartSlice.caseReducers.calculateTotals(state);
    },
    
    // 本地清空购物车
    clearCartLocal: (state) => {
      state.items = [];
      state.total_amount = '0.00';
      state.total_quantity = 0;
    },
    
    // 全选/取消全选
    toggleSelectAll: (state, action: PayloadAction<boolean>) => {
      state.items.forEach(item => {
        item.selected = action.payload;
      });
      cartSlice.caseReducers.calculateTotals(state);
    },
    
    // 计算总计
    calculateTotals: (state) => {
      const selectedItems = state.items.filter(item => item.selected);
      
      state.total_quantity = selectedItems.reduce((total, item) => total + item.quantity, 0);
      
      const totalAmount = selectedItems.reduce((total, item) => {
        return total + (parseFloat(item.price) * item.quantity);
      }, 0);
      
      state.total_amount = totalAmount.toFixed(2);
    },
    
    // 设置购物车数据
    setCartData: (state, action: PayloadAction<Cart>) => {
      state.items = action.payload.items;
      state.total_amount = action.payload.total_amount;
      state.total_quantity = action.payload.total_quantity;
    },
  },
  extraReducers: (builder) => {
    // 获取购物车
    builder
      .addCase(fetchCartAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchCartAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.items = action.payload.items;
        state.total_amount = action.payload.total_amount;
        state.total_quantity = action.payload.total_quantity;
      })
      .addCase(fetchCartAsync.rejected, (state, action) => {
        state.loading = false;
        message.error(action.payload as string);
      });
    
    // 添加到购物车
    builder
      .addCase(addToCartAsync.pending, (state) => {
        state.loading = true;
      })
      .addCase(addToCartAsync.fulfilled, (state, action) => {
        state.loading = false;
        
        const existingItem = state.items.find(
          item => item.product_id === action.payload.product_id && 
                  item.sku_id === action.payload.sku_id
        );
        
        if (existingItem) {
          existingItem.quantity += action.payload.quantity;
        } else {
          state.items.push(action.payload);
        }
        
        cartSlice.caseReducers.calculateTotals(state);
      })
      .addCase(addToCartAsync.rejected, (state, action) => {
        state.loading = false;
        message.error(action.payload as string);
      });
    
    // 更新购物车商品
    builder
      .addCase(updateCartItemAsync.fulfilled, (state, action) => {
        const item = state.items.find(item => item.id === action.payload.id);
        if (item) {
          Object.assign(item, action.payload);
        }
        cartSlice.caseReducers.calculateTotals(state);
      })
      .addCase(updateCartItemAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });
    
    // 删除购物车商品
    builder
      .addCase(removeCartItemAsync.fulfilled, (state, action) => {
        state.items = state.items.filter(item => item.id !== action.payload);
        cartSlice.caseReducers.calculateTotals(state);
      })
      .addCase(removeCartItemAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });
    
    // 清空购物车
    builder
      .addCase(clearCartAsync.fulfilled, (state) => {
        state.items = [];
        state.total_amount = '0.00';
        state.total_quantity = 0;
      })
      .addCase(clearCartAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });
    
    // 同步购物车
    builder
      .addCase(syncCartAsync.pending, (state) => {
        state.syncing = true;
      })
      .addCase(syncCartAsync.fulfilled, (state, action) => {
        state.syncing = false;
        state.items = action.payload.items;
        state.total_amount = action.payload.total_amount;
        state.total_quantity = action.payload.total_quantity;
      })
      .addCase(syncCartAsync.rejected, (state, action) => {
        state.syncing = false;
        console.error('购物车同步失败:', action.payload);
      });
  },
});

// 导出actions
export const {
  addItemLocal,
  updateItemLocal,
  removeItemLocal,
  clearCartLocal,
  toggleSelectAll,
  calculateTotals,
  setCartData,
} = cartSlice.actions;

// 选择器
export const selectCart = (state: { cart: CartState }) => state.cart;
export const selectCartItems = (state: { cart: CartState }) => state.cart.items;
export const selectSelectedCartItems = (state: { cart: CartState }) => 
  state.cart.items.filter(item => item.selected);
export const selectCartTotal = (state: { cart: CartState }) => ({
  amount: state.cart.total_amount,
  quantity: state.cart.total_quantity,
});
export const selectCartLoading = (state: { cart: CartState }) => state.cart.loading;
export const selectCartItemCount = (state: { cart: CartState }) => 
  state.cart.items.reduce((total, item) => total + item.quantity, 0);

// 导出reducer
export default cartSlice.reducer;
