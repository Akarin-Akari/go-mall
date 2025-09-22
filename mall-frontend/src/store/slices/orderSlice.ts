import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Order, PageResult, PaginationParams, Address } from '@/types';
import { orderAPI } from '@/services/api';
import { message } from 'antd';

// 订单状态接口
interface OrderState {
  // 订单列表
  orders: Order[];
  total: number;
  loading: boolean;

  // 当前订单详情
  currentOrder: Order | null;
  orderLoading: boolean;

  // 创建订单
  creating: boolean;

  // 搜索和筛选
  searchParams: {
    status?: string;
    start_date?: string;
    end_date?: string;
    page: number;
    page_size: number;
  };
}

// 初始状态
const initialState: OrderState = {
  orders: [],
  total: 0,
  loading: false,

  currentOrder: null,
  orderLoading: false,

  creating: false,

  searchParams: {
    page: 1,
    page_size: 10,
  },
};

// 异步actions
export const fetchOrdersAsync = createAsyncThunk(
  'order/fetchOrders',
  async (
    params: PaginationParams & {
      status?: string;
      start_date?: string;
      end_date?: string;
    },
    { rejectWithValue }
  ) => {
    try {
      const response = await orderAPI.getOrders(params);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取订单列表失败');
    }
  }
);

export const fetchOrderDetailAsync = createAsyncThunk(
  'order/fetchOrderDetail',
  async (id: number, { rejectWithValue }) => {
    try {
      const response = await orderAPI.getOrderDetail(id);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取订单详情失败');
    }
  }
);

export const createOrderAsync = createAsyncThunk(
  'order/createOrder',
  async (
    orderData: {
      items: {
        product_id: number;
        sku_id?: number;
        quantity: number;
        price: string;
      }[];
      shipping_address: Address;
      remark?: string;
    },
    { rejectWithValue }
  ) => {
    try {
      const response = await orderAPI.createOrder(orderData);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建订单失败');
    }
  }
);

export const updateOrderStatusAsync = createAsyncThunk(
  'order/updateOrderStatus',
  async (
    { id, status, remark }: { id: number; status: string; remark?: string },
    { rejectWithValue }
  ) => {
    try {
      const response = await orderAPI.updateOrderStatus(id, status, remark);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新订单状态失败');
    }
  }
);

export const cancelOrderAsync = createAsyncThunk(
  'order/cancelOrder',
  async (
    { id, reason }: { id: number; reason?: string },
    { rejectWithValue }
  ) => {
    try {
      const response = await orderAPI.cancelOrder(id, reason);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '取消订单失败');
    }
  }
);

// 创建slice
const orderSlice = createSlice({
  name: 'order',
  initialState,
  reducers: {
    // 设置搜索参数
    setSearchParams: (
      state,
      action: PayloadAction<Partial<OrderState['searchParams']>>
    ) => {
      state.searchParams = { ...state.searchParams, ...action.payload };
    },

    // 重置搜索参数
    resetSearchParams: state => {
      state.searchParams = {
        page: 1,
        page_size: 10,
      };
    },

    // 清除当前订单
    clearCurrentOrder: state => {
      state.currentOrder = null;
    },

    // 更新订单状态（本地更新）
    updateOrderStatusLocal: (
      state,
      action: PayloadAction<{ id: number; status: string }>
    ) => {
      const order = state.orders.find(o => o.id === action.payload.id);
      if (order) {
        order.status = action.payload.status as any;
      }

      if (state.currentOrder && state.currentOrder.id === action.payload.id) {
        state.currentOrder.status = action.payload.status as any;
      }
    },
  },
  extraReducers: builder => {
    // 获取订单列表
    builder
      .addCase(fetchOrdersAsync.pending, state => {
        state.loading = true;
      })
      .addCase(fetchOrdersAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.orders = action.payload.list;
        state.total = action.payload.total;
      })
      .addCase(fetchOrdersAsync.rejected, (state, action) => {
        state.loading = false;
        message.error(action.payload as string);
      });

    // 获取订单详情
    builder
      .addCase(fetchOrderDetailAsync.pending, state => {
        state.orderLoading = true;
      })
      .addCase(fetchOrderDetailAsync.fulfilled, (state, action) => {
        state.orderLoading = false;
        state.currentOrder = action.payload;
      })
      .addCase(fetchOrderDetailAsync.rejected, (state, action) => {
        state.orderLoading = false;
        state.currentOrder = null;
        message.error(action.payload as string);
      });

    // 创建订单
    builder
      .addCase(createOrderAsync.pending, state => {
        state.creating = true;
      })
      .addCase(createOrderAsync.fulfilled, (state, action) => {
        state.creating = false;
        state.orders.unshift(action.payload);
        state.total += 1;
        state.currentOrder = action.payload;
      })
      .addCase(createOrderAsync.rejected, (state, action) => {
        state.creating = false;
        message.error(action.payload as string);
      });

    // 更新订单状态
    builder
      .addCase(updateOrderStatusAsync.fulfilled, (state, action) => {
        const index = state.orders.findIndex(o => o.id === action.payload.id);
        if (index !== -1) {
          state.orders[index] = action.payload;
        }

        if (state.currentOrder && state.currentOrder.id === action.payload.id) {
          state.currentOrder = action.payload;
        }
      })
      .addCase(updateOrderStatusAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });

    // 取消订单
    builder
      .addCase(cancelOrderAsync.fulfilled, (state, action) => {
        const index = state.orders.findIndex(o => o.id === action.payload.id);
        if (index !== -1) {
          state.orders[index] = action.payload;
        }

        if (state.currentOrder && state.currentOrder.id === action.payload.id) {
          state.currentOrder = action.payload;
        }
      })
      .addCase(cancelOrderAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });
  },
});

// 导出actions
export const {
  setSearchParams,
  resetSearchParams,
  clearCurrentOrder,
  updateOrderStatusLocal,
} = orderSlice.actions;

// 选择器
export const selectOrder = (state: { order: OrderState }) => state.order;
export const selectOrders = (state: { order: OrderState }) =>
  state.order.orders;
export const selectCurrentOrder = (state: { order: OrderState }) =>
  state.order.currentOrder;
export const selectOrderLoading = (state: { order: OrderState }) =>
  state.order.loading;
export const selectOrderDetailLoading = (state: { order: OrderState }) =>
  state.order.orderLoading;
export const selectOrderCreating = (state: { order: OrderState }) =>
  state.order.creating;
export const selectOrderSearchParams = (state: { order: OrderState }) =>
  state.order.searchParams;
export const selectOrderTotal = (state: { order: OrderState }) =>
  state.order.total;

// 导出reducer
export default orderSlice.reducer;
