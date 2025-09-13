# 第3章：状态管理策略与最佳实践 🗃️

> *"选择合适的状态管理方案，是构建可维护大型应用的关键！"* 🎯

## 📚 本章导览

在现代React应用中，状态管理是一个核心话题。随着应用复杂度的增加，如何高效地管理状态、保持数据一致性、优化性能成为了关键挑战。本章将深入探讨各种状态管理策略，从React内置的状态管理到第三方解决方案，帮你在Mall-Frontend项目中做出最佳选择。

### 🎯 学习目标

通过本章学习，你将掌握：

- **状态管理基础** - 理解不同类型的状态和管理策略
- **React内置方案** - useState、useReducer、Context API的深度应用
- **Redux Toolkit** - 现代Redux的最佳实践
- **Zustand轻量方案** - 简单高效的状态管理
- **React Query** - 服务端状态管理的最佳选择
- **状态设计模式** - 状态规范化、派生状态等高级概念
- **性能优化** - 避免不必要的重渲染和状态更新
- **实战应用** - 在Mall-Frontend中的状态管理架构

### 🛠️ 技术栈概览

```typescript
{
  "react": "React 19.1.0",
  "stateManagement": ["Redux Toolkit", "Zustand", "React Query"],
  "patterns": ["Flux", "Observer", "State Machine"],
  "optimization": ["Selector", "Memoization", "Normalization"]
}
```

### 📖 本章目录

