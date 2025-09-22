// 閫氱敤绫诲瀷瀹氫箟
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

// 鐢ㄦ埛鐩稿叧绫诲瀷
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
  email: string;
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
  error: string | null;
}

// 鍟嗗搧鐩稿叧绫诲瀷
export interface Product {
  id: number;
  name: string;
  description: string;
  price: string;
  discount_price?: string; // 鎶樻墸浠锋牸
  original_price?: string; // 鍘熶环锛堝吋瀹癸級
  stock: number;
  sold_count?: number; // 宸插敭鏁伴噺锛堝吋瀹癸級
  sales_count?: number; // 閿€鍞暟閲?
  rating?: number; // 璇勫垎
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
  icon?: string;
  product_count?: number;
  children?: Category[];
}

// 璐墿杞︾浉鍏崇被鍨?
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

// 璁㈠崟鐩稿叧绫诲瀷
export interface Order {
  id: number;
  order_no: string;
  user_id: number;
  status:
    | 'pending'
    | 'paid'
    | 'shipped'
    | 'delivered'
    | 'completed'
    | 'cancelled';
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

export interface CreateOrderRequest {
  cart_item_ids: number[];
  receiver_name: string;
  receiver_phone: string;
  receiver_address: string;
  province: string;
  city: string;
  district: string;
  shipping_method: string;
  payment_method: string;
  buyer_message?: string;
}

// 鏀粯鐩稿叧绫诲瀷
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

// 鏂囦欢涓婁紶鐩稿叧绫诲瀷
export interface UploadFile {
  uid: string;
  name: string;
  status: 'uploading' | 'done' | 'error';
  url?: string;
  response?: any;
}

// 閫氱敤缁勪欢Props绫诲瀷
export interface BaseComponentProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

// 琛ㄥ崟鐩稿叧绫诲瀷
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

// 璺敱鐩稿叧绫诲瀷
export interface RouteConfig {
  path: string;
  component: React.ComponentType;
  exact?: boolean;
  auth?: boolean;
  roles?: string[];
  title?: string;
  icon?: string;
}

// 鑿滃崟鐩稿叧绫诲瀷
export interface MenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  path?: string;
  children?: MenuItem[];
  roles?: string[];
}

// 涓婚鐩稿叧绫诲瀷
export interface ThemeConfig {
  primaryColor: string;
  borderRadius: number;
  colorBgContainer: string;
}

// 搴旂敤鐘舵€佺被鍨?
export interface AppState {
  theme: 'light' | 'dark';
  collapsed: boolean;
  loading: boolean;
  locale: 'zh-CN' | 'en-US';
}

// 閿欒绫诲瀷
export interface AppError {
  code: string;
  message: string;
  details?: any;
}

// 鐜鍙橀噺绫诲瀷
export interface EnvConfig {
  API_BASE_URL: string;
  APP_NAME: string;
  APP_VERSION: string;
  DEBUG: boolean;
  LOG_LEVEL: string;
}

// 澶栧崠鐩稿叧绫诲瀷锛堜负鏈潵鎵╁睍鍑嗗锛?
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

// React Native鍏煎绫诲瀷锛堜负璺ㄥ钩鍙板噯澶囷級
export interface PlatformConfig {
  isWeb: boolean;
  isMobile: boolean;
  isNative: boolean;
  platform: 'web' | 'ios' | 'android';
}
