# 第2章：Hooks深度应用与自定义Hooks 🎣

> _"Hooks是React的革命性特性，让函数组件拥有了类组件的所有能力！"_ ⚡

## 📚 本章导览

React Hooks彻底改变了我们编写React组件的方式，让状态管理和副作用处理变得更加简洁和强大。在Mall-Frontend项目中，我们将深入探索Hooks的高级用法，学会设计可复用的自定义Hooks，构建更优雅的React应用。

### 🎯 学习目标

通过本章学习，你将掌握：

- **内置Hooks深度应用** - useState、useEffect、useContext等的高级用法
- **性能优化Hooks** - useMemo、useCallback、useRef的最佳实践
- **自定义Hooks设计** - 抽象可复用的状态逻辑
- **Hooks组合模式** - 多个Hooks的协同工作
- **异步Hooks处理** - 处理异步操作和竞态条件
- **Hooks最佳实践** - 避免常见陷阱和性能问题
- **实战应用** - 在Mall-Frontend中的Hooks应用案例

### 🛠️ 技术栈概览

```typescript
{
  "hooks": "React 19.1.0",
  "patterns": ["Custom Hooks", "Compound Hooks", "State Machines"],
  "optimization": ["useMemo", "useCallback", "React.memo"],
  "async": ["useEffect", "useLayoutEffect", "Suspense"]
}
```

### 📖 本章目录