- [状态管理基础概念](#状态管理基础概念)
- [React内置状态管理](#react内置状态管理)
- [Redux Toolkit现代实践](#redux-toolkit现代实践)
- [Zustand轻量级方案](#zustand轻量级方案)
- [React Query服务端状态](#react-query服务端状态)
- [状态设计模式](#状态设计模式)
- [性能优化策略](#性能优化策略)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🧠 状态管理基础概念

### 状态的分类

在React应用中，我们通常将状态分为以下几类：

```typescript
// 1. 本地组件状态 (Local State)
function ProductCard({ product }: { product: Product }) {
  const [isLiked, setIsLiked] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  return (
    <div className="product-card">
      <button onClick={() => setIsLiked(!isLiked)}>
        {isLiked ? '❤️' : '🤍'}
      </button>
      <button onClick={() => setShowDetails(!showDetails)}>
        {showDetails ? '收起' : '详情'}
      </button>
    </div>
  );
}

// 2. 共享状态 (Shared State)
interface AppState {
  user: User | null;
  cart: CartItem[];
  theme: 'light' | 'dark';
  language: 'zh' | 'en';
}

// 3. 服务端状态 (Server State)
interface ServerState {
  products: Product[];
  categories: Category[];
  orders: Order[];
  // 特点：异步、可能过期、需要同步
}

// 4. URL状态 (URL State)
interface URLState {
  page: number;
  search: string;
  filters: ProductFilters;
  sortBy: string;
}

// 5. 表单状态 (Form State)
interface FormState {
  values: Record<string, any>;
  errors: Record<string, string>;
  touched: Record<string, boolean>;
  isSubmitting: boolean;
}
```

### 🔄 框架对比：状态管理方式

```vue
<!-- Vue 3 - Composition API + Pinia -->
<template>
  <div class="product-card">
    <button @click="toggleLike">
      {{ isLiked ? '❤️' : '🤍' }}
    </button>
    <button @click="toggleDetails">
      {{ showDetails ? '收起' : '详情' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useUserStore } from '@/stores/user';

// 本地状态
const isLiked = ref(false);
const showDetails = ref(false);

// 全局状态 (Pinia Store)
const userStore = useUserStore();

const toggleLike = () => {
  isLiked.value = !isLiked.value;
};

const toggleDetails = () => {
  showDetails.value = !showDetails.value;
};
</script>

<!-- Pinia Store 定义 -->
<script lang="ts">
// stores/user.ts
import { defineStore } from 'pinia';

interface User {
  id: number;
  name: string;
  email: string;
}

export const useUserStore = defineStore('user', {
  state: () => ({
    user: null as User | null,
    cart: [] as CartItem[],
    theme: 'light' as 'light' | 'dark',
    language: 'zh' as 'zh' | 'en'
  }),

  getters: {
    cartItemCount: (state) => state.cart.length,
    isLoggedIn: (state) => state.user !== null
  },

  actions: {
    setUser(user: User) {
      this.user = user;
    },

    addToCart(item: CartItem) {
      this.cart.push(item);
    }
  }
});
</script>
```

```typescript
// Angular - 服务 + RxJS
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

interface AppState {
  user: User | null;
  cart: CartItem[];
  theme: 'light' | 'dark';
  language: 'zh' | 'en';
}

@Injectable({
  providedIn: 'root'
})
export class StateService {
  private stateSubject = new BehaviorSubject<AppState>({
    user: null,
    cart: [],
    theme: 'light',
    language: 'zh'
  });

  public state$: Observable<AppState> = this.stateSubject.asObservable();

  get currentState(): AppState {
    return this.stateSubject.value;
  }

  updateUser(user: User | null): void {
    this.stateSubject.next({
      ...this.currentState,
      user
    });
  }

  addToCart(item: CartItem): void {
    this.stateSubject.next({
      ...this.currentState,
      cart: [...this.currentState.cart, item]
    });
  }

  setTheme(theme: 'light' | 'dark'): void {
    this.stateSubject.next({
      ...this.currentState,
      theme
    });
  }
}

// 组件中使用
@Component({
  selector: 'app-product-card',
  template: `
    <div class="product-card">
      <button (click)="toggleLike()">
        {{ isLiked ? '❤️' : '🤍' }}
      </button>
      <button (click)="toggleDetails()">
        {{ showDetails ? '收起' : '详情' }}
      </button>
    </div>
  `
})
export class ProductCardComponent {
  isLiked = false;
  showDetails = false;

  constructor(private stateService: StateService) {}

  toggleLike(): void {
    this.isLiked = !this.isLiked;
  }

  toggleDetails(): void {
    this.showDetails = !this.showDetails;
  }
}
```

```svelte
<!-- Svelte - Stores -->
<script lang="ts">
  import { writable, derived } from 'svelte/store';

  // 本地状态
  let isLiked = false;
  let showDetails = false;

  // 全局状态 (Svelte Stores)
  interface AppState {
    user: User | null;
    cart: CartItem[];
    theme: 'light' | 'dark';
    language: 'zh' | 'en';
  }

  // 创建可写store
  export const appState = writable<AppState>({
    user: null,
    cart: [],
    theme: 'light',
    language: 'zh'
  });

  // 派生状态
  export const cartItemCount = derived(
    appState,
    $appState => $appState.cart.length
  );

  export const isLoggedIn = derived(
    appState,
    $appState => $appState.user !== null
  );

  // 状态更新函数
  export function setUser(user: User | null) {
    appState.update(state => ({
      ...state,
      user
    }));
  }

  export function addToCart(item: CartItem) {
    appState.update(state => ({
      ...state,
      cart: [...state.cart, item]
    }));
  }

  function toggleLike() {
    isLiked = !isLiked;
  }

  function toggleDetails() {
    showDetails = !showDetails;
  }
</script>

<div class="product-card">
  <button on:click={toggleLike}>
    {isLiked ? '❤️' : '🤍'}
  </button>
  <button on:click={toggleDetails}>
    {showDetails ? '收起' : '详情'}
  </button>
</div>
```

```dart
// Flutter - Provider + ChangeNotifier
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

// 状态模型
class AppState extends ChangeNotifier {
  User? _user;
  List<CartItem> _cart = [];
  String _theme = 'light';
  String _language = 'zh';

  User? get user => _user;
  List<CartItem> get cart => _cart;
  String get theme => _theme;
  String get language => _language;

  int get cartItemCount => _cart.length;
  bool get isLoggedIn => _user != null;

  void setUser(User? user) {
    _user = user;
    notifyListeners();
  }

  void addToCart(CartItem item) {
    _cart.add(item);
    notifyListeners();
  }

  void setTheme(String theme) {
    _theme = theme;
    notifyListeners();
  }
}

// 组件使用
class ProductCard extends StatefulWidget {
  final Product product;

  const ProductCard({Key? key, required this.product}) : super(key: key);

  @override
  _ProductCardState createState() => _ProductCardState();
}

class _ProductCardState extends State<ProductCard> {
  bool isLiked = false;
  bool showDetails = false;

  @override
  Widget build(BuildContext context) {
    return Consumer<AppState>(
      builder: (context, appState, child) {
        return Container(
          child: Column(
            children: [
              IconButton(
                icon: Icon(isLiked ? Icons.favorite : Icons.favorite_border),
                onPressed: () {
                  setState(() {
                    isLiked = !isLiked;
                  });
                },
              ),
              TextButton(
                child: Text(showDetails ? '收起' : '详情'),
                onPressed: () {
                  setState(() {
                    showDetails = !showDetails;
                  });
                },
              ),
            ],
          ),
        );
      },
    );
  }
}
```

**💡 状态管理对比：**

| 特性 | React | Vue 3 | Angular | Svelte | Flutter |
|------|-------|-------|---------|--------|---------|
| **本地状态** | `useState` | `ref/reactive` | 组件属性 | 变量 | `setState` |
| **全局状态** | Context/Redux | Pinia | 服务+RxJS | Stores | Provider |
| **状态更新** | `setState` | `.value =` | `next()` | `update()` | `notifyListeners()` |
| **派生状态** | `useMemo` | `computed` | `map/filter` | `derived` | `get` 方法 |
| **异步状态** | useEffect | `watch` | Observable | `$:` | FutureBuilder |
| **性能优化** | memo/callback | `shallowRef` | OnPush | 自动优化 | `const` 构造 |

### 状态管理的挑战

```typescript
// 常见的状态管理问题

// 1. 状态提升问题
function App() {
  const [user, setUser] = useState<User | null>(null);
  const [cart, setCart] = useState<CartItem[]>([]);

  // 状态需要在多个深层组件间共享
  return (
    <div>
      <Header user={user} cartCount={cart.length} />
      <ProductList onAddToCart={(item) => setCart([...cart, item])} />
      <Cart items={cart} onUpdateCart={setCart} />
      <Footer user={user} />
    </div>
  );
}

// 2. 状态同步问题
function useUserProfile() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);

  // 多个组件都需要用户信息，如何保持同步？
  useEffect(() => {
    setLoading(true);
    fetchUser().then(setUser).finally(() => setLoading(false));
  }, []);

  return { user, loading, setUser };
}

// 3. 状态更新复杂性
function cartReducer(state: CartState, action: CartAction): CartState {
  switch (action.type) {
    case 'ADD_ITEM':
      // 复杂的状态更新逻辑
      const existingItem = state.items.find(item => item.id === action.payload.id);
      if (existingItem) {
        return {
          ...state,
          items: state.items.map(item =>
            item.id === action.payload.id
              ? { ...item, quantity: item.quantity + action.payload.quantity }
              : item
          ),
        };
      }
      return {
        ...state,
        items: [...state.items, action.payload],
      };
    // 更多复杂的case...
  }
}
```

---

## ⚛️ React内置状态管理

### useState的高级模式

```typescript
import { useState, useCallback, useMemo } from 'react';

// 1. 状态工厂模式
function createInitialState() {
  return {
    products: [],
    filters: {
      category: '',
      priceRange: [0, 1000],
      rating: 0,
    },
    pagination: {
      page: 1,
      pageSize: 20,
      total: 0,
    },
    loading: false,
    error: null,
  };
}

function useProductList() {
  const [state, setState] = useState(createInitialState);

  // 2. 状态更新器模式
  const updateFilters = useCallback((newFilters: Partial<ProductFilters>) => {
    setState(prev => ({
      ...prev,
      filters: { ...prev.filters, ...newFilters },
      pagination: { ...prev.pagination, page: 1 }, // 重置页码
    }));
  }, []);

  const updatePagination = useCallback((newPagination: Partial<Pagination>) => {
    setState(prev => ({
      ...prev,
      pagination: { ...prev.pagination, ...newPagination },
    }));
  }, []);

  // 3. 派生状态
  const filteredProducts = useMemo(() => {
    return state.products.filter(product => {
      if (state.filters.category && product.category !== state.filters.category) {
        return false;
      }
      if (product.price < state.filters.priceRange[0] ||
          product.price > state.filters.priceRange[1]) {
        return false;
      }
      if (product.rating < state.filters.rating) {
        return false;
      }
      return true;
    });
  }, [state.products, state.filters]);

  return {
    state,
    filteredProducts,
    updateFilters,
    updatePagination,
  };
}
```

### useReducer的企业级应用

```typescript
import { useReducer, useCallback, useContext, createContext, ReactNode } from 'react';

// 1. 复杂状态管理
interface AppState {
  user: User | null;
  cart: {
    items: CartItem[];
    totalItems: number;
    totalPrice: number;
  };
  ui: {
    sidebarOpen: boolean;
    theme: 'light' | 'dark';
    language: 'zh' | 'en';
    loading: boolean;
    notifications: Notification[];
  };
  products: {
    list: Product[];
    categories: Category[];
    filters: ProductFilters;
    pagination: Pagination;
  };
}

type AppAction =
  | { type: 'SET_USER'; payload: User | null }
  | { type: 'ADD_TO_CART'; payload: { product: Product; quantity: number } }
  | { type: 'REMOVE_FROM_CART'; payload: { itemId: string } }
  | { type: 'UPDATE_CART_QUANTITY'; payload: { itemId: string; quantity: number } }
  | { type: 'TOGGLE_SIDEBAR' }
  | { type: 'SET_THEME'; payload: 'light' | 'dark' }
  | { type: 'SET_LANGUAGE'; payload: 'zh' | 'en' }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'ADD_NOTIFICATION'; payload: Notification }
  | { type: 'REMOVE_NOTIFICATION'; payload: { id: string } }
  | { type: 'SET_PRODUCTS'; payload: Product[] }
  | { type: 'SET_CATEGORIES'; payload: Category[] }
  | { type: 'UPDATE_FILTERS'; payload: Partial<ProductFilters> }
  | { type: 'UPDATE_PAGINATION'; payload: Partial<Pagination> };

// 2. Reducer组合模式
function cartReducer(state: AppState['cart'], action: AppAction): AppState['cart'] {
  switch (action.type) {
    case 'ADD_TO_CART': {
      const { product, quantity } = action.payload;
      const existingItemIndex = state.items.findIndex(item => item.product.id === product.id);

      let newItems: CartItem[];
      if (existingItemIndex > -1) {
        newItems = state.items.map((item, index) =>
          index === existingItemIndex
            ? { ...item, quantity: item.quantity + quantity }
            : item
        );
      } else {
        newItems = [...state.items, {
          id: `${product.id}-${Date.now()}`,
          product,
          quantity,
          selected: true,
        }];
      }

      return {
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce((sum, item) =>
          sum + parseFloat(item.product.price) * item.quantity, 0
        ),
      };
    }

    case 'REMOVE_FROM_CART': {
      const newItems = state.items.filter(item => item.id !== action.payload.itemId);
      return {
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce((sum, item) =>
          sum + parseFloat(item.product.price) * item.quantity, 0
        ),
      };
    }

    case 'UPDATE_CART_QUANTITY': {
      const { itemId, quantity } = action.payload;
      if (quantity <= 0) {
        return cartReducer(state, { type: 'REMOVE_FROM_CART', payload: { itemId } });
      }

      const newItems = state.items.map(item =>
        item.id === itemId ? { ...item, quantity } : item
      );

      return {
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce((sum, item) =>
          sum + parseFloat(item.product.price) * item.quantity, 0
        ),
      };
    }

    default:
      return state;
  }
}

function uiReducer(state: AppState['ui'], action: AppAction): AppState['ui'] {
  switch (action.type) {
    case 'TOGGLE_SIDEBAR':
      return { ...state, sidebarOpen: !state.sidebarOpen };

    case 'SET_THEME':
      return { ...state, theme: action.payload };

    case 'SET_LANGUAGE':
      return { ...state, language: action.payload };

    case 'SET_LOADING':
      return { ...state, loading: action.payload };

    case 'ADD_NOTIFICATION':
      return {
        ...state,
        notifications: [...state.notifications, action.payload],
      };

    case 'REMOVE_NOTIFICATION':
      return {
        ...state,
        notifications: state.notifications.filter(n => n.id !== action.payload.id),
      };

    default:
      return state;
  }
}

function productsReducer(state: AppState['products'], action: AppAction): AppState['products'] {
  switch (action.type) {
    case 'SET_PRODUCTS':
      return { ...state, list: action.payload };

    case 'SET_CATEGORIES':
      return { ...state, categories: action.payload };

    case 'UPDATE_FILTERS':
      return {
        ...state,
        filters: { ...state.filters, ...action.payload },
        pagination: { ...state.pagination, page: 1 }, // 重置页码
      };

    case 'UPDATE_PAGINATION':
      return {
        ...state,
        pagination: { ...state.pagination, ...action.payload },
      };

    default:
      return state;
  }
}

// 3. 主Reducer
function appReducer(state: AppState, action: AppAction): AppState {
  return {
    user: action.type === 'SET_USER' ? action.payload : state.user,
    cart: cartReducer(state.cart, action),
    ui: uiReducer(state.ui, action),
    products: productsReducer(state.products, action),
  };
}

// 4. Context Provider
interface AppContextValue {
  state: AppState;
  dispatch: React.Dispatch<AppAction>;
  // 便捷的action creators
  setUser: (user: User | null) => void;
  addToCart: (product: Product, quantity: number) => void;
  removeFromCart: (itemId: string) => void;
  updateCartQuantity: (itemId: string, quantity: number) => void;
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  setLanguage: (language: 'zh' | 'en') => void;
  addNotification: (notification: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
}

const AppContext = createContext<AppContextValue | undefined>(undefined);

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, {
    user: null,
    cart: {
      items: [],
      totalItems: 0,
      totalPrice: 0,
    },
    ui: {
      sidebarOpen: false,
      theme: 'light',
      language: 'zh',
      loading: false,
      notifications: [],
    },
    products: {
      list: [],
      categories: [],
      filters: {
        category: '',
        priceRange: [0, 1000],
        rating: 0,
        search: '',
      },
      pagination: {
        page: 1,
        pageSize: 20,
        total: 0,
      },
    },
  });

  // Action creators
  const setUser = useCallback((user: User | null) => {
    dispatch({ type: 'SET_USER', payload: user });
  }, []);

  const addToCart = useCallback((product: Product, quantity: number) => {
    dispatch({ type: 'ADD_TO_CART', payload: { product, quantity } });
  }, []);

  const removeFromCart = useCallback((itemId: string) => {
    dispatch({ type: 'REMOVE_FROM_CART', payload: { itemId } });
  }, []);

  const updateCartQuantity = useCallback((itemId: string, quantity: number) => {
    dispatch({ type: 'UPDATE_CART_QUANTITY', payload: { itemId, quantity } });
  }, []);

  const toggleSidebar = useCallback(() => {
    dispatch({ type: 'TOGGLE_SIDEBAR' });
  }, []);

  const setTheme = useCallback((theme: 'light' | 'dark') => {
    dispatch({ type: 'SET_THEME', payload: theme });
  }, []);

  const setLanguage = useCallback((language: 'zh' | 'en') => {
    dispatch({ type: 'SET_LANGUAGE', payload: language });
  }, []);

  const addNotification = useCallback((notification: Omit<Notification, 'id'>) => {
    const id = Date.now().toString();
    dispatch({ type: 'ADD_NOTIFICATION', payload: { ...notification, id } });

    // 自动移除通知
    setTimeout(() => {
      dispatch({ type: 'REMOVE_NOTIFICATION', payload: { id } });
    }, 5000);
  }, []);

  const removeNotification = useCallback((id: string) => {
    dispatch({ type: 'REMOVE_NOTIFICATION', payload: { id } });
  }, []);

  const value: AppContextValue = {
    state,
    dispatch,
    setUser,
    addToCart,
    removeFromCart,
    updateCartQuantity,
    toggleSidebar,
    setTheme,
    setLanguage,
    addNotification,
    removeNotification,
  };

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

// 5. 自定义Hook
export function useApp(): AppContextValue {
  const context = useContext(AppContext);
  if (!context) {
    throw new Error('useApp must be used within an AppProvider');
  }
  return context;
}

// 6. 选择器Hook
export function useAppSelector<T>(selector: (state: AppState) => T): T {
  const { state } = useApp();
  return useMemo(() => selector(state), [state, selector]);
}

// 使用示例
function CartIcon() {
  const totalItems = useAppSelector(state => state.cart.totalItems);
  const { toggleSidebar } = useApp();

  return (
    <button onClick={toggleSidebar}>
      🛒 {totalItems > 0 && <span>{totalItems}</span>}
    </button>
  );
}
```

---

## 🔄 Redux Toolkit现代实践

### RTK的基础设置

```typescript
// store/index.ts
import { configureStore } from '@reduxjs/toolkit';
import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import userSlice from './slices/userSlice';
import cartSlice from './slices/cartSlice';
import productsSlice from './slices/productsSlice';
import uiSlice from './slices/uiSlice';

// 配置store
export const store = configureStore({
  reducer: {
    user: userSlice,
    cart: cartSlice,
    products: productsSlice,
    ui: uiSlice,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }),
  devTools: process.env.NODE_ENV !== 'production',
});

// 类型定义
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// 类型化的hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;
```

### 创建Slice

```typescript
// store/slices/cartSlice.ts
import { createSlice, PayloadAction, createSelector } from '@reduxjs/toolkit';
import type { RootState } from '../index';

interface CartItem {
  id: string;
  product: Product;
  quantity: number;
  selected: boolean;
}

interface CartState {
  items: CartItem[];
  loading: boolean;
  error: string | null;
}

const initialState: CartState = {
  items: [],
  loading: false,
  error: null,
};

const cartSlice = createSlice({
  name: 'cart',
  initialState,
  reducers: {
    addItem: (state, action: PayloadAction<{ product: Product; quantity: number }>) => {
      const { product, quantity } = action.payload;
      const existingItem = state.items.find(item => item.product.id === product.id);

      if (existingItem) {
        existingItem.quantity += quantity;
      } else {
        state.items.push({
          id: `${product.id}-${Date.now()}`,
          product,
          quantity,
          selected: true,
        });
      }
    },

    removeItem: (state, action: PayloadAction<{ itemId: string }>) => {
      state.items = state.items.filter(item => item.id !== action.payload.itemId);
    },

    updateQuantity: (state, action: PayloadAction<{ itemId: string; quantity: number }>) => {
      const { itemId, quantity } = action.payload;
      const item = state.items.find(item => item.id === itemId);

      if (item) {
        if (quantity <= 0) {
          state.items = state.items.filter(item => item.id !== itemId);
        } else {
          item.quantity = quantity;
        }
      }
    },

    toggleSelect: (state, action: PayloadAction<{ itemId: string }>) => {
      const item = state.items.find(item => item.id === action.payload.itemId);
      if (item) {
        item.selected = !item.selected;
      }
    },

    selectAll: (state, action: PayloadAction<{ selected: boolean }>) => {
      state.items.forEach(item => {
        item.selected = action.payload.selected;
      });
    },

    clearCart: (state) => {
      state.items = [];
    },

    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },

    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

// Action creators
export const {
  addItem,
  removeItem,
  updateQuantity,
  toggleSelect,
  selectAll,
  clearCart,
  setLoading,
  setError,
} = cartSlice.actions;

// Selectors
export const selectCartItems = (state: RootState) => state.cart.items;
export const selectCartLoading = (state: RootState) => state.cart.loading;
export const selectCartError = (state: RootState) => state.cart.error;

// Memoized selectors
export const selectCartTotalItems = createSelector(
  [selectCartItems],
  (items) => items.reduce((total, item) => total + item.quantity, 0)
);

export const selectCartTotalPrice = createSelector(
  [selectCartItems],
  (items) => items.reduce((total, item) => {
    const price = parseFloat(item.product.discount_price || item.product.price);
    return total + price * item.quantity;
  }, 0)
);

export const selectSelectedItems = createSelector(
  [selectCartItems],
  (items) => items.filter(item => item.selected)
);

export const selectSelectedTotalPrice = createSelector(
  [selectSelectedItems],
  (items) => items.reduce((total, item) => {
    const price = parseFloat(item.product.discount_price || item.product.price);
    return total + price * item.quantity;
  }, 0)
);

export default cartSlice.reducer;
```

### 异步Actions (Thunks)

```typescript
// store/slices/productsSlice.ts
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import type { RootState } from '../index';

interface ProductsState {
  list: Product[];
  categories: Category[];
  filters: ProductFilters;
  pagination: Pagination;
  loading: boolean;
  error: string | null;
}

const initialState: ProductsState = {
  list: [],
  categories: [],
  filters: {
    category: '',
    priceRange: [0, 1000],
    rating: 0,
    search: '',
  },
  pagination: {
    page: 1,
    pageSize: 20,
    total: 0,
  },
  loading: false,
  error: null,
};

// 异步thunk
export const fetchProducts = createAsyncThunk(
  'products/fetchProducts',
  async (params: {
    page?: number;
    pageSize?: number;
    category?: string;
    search?: string;
  }, { rejectWithValue }) => {
    try {
      const searchParams = new URLSearchParams();
      if (params.page) searchParams.set('page', params.page.toString());
      if (params.pageSize) searchParams.set('pageSize', params.pageSize.toString());
      if (params.category) searchParams.set('category', params.category);
      if (params.search) searchParams.set('search', params.search);

      const response = await fetch(`/api/products?${searchParams}`);

      if (!response.ok) {
        throw new Error('Failed to fetch products');
      }

      const data = await response.json();
      return data;
    } catch (error: any) {
      return rejectWithValue(error.message);
    }
  }
);

export const fetchCategories = createAsyncThunk(
  'products/fetchCategories',
  async (_, { rejectWithValue }) => {
    try {
      const response = await fetch('/api/categories');

      if (!response.ok) {
        throw new Error('Failed to fetch categories');
      }

      const data = await response.json();
      return data.categories;
    } catch (error: any) {
      return rejectWithValue(error.message);
    }
  }
);

const productsSlice = createSlice({
  name: 'products',
  initialState,
  reducers: {
    updateFilters: (state, action: PayloadAction<Partial<ProductFilters>>) => {
      state.filters = { ...state.filters, ...action.payload };
      state.pagination.page = 1; // 重置页码
    },

    updatePagination: (state, action: PayloadAction<Partial<Pagination>>) => {
      state.pagination = { ...state.pagination, ...action.payload };
    },

    clearFilters: (state) => {
      state.filters = initialState.filters;
      state.pagination.page = 1;
    },
  },
  extraReducers: (builder) => {
    // fetchProducts
    builder
      .addCase(fetchProducts.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchProducts.fulfilled, (state, action) => {
        state.loading = false;
        state.list = action.payload.products;
        state.pagination.total = action.payload.total;
      })
      .addCase(fetchProducts.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });

    // fetchCategories
    builder
      .addCase(fetchCategories.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchCategories.fulfilled, (state, action) => {
        state.loading = false;
        state.categories = action.payload;
      })
      .addCase(fetchCategories.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      });
  },
});

export const { updateFilters, updatePagination, clearFilters } = productsSlice.actions;

// Selectors
export const selectProducts = (state: RootState) => state.products.list;
export const selectCategories = (state: RootState) => state.products.categories;
export const selectFilters = (state: RootState) => state.products.filters;
export const selectPagination = (state: RootState) => state.products.pagination;
export const selectProductsLoading = (state: RootState) => state.products.loading;
export const selectProductsError = (state: RootState) => state.products.error;

export default productsSlice.reducer;
```

---

## 🪶 Zustand轻量级方案

### 基础Store设计

```typescript
// store/useCartStore.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface CartItem {
  id: string;
  product: Product;
  quantity: number;
  selected: boolean;
}

interface CartStore {
  // State
  items: CartItem[];
  loading: boolean;
  error: string | null;

  // Computed values
  totalItems: number;
  totalPrice: number;
  selectedItems: CartItem[];
  selectedTotalPrice: number;

  // Actions
  addItem: (product: Product, quantity?: number) => void;
  removeItem: (itemId: string) => void;
  updateQuantity: (itemId: string, quantity: number) => void;
  toggleSelect: (itemId: string) => void;
  selectAll: (selected: boolean) => void;
  clearCart: () => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;

  // Async actions
  syncWithServer: () => Promise<void>;
}

export const useCartStore = create<CartStore>()(
  devtools(
    persist(
      immer((set, get) => ({
        // Initial state
        items: [],
        loading: false,
        error: null,

        // Computed values (getters)
        get totalItems() {
          return get().items.reduce((total, item) => total + item.quantity, 0);
        },

        get totalPrice() {
          return get().items.reduce((total, item) => {
            const price = parseFloat(item.product.discount_price || item.product.price);
            return total + price * item.quantity;
          }, 0);
        },

        get selectedItems() {
          return get().items.filter(item => item.selected);
        },

        get selectedTotalPrice() {
          return get().selectedItems.reduce((total, item) => {
            const price = parseFloat(item.product.discount_price || item.product.price);
            return total + price * item.quantity;
          }, 0);
        },

        // Actions
        addItem: (product, quantity = 1) => {
          set((state) => {
            const existingItemIndex = state.items.findIndex(
              item => item.product.id === product.id
            );

            if (existingItemIndex > -1) {
              state.items[existingItemIndex].quantity += quantity;
            } else {
              state.items.push({
                id: `${product.id}-${Date.now()}`,
                product,
                quantity,
                selected: true,
              });
            }
          });
        },

        removeItem: (itemId) => {
          set((state) => {
            state.items = state.items.filter(item => item.id !== itemId);
          });
        },

        updateQuantity: (itemId, quantity) => {
          if (quantity <= 0) {
            get().removeItem(itemId);
            return;
          }

          set((state) => {
            const item = state.items.find(item => item.id === itemId);
            if (item) {
              item.quantity = quantity;
            }
          });
        },

        toggleSelect: (itemId) => {
          set((state) => {
            const item = state.items.find(item => item.id === itemId);
            if (item) {
              item.selected = !item.selected;
            }
          });
        },

        selectAll: (selected) => {
          set((state) => {
            state.items.forEach(item => {
              item.selected = selected;
            });
          });
        },

        clearCart: () => {
          set((state) => {
            state.items = [];
          });
        },

        setLoading: (loading) => {
          set((state) => {
            state.loading = loading;
          });
        },

        setError: (error) => {
          set((state) => {
            state.error = error;
          });
        },

        // Async actions
        syncWithServer: async () => {
          const { setLoading, setError } = get();

          try {
            setLoading(true);
            setError(null);

            const response = await fetch('/api/cart/sync', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ items: get().items }),
            });

            if (!response.ok) {
              throw new Error('Failed to sync cart');
            }

            const data = await response.json();

            set((state) => {
              state.items = data.items;
            });
          } catch (error: any) {
            setError(error.message);
          } finally {
            setLoading(false);
          }
        },
      })),
      {
        name: 'mall-cart',
        partialize: (state) => ({ items: state.items }), // 只持久化items
      }
    ),
    { name: 'cart-store' }
  )
);

// 选择器hooks
export const useCartItems = () => useCartStore(state => state.items);
export const useCartTotalItems = () => useCartStore(state => state.totalItems);
export const useCartTotalPrice = () => useCartStore(state => state.totalPrice);
export const useCartActions = () => useCartStore(state => ({
  addItem: state.addItem,
  removeItem: state.removeItem,
  updateQuantity: state.updateQuantity,
  toggleSelect: state.toggleSelect,
  selectAll: state.selectAll,
  clearCart: state.clearCart,
}));
```

### 组合多个Store

```typescript
// store/useAppStore.ts
import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

interface User {
  id: number;
  username: string;
  email: string;
  avatar?: string;
}

interface AppStore {
  // User state
  user: User | null;
  isAuthenticated: boolean;

  // UI state
  theme: 'light' | 'dark';
  language: 'zh' | 'en';
  sidebarOpen: boolean;

  // Actions
  setUser: (user: User | null) => void;
  logout: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  setLanguage: (language: 'zh' | 'en') => void;
  toggleSidebar: () => void;
}

export const useAppStore = create<AppStore>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    user: null,
    isAuthenticated: false,
    theme: 'light',
    language: 'zh',
    sidebarOpen: false,

    // Actions
    setUser: (user) => {
      set({ user, isAuthenticated: !!user });
    },

    logout: () => {
      set({ user: null, isAuthenticated: false });
      // 清除其他相关状态
      useCartStore.getState().clearCart();
    },

    setTheme: (theme) => {
      set({ theme });
      // 更新CSS变量或类名
      document.documentElement.setAttribute('data-theme', theme);
    },

    setLanguage: (language) => {
      set({ language });
    },

    toggleSidebar: () => {
      set((state) => ({ sidebarOpen: !state.sidebarOpen }));
    },
  }))
);

// 监听状态变化
useAppStore.subscribe(
  (state) => state.theme,
  (theme) => {
    localStorage.setItem('theme', theme);
  }
);

useAppStore.subscribe(
  (state) => state.language,
  (language) => {
    localStorage.setItem('language', language);
  }
);
```

---

## 🌐 React Query服务端状态

### 基础配置

```typescript
// lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5分钟
      gcTime: 10 * 60 * 1000, // 10分钟 (原cacheTime)
      retry: (failureCount, error: any) => {
        // 4xx错误不重试
        if (error?.status >= 400 && error?.status < 500) {
          return false;
        }
        return failureCount < 3;
      },
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1,
    },
  },
});

// app/layout.tsx
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <QueryClientProvider client={queryClient}>
          {children}
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </body>
    </html>
  );
}
```

### 查询操作

```typescript
// hooks/useProducts.ts
import { useQuery, useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';

// 产品查询
export function useProducts(params: {
  page?: number;
  pageSize?: number;
  category?: string;
  search?: string;
}) {
  return useQuery({
    queryKey: ['products', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.set(key, value.toString());
        }
      });

      const response = await fetch(`/api/products?${searchParams}`);
      if (!response.ok) {
        throw new Error('Failed to fetch products');
      }
      return response.json();
    },
    staleTime: 5 * 60 * 1000, // 5分钟内不重新获取
    placeholderData: { products: [], total: 0 }, // 占位数据
  });
}

// 无限滚动查询
export function useInfiniteProducts(filters: ProductFilters) {
  return useInfiniteQuery({
    queryKey: ['products', 'infinite', filters],
    queryFn: async ({ pageParam = 1 }) => {
      const searchParams = new URLSearchParams({
        page: pageParam.toString(),
        pageSize: '20',
        ...filters,
      });

      const response = await fetch(`/api/products?${searchParams}`);
      if (!response.ok) {
        throw new Error('Failed to fetch products');
      }
      return response.json();
    },
    getNextPageParam: (lastPage, allPages) => {
      const hasMore = lastPage.products.length === 20;
      return hasMore ? allPages.length + 1 : undefined;
    },
    initialPageParam: 1,
  });
}

// 单个产品查询
export function useProduct(id: number) {
  return useQuery({
    queryKey: ['product', id],
    queryFn: async () => {
      const response = await fetch(`/api/products/${id}`);
      if (!response.ok) {
        throw new Error('Failed to fetch product');
      }
      return response.json();
    },
    enabled: !!id, // 只有当id存在时才执行查询
    staleTime: 10 * 60 * 1000, // 10分钟
  });
}

// 产品分类查询
export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const response = await fetch('/api/categories');
      if (!response.ok) {
        throw new Error('Failed to fetch categories');
      }
      const data = await response.json();
      return data.categories;
    },
    staleTime: 30 * 60 * 1000, // 30分钟，分类数据变化较少
  });
}
```

### 变更操作

```typescript
// hooks/useProductMutations.ts
import { useMutation, useQueryClient } from '@tanstack/react-query';

// 添加到购物车
export function useAddToCart() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ productId, quantity }: { productId: number; quantity: number }) => {
      const response = await fetch('/api/cart/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ productId, quantity }),
      });

      if (!response.ok) {
        throw new Error('Failed to add to cart');
      }

      return response.json();
    },
    onSuccess: () => {
      // 使购物车查询失效，触发重新获取
      queryClient.invalidateQueries({ queryKey: ['cart'] });

      // 显示成功提示
      useAppStore.getState().addNotification({
        type: 'success',
        message: '商品已添加到购物车',
      });
    },
    onError: (error: Error) => {
      useAppStore.getState().addNotification({
        type: 'error',
        message: error.message,
      });
    },
  });
}

// 更新产品
export function useUpdateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Partial<Product> }) => {
      const response = await fetch(`/api/products/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        throw new Error('Failed to update product');
      }

      return response.json();
    },
    onSuccess: (updatedProduct, { id }) => {
      // 更新缓存中的产品数据
      queryClient.setQueryData(['product', id], updatedProduct);

      // 使产品列表查询失效
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
    onError: (error: Error) => {
      console.error('Update product failed:', error);
    },
  });
}

// 乐观更新示例
export function useToggleFavorite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ productId, isFavorite }: { productId: number; isFavorite: boolean }) => {
      const response = await fetch(`/api/products/${productId}/favorite`, {
        method: isFavorite ? 'POST' : 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Failed to toggle favorite');
      }

      return { productId, isFavorite };
    },
    onMutate: async ({ productId, isFavorite }) => {
      // 取消正在进行的查询
      await queryClient.cancelQueries({ queryKey: ['product', productId] });

      // 获取当前数据
      const previousProduct = queryClient.getQueryData(['product', productId]);

      // 乐观更新
      queryClient.setQueryData(['product', productId], (old: any) => ({
        ...old,
        isFavorite,
      }));

      // 返回回滚数据
      return { previousProduct, productId };
    },
    onError: (error, variables, context) => {
      // 回滚到之前的状态
      if (context?.previousProduct) {
        queryClient.setQueryData(['product', context.productId], context.previousProduct);
      }
    },
    onSettled: (data, error, { productId }) => {
      // 无论成功失败都重新获取数据
      queryClient.invalidateQueries({ queryKey: ['product', productId] });
    },
  });
}
```

---

## 🎯 面试常考知识点

### 1. 状态管理方案选择

**Q: 如何选择合适的状态管理方案？**

**A: 选择标准：**

| 场景 | 推荐方案 | 理由 |
|------|----------|------|
| **小型应用** | useState + useContext | 简单直接，无额外依赖 |
| **中型应用** | Zustand | 轻量级，易于使用 |
| **大型应用** | Redux Toolkit | 成熟生态，强大的开发工具 |
| **服务端状态** | React Query | 专门处理异步数据 |
| **表单状态** | React Hook Form | 专门的表单解决方案 |

```typescript
// 决策树
function chooseStateManagement(appSize: string, complexity: string) {
  if (appSize === 'small' && complexity === 'low') {
    return 'useState + useContext';
  }

  if (appSize === 'medium' || complexity === 'medium') {
    return 'Zustand';
  }

  if (appSize === 'large' || complexity === 'high') {
    return 'Redux Toolkit';
  }

  return 'Hybrid approach';
}
```

### 2. Redux vs Zustand

**Q: Redux和Zustand有什么区别？**

**A: 主要区别：**

```typescript
// Redux Toolkit - 更多样板代码，但更规范
const counterSlice = createSlice({
  name: 'counter',
  initialState: { value: 0 },
  reducers: {
    increment: (state) => {
      state.value += 1;
    },
  },
});

