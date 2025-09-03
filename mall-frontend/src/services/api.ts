import { http } from '@/utils/request';
import { API_ENDPOINTS } from '@/constants';
import {
  ApiResponse,
  PageResult,
  PaginationParams,
  User,
  LoginRequest,
  RegisterRequest,
  Product,
  Category,
  CartItem,
  Cart,
  Order,
  OrderItem,
  Address,
  Payment,
  PaymentRequest,
} from '@/types';

// 认证相关API
export const authAPI = {
  // 用户登录
  login: (data: LoginRequest): Promise<ApiResponse<{
    user: User;
    token: string;
    refresh_token: string;
  }>> => {
    return http.post(API_ENDPOINTS.AUTH.LOGIN, data, {
      skipAuth: true,
      showSuccessMessage: true,
      successMessage: '登录成功',
    });
  },

  // 用户注册
  register: (data: RegisterRequest): Promise<ApiResponse<{
    user: User;
    token: string;
    refresh_token: string;
  }>> => {
    return http.post(API_ENDPOINTS.AUTH.REGISTER, data, {
      skipAuth: true,
      showSuccessMessage: true,
      successMessage: '注册成功',
    });
  },

  // 用户登出
  logout: (): Promise<ApiResponse<null>> => {
    return http.post(API_ENDPOINTS.AUTH.LOGOUT, {}, {
      showSuccessMessage: true,
      successMessage: '登出成功',
    });
  },

  // 获取用户信息
  getProfile: (): Promise<ApiResponse<User>> => {
    return http.get(API_ENDPOINTS.AUTH.PROFILE);
  },

  // 刷新token
  refreshToken: (refreshToken: string): Promise<ApiResponse<{
    token: string;
    refresh_token: string;
  }>> => {
    return http.post(API_ENDPOINTS.AUTH.REFRESH_TOKEN, {
      refresh_token: refreshToken,
    }, {
      skipAuth: true,
    });
  },
};

// 用户管理API
export const userAPI = {
  // 获取用户列表
  getUsers: (params: PaginationParams): Promise<ApiResponse<PageResult<User>>> => {
    return http.get(API_ENDPOINTS.USERS.LIST, { params });
  },

  // 获取用户详情
  getUserDetail: (id: number): Promise<ApiResponse<User>> => {
    return http.get(API_ENDPOINTS.USERS.DETAIL(id));
  },

  // 更新用户信息
  updateUser: (id: number, data: Partial<User>): Promise<ApiResponse<User>> => {
    return http.put(API_ENDPOINTS.USERS.UPDATE(id), data, {
      showSuccessMessage: true,
      successMessage: '用户信息更新成功',
    });
  },

  // 删除用户
  deleteUser: (id: number): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.USERS.DELETE(id), {
      showSuccessMessage: true,
      successMessage: '用户删除成功',
    });
  },
};

// 商品相关API
export const productAPI = {
  // 获取商品列表
  getProducts: (params: PaginationParams & {
    category_id?: number;
    status?: string;
    min_price?: number;
    max_price?: number;
  }): Promise<ApiResponse<PageResult<Product>>> => {
    return http.get(API_ENDPOINTS.PRODUCTS.LIST, { params });
  },

  // 获取商品详情
  getProductDetail: (id: number): Promise<ApiResponse<Product>> => {
    return http.get(API_ENDPOINTS.PRODUCTS.DETAIL(id));
  },

  // 创建商品
  createProduct: (data: Omit<Product, 'id' | 'created_at' | 'updated_at'>): Promise<ApiResponse<Product>> => {
    return http.post(API_ENDPOINTS.PRODUCTS.CREATE, data, {
      showSuccessMessage: true,
      successMessage: '商品创建成功',
    });
  },

  // 更新商品
  updateProduct: (id: number, data: Partial<Product>): Promise<ApiResponse<Product>> => {
    return http.put(API_ENDPOINTS.PRODUCTS.UPDATE(id), data, {
      showSuccessMessage: true,
      successMessage: '商品更新成功',
    });
  },

  // 删除商品
  deleteProduct: (id: number): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.PRODUCTS.DELETE(id), {
      showSuccessMessage: true,
      successMessage: '商品删除成功',
    });
  },

  // 获取商品分类
  getCategories: (): Promise<ApiResponse<Category[]>> => {
    return http.get(API_ENDPOINTS.PRODUCTS.CATEGORIES);
  },
};

// 购物车相关API
export const cartAPI = {
  // 获取购物车
  getCart: (): Promise<ApiResponse<Cart>> => {
    return http.get(API_ENDPOINTS.CART.LIST);
  },

  // 添加到购物车
  addToCart: (data: {
    product_id: number;
    sku_id?: number;
    quantity: number;
  }): Promise<ApiResponse<CartItem>> => {
    return http.post(API_ENDPOINTS.CART.ADD, data, {
      showSuccessMessage: true,
      successMessage: '已添加到购物车',
    });
  },

  // 更新购物车商品
  updateCartItem: (data: {
    id: number;
    quantity: number;
    selected?: boolean;
  }): Promise<ApiResponse<CartItem>> => {
    return http.put(API_ENDPOINTS.CART.UPDATE, data);
  },

  // 删除购物车商品
  removeCartItem: (id: number): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.CART.REMOVE, {
      data: { id },
      showSuccessMessage: true,
      successMessage: '商品已从购物车移除',
    });
  },

  // 清空购物车
  clearCart: (): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.CART.CLEAR, {
      showSuccessMessage: true,
      successMessage: '购物车已清空',
    });
  },

  // 同步购物车
  syncCart: (items: CartItem[]): Promise<ApiResponse<Cart>> => {
    return http.post(API_ENDPOINTS.CART.SYNC, { items });
  },
};