- [内置Hooks深度应用](#内置hooks深度应用)
- [自定义Hooks设计模式](#自定义hooks设计模式)
- [性能优化Hooks](#性能优化hooks)
- [异步处理与竞态条件](#异步处理与竞态条件)
- [Hooks组合模式](#hooks组合模式)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🔧 内置Hooks深度应用

### useState的高级用法

```typescript
import { useState, useCallback, Dispatch, SetStateAction } from 'react';

// 复杂状态管理
interface UserFormState {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  agreeTerms: boolean;
  errors: Record<string, string>;
}

// 状态更新器类型
type StateUpdater<T> = Dispatch<SetStateAction<T>>;

// 使用函数式更新
function useUserForm() {
  const [formState, setFormState] = useState<UserFormState>({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    agreeTerms: false,
    errors: {},
  });

  // 字段更新函数
  const updateField = useCallback((field: keyof UserFormState, value: any) => {
    setFormState(prev => ({
      ...prev,
      [field]: value,
      errors: {
        ...prev.errors,
        [field]: '', // 清除该字段的错误
      },
    }));
  }, []);

  // 批量更新
  const updateFields = useCallback((updates: Partial<UserFormState>) => {
    setFormState(prev => ({ ...prev, ...updates }));
  }, []);

  // 重置表单
  const resetForm = useCallback(() => {
    setFormState({
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      agreeTerms: false,
      errors: {},
    });
  }, []);

  // 设置错误
  const setErrors = useCallback((errors: Record<string, string>) => {
    setFormState(prev => ({ ...prev, errors }));
  }, []);

  return {
    formState,
    updateField,
    updateFields,
    resetForm,
    setErrors,
  };
}

// 使用示例
const RegisterForm: React.FC = () => {
  const { formState, updateField, resetForm, setErrors } = useUserForm();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await submitRegistration(formState);
      resetForm();
    } catch (error: any) {
      setErrors(error.fieldErrors || { general: error.message });
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        value={formState.username}
        onChange={(e) => updateField('username', e.target.value)}
        placeholder="用户名"
      />
      {formState.errors.username && (
        <span className="error">{formState.errors.username}</span>
      )}
      {/* 其他表单字段... */}
    </form>
  );
};
```

### useEffect的最佳实践

```typescript
import { useEffect, useRef, useCallback, useState } from 'react';

// 防止内存泄漏的useEffect
function useAsyncEffect<T>(
  asyncFn: () => Promise<T>,
  deps: React.DependencyList,
  onSuccess?: (data: T) => void,
  onError?: (error: Error) => void
) {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const cancelRef = useRef<boolean>(false);

  useEffect(() => {
    let isCancelled = false;
    cancelRef.current = false;

    const executeAsync = async () => {
      try {
        setLoading(true);
        setError(null);

        const result = await asyncFn();

        if (!isCancelled && !cancelRef.current) {
          setData(result);
          onSuccess?.(result);
        }
      } catch (err) {
        if (!isCancelled && !cancelRef.current) {
          const error = err as Error;
          setError(error);
          onError?.(error);
        }
      } finally {
        if (!isCancelled && !cancelRef.current) {
          setLoading(false);
        }
      }
    };

    executeAsync();

    return () => {
      isCancelled = true;
      cancelRef.current = true;
    };
  }, deps);

  const cancel = useCallback(() => {
    cancelRef.current = true;
  }, []);

  return { loading, data, error, cancel };
}

// Mall-Frontend中的商品详情获取
function useProductDetail(productId: number) {
  const {
    loading,
    data: product,
    error,
  } = useAsyncEffect(
    () => fetch(`/api/products/${productId}`).then(res => res.json()),
    [productId],
    product => {
      console.log('商品详情加载成功:', product.name);
    },
    error => {
      console.error('商品详情加载失败:', error);
    }
  );

  return { loading, product, error };
}
```

### useContext的高级模式

```typescript
import { createContext, useContext, useReducer, ReactNode } from 'react';

// 购物车状态管理
interface CartItem {
  id: number;
  product: Product;
  quantity: number;
  selected: boolean;
}

interface CartState {
  items: CartItem[];
  totalItems: number;
  totalPrice: number;
  loading: boolean;
  error: string | null;
}

type CartAction =
  | { type: 'ADD_ITEM'; payload: { product: Product; quantity: number } }
  | { type: 'REMOVE_ITEM'; payload: { itemId: number } }
  | { type: 'UPDATE_QUANTITY'; payload: { itemId: number; quantity: number } }
  | { type: 'TOGGLE_SELECT'; payload: { itemId: number } }
  | { type: 'SELECT_ALL'; payload: { selected: boolean } }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'CLEAR_CART' };

// Reducer函数
function cartReducer(state: CartState, action: CartAction): CartState {
  switch (action.type) {
    case 'ADD_ITEM': {
      const { product, quantity } = action.payload;
      const existingItemIndex = state.items.findIndex(
        item => item.product.id === product.id
      );

      let newItems: CartItem[];
      if (existingItemIndex > -1) {
        newItems = state.items.map((item, index) =>
          index === existingItemIndex
            ? { ...item, quantity: item.quantity + quantity }
            : item
        );
      } else {
        newItems = [
          ...state.items,
          {
            id: Date.now(),
            product,
            quantity,
            selected: true,
          },
        ];
      }

      return {
        ...state,
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce(
          (sum, item) => sum + parseFloat(item.product.price) * item.quantity,
          0
        ),
      };
    }

    case 'REMOVE_ITEM': {
      const newItems = state.items.filter(item => item.id !== action.payload.itemId);
      return {
        ...state,
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce(
          (sum, item) => sum + parseFloat(item.product.price) * item.quantity,
          0
        ),
      };
    }

    case 'UPDATE_QUANTITY': {
      const { itemId, quantity } = action.payload;
      if (quantity <= 0) {
        return cartReducer(state, { type: 'REMOVE_ITEM', payload: { itemId } });
      }

      const newItems = state.items.map(item =>
        item.id === itemId ? { ...item, quantity } : item
      );

      return {
        ...state,
        items: newItems,
        totalItems: newItems.reduce((sum, item) => sum + item.quantity, 0),
        totalPrice: newItems.reduce(
          (sum, item) => sum + parseFloat(item.product.price) * item.quantity,
          0
        ),
      };
    }

    case 'TOGGLE_SELECT': {
      const newItems = state.items.map(item =>
        item.id === action.payload.itemId
          ? { ...item, selected: !item.selected }
          : item
      );

      return { ...state, items: newItems };
    }

    case 'SELECT_ALL': {
      const newItems = state.items.map(item => ({
        ...item,
        selected: action.payload.selected,
      }));

      return { ...state, items: newItems };
    }

    case 'SET_LOADING':
      return { ...state, loading: action.payload };

    case 'SET_ERROR':
      return { ...state, error: action.payload };

    case 'CLEAR_CART':
      return {
        ...state,
        items: [],
        totalItems: 0,
        totalPrice: 0,
      };

    default:
      return state;
  }
}

// Context创建
interface CartContextValue {
  state: CartState;
  addItem: (product: Product, quantity: number) => void;
  removeItem: (itemId: number) => void;
  updateQuantity: (itemId: number, quantity: number) => void;
  toggleSelect: (itemId: number) => void;
  selectAll: (selected: boolean) => void;
  clearCart: () => void;
}

const CartContext = createContext<CartContextValue | undefined>(undefined);

// Provider组件
export function CartProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(cartReducer, {
    items: [],
    totalItems: 0,
    totalPrice: 0,
    loading: false,
    error: null,
  });

  const addItem = useCallback((product: Product, quantity: number) => {
    dispatch({ type: 'ADD_ITEM', payload: { product, quantity } });
  }, []);

  const removeItem = useCallback((itemId: number) => {
    dispatch({ type: 'REMOVE_ITEM', payload: { itemId } });
  }, []);

  const updateQuantity = useCallback((itemId: number, quantity: number) => {
    dispatch({ type: 'UPDATE_QUANTITY', payload: { itemId, quantity } });
  }, []);

  const toggleSelect = useCallback((itemId: number) => {
    dispatch({ type: 'TOGGLE_SELECT', payload: { itemId } });
  }, []);

  const selectAll = useCallback((selected: boolean) => {
    dispatch({ type: 'SELECT_ALL', payload: { selected } });
  }, []);

  const clearCart = useCallback(() => {
    dispatch({ type: 'CLEAR_CART' });
  }, []);

  const value: CartContextValue = {
    state,
    addItem,
    removeItem,
    updateQuantity,
    toggleSelect,
    selectAll,
    clearCart,
  };

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

// Hook使用
export function useCart(): CartContextValue {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
```

---

## 🎨 自定义Hooks设计模式

### 数据获取Hook

```typescript
import { useState, useEffect, useCallback, useRef } from 'react';

// 通用数据获取Hook
interface UseApiOptions<T> {
  initialData?: T;
  immediate?: boolean;
  onSuccess?: (data: T) => void;
  onError?: (error: Error) => void;
  transform?: (data: any) => T;
}

interface UseApiReturn<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
  execute: (...args: any[]) => Promise<T>;
  reset: () => void;
}

function useApi<T = any>(
  apiFunction: (...args: any[]) => Promise<any>,
  options: UseApiOptions<T> = {}
): UseApiReturn<T> {
  const {
    initialData = null,
    immediate = false,
    onSuccess,
    onError,
    transform,
  } = options;

  const [data, setData] = useState<T | null>(initialData);
  const [loading, setLoading] = useState(immediate);
  const [error, setError] = useState<Error | null>(null);
  const cancelRef = useRef<boolean>(false);

  const execute = useCallback(
    async (...args: any[]): Promise<T> => {
      try {
        setLoading(true);
        setError(null);
        cancelRef.current = false;

        const response = await apiFunction(...args);

        if (cancelRef.current) {
          throw new Error('Request cancelled');
        }

        const transformedData = transform ? transform(response) : response;
        setData(transformedData);
        onSuccess?.(transformedData);

        return transformedData;
      } catch (err) {
        if (!cancelRef.current) {
          const error = err as Error;
          setError(error);
          onError?.(error);
          throw error;
        }
        throw err;
      } finally {
        if (!cancelRef.current) {
          setLoading(false);
        }
      }
    },
    [apiFunction, transform, onSuccess, onError]
  );

  const reset = useCallback(() => {
    setData(initialData);
    setLoading(false);
    setError(null);
    cancelRef.current = true;
  }, [initialData]);

  useEffect(() => {
    if (immediate) {
      execute();
    }

    return () => {
      cancelRef.current = true;
    };
  }, [execute, immediate]);

  return { data, loading, error, execute, reset };
}

// Mall-Frontend中的商品API Hook
function useProducts() {
  const {
    data: products,
    loading,
    error,
    execute: fetchProducts,
  } = useApi(
    async (params: { category?: string; page?: number; search?: string }) => {
      const searchParams = new URLSearchParams();
      if (params.category) searchParams.set('category', params.category);
      if (params.page) searchParams.set('page', params.page.toString());
      if (params.search) searchParams.set('search', params.search);

      const response = await fetch(`/api/products?${searchParams}`);
      if (!response.ok) throw new Error('Failed to fetch products');
      return response.json();
    },
    {
      initialData: [],
      transform: response => response.data || [],
      onError: error => console.error('获取商品失败:', error),
    }
  );

  return { products, loading, error, fetchProducts };
}
```

### 表单处理Hook

```typescript
import { useState, useCallback, useMemo } from 'react';

// 验证规则类型
type ValidationRule<T> = (value: T) => string | null;

interface UseFormOptions<T> {
  initialValues: T;
  validationRules?: Partial<Record<keyof T, ValidationRule<any>[]>>;
  onSubmit?: (values: T) => Promise<void> | void;
}

interface UseFormReturn<T> {
  values: T;
  errors: Partial<Record<keyof T, string>>;
  touched: Partial<Record<keyof T, boolean>>;
  isValid: boolean;
  isSubmitting: boolean;
  setValue: (field: keyof T, value: any) => void;
  setValues: (values: Partial<T>) => void;
  setError: (field: keyof T, error: string) => void;
  setErrors: (errors: Partial<Record<keyof T, string>>) => void;
  handleChange: (
    field: keyof T
  ) => (e: React.ChangeEvent<HTMLInputElement>) => void;
  handleBlur: (field: keyof T) => () => void;
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  reset: () => void;
  validateField: (field: keyof T) => void;
  validateForm: () => boolean;
}

function useForm<T extends Record<string, any>>(
  options: UseFormOptions<T>
): UseFormReturn<T> {
  const { initialValues, validationRules = {}, onSubmit } = options;

  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 验证单个字段
  const validateField = useCallback(
    (field: keyof T) => {
      const rules = validationRules[field];
      if (!rules) return;

      const value = values[field];
      for (const rule of rules) {
        const error = rule(value);
        if (error) {
          setErrors(prev => ({ ...prev, [field]: error }));
          return;
        }
      }

      setErrors(prev => ({ ...prev, [field]: undefined }));
    },
    [values, validationRules]
  );

  // 验证整个表单
  const validateForm = useCallback(() => {
    const newErrors: Partial<Record<keyof T, string>> = {};
    let isValid = true;

    Object.keys(validationRules).forEach(field => {
      const rules = validationRules[field as keyof T];
      if (!rules) return;

      const value = values[field as keyof T];
      for (const rule of rules) {
        const error = rule(value);
        if (error) {
          newErrors[field as keyof T] = error;
          isValid = false;
          break;
        }
      }
    });

    setErrors(newErrors);
    return isValid;
  }, [values, validationRules]);

  // 设置字段值
  const setValue = useCallback((field: keyof T, value: any) => {
    setValues(prev => ({ ...prev, [field]: value }));
  }, []);

  // 批量设置值
  const setValuesCallback = useCallback((newValues: Partial<T>) => {
    setValues(prev => ({ ...prev, ...newValues }));
  }, []);

  // 设置字段错误
  const setError = useCallback((field: keyof T, error: string) => {
    setErrors(prev => ({ ...prev, [field]: error }));
  }, []);

  // 批量设置错误
  const setErrorsCallback = useCallback(
    (newErrors: Partial<Record<keyof T, string>>) => {
      setErrors(prev => ({ ...prev, ...newErrors }));
    },
    []
  );

  // 处理输入变化
  const handleChange = useCallback(
    (field: keyof T) => (e: React.ChangeEvent<HTMLInputElement>) => {
      const value =
        e.target.type === 'checkbox' ? e.target.checked : e.target.value;
      setValue(field, value);
    },
    [setValue]
  );

  // 处理失焦
  const handleBlur = useCallback(
    (field: keyof T) => () => {
      setTouched(prev => ({ ...prev, [field]: true }));
      validateField(field);
    },
    [validateField]
  );

  // 处理提交
  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();

      if (!validateForm()) {
        return;
      }

      if (!onSubmit) return;

      try {
        setIsSubmitting(true);
        await onSubmit(values);
      } catch (error) {
        console.error('表单提交失败:', error);
      } finally {
        setIsSubmitting(false);
      }
    },
    [validateForm, onSubmit, values]
  );

  // 重置表单
  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
    setIsSubmitting(false);
  }, [initialValues]);

  // 计算表单是否有效
  const isValid = useMemo(() => {
    return Object.keys(errors).length === 0;
  }, [errors]);

  return {
    values,
    errors,
    touched,
    isValid,
    isSubmitting,
    setValue,
    setValues: setValuesCallback,
    setError,
    setErrors: setErrorsCallback,
    handleChange,
    handleBlur,
    handleSubmit,
    reset,
    validateField,
    validateForm,
  };
}

// 验证规则
const validationRules = {
  required:
    (message = '此字段为必填项') =>
    (value: any) =>
      !value || (typeof value === 'string' && !value.trim()) ? message : null,

  minLength: (length: number, message?: string) => (value: string) =>
    value && value.length < length
      ? message || `最少需要${length}个字符`
      : null,

  email:
    (message = '请输入有效的邮箱地址') =>
    (value: string) =>
      value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) ? message : null,

  phone:
    (message = '请输入有效的手机号') =>
    (value: string) =>
      value && !/^1[3-9]\d{9}$/.test(value) ? message : null,
};

// 使用示例：用户注册表单
interface RegisterFormData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  agreeTerms: boolean;
}

function useRegisterForm() {
  return useForm<RegisterFormData>({
    initialValues: {
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      agreeTerms: false,
    },
    validationRules: {
      username: [validationRules.required(), validationRules.minLength(3)],
      email: [validationRules.required(), validationRules.email()],
      password: [validationRules.required(), validationRules.minLength(6)],
      confirmPassword: [
        validationRules.required(),
        (value: string, values: RegisterFormData) =>
          value !== values.password ? '两次输入的密码不一致' : null,
      ],
      agreeTerms: [(value: boolean) => (!value ? '请同意用户协议' : null)],
    },
    onSubmit: async values => {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      if (!response.ok) {
        throw new Error('注册失败');
      }
    },
  });
}
```

---

## 🚀 性能优化Hooks

### useMemo和useCallback的最佳实践

```typescript
import { useMemo, useCallback, useState, useEffect } from 'react';

// 复杂计算的memoization
function useExpensiveCalculation(data: any[], filters: any) {
  return useMemo(() => {
    console.log('执行复杂计算...');

    // 模拟复杂的数据处理
    return data
      .filter(item => {
        if (filters.category && item.category !== filters.category)
          return false;
        if (filters.minPrice && item.price < filters.minPrice) return false;
        if (filters.maxPrice && item.price > filters.maxPrice) return false;
        if (
          filters.search &&
          !item.name.toLowerCase().includes(filters.search.toLowerCase())
        )
          return false;
        return true;
      })
      .sort((a, b) => {
        switch (filters.sortBy) {
          case 'price':
            return filters.sortOrder === 'asc'
              ? a.price - b.price
              : b.price - a.price;
          case 'name':
            return filters.sortOrder === 'asc'
              ? a.name.localeCompare(b.name)
              : b.name.localeCompare(a.name);
          default:
            return 0;
        }
      });
  }, [data, filters]);
}

// 事件处理函数的优化
function useOptimizedEventHandlers() {
  const [selectedItems, setSelectedItems] = useState<number[]>([]);
  const [sortConfig, setSortConfig] = useState({ field: 'name', order: 'asc' });

  // 使用useCallback优化事件处理函数
  const handleItemSelect = useCallback((itemId: number, selected: boolean) => {
    setSelectedItems(prev =>
      selected ? [...prev, itemId] : prev.filter(id => id !== itemId)
    );
  }, []);

  const handleSelectAll = useCallback((items: any[], selected: boolean) => {
    setSelectedItems(selected ? items.map(item => item.id) : []);
  }, []);

  const handleSort = useCallback((field: string) => {
    setSortConfig(prev => ({
      field,
      order: prev.field === field && prev.order === 'asc' ? 'desc' : 'asc',
    }));
  }, []);

  // 批量操作
  const handleBatchDelete = useCallback(async () => {
    if (selectedItems.length === 0) return;

    try {
      await Promise.all(
        selectedItems.map(id => fetch(`/api/items/${id}`, { method: 'DELETE' }))
      );
      setSelectedItems([]);
    } catch (error) {
      console.error('批量删除失败:', error);
    }
  }, [selectedItems]);

  return {
    selectedItems,
    sortConfig,
    handleItemSelect,
    handleSelectAll,
    handleSort,
    handleBatchDelete,
  };
}
```

### 防抖和节流Hook

```typescript
import { useCallback, useEffect, useRef, useState } from 'react';

// 防抖Hook
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

// 防抖回调Hook
function useDebouncedCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const callbackRef = useRef(callback);
  const timeoutRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  return useCallback(
    ((...args) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        callbackRef.current(...args);
      }, delay);
    }) as T,
    [delay]
  );
}

// 节流Hook
function useThrottledCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const callbackRef = useRef(callback);
  const lastCallRef = useRef<number>(0);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  return useCallback(
    ((...args) => {
      const now = Date.now();
      if (now - lastCallRef.current >= delay) {
        lastCallRef.current = now;
        callbackRef.current(...args);
      }
    }) as T,
    [delay]
  );
}

// Mall-Frontend中的搜索Hook
function useProductSearch() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);

  // 防抖搜索查询
  const debouncedQuery = useDebounce(query, 300);

  // 防抖搜索函数
  const searchProducts = useDebouncedCallback(async (searchQuery: string) => {
    if (!searchQuery.trim()) {
      setResults([]);
      return;
    }

    try {
      setLoading(true);
      const response = await fetch(
        `/api/products/search?q=${encodeURIComponent(searchQuery)}`
      );
      const data = await response.json();
      setResults(data.products || []);
    } catch (error) {
      console.error('搜索失败:', error);
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, 300);

  // 当防抖查询改变时执行搜索
  useEffect(() => {
    searchProducts(debouncedQuery);
  }, [debouncedQuery, searchProducts]);

  // 节流的滚动处理
  const handleScroll = useThrottledCallback(() => {
    // 处理滚动加载更多
    const { scrollTop, scrollHeight, clientHeight } = document.documentElement;
    if (scrollTop + clientHeight >= scrollHeight - 100) {
      // 加载更多逻辑
      console.log('加载更多商品...');
    }
  }, 200);

  useEffect(() => {
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [handleScroll]);

  return {
    query,
    setQuery,
    results,
    loading,
    debouncedQuery,
  };
}
```

---

## 🎯 面试常考知识点

### 1. Hooks的规则和原理

**Q: React Hooks有哪些使用规则？为什么有这些规则？**

**A: Hooks的两个基本规则：**

1. **只在顶层调用Hooks** - 不要在循环、条件或嵌套函数中调用
2. **只在React函数中调用Hooks** - 函数组件或自定义Hooks中

```typescript
// ❌ 错误用法
function BadComponent({ condition }: { condition: boolean }) {
  if (condition) {
    const [state, setState] = useState(0); // 违反规则1
  }

  for (let i = 0; i < 3; i++) {
    useEffect(() => {}); // 违反规则1
  }

  return <div />;
}

// ✅ 正确用法
function GoodComponent({ condition }: { condition: boolean }) {
  const [state, setState] = useState(0);

  useEffect(() => {
    if (condition) {
      // 条件逻辑放在Hook内部
      setState(1);
    }
  }, [condition]);

  return <div />;
}
```

**原理**: React依赖Hooks的调用顺序来正确地将状态与组件实例关联。

### 2. useEffect的依赖数组

**Q: useEffect的依赖数组如何正确使用？**

**A: 依赖数组的最佳实践：**

```typescript
// ✅ 正确的依赖数组
function Component({ userId }: { userId: number }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    fetchUser(userId).then(setUser);
  }, [userId]); // 依赖userId

  // 使用useCallback避免不必要的重新渲染
  const handleClick = useCallback(() => {
    console.log(user);
  }, [user]); // 依赖user

  return <button onClick={handleClick}>Click</button>;
}

// ❌ 常见错误
function BadComponent({ userId }: { userId: number }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    fetchUser(userId).then(setUser);
  }, []); // 缺少userId依赖

  useEffect(() => {
    fetchUser(userId).then(setUser);
  }); // 缺少依赖数组，每次渲染都执行

  return <div />;
}
```

### 3. 自定义Hooks的设计原则

**Q: 如何设计一个好的自定义Hook？**

**A: 自定义Hooks设计原则：**

1. **单一职责** - 每个Hook只负责一个功能
2. **可复用性** - 抽象通用逻辑
3. **类型安全** - 完整的TypeScript类型定义
4. **错误处理** - 优雅的错误处理机制

```typescript
// ✅ 好的自定义Hook设计
function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((prev: T) => T)) => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.error(`Error reading localStorage key "${key}":`, error);
      return initialValue;
    }
  });

  const setValue = useCallback(
    (value: T | ((prev: T) => T)) => {
      try {
        const valueToStore =
          value instanceof Function ? value(storedValue) : value;
        setStoredValue(valueToStore);
        window.localStorage.setItem(key, JSON.stringify(valueToStore));
      } catch (error) {
        console.error(`Error setting localStorage key "${key}":`, error);
      }
    },
    [key, storedValue]
  );

  return [storedValue, setValue];
}
```

---

## 🏋️ 实战练习

### 练习1: 实现一个完整的购物车Hook

**题目**: 为Mall-Frontend实现一个功能完整的购物车管理Hook

**要求**:

1. 支持添加、删除、修改商品
2. 支持批量选择和操作
3. 本地存储持久化
4. 优化性能，避免不必要的重渲染
5. 完整的TypeScript类型支持

**解决方案**:

```typescript
import { useState, useCallback, useEffect, useMemo } from 'react';

interface CartItem {
  id: string;
  product: Product;
  quantity: number;
  selected: boolean;
  addedAt: number;
}

interface UseCartReturn {
  items: CartItem[];
  totalItems: number;
  totalPrice: number;
  selectedItems: CartItem[];
  selectedTotalPrice: number;
  addItem: (product: Product, quantity?: number) => void;
  removeItem: (itemId: string) => void;
  updateQuantity: (itemId: string, quantity: number) => void;
  toggleSelect: (itemId: string) => void;
  selectAll: (selected: boolean) => void;
  clearCart: () => void;
  clearSelected: () => void;
}

function useCart(): UseCartReturn {
  const [items, setItems] = useState<CartItem[]>([]);

  // 从localStorage加载购物车数据
  useEffect(() => {
    try {
      const savedCart = localStorage.getItem('mall-cart');
      if (savedCart) {
        const parsedCart = JSON.parse(savedCart);
        setItems(parsedCart);
      }
    } catch (error) {
      console.error('加载购物车数据失败:', error);
    }
  }, []);

  // 保存购物车数据到localStorage
  useEffect(() => {
    try {
      localStorage.setItem('mall-cart', JSON.stringify(items));
    } catch (error) {
      console.error('保存购物车数据失败:', error);
    }
  }, [items]);

  // 添加商品到购物车
  const addItem = useCallback((product: Product, quantity = 1) => {
    setItems(prevItems => {
      const existingItemIndex = prevItems.findIndex(
        item => item.product.id === product.id
      );

      if (existingItemIndex > -1) {
        // 更新现有商品数量
        return prevItems.map((item, index) =>
          index === existingItemIndex
            ? { ...item, quantity: item.quantity + quantity }
            : item
        );
      } else {
        // 添加新商品
        const newItem: CartItem = {
          id: `${product.id}-${Date.now()}`,
          product,
          quantity,
          selected: true,
          addedAt: Date.now(),
        };
        return [...prevItems, newItem];
      }
    });
  }, []);

  // 移除商品
  const removeItem = useCallback((itemId: string) => {
    setItems(prevItems => prevItems.filter(item => item.id !== itemId));
  }, []);

  // 更新商品数量
  const updateQuantity = useCallback(
    (itemId: string, quantity: number) => {
      if (quantity <= 0) {
        removeItem(itemId);
        return;
      }

      setItems(prevItems =>
        prevItems.map(item =>
          item.id === itemId ? { ...item, quantity } : item
        )
      );
    },
    [removeItem]
  );

  // 切换商品选中状态
  const toggleSelect = useCallback((itemId: string) => {
    setItems(prevItems =>
      prevItems.map(item =>
        item.id === itemId ? { ...item, selected: !item.selected } : item
      )
    );
  }, []);

  // 全选/取消全选
  const selectAll = useCallback((selected: boolean) => {
    setItems(prevItems => prevItems.map(item => ({ ...item, selected })));
  }, []);

  // 清空购物车
  const clearCart = useCallback(() => {
    setItems([]);
  }, []);

  // 清空选中的商品
  const clearSelected = useCallback(() => {
    setItems(prevItems => prevItems.filter(item => !item.selected));
  }, []);

  // 计算总商品数量
  const totalItems = useMemo(() => {
    return items.reduce((total, item) => total + item.quantity, 0);
  }, [items]);

  // 计算总价格
  const totalPrice = useMemo(() => {
    return items.reduce((total, item) => {
      const price = parseFloat(
        item.product.discount_price || item.product.price
      );
      return total + price * item.quantity;
    }, 0);
  }, [items]);

  // 获取选中的商品
  const selectedItems = useMemo(() => {
    return items.filter(item => item.selected);
  }, [items]);

  // 计算选中商品的总价格
  const selectedTotalPrice = useMemo(() => {
    return selectedItems.reduce((total, item) => {
      const price = parseFloat(
        item.product.discount_price || item.product.price
      );
      return total + price * item.quantity;
    }, 0);
  }, [selectedItems]);

  return {
    items,
    totalItems,
    totalPrice,
    selectedItems,
    selectedTotalPrice,
    addItem,
    removeItem,
    updateQuantity,
    toggleSelect,
    selectAll,
    clearCart,
    clearSelected,
  };
}

export default useCart;
```

这个练习展示了：

1. **完整的状态管理** - 购物车的所有操作
2. **性能优化** - 使用useMemo和useCallback
3. **持久化存储** - localStorage集成
4. **类型安全** - 完整的TypeScript类型定义
5. **实际应用** - 真实的电商购物车功能

---

## 📚 本章总结

通过本章学习，我们深入掌握了React Hooks的高级应用：

### 🎯 核心收获

1. **内置Hooks深度应用** 🔧
   - 掌握了useState、useEffect、useContext的高级用法
   - 学会了处理复杂状态和异步操作
   - 理解了Hooks的执行时机和依赖管理

2. **自定义Hooks设计** 🎨
   - 学会了抽象可复用的状态逻辑
   - 掌握了数据获取、表单处理等常见模式
   - 理解了Hooks组合和设计原则

3. **性能优化技巧** 🚀
   - 掌握了useMemo、useCallback的正确使用
   - 学会了防抖、节流等性能优化技术
   - 理解了如何避免不必要的重渲染

4. **实战应用** 💼
   - 在Mall-Frontend项目中应用Hooks模式
   - 构建了完整的购物车管理系统
   - 掌握了企业级Hooks的设计和实现

### 🚀 技术进阶

- **下一步学习**: 状态管理策略与最佳实践
- **实践建议**: 在项目中应用自定义Hooks抽象业务逻辑
- **深入方向**: React并发特性和Suspense

Hooks让React开发变得更加简洁和强大，是现代React开发的核心技能！ 🎉

---

_下一章我们将学习《状态管理策略与最佳实践》，探索复杂应用的状态管理解决方案！_ 🚀