// Zustand - 更简洁，但需要自律
const useCounterStore = create((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
}));
```

**对比表：**

| 特性 | Redux Toolkit | Zustand |
|------|---------------|---------|
| **学习曲线** | 陡峭 | 平缓 |
| **样板代码** | 多 | 少 |
| **类型安全** | 需要配置 | 内置支持 |
| **开发工具** | 强大 | 基础 |
| **生态系统** | 丰富 | 简单 |
| **包大小** | 较大 | 很小 |

### 3. React Query的核心概念

**Q: React Query解决了什么问题？核心概念是什么？**

**A: 解决的问题：**

1. **服务端状态同步** - 自动同步服务端数据
2. **缓存管理** - 智能缓存和失效策略
3. **后台更新** - 自动后台刷新数据
4. **乐观更新** - 提升用户体验
5. **错误处理** - 统一的错误处理机制

**核心概念：**

```typescript
// 1. 查询键 (Query Keys)
const queryKey = ['products', { category: 'electronics', page: 1 }];

// 2. 查询函数 (Query Function)
const queryFn = () => fetch('/api/products').then(res => res.json());

// 3. 缓存时间 (Cache Time)
const cacheTime = 5 * 60 * 1000; // 5分钟

// 4. 过期时间 (Stale Time)
const staleTime = 30 * 1000; // 30秒

