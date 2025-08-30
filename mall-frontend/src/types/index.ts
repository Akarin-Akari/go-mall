// 通用类型定义
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

export interface PageResult<T = any> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
  keyword?: string;
}

// 用户相关类型
export interface User {
  id: number;
  username: string;
  email: string;
  nickname: string;
  avatar?: string;
  phone?: string;
  role: string;
  status: 'active' | 'inactive' | 'locked';
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
  remember?: boolean;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  nickname: string;
  phone?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
}

// 商品相关类型
export interface Product {
  id: number;
  name: string;
  description: string;
  price: string;
  original_price?: string;
  stock: number;
  sold_count: number;
  category_id: number;
  category_name?: string;
  images: string[];
  status: 'active' | 'inactive' | 'draft';
  created_at: string;
  updated_at: string;
}

export interface ProductSKU {
  id: number;
  product_id: number;
  sku_code: string;
  name: string;
  price: string;
  stock: number;
  image?: string;
  attributes: Record<string, string>;
  status: 'active' | 'inactive';
}

export interface Category {
  id: number;
  name: string;
  description?: string;
  parent_id?: number;
  sort_order: number;
  status: 'active' | 'inactive';
  children?: Category[];
}

// 购物车相关类型
export interface CartItem {
  id: number;
  product_id: number;
  sku_id?: number;
  product_name: string;
  sku_name?: string;
  price: string;
  quantity: number;
  image?: string;
  selected: boolean;
}

export interface Cart {
  items: CartItem[];
  total_amount: string;
  total_quantity: number;
}

// 订单相关类型
export interface Order {
  id: number;
  order_no: string;
  user_id: number;
  status: 'pending' | 'paid' | 'shipped' | 'delivered' | 'completed' | 'cancelled';
  payment_status: 'pending' | 'paid' | 'failed' | 'refunded';
  shipping_status: 'pending' | 'shipped' | 'delivered';
  total_amount: string;
  payable_amount: string;
  paid_amount: string;
  shipping_address: Address;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: number;
  order_id: number;
  product_id: number;
  sku_id?: number;
  product_name: string;
  sku_name?: string;
  price: string;
  quantity: number;
  total_amount: string;
  image?: string;
}

export interface Address {
  id?: number;
  user_id?: number;
  name: string;
  phone: string;
  province: string;
  city: string;
  district: string;
  detail: string;
  postal_code?: string;
  is_default: boolean;
}

// 支付相关类型
export interface Payment {
  id: number;
  order_id: number;
  payment_no: string;
  payment_method: 'alipay' | 'wechat' | 'balance' | 'unionpay';
  amount: string;
  status: 'pending' | 'success' | 'failed' | 'cancelled';
  third_party_id?: string;
  created_at: string;
  updated_at: string;
}

export interface PaymentRequest {
  order_id: number;
  payment_method: string;
  return_url?: string;
  notify_url?: string;
}

// 文件上传相关类型
export interface UploadFile {
  uid: string;
  name: string;
  status: 'uploading' | 'done' | 'error';
  url?: string;
  response?: any;
}

// 通用组件Props类型
export interface BaseComponentProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

// 表单相关类型
export interface FormFieldError {
  field: string;
  message: string;
}

export interface FormState<T = any> {
  values: T;
  errors: FormFieldError[];
  loading: boolean;
  touched: Record<string, boolean>;
}

// 路由相关类型
export interface RouteConfig {
  path: string;
  component: React.ComponentType;
  exact?: boolean;
  auth?: boolean;
  roles?: string[];
  title?: string;
  icon?: string;
}

// 菜单相关类型
export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
  roles?: string[];
}

// 主题相关类型
export interface ThemeConfig {
  primaryColor: string;
  borderRadius: number;
  colorBgContainer: string;
}

// 应用状态类型
export interface AppState {
  theme: 'light' | 'dark';
  collapsed: boolean;
  loading: boolean;
  locale: 'zh-CN' | 'en-US';
}

// 错误类型
export interface AppError {
  code: string;
  message: string;
  details?: any;
}

// 环境变量类型
export interface EnvConfig {
  API_BASE_URL: string;
  APP_NAME: string;
  APP_VERSION: string;
  DEBUG: boolean;
  LOG_LEVEL: string;
}

// 外卖相关类型（为未来扩展准备）
export interface Restaurant {
  id: number;
  name: string;
  description: string;
  address: string;
  phone: string;
  rating: number;
  delivery_fee: string;
  min_order_amount: string;
  delivery_time: string;
  status: 'open' | 'closed' | 'busy';
  images: string[];
  categories: Category[];
}

export interface DeliveryAddress extends Address {
  latitude?: number;
  longitude?: number;
  distance?: number;
}

// React Native兼容类型（为跨平台准备）
export interface PlatformConfig {
  isWeb: boolean;
  isMobile: boolean;
  isNative: boolean;
  platform: 'web' | 'ios' | 'android';
}