// 订单相关API
export const orderAPI = {
  // 获取订单列表
  getOrders: (params: PaginationParams & {
    status?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<ApiResponse<PageResult<Order>>> => {
    return http.get(API_ENDPOINTS.ORDERS.LIST, { params });
  },

  // 获取订单详情
  getOrderDetail: (id: number): Promise<ApiResponse<Order>> => {
    return http.get(API_ENDPOINTS.ORDERS.DETAIL(id));
  },

  // 创建订单
  createOrder: (data: {
    items: {
      product_id: number;
      sku_id?: number;
      quantity: number;
      price: string;
    }[];
    shipping_address: Address;
    remark?: string;
  }): Promise<ApiResponse<Order>> => {
    return http.post(API_ENDPOINTS.ORDERS.CREATE, data, {
      showSuccessMessage: true,
      successMessage: '订单创建成功',
    });
  },

  // 更新订单状态
  updateOrderStatus: (id: number, status: string, remark?: string): Promise<ApiResponse<Order>> => {
    return http.put(API_ENDPOINTS.ORDERS.UPDATE_STATUS(id), {
      status,
      remark,
    }, {
      showSuccessMessage: true,
      successMessage: '订单状态更新成功',
    });
  },

  // 取消订单
  cancelOrder: (id: number, reason?: string): Promise<ApiResponse<Order>> => {
    return http.put(API_ENDPOINTS.ORDERS.CANCEL(id), {
      reason,
    }, {
      showSuccessMessage: true,
      successMessage: '订单已取消',
    });
  },
};

// 支付相关API
export const paymentAPI = {
  // 创建支付
  createPayment: (data: PaymentRequest): Promise<ApiResponse<{
    payment_id: number;
    payment_url?: string;
    qr_code?: string;
  }>> => {
    return http.post(API_ENDPOINTS.PAYMENT.CREATE, data);
  },

  // 查询支付状态
  queryPayment: (id: number): Promise<ApiResponse<Payment>> => {
    return http.get(API_ENDPOINTS.PAYMENT.QUERY(id));
  },

  // 申请退款
  refundPayment: (data: {
    payment_id: number;
    amount: string;
    reason: string;
  }): Promise<ApiResponse<{
    refund_id: number;
    status: string;
  }>> => {
    return http.post(API_ENDPOINTS.PAYMENT.REFUND, data, {
      showSuccessMessage: true,
      successMessage: '退款申请已提交',
    });
  },
};

// 地址管理API
export const addressAPI = {
  // 获取地址列表
  getAddresses: (): Promise<ApiResponse<Address[]>> => {
    return http.get(API_ENDPOINTS.ADDRESS.LIST);
  },

  // 创建地址
  createAddress: (data: Omit<Address, 'id'>): Promise<ApiResponse<Address>> => {
    return http.post(API_ENDPOINTS.ADDRESS.CREATE, data, {
      showSuccessMessage: true,
      successMessage: '地址添加成功',
    });
  },

  // 更新地址
  updateAddress: (id: number, data: Partial<Address>): Promise<ApiResponse<Address>> => {
    return http.put(API_ENDPOINTS.ADDRESS.UPDATE(id), data, {
      showSuccessMessage: true,
      successMessage: '地址更新成功',
    });
  },

  // 删除地址
  deleteAddress: (id: number): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.ADDRESS.DELETE(id), {
      showSuccessMessage: true,
      successMessage: '地址删除成功',
    });
  },

  // 设置默认地址
  setDefaultAddress: (id: number): Promise<ApiResponse<Address>> => {
    return http.put(API_ENDPOINTS.ADDRESS.SET_DEFAULT(id), {}, {
      showSuccessMessage: true,
      successMessage: '默认地址设置成功',
    });
  },
};

// 文件上传API
export const uploadAPI = {
  // 上传图片
  uploadImage: (file: FormData): Promise<ApiResponse<{
    url: string;
    filename: string;
    size: number;
  }>> => {
    return http.upload(API_ENDPOINTS.UPLOAD.IMAGE, file);
  },

  // 上传文件
  uploadFile: (file: FormData): Promise<ApiResponse<{
    url: string;
    filename: string;
    size: number;
  }>> => {
    return http.upload(API_ENDPOINTS.UPLOAD.FILE, file);
  },

  // 删除文件
  deleteFile: (id: string): Promise<ApiResponse<null>> => {
    return http.delete(API_ENDPOINTS.UPLOAD.DELETE(id));
  },
};

// 统计相关API（管理后台使用）
export const statisticsAPI = {
  // 获取概览统计
  getOverview: (): Promise<ApiResponse<{
    total_users: number;
    total_products: number;
    total_orders: number;
    total_revenue: string;
    today_orders: number;
    today_revenue: string;
  }>> => {
    return http.get('/api/v1/statistics/overview');
  },

  // 获取销售统计
  getSalesStatistics: (params: {
    start_date: string;
    end_date: string;
    type: 'daily' | 'weekly' | 'monthly';
  }): Promise<ApiResponse<{
    labels: string[];
    data: number[];
  }>> => {
    return http.get('/api/v1/statistics/sales', { params });
  },

  // 获取商品销售排行
  getProductRanking: (params: {
    start_date: string;
    end_date: string;
    limit?: number;
  }): Promise<ApiResponse<{
    product_id: number;
    product_name: string;
    sales_count: number;
    sales_amount: string;
  }[]>> => {
    return http.get('/api/v1/statistics/products', { params });
  },
};

// 导出所有API
export const api = {
  auth: authAPI,
  user: userAPI,
  product: productAPI,
  cart: cartAPI,
  order: orderAPI,
  payment: paymentAPI,
  address: addressAPI,
  upload: uploadAPI,
  statistics: statisticsAPI,
};

export default api;