// 5. 重新获取策略
const refetchOnWindowFocus = true;
const refetchOnReconnect = true;
```

---

## 🏋️ 实战练习

### 练习1: 设计一个完整的电商状态管理系统

**题目**: 为Mall-Frontend设计一个混合状态管理架构

**要求**:
1. 使用不同的状态管理方案处理不同类型的状态
2. 实现状态持久化
3. 优化性能，避免不必要的重渲染
4. 提供完整的TypeScript类型支持
5. 包含错误处理和加载状态

**解决方案**:

```typescript
// 1. 全局应用状态 - Zustand
interface AppState {
  user: User | null;
  theme: 'light' | 'dark';
  language: 'zh' | 'en';
  notifications: Notification[];
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      user: null,
      theme: 'light',
      language: 'zh',
      notifications: [],

      setUser: (user: User | null) => set({ user }),
      setTheme: (theme: 'light' | 'dark') => set({ theme }),
      setLanguage: (language: 'zh' | 'en') => set({ language }),
      addNotification: (notification: Omit<Notification, 'id'>) => {
        const id = Date.now().toString();
        set(state => ({
          notifications: [...state.notifications, { ...notification, id }]
        }));

        // 自动移除
        setTimeout(() => {
          set(state => ({
            notifications: state.notifications.filter(n => n.id !== id)
          }));
        }, 5000);
      },
    }),
    { name: 'app-store' }
  )
);

// 2. 购物车状态 - Zustand with persistence
export const useCartStore = create<CartStore>()(
  persist(
    immer((set, get) => ({
      items: [],

      addItem: (product: Product, quantity = 1) => {
        set(state => {
          const existingItem = state.items.find(item => item.product.id === product.id);
          if (existingItem) {
            existingItem.quantity += quantity;
          } else {
            state.items.push({
              id: `${product.id}-${Date.now()}`,
              product,
              quantity,
              selected: true,
            });
          }
        });

        // 同步到服务器
        get().syncWithServer();
      },

      syncWithServer: async () => {
        try {
          await fetch('/api/cart/sync', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ items: get().items }),
          });
        } catch (error) {
          console.error('Cart sync failed:', error);
        }
      },
    })),
    { name: 'cart-store' }
  )
);

// 3. 服务端状态 - React Query
export function useProducts(filters: ProductFilters) {
  return useQuery({
    queryKey: ['products', filters],
    queryFn: () => fetchProducts(filters),
    staleTime: 5 * 60 * 1000,
  });
}

export function useProductMutations() {
  const queryClient = useQueryClient();

  const addToCart = useMutation({
    mutationFn: ({ productId, quantity }: { productId: number; quantity: number }) =>
      fetch('/api/cart/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ productId, quantity }),
      }).then(res => res.json()),

    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cart'] });
      useAppStore.getState().addNotification({
        type: 'success',
        message: '商品已添加到购物车',
      });
    },
  });

  return { addToCart };
}

// 4. 表单状态 - 自定义Hook
export function useCheckoutForm() {
  const [formData, setFormData] = useState({
    shippingAddress: '',
    paymentMethod: '',
    couponCode: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateForm = useCallback(() => {
    const newErrors: Record<string, string> = {};

    if (!formData.shippingAddress.trim()) {
      newErrors.shippingAddress = '请填写收货地址';
    }

    if (!formData.paymentMethod) {
      newErrors.paymentMethod = '请选择支付方式';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData]);

  const handleSubmit = useCallback(async () => {
    if (!validateForm()) return;

    setIsSubmitting(true);
    try {
      const response = await fetch('/api/orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (!response.ok) throw new Error('提交失败');

      // 清空购物车
      useCartStore.getState().clearCart();

      // 显示成功消息
      useAppStore.getState().addNotification({
        type: 'success',
        message: '订单提交成功',
      });

      return await response.json();
    } catch (error: any) {
      useAppStore.getState().addNotification({
        type: 'error',
        message: error.message,
      });
      throw error;
    } finally {
      setIsSubmitting(false);
    }
  }, [formData, validateForm]);

  return {
    formData,
    setFormData,
    errors,
    isSubmitting,
    handleSubmit,
    validateForm,
  };
}

// 5. 组件使用示例
function ProductCard({ product }: { product: Product }) {
  const addToCart = useCartStore(state => state.addItem);
  const { addToCart: addToCartMutation } = useProductMutations();

  const handleAddToCart = () => {
    // 本地状态立即更新
    addToCart(product, 1);

    // 同步到服务器
    addToCartMutation.mutate({ productId: product.id, quantity: 1 });
  };

  return (
    <div className="product-card">
      <h3>{product.name}</h3>
      <p>¥{product.price}</p>
      <button onClick={handleAddToCart}>
        加入购物车
      </button>
    </div>
  );
}
```

这个练习展示了：

1. **混合架构** - 不同状态使用不同的管理方案
2. **状态持久化** - 重要状态的本地存储
3. **性能优化** - 选择器和memoization
4. **类型安全** - 完整的TypeScript支持
5. **用户体验** - 乐观更新和错误处理

---

## 📚 本章总结

通过本章学习，我们深入掌握了现代React应用的状态管理：

### 🎯 核心收获

1. **状态管理基础** 🧠
   - 理解了不同类型状态的特点和管理策略
   - 掌握了状态管理方案的选择标准
   - 学会了分析应用的状态管理需求

2. **多种解决方案** 🛠️
   - 掌握了React内置的状态管理方案
   - 学会了Redux Toolkit的现代用法
   - 理解了Zustand的轻量级优势
   - 掌握了React Query的服务端状态管理

3. **实战应用** 💼
   - 设计了完整的电商状态管理架构
   - 实现了状态持久化和同步机制
   - 掌握了性能优化和错误处理策略

4. **最佳实践** 💡
   - 学会了混合使用多种状态管理方案
   - 掌握了状态规范化和派生状态设计
   - 理解了乐观更新和缓存策略

### 🚀 技术进阶

- **下一步学习**: React性能优化技巧
- **实践建议**: 在项目中应用合适的状态管理方案
- **深入方向**: 状态机和复杂状态管理模式

选择合适的状态管理方案是构建可维护React应用的关键！ 🎉

---

*下一章我们将学习《React性能优化技巧》，探索提升React应用性能的各种策略！* 🚀
```
```
```