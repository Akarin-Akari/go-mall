# 📘 第3章：React组件设计与Hooks应用

> 掌握现代React开发，构建高质量的TypeScript组件

## 🎯 学习目标

通过本章学习，你将掌握：

- React组件的TypeScript类型定义
- 函数组件与类组件的最佳实践
- Hooks的深度应用和自定义Hooks
- 组件间通信和状态管理
- 性能优化技巧
- Mall-Frontend项目中的组件设计模式

## 📖 目录

- [React组件类型基础](#react组件类型基础)
- [函数组件与Props类型](#函数组件与props类型)
- [Hooks深度应用](#hooks深度应用)
- [自定义Hooks设计](#自定义hooks设计)
- [组件间通信](#组件间通信)
- [性能优化技巧](#性能优化技巧)
- [Mall-Frontend实战案例](#mall-frontend实战案例)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## ⚛️ React组件类型基础

### React组件的TypeScript类型

React组件在TypeScript中有多种类型定义方式：

```typescript
import React, { FC, Component, ReactNode, PropsWithChildren } from 'react';

// 1. 函数组件类型定义
interface ButtonProps {
  children: ReactNode;
  onClick: () => void;
  disabled?: boolean;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'small' | 'medium' | 'large';
}

// 方式1：使用FC（FunctionComponent）
const Button: FC<ButtonProps> = ({ children, onClick, disabled = false, variant = 'primary', size = 'medium' }) => {
  return (
    <button
      className={`btn btn-${variant} btn-${size}`}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
};

// 方式2：直接函数声明（推荐）
function Button2(props: ButtonProps): JSX.Element {
  const { children, onClick, disabled = false, variant = 'primary', size = 'medium' } = props;

  return (
    <button
      className={`btn btn-${variant} btn-${size}`}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

// 方式3：箭头函数（推荐）
const Button3 = (props: ButtonProps): JSX.Element => {
  // 组件实现
  return <button>...</button>;
};
```

### 🔄 框架对比：组件定义方式

```vue
<!-- Vue 3 + TypeScript - 单文件组件 -->
<template>
  <button
    :class="`btn btn-${variant} btn-${size}`"
    :disabled="disabled"
    @click="onClick"
  >
    <slot></slot>
  </button>
</template>

<script setup lang="ts">
interface ButtonProps {
  disabled?: boolean;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'small' | 'medium' | 'large';
}

// Props定义
const props = withDefaults(defineProps<ButtonProps>(), {
  disabled: false,
  variant: 'primary',
  size: 'medium',
});

// 事件定义
const emit = defineEmits<{
  click: [];
}>();

const onClick = () => {
  emit('click');
};
</script>

<style scoped>
.btn {
  /* 样式定义 */
}
</style>
```

```typescript
// Angular - 组件装饰器
import { Component, Input, Output, EventEmitter } from '@angular/core';

interface ButtonProps {
  disabled?: boolean;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'small' | 'medium' | 'large';
}

@Component({
  selector: 'app-button',
  template: `
    <button
      [class]="'btn btn-' + variant + ' btn-' + size"
      [disabled]="disabled"
      (click)="onClick()"
    >
      <ng-content></ng-content>
    </button>
  `,
  styleUrls: ['./button.component.css'],
})
export class ButtonComponent implements ButtonProps {
  @Input() disabled: boolean = false;
  @Input() variant: 'primary' | 'secondary' | 'danger' = 'primary';
  @Input() size: 'small' | 'medium' | 'large' = 'medium';

  @Output() buttonClick = new EventEmitter<void>();

  onClick(): void {
    this.buttonClick.emit();
  }
}
```

```svelte
<!-- Svelte + TypeScript -->
<script lang="ts">
  interface ButtonProps {
    disabled?: boolean;
    variant?: 'primary' | 'secondary' | 'danger';
    size?: 'small' | 'medium' | 'large';
  }

  export let disabled: boolean = false;
  export let variant: 'primary' | 'secondary' | 'danger' = 'primary';
  export let size: 'small' | 'medium' | 'large' = 'medium';

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function onClick() {
    dispatch('click');
  }
</script>

<button
  class="btn btn-{variant} btn-{size}"
  {disabled}
  on:click={onClick}
>
  <slot></slot>
</button>

<style>
  .btn {
    /* 样式定义 */
  }
</style>
```

```dart
// Flutter - Widget类
import 'package:flutter/material.dart';

enum ButtonVariant { primary, secondary, danger }
enum ButtonSize { small, medium, large }

class CustomButton extends StatelessWidget {
  final Widget child;
  final VoidCallback? onPressed;
  final bool disabled;
  final ButtonVariant variant;
  final ButtonSize size;

  const CustomButton({
    Key? key,
    required this.child,
    this.onPressed,
    this.disabled = false,
    this.variant = ButtonVariant.primary,
    this.size = ButtonSize.medium,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: disabled ? null : onPressed,
      style: ElevatedButton.styleFrom(
        backgroundColor: _getBackgroundColor(),
        padding: _getPadding(),
      ),
      child: child,
    );
  }

  Color _getBackgroundColor() {
    switch (variant) {
      case ButtonVariant.primary:
        return Colors.blue;
      case ButtonVariant.secondary:
        return Colors.grey;
      case ButtonVariant.danger:
        return Colors.red;
    }
  }

  EdgeInsets _getPadding() {
    switch (size) {
      case ButtonSize.small:
        return const EdgeInsets.symmetric(horizontal: 8, vertical: 4);
      case ButtonSize.medium:
        return const EdgeInsets.symmetric(horizontal: 16, vertical: 8);
      case ButtonSize.large:
        return const EdgeInsets.symmetric(horizontal: 24, vertical: 12);
    }
  }
}
```

**💡 组件系统对比：**

| 特性          | React + TS  | Vue 3 + TS         | Angular           | Svelte                  | Flutter    |
| ------------- | ----------- | ------------------ | ----------------- | ----------------------- | ---------- |
| **组件定义**  | 函数/类     | SFC/Composition    | 装饰器类          | 单文件                  | Widget类   |
| **Props类型** | 接口定义    | `defineProps<T>()` | `@Input()`        | `export let`            | 构造参数   |
| **事件处理**  | 回调函数    | `defineEmits<T>()` | `@Output()`       | `createEventDispatcher` | 回调函数   |
| **插槽/内容** | `children`  | `<slot>`           | `<ng-content>`    | `<slot>`                | `child`    |
| **样式隔离**  | CSS Modules | `scoped`           | ViewEncapsulation | 自动隔离                | Widget样式 |
| **类型安全**  | 编译时      | 编译时             | 编译时            | 编译时                  | 编译时     |

### 类组件的TypeScript定义

```typescript
// 类组件的Props和State类型
interface CounterProps {
  initialValue?: number;
  onValueChange?: (value: number) => void;
}

interface CounterState {
  count: number;
  isLoading: boolean;
}

// 类组件定义
class Counter extends Component<CounterProps, CounterState> {
  constructor(props: CounterProps) {
    super(props);
    this.state = {
      count: props.initialValue || 0,
      isLoading: false,
    };
  }

  handleIncrement = (): void => {
    this.setState(
      (prevState) => ({ count: prevState.count + 1 }),
      () => {
        // 状态更新后的回调
        this.props.onValueChange?.(this.state.count);
      }
    );
  };

  handleDecrement = (): void => {
    this.setState((prevState) => ({ count: prevState.count - 1 }));
  };

  render(): ReactNode {
    const { count, isLoading } = this.state;

    return (
      <div className="counter">
        <button onClick={this.handleDecrement} disabled={isLoading}>
          -
        </button>
        <span className="count">{count}</span>
        <button onClick={this.handleIncrement} disabled={isLoading}>
          +
        </button>
      </div>
    );
  }
}
```

### 组件Props的高级类型定义

```typescript
// 扩展HTML元素属性
interface CustomInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  helperText?: string;
}

const CustomInput: FC<CustomInputProps> = ({ label, error, helperText, className, ...inputProps }) => {
  return (
    <div className={`input-group ${className || ''}`}>
      <label className="input-label">{label}</label>
      <input
        className={`input ${error ? 'input-error' : ''}`}
        {...inputProps}
      />
      {error && <span className="error-text">{error}</span>}
      {helperText && <span className="helper-text">{helperText}</span>}
    </div>
  );
};

// 使用示例
<CustomInput
  label="用户名"
  placeholder="请输入用户名"
  value={username}
  onChange={(e) => setUsername(e.target.value)}
  error={usernameError}
  helperText="用户名长度为3-20个字符"
  required
/>
```

---

## 🎣 函数组件与Props类型

### Props类型的最佳实践

```typescript
// 1. 基础Props类型
interface UserCardProps {
  user: User;
  showEmail?: boolean;
  onEdit?: (user: User) => void;
  onDelete?: (userId: number) => void;
}

// 2. 带有children的Props
interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  size?: 'small' | 'medium' | 'large';
}

// 3. 泛型Props
interface ListProps<T> {
  items: T[];
  renderItem: (item: T, index: number) => ReactNode;
  keyExtractor: (item: T) => string | number;
  loading?: boolean;
  emptyText?: string;
}

function List<T>({ items, renderItem, keyExtractor, loading, emptyText }: ListProps<T>): JSX.Element {
  if (loading) {
    return <div className="loading">加载中...</div>;
  }

  if (items.length === 0) {
    return <div className="empty">{emptyText || '暂无数据'}</div>;
  }

  return (
    <div className="list">
      {items.map((item, index) => (
        <div key={keyExtractor(item)} className="list-item">
          {renderItem(item, index)}
        </div>
      ))}
    </div>
  );
}

// 使用泛型组件
<List<User>
  items={users}
  renderItem={(user) => <UserCard user={user} />}
  keyExtractor={(user) => user.id}
  loading={isLoading}
  emptyText="暂无用户数据"
/>
```

### 条件Props类型

```typescript
// 条件Props：根据某个属性决定其他属性是否必需
type ButtonProps =
  | {
      variant: 'link';
      href: string;
      onClick?: never;
    }
  | {
      variant?: 'primary' | 'secondary' | 'danger';
      href?: never;
      onClick: () => void;
    };

interface BaseButtonProps {
  children: ReactNode;
  disabled?: boolean;
  loading?: boolean;
}

type ConditionalButtonProps = BaseButtonProps & ButtonProps;

const ConditionalButton: FC<ConditionalButtonProps> = (props) => {
  if (props.variant === 'link') {
    return (
      <a href={props.href} className="btn-link" aria-disabled={props.disabled}>
        {props.children}
      </a>
    );
  }

  return (
    <button
      className={`btn btn-${props.variant || 'primary'}`}
      onClick={props.onClick}
      disabled={props.disabled || props.loading}
    >
      {props.loading && <span className="spinner" />}
      {props.children}
    </button>
  );
};

// 使用示例
<ConditionalButton variant="link" href="/home">
  首页
</ConditionalButton>

<ConditionalButton variant="primary" onClick={handleSubmit}>
  提交
</ConditionalButton>
```

### 组件Ref的类型定义

```typescript
import { forwardRef, useImperativeHandle, useRef } from 'react';

// 定义组件暴露的方法
interface InputRef {
  focus: () => void;
  blur: () => void;
  getValue: () => string;
  setValue: (value: string) => void;
}

interface InputProps {
  placeholder?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
}

// 使用forwardRef创建可引用的组件
const Input = forwardRef<InputRef, InputProps>(({ placeholder, defaultValue, onChange }, ref) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [value, setValue] = useState(defaultValue || '');

  // 暴露方法给父组件
  useImperativeHandle(ref, () => ({
    focus: () => inputRef.current?.focus(),
    blur: () => inputRef.current?.blur(),
    getValue: () => value,
    setValue: (newValue: string) => {
      setValue(newValue);
      onChange?.(newValue);
    },
  }));

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setValue(newValue);
    onChange?.(newValue);
  };

  return (
    <input
      ref={inputRef}
      value={value}
      onChange={handleChange}
      placeholder={placeholder}
    />
  );
});

// 父组件中使用
const ParentComponent: FC = () => {
  const inputRef = useRef<InputRef>(null);

  const handleFocus = () => {
    inputRef.current?.focus();
  };

  const handleGetValue = () => {
    const value = inputRef.current?.getValue();
    console.log('当前值:', value);
  };

  return (
    <div>
      <Input ref={inputRef} placeholder="请输入内容" />
      <button onClick={handleFocus}>聚焦输入框</button>
      <button onClick={handleGetValue}>获取值</button>
    </div>
  );
};
```

---

## 🎣 Hooks深度应用

### useState的类型应用

```typescript
import { useState, useCallback } from 'react';

// 1. 基础类型状态
const [count, setCount] = useState<number>(0);
const [name, setName] = useState<string>('');
const [isLoading, setIsLoading] = useState<boolean>(false);

// 2. 对象类型状态
interface UserForm {
  username: string;
  email: string;
  age: number;
}

const [userForm, setUserForm] = useState<UserForm>({
  username: '',
  email: '',
  age: 0,
});

// 更新对象状态的最佳实践
const updateUserForm = useCallback((updates: Partial<UserForm>) => {
  setUserForm(prev => ({ ...prev, ...updates }));
}, []);

// 3. 数组类型状态
const [users, setUsers] = useState<User[]>([]);

// 添加用户
const addUser = useCallback((user: User) => {
  setUsers(prev => [...prev, user]);
}, []);

// 更新用户
const updateUser = useCallback((userId: number, updates: Partial<User>) => {
  setUsers(prev =>
    prev.map(user => (user.id === userId ? { ...user, ...updates } : user))
  );
}, []);

// 删除用户
const removeUser = useCallback((userId: number) => {
  setUsers(prev => prev.filter(user => user.id !== userId));
}, []);

// 4. 联合类型状态
type LoadingState = 'idle' | 'loading' | 'success' | 'error';
const [loadingState, setLoadingState] = useState<LoadingState>('idle');

// 5. 可选类型状态
const [selectedUser, setSelectedUser] = useState<User | null>(null);
```

### useEffect的类型应用

```typescript
import { useEffect, useRef, DependencyList } from 'react';

// 1. 基础useEffect
useEffect(() => {
  // 副作用逻辑
  console.log('组件挂载或更新');

  // 清理函数
  return () => {
    console.log('组件卸载或依赖变化');
  };
}, []); // 依赖数组

// 2. 数据获取的useEffect
useEffect(() => {
  let isCancelled = false;

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const response = await api.getUsers();

      if (!isCancelled) {
        setUsers(response.data);
      }
    } catch (error) {
      if (!isCancelled) {
        setError(error.message);
      }
    } finally {
      if (!isCancelled) {
        setIsLoading(false);
      }
    }
  };

  fetchData();

  return () => {
    isCancelled = true;
  };
}, []);

// 3. 事件监听的useEffect
useEffect(() => {
  const handleResize = () => {
    setWindowSize({
      width: window.innerWidth,
      height: window.innerHeight,
    });
  };

  window.addEventListener('resize', handleResize);

  return () => {
    window.removeEventListener('resize', handleResize);
  };
}, []);

// 4. 定时器的useEffect
useEffect(() => {
  const timer = setInterval(() => {
    setCurrentTime(new Date());
  }, 1000);

  return () => {
    clearInterval(timer);
  };
}, []);
```

### useCallback和useMemo的类型应用

```typescript
import { useCallback, useMemo, useState } from 'react';

interface Product {
  id: number;
  name: string;
  price: number;
  category: string;
}

const ProductList: FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [sortBy, setSortBy] = useState<keyof Product>('name');

  // useCallback：缓存函数
  const handleSearch = useCallback((term: string) => {
    setSearchTerm(term);
  }, []);

  const handleSort = useCallback((field: keyof Product) => {
    setSortBy(field);
  }, []);

  const handleAddProduct = useCallback((product: Omit<Product, 'id'>) => {
    const newProduct: Product = {
      ...product,
      id: Date.now(),
    };
    setProducts(prev => [...prev, newProduct]);
  }, []);

  // useMemo：缓存计算结果
  const filteredProducts = useMemo(() => {
    return products.filter(product =>
      product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      product.category.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [products, searchTerm]);

  const sortedProducts = useMemo(() => {
    return [...filteredProducts].sort((a, b) => {
      const aValue = a[sortBy];
      const bValue = b[sortBy];

      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return aValue.localeCompare(bValue);
      }

      if (typeof aValue === 'number' && typeof bValue === 'number') {
        return aValue - bValue;
      }

      return 0;
    });
  }, [filteredProducts, sortBy]);

  const totalValue = useMemo(() => {
    return sortedProducts.reduce((sum, product) => sum + product.price, 0);
  }, [sortedProducts]);

  return (
    <div>
      <SearchInput onSearch={handleSearch} />
      <SortSelector onSort={handleSort} />
      <div>总价值: ¥{totalValue.toFixed(2)}</div>
      <ProductGrid products={sortedProducts} />
    </div>
  );
};
```

### useReducer的类型应用

```typescript
import { useReducer, Reducer } from 'react';

// 定义状态类型
interface TodoState {
  todos: Todo[];
  filter: 'all' | 'active' | 'completed';
  isLoading: boolean;
  error: string | null;
}

// 定义Action类型
type TodoAction =
  | { type: 'ADD_TODO'; payload: { text: string } }
  | { type: 'TOGGLE_TODO'; payload: { id: number } }
  | { type: 'DELETE_TODO'; payload: { id: number } }
  | { type: 'SET_FILTER'; payload: { filter: TodoState['filter'] } }
  | { type: 'SET_LOADING'; payload: { isLoading: boolean } }
  | { type: 'SET_ERROR'; payload: { error: string | null } }
  | { type: 'LOAD_TODOS_SUCCESS'; payload: { todos: Todo[] } };

// 定义Reducer
const todoReducer: Reducer<TodoState, TodoAction> = (state, action) => {
  switch (action.type) {
    case 'ADD_TODO':
      return {
        ...state,
        todos: [
          ...state.todos,
          {
            id: Date.now(),
            text: action.payload.text,
            completed: false,
            createdAt: new Date(),
          },
        ],
      };

    case 'TOGGLE_TODO':
      return {
        ...state,
        todos: state.todos.map(todo =>
          todo.id === action.payload.id
            ? { ...todo, completed: !todo.completed }
            : todo
        ),
      };

    case 'DELETE_TODO':
      return {
        ...state,
        todos: state.todos.filter(todo => todo.id !== action.payload.id),
      };

    case 'SET_FILTER':
      return {
        ...state,
        filter: action.payload.filter,
      };

    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload.isLoading,
      };

    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload.error,
      };

    case 'LOAD_TODOS_SUCCESS':
      return {
        ...state,
        todos: action.payload.todos,
        isLoading: false,
        error: null,
      };

    default:
      return state;
  }
};

// 初始状态
const initialState: TodoState = {
  todos: [],
  filter: 'all',
  isLoading: false,
  error: null,
};

// 使用useReducer
const TodoApp: FC = () => {
  const [state, dispatch] = useReducer(todoReducer, initialState);

  const addTodo = useCallback((text: string) => {
    dispatch({ type: 'ADD_TODO', payload: { text } });
  }, []);

  const toggleTodo = useCallback((id: number) => {
    dispatch({ type: 'TOGGLE_TODO', payload: { id } });
  }, []);

  const deleteTodo = useCallback((id: number) => {
    dispatch({ type: 'DELETE_TODO', payload: { id } });
  }, []);

  const setFilter = useCallback((filter: TodoState['filter']) => {
    dispatch({ type: 'SET_FILTER', payload: { filter } });
  }, []);

  // 过滤后的todos
  const filteredTodos = useMemo(() => {
    switch (state.filter) {
      case 'active':
        return state.todos.filter(todo => !todo.completed);
      case 'completed':
        return state.todos.filter(todo => todo.completed);
      default:
        return state.todos;
    }
  }, [state.todos, state.filter]);

  return (
    <div className="todo-app">
      <TodoInput onAdd={addTodo} />
      <TodoFilter currentFilter={state.filter} onFilterChange={setFilter} />
      <TodoList
        todos={filteredTodos}
        onToggle={toggleTodo}
        onDelete={deleteTodo}
      />
      {state.isLoading && <div>加载中...</div>}
      {state.error && <div className="error">{state.error}</div>}
    </div>
  );
};
```

---

## 🔧 自定义Hooks设计

### 数据获取Hook

````typescript
import { useState, useEffect, useCallback } from 'react';

// 通用数据获取Hook
interface UseApiOptions<T> {
  initialData?: T;
  immediate?: boolean;
  onSuccess?: (data: T) => void;
  onError?: (error: Error) => void;
}

interface UseApiReturn<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
  execute: () => Promise<void>;
  reset: () => void;
}

function useApi<T>(
  apiFunction: () => Promise<T>,
  options: UseApiOptions<T> = {}
): UseApiReturn<T> {
  const { initialData = null, immediate = true, onSuccess, onError } = options;

  const [data, setData] = useState<T | null>(initialData);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const result = await apiFunction();
      setData(result);
      onSuccess?.(result);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error');
      setError(error);
      onError?.(error);
    } finally {
      setLoading(false);
    }
  }, [apiFunction, onSuccess, onError]);

  const reset = useCallback(() => {
    setData(initialData);
    setLoading(false);
    setError(null);
  }, [initialData]);

  useEffect(() => {
    if (immediate) {
      execute();
    }
  }, [execute, immediate]);

  return { data, loading, error, execute, reset };
}

// 使用示例
const UserProfile: FC<{ userId: number }> = ({ userId }) => {
  const {
    data: user,
    loading,
    error,
    execute: refetchUser,
  } = useApi(
    () => api.getUserDetail(userId),
    {
      onSuccess: (user) => {
        console.log('用户数据加载成功:', user);
      },
      onError: (error) => {
        console.error('用户数据加载失败:', error);
      },
    }
  );

  if (loading) return <div>加载中...</div>;
  if (error) return <div>错误: {error.message}</div>;
  if (!user) return <div>用户不存在</div>;

  return (
    <div>
      <h1>{user.username}</h1>
      <p>{user.email}</p>
      <button onClick={refetchUser}>刷新</button>
    </div>
  );
};

### 表单处理Hook

```typescript
import { useState, useCallback, ChangeEvent } from 'react';

// 表单验证规则类型
type ValidationRule<T> = (value: T) => string | null;

// 表单配置类型
interface FormConfig<T> {
  initialValues: T;
  validationRules?: {
    [K in keyof T]?: ValidationRule<T[K]>[];
  };
  onSubmit?: (values: T) => Promise<void> | void;
}

// 表单Hook返回类型
interface UseFormReturn<T> {
  values: T;
  errors: Partial<Record<keyof T, string>>;
  touched: Partial<Record<keyof T, boolean>>;
  isSubmitting: boolean;
  isValid: boolean;
  handleChange: (field: keyof T) => (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
  handleBlur: (field: keyof T) => () => void;
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  setFieldValue: (field: keyof T, value: T[keyof T]) => void;
  setFieldError: (field: keyof T, error: string) => void;
  resetForm: () => void;
  validateField: (field: keyof T) => void;
  validateForm: () => boolean;
}

function useForm<T extends Record<string, any>>(config: FormConfig<T>): UseFormReturn<T> {
  const { initialValues, validationRules = {}, onSubmit } = config;

  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // 验证单个字段
  const validateField = useCallback((field: keyof T) => {
    const rules = validationRules[field];
    if (!rules) return;

    const value = values[field];
    let error: string | null = null;

    for (const rule of rules) {
      error = rule(value);
      if (error) break;
    }

    setErrors(prev => ({
      ...prev,
      [field]: error,
    }));
  }, [values, validationRules]);

  // 验证整个表单
  const validateForm = useCallback((): boolean => {
    const newErrors: Partial<Record<keyof T, string>> = {};
    let isValid = true;

    Object.keys(validationRules).forEach((field) => {
      const rules = validationRules[field as keyof T];
      if (!rules) return;

      const value = values[field as keyof T];
      let error: string | null = null;

      for (const rule of rules) {
        error = rule(value);
        if (error) break;
      }

      if (error) {
        newErrors[field as keyof T] = error;
        isValid = false;
      }
    });

    setErrors(newErrors);
    return isValid;
  }, [values, validationRules]);

  // 处理字段变化
  const handleChange = useCallback((field: keyof T) => {
    return (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = e.target.type === 'checkbox'
        ? (e.target as HTMLInputElement).checked
        : e.target.value;

      setValues(prev => ({
        ...prev,
        [field]: value,
      }));

      // 如果字段已经被触摸过，立即验证
      if (touched[field]) {
        setTimeout(() => validateField(field), 0);
      }
    };
  }, [touched, validateField]);

  // 处理字段失焦
  const handleBlur = useCallback((field: keyof T) => {
    return () => {
      setTouched(prev => ({
        ...prev,
        [field]: true,
      }));
      validateField(field);
    };
  }, [validateField]);

  // 设置字段值
  const setFieldValue = useCallback((field: keyof T, value: T[keyof T]) => {
    setValues(prev => ({
      ...prev,
      [field]: value,
    }));
  }, []);

  // 设置字段错误
  const setFieldError = useCallback((field: keyof T, error: string) => {
    setErrors(prev => ({
      ...prev,
      [field]: error,
    }));
  }, []);

  // 重置表单
  const resetForm = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
    setIsSubmitting(false);
  }, [initialValues]);

  // 处理表单提交
  const handleSubmit = useCallback(async (e: React.FormEvent) => {
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
  }, [validateForm, onSubmit, values]);

  // 计算表单是否有效
  const isValid = Object.keys(errors).length === 0;

  return {
    values,
    errors,
    touched,
    isSubmitting,
    isValid,
    handleChange,
    handleBlur,
    handleSubmit,
    setFieldValue,
    setFieldError,
    resetForm,
    validateField,
    validateForm,
  };
}

// 常用验证规则
export const validationRules = {
  required: <T>(message = '此字段为必填项') => (value: T): string | null => {
    if (value === null || value === undefined || value === '') {
      return message;
    }
    return null;
  },

  minLength: (min: number, message?: string) => (value: string): string | null => {
    if (value && value.length < min) {
      return message || `最少需要${min}个字符`;
    }
    return null;
  },

  maxLength: (max: number, message?: string) => (value: string): string | null => {
    if (value && value.length > max) {
      return message || `最多允许${max}个字符`;
    }
    return null;
  },

  email: (message = '请输入有效的邮箱地址') => (value: string): string | null => {
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      return message;
    }
    return null;
  },

  pattern: (regex: RegExp, message: string) => (value: string): string | null => {
    if (value && !regex.test(value)) {
      return message;
    }
    return null;
  },
};

// 使用示例
interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

const LoginComponent: FC = () => {
  const form = useForm<LoginForm>({
    initialValues: {
      username: '',
      password: '',
      remember: false,
    },
    validationRules: {
      username: [
        validationRules.required('用户名不能为空'),
        validationRules.minLength(3, '用户名至少3个字符'),
      ],
      password: [
        validationRules.required('密码不能为空'),
        validationRules.minLength(6, '密码至少6个字符'),
      ],
    },
    onSubmit: async (values) => {
      console.log('提交表单:', values);
      // 调用登录API
      await api.auth.login(values);
    },
  });

  return (
    <form onSubmit={form.handleSubmit}>
      <div>
        <input
          type="text"
          placeholder="用户名"
          value={form.values.username}
          onChange={form.handleChange('username')}
          onBlur={form.handleBlur('username')}
        />
        {form.touched.username && form.errors.username && (
          <span className="error">{form.errors.username}</span>
        )}
      </div>

      <div>
        <input
          type="password"
          placeholder="密码"
          value={form.values.password}
          onChange={form.handleChange('password')}
          onBlur={form.handleBlur('password')}
        />
        {form.touched.password && form.errors.password && (
          <span className="error">{form.errors.password}</span>
        )}
      </div>

      <div>
        <label>
          <input
            type="checkbox"
            checked={form.values.remember}
            onChange={form.handleChange('remember')}
          />
          记住我
        </label>
      </div>

      <button type="submit" disabled={!form.isValid || form.isSubmitting}>
        {form.isSubmitting ? '登录中...' : '登录'}
      </button>
    </form>
  );
};

### 本地存储Hook

```typescript
import { useState, useEffect, useCallback } from 'react';

// 本地存储Hook
function useLocalStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((prev: T) => T)) => void, () => void] {
  // 获取初始值
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.error(`Error reading localStorage key "${key}":`, error);
      return initialValue;
    }
  });

  // 设置值
  const setValue = useCallback((value: T | ((prev: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.error(`Error setting localStorage key "${key}":`, error);
    }
  }, [key, storedValue]);

  // 删除值
  const removeValue = useCallback(() => {
    try {
      window.localStorage.removeItem(key);
      setStoredValue(initialValue);
    } catch (error) {
      console.error(`Error removing localStorage key "${key}":`, error);
    }
  }, [key, initialValue]);

  return [storedValue, setValue, removeValue];
}

// 会话存储Hook
function useSessionStorage<T>(
  key: string,
  initialValue: T
): [T, (value: T | ((prev: T) => T)) => void, () => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.sessionStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.error(`Error reading sessionStorage key "${key}":`, error);
      return initialValue;
    }
  });

  const setValue = useCallback((value: T | ((prev: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.sessionStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.error(`Error setting sessionStorage key "${key}":`, error);
    }
  }, [key, storedValue]);

  const removeValue = useCallback(() => {
    try {
      window.sessionStorage.removeItem(key);
      setStoredValue(initialValue);
    } catch (error) {
      console.error(`Error removing sessionStorage key "${key}":`, error);
    }
  }, [key, initialValue]);

  return [storedValue, setValue, removeValue];
}

// 使用示例
const UserPreferences: FC = () => {
  const [theme, setTheme, removeTheme] = useLocalStorage<'light' | 'dark'>('theme', 'light');
  const [language, setLanguage] = useLocalStorage<string>('language', 'zh-CN');
  const [tempData, setTempData, removeTempData] = useSessionStorage<any>('tempData', null);

  return (
    <div>
      <div>
        <label>主题:</label>
        <select value={theme} onChange={(e) => setTheme(e.target.value as 'light' | 'dark')}>
          <option value="light">浅色</option>
          <option value="dark">深色</option>
        </select>
        <button onClick={removeTheme}>重置主题</button>
      </div>

      <div>
        <label>语言:</label>
        <select value={language} onChange={(e) => setLanguage(e.target.value)}>
          <option value="zh-CN">中文</option>
          <option value="en-US">English</option>
        </select>
      </div>
    </div>
  );
};

### 防抖和节流Hook

```typescript
import { useCallback, useEffect, useRef } from 'react';

// 防抖Hook
function useDebounce<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout>();

  const debouncedCallback = useCallback((...args: Parameters<T>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(() => {
      callback(...args);
    }, delay);
  }, [callback, delay]) as T;

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return debouncedCallback;
}

// 节流Hook
function useThrottle<T extends (...args: any[]) => any>(
  callback: T,
  delay: number
): T {
  const lastCallRef = useRef<number>(0);

  const throttledCallback = useCallback((...args: Parameters<T>) => {
    const now = Date.now();

    if (now - lastCallRef.current >= delay) {
      lastCallRef.current = now;
      callback(...args);
    }
  }, [callback, delay]) as T;

  return throttledCallback;
}

// 使用示例
const SearchComponent: FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [results, setResults] = useState<any[]>([]);

  // 防抖搜索
  const debouncedSearch = useDebounce(async (term: string) => {
    if (term.trim()) {
      const response = await api.search(term);
      setResults(response.data);
    } else {
      setResults([]);
    }
  }, 300);

  // 节流滚动处理
  const throttledScrollHandler = useThrottle(() => {
    console.log('滚动事件处理');
  }, 100);

  useEffect(() => {
    debouncedSearch(searchTerm);
  }, [searchTerm, debouncedSearch]);

  useEffect(() => {
    window.addEventListener('scroll', throttledScrollHandler);
    return () => {
      window.removeEventListener('scroll', throttledScrollHandler);
    };
  }, [throttledScrollHandler]);

  return (
    <div>
      <input
        type="text"
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
        placeholder="搜索..."
      />
      <div>
        {results.map((result, index) => (
          <div key={index}>{result.name}</div>
        ))}
      </div>
    </div>
  );
};

---

## 🔗 组件间通信

### Props传递

```typescript
// 父子组件通信
interface ParentProps {
  initialCount?: number;
}

const Parent: FC<ParentProps> = ({ initialCount = 0 }) => {
  const [count, setCount] = useState(initialCount);
  const [message, setMessage] = useState('');

  const handleCountChange = useCallback((newCount: number) => {
    setCount(newCount);
    setMessage(`计数已更新为: ${newCount}`);
  }, []);

  return (
    <div>
      <h2>父组件</h2>
      <p>当前计数: {count}</p>
      <p>消息: {message}</p>

      <Child
        count={count}
        onCountChange={handleCountChange}
        disabled={count >= 10}
      />
    </div>
  );
};

interface ChildProps {
  count: number;
  onCountChange: (count: number) => void;
  disabled?: boolean;
}

const Child: FC<ChildProps> = ({ count, onCountChange, disabled = false }) => {
  const handleIncrement = () => {
    onCountChange(count + 1);
  };

  const handleDecrement = () => {
    onCountChange(count - 1);
  };

  return (
    <div>
      <h3>子组件</h3>
      <button onClick={handleDecrement} disabled={disabled || count <= 0}>
        -
      </button>
      <span>{count}</span>
      <button onClick={handleIncrement} disabled={disabled}>
        +
      </button>
    </div>
  );
};

### Context API通信

```typescript
import { createContext, useContext, ReactNode } from 'react';

// 定义Context类型
interface ThemeContextType {
  theme: 'light' | 'dark';
  toggleTheme: () => void;
  colors: {
    primary: string;
    secondary: string;
    background: string;
    text: string;
  };
}

// 创建Context
const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

// Context Provider组件
interface ThemeProviderProps {
  children: ReactNode;
}

const ThemeProvider: FC<ThemeProviderProps> = ({ children }) => {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  const toggleTheme = useCallback(() => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light');
  }, []);

  const colors = useMemo(() => {
    return theme === 'light'
      ? {
          primary: '#007bff',
          secondary: '#6c757d',
          background: '#ffffff',
          text: '#333333',
        }
      : {
          primary: '#0d6efd',
          secondary: '#6c757d',
          background: '#1a1a1a',
          text: '#ffffff',
        };
  }, [theme]);

  const value: ThemeContextType = {
    theme,
    toggleTheme,
    colors,
  };

  return (
    <ThemeContext.Provider value={value}>
      <div style={{ backgroundColor: colors.background, color: colors.text }}>
        {children}
      </div>
    </ThemeContext.Provider>
  );
};

// 自定义Hook使用Context
const useTheme = (): ThemeContextType => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
};

// 使用Context的组件
const ThemedButton: FC<{ children: ReactNode; onClick?: () => void }> = ({ children, onClick }) => {
  const { colors } = useTheme();

  return (
    <button
      onClick={onClick}
      style={{
        backgroundColor: colors.primary,
        color: colors.background,
        border: 'none',
        padding: '8px 16px',
        borderRadius: '4px',
        cursor: 'pointer',
      }}
    >
      {children}
    </button>
  );
};

const ThemeToggle: FC = () => {
  const { theme, toggleTheme } = useTheme();

  return (
    <ThemedButton onClick={toggleTheme}>
      切换到{theme === 'light' ? '深色' : '浅色'}主题
    </ThemedButton>
  );
};

// 应用根组件
const App: FC = () => {
  return (
    <ThemeProvider>
      <div>
        <h1>主题切换示例</h1>
        <ThemeToggle />
        <ThemedButton>普通按钮</ThemedButton>
      </div>
    </ThemeProvider>
  );
};
````

### 事件总线通信

````typescript
// 事件总线类型定义
interface EventBusEvents {
  'user:login': { user: User };
  'user:logout': { userId: number };
  'cart:update': { itemCount: number };
  'notification:show': { message: string; type: 'success' | 'error' | 'warning' };
}

// 事件总线Hook
function useEventBus() {
  const eventBus = useRef<{
    listeners: Map<string, Function[]>;
    emit: <K extends keyof EventBusEvents>(event: K, data: EventBusEvents[K]) => void;
    on: <K extends keyof EventBusEvents>(event: K, callback: (data: EventBusEvents[K]) => void) => () => void;
    off: (event: string, callback: Function) => void;
  }>();

  if (!eventBus.current) {
    const listeners = new Map<string, Function[]>();

    eventBus.current = {
      listeners,
      emit: (event, data) => {
        const eventListeners = listeners.get(event);
        if (eventListeners) {
          eventListeners.forEach(callback => callback(data));
        }
      },
      on: (event, callback) => {
        if (!listeners.has(event)) {
          listeners.set(event, []);
        }
        listeners.get(event)!.push(callback);

        return () => {
          const eventListeners = listeners.get(event);
          if (eventListeners) {
            const index = eventListeners.indexOf(callback);
            if (index > -1) {
              eventListeners.splice(index, 1);
            }
          }
        };
      },
      off: (event, callback) => {
        const eventListeners = listeners.get(event);
        if (eventListeners) {
          const index = eventListeners.indexOf(callback);
          if (index > -1) {
            eventListeners.splice(index, 1);
          }
        }
      },
    };
  }

  return eventBus.current;
}

// 使用事件总线的组件
const LoginComponent: FC = () => {
  const eventBus = useEventBus();

  const handleLogin = async (userData: LoginRequest) => {
    try {
      const response = await api.auth.login(userData);
      const user = response.data.user;

      // 发布登录成功事件
      eventBus.emit('user:login', { user });
      eventBus.emit('notification:show', {
        message: '登录成功',
        type: 'success'
      });
    } catch (error) {
      eventBus.emit('notification:show', {
        message: '登录失败',
        type: 'error'
      });
    }
  };

  return (
    <div>
      {/* 登录表单 */}
    </div>
  );
};

const NotificationComponent: FC = () => {
  const [notifications, setNotifications] = useState<Array<{
    id: string;
    message: string;
    type: 'success' | 'error' | 'warning';
  }>>([]);
  const eventBus = useEventBus();

  useEffect(() => {
    const unsubscribe = eventBus.on('notification:show', ({ message, type }) => {
      const id = Math.random().toString(36).substr(2, 9);
      setNotifications(prev => [...prev, { id, message, type }]);

      // 3秒后自动移除通知
      setTimeout(() => {
        setNotifications(prev => prev.filter(n => n.id !== id));
      }, 3000);
    });

    return unsubscribe;
  }, [eventBus]);

  return (
    <div className="notifications">
      {notifications.map(notification => (
        <div key={notification.id} className={`notification notification-${notification.type}`}>
          {notification.message}
        </div>
      ))}
    </div>
  );
};

---

## ⚡ 性能优化技巧

### React.memo优化

```typescript
import { memo, useMemo } from 'react';

// 基础memo使用
interface UserCardProps {
  user: User;
  onEdit?: (user: User) => void;
  onDelete?: (userId: number) => void;
}

const UserCard = memo<UserCardProps>(({ user, onEdit, onDelete }) => {
  console.log('UserCard渲染:', user.username);

  return (
    <div className="user-card">
      <h3>{user.username}</h3>
      <p>{user.email}</p>
      {onEdit && (
        <button onClick={() => onEdit(user)}>编辑</button>
      )}
      {onDelete && (
        <button onClick={() => onDelete(user.id)}>删除</button>
      )}
    </div>
  );
});

// 自定义比较函数的memo
interface ProductCardProps {
  product: Product;
  isSelected: boolean;
  onSelect: (productId: number) => void;
}

const ProductCard = memo<ProductCardProps>(
  ({ product, isSelected, onSelect }) => {
    return (
      <div className={`product-card ${isSelected ? 'selected' : ''}`}>
        <h3>{product.name}</h3>
        <p>¥{product.price}</p>
        <button onClick={() => onSelect(product.id)}>
          {isSelected ? '取消选择' : '选择'}
        </button>
      </div>
    );
  },
  (prevProps, nextProps) => {
    // 自定义比较逻辑
    return (
      prevProps.product.id === nextProps.product.id &&
      prevProps.product.name === nextProps.product.name &&
      prevProps.product.price === nextProps.product.price &&
      prevProps.isSelected === nextProps.isSelected
    );
  }
);

// 使用memo的列表组件
interface ProductListProps {
  products: Product[];
  selectedIds: number[];
  onSelectProduct: (productId: number) => void;
}

const ProductList: FC<ProductListProps> = ({ products, selectedIds, onSelectProduct }) => {
  // 使用useMemo优化选择状态计算
  const selectedSet = useMemo(() => new Set(selectedIds), [selectedIds]);

  return (
    <div className="product-list">
      {products.map(product => (
        <ProductCard
          key={product.id}
          product={product}
          isSelected={selectedSet.has(product.id)}
          onSelect={onSelectProduct}
        />
      ))}
    </div>
  );
};
````

### 虚拟滚动优化

```typescript
import { useState, useEffect, useMemo, useCallback } from 'react';

interface VirtualScrollProps<T> {
  items: T[];
  itemHeight: number;
  containerHeight: number;
  renderItem: (item: T, index: number) => ReactNode;
  overscan?: number;
}

function VirtualScroll<T>({
  items,
  itemHeight,
  containerHeight,
  renderItem,
  overscan = 5,
}: VirtualScrollProps<T>): JSX.Element {
  const [scrollTop, setScrollTop] = useState(0);

  // 计算可见范围
  const visibleRange = useMemo(() => {
    const startIndex = Math.floor(scrollTop / itemHeight);
    const endIndex = Math.min(
      startIndex + Math.ceil(containerHeight / itemHeight),
      items.length - 1
    );

    return {
      start: Math.max(0, startIndex - overscan),
      end: Math.min(items.length - 1, endIndex + overscan),
    };
  }, [scrollTop, itemHeight, containerHeight, items.length, overscan]);

  // 可见项目
  const visibleItems = useMemo(() => {
    return items.slice(visibleRange.start, visibleRange.end + 1);
  }, [items, visibleRange]);

  // 滚动处理
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  }, []);

  // 总高度
  const totalHeight = items.length * itemHeight;

  // 偏移量
  const offsetY = visibleRange.start * itemHeight;

  return (
    <div
      style={{
        height: containerHeight,
        overflow: 'auto',
      }}
      onScroll={handleScroll}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        <div
          style={{
            transform: `translateY(${offsetY}px)`,
          }}
        >
          {visibleItems.map((item, index) => (
            <div
              key={visibleRange.start + index}
              style={{
                height: itemHeight,
                overflow: 'hidden',
              }}
            >
              {renderItem(item, visibleRange.start + index)}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// 使用虚拟滚动的大列表
const LargeUserList: FC = () => {
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    // 模拟大量数据
    const mockUsers: User[] = Array.from({ length: 10000 }, (_, index) => ({
      id: index + 1,
      username: `user${index + 1}`,
      email: `user${index + 1}@example.com`,
      // ... 其他属性
    }));
    setUsers(mockUsers);
  }, []);

  const renderUserItem = useCallback((user: User, index: number) => (
    <div style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
      <strong>{user.username}</strong>
      <div>{user.email}</div>
    </div>
  ), []);

  return (
    <div>
      <h2>用户列表 ({users.length} 个用户)</h2>
      <VirtualScroll
        items={users}
        itemHeight={60}
        containerHeight={400}
        renderItem={renderUserItem}
        overscan={10}
      />
    </div>
  );
};
```

### 懒加载组件

`````typescript
import { lazy, Suspense, ComponentType } from 'react';

// 懒加载组件
const LazyUserProfile = lazy(() => import('./UserProfile'));
const LazyProductDetail = lazy(() => import('./ProductDetail'));
const LazyOrderHistory = lazy(() => import('./OrderHistory'));

// 带有错误边界的懒加载包装器
interface LazyWrapperProps {
  children: ReactNode;
  fallback?: ReactNode;
  errorFallback?: ReactNode;
}

class ErrorBoundary extends Component<
  { children: ReactNode; fallback: ReactNode },
  { hasError: boolean }
> {
  constructor(props: { children: ReactNode; fallback: ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): { hasError: boolean } {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('懒加载组件错误:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback;
    }

    return this.props.children;
  }
}

const LazyWrapper: FC<LazyWrapperProps> = ({
  children,
  fallback = <div>加载中...</div>,
  errorFallback = <div>加载失败，请刷新重试</div>
}) => {
  return (
    <ErrorBoundary fallback={errorFallback}>
      <Suspense fallback={fallback}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
};

// 路由级别的懒加载
const AppRouter: FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route
          path="/profile"
          element={
            <LazyWrapper fallback={<div>加载用户资料...</div>}>
              <LazyUserProfile />
            </LazyWrapper>
          }
        />
        <Route
          path="/product/:id"
          element={
            <LazyWrapper fallback={<div>加载商品详情...</div>}>
              <LazyProductDetail />
            </LazyWrapper>
          }
        />
        <Route
          path="/orders"
          element={
            <LazyWrapper fallback={<div>加载订单历史...</div>}>
              <LazyOrderHistory />
            </LazyWrapper>
          }
        />
      </Routes>
    </Router>
  );
};

// 条件懒加载Hook
function useConditionalLazyLoad<T extends ComponentType<any>>(
  condition: boolean,
  importFn: () => Promise<{ default: T }>
): T | null {
  const [Component, setComponent] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (condition && !Component && !loading) {
      setLoading(true);
      importFn()
        .then(module => {
          setComponent(() => module.default);
        })
        .catch(error => {
          console.error('懒加载失败:', error);
        })
        .finally(() => {
          setLoading(false);
        });
    }
  }, [condition, Component, loading, importFn]);

  return Component;
}

// 使用条件懒加载
const ConditionalFeature: FC<{ showAdvanced: boolean }> = ({ showAdvanced }) => {
  const AdvancedComponent = useConditionalLazyLoad(
    showAdvanced,
    () => import('./AdvancedFeature')
  );

  return (
    <div>
      <h2>基础功能</h2>
      <p>这是基础功能内容</p>

      {showAdvanced && (
        <div>
          <h3>高级功能</h3>
          {AdvancedComponent ? (
            <AdvancedComponent />
          ) : (
            <div>加载高级功能中...</div>
          )}
        </div>
      )}
    </div>
  );
};

---

## 🛍️ Mall-Frontend实战案例

### 商品卡片组件设计

基于Mall-Frontend项目中的实际ProductCard组件，我们来分析一个完整的企业级React组件设计：

<augment_code_snippet path="mall-frontend/src/components/business/ProductCard.tsx" mode="EXCERPT">
````typescript
interface ProductCardProps {
  product: Product;
  loading?: boolean;
  showActions?: boolean;
  showBadge?: boolean | string;
  badgeColor?: string;
  size?: 'small' | 'default' | 'large';
  onAddToCart?: (product: Product) => void;
  onToggleFavorite?: (product: Product) => void;
  onShare?: (product: Product) => void;
  onViewDetail?: (productId: number) => void;
  className?: string;
  style?: React.CSSProperties;
}

const ProductCard: React.FC<ProductCardProps> = ({
  product,
  loading = false,
  showActions = true,
  showBadge = true,
  badgeColor,
  size = 'default',
  onAddToCart,
  onToggleFavorite,
  onShare,
  onViewDetail,
  className,
  style,
}) => {
  const [imageLoading, setImageLoading] = useState(true);
  const [imageError, setImageError] = useState(false);
  const [isFavorited, setIsFavorited] = useState(false);
  const [addingToCart, setAddingToCart] = useState(false);

  const router = useRouter();

  // 处理添加到购物车
  const handleAddToCart = useCallback(async (e: React.MouseEvent) => {
    e.stopPropagation(); // 阻止事件冒泡

    if (!onAddToCart) return;

    try {
      setAddingToCart(true);
      await onAddToCart(product);
      message.success('已添加到购物车');
    } catch (error) {
      message.error('添加失败，请重试');
    } finally {
      setAddingToCart(false);
    }
  }, [onAddToCart, product]);
`````

</augment_code_snippet>

### 组件设计的最佳实践分析

从ProductCard组件中，我们可以学到以下最佳实践：

#### 1. 完善的Props类型定义

```typescript
// 良好的Props设计原则
interface ComponentProps {
  // 必需的核心数据
  product: Product;

  // 可选的行为控制
  loading?: boolean;
  showActions?: boolean;
  showBadge?: boolean | string; // 支持布尔值和自定义文本

  // 样式和尺寸控制
  size?: 'small' | 'default' | 'large';
  className?: string;
  style?: React.CSSProperties;

  // 事件回调函数
  onAddToCart?: (product: Product) => void;
  onToggleFavorite?: (product: Product) => void;
  onShare?: (product: Product) => void;
  onViewDetail?: (productId: number) => void;
}
```

#### 2. 状态管理的最佳实践

```typescript
// 组件内部状态管理
const ProductCard: FC<ProductCardProps> = props => {
  // UI状态
  const [imageLoading, setImageLoading] = useState(true);
  const [imageError, setImageError] = useState(false);

  // 交互状态
  const [isFavorited, setIsFavorited] = useState(false);
  const [addingToCart, setAddingToCart] = useState(false);

  // 使用useCallback优化事件处理函数
  const handleAddToCart = useCallback(
    async (e: React.MouseEvent) => {
      e.stopPropagation(); // 防止事件冒泡

      if (!onAddToCart) return;

      try {
        setAddingToCart(true);
        await onAddToCart(product);
        // 成功反馈
        message.success('已添加到购物车');
      } catch (error) {
        // 错误处理
        message.error('添加失败，请重试');
      } finally {
        // 状态重置
        setAddingToCart(false);
      }
    },
    [onAddToCart, product]
  );

  // 其他事件处理函数...
};
```

#### 3. 响应式设计和尺寸适配

```typescript
// 响应式尺寸配置
const getSizeConfig = () => {
  switch (size) {
    case 'small':
      return {
        cardStyle: { width: 200 },
        imageHeight: 150,
        titleLevel: 5 as const,
        showDescription: false,
      };
    case 'large':
      return {
        cardStyle: { width: 320 },
        imageHeight: 240,
        titleLevel: 4 as const,
        showDescription: true,
      };
    default:
      return {
        cardStyle: { width: 260 },
        imageHeight: 200,
        titleLevel: 5 as const,
        showDescription: true,
      };
  }
};

// 使用配置
const sizeConfig = getSizeConfig();
```

#### 4. 条件渲染和动态内容

```typescript
// 智能标签显示
const getStatusBadge = () => {
  if (!showBadge) return null;

  // 自定义标签
  if (typeof showBadge === 'string') {
    return (
      <Badge.Ribbon text={showBadge} color={badgeColor || 'blue'}>
        <div />
      </Badge.Ribbon>
    );
  }

  // 动态标签逻辑
  const badges = [];

  // 热销标签
  if (product.sales_count && product.sales_count > 1000) {
    badges.push(
      <Badge.Ribbon key="hot" text="热销" color="red">
        <div />
      </Badge.Ribbon>
    );
  }

  // 新品标签
  const isNew = product.created_at &&
    new Date(product.created_at).getTime() > Date.now() - 7 * 24 * 60 * 60 * 1000;
  if (isNew) {
    badges.push(
      <Badge.Ribbon key="new" text="新品" color="green">
        <div />
      </Badge.Ribbon>
    );
  }

  // 优惠标签
  if (product.discount_price &&
      parseFloat(product.discount_price) < parseFloat(product.price)) {
    badges.push(
      <Badge.Ribbon key="discount" text="限时优惠" color="orange">
        <div />
      </Badge.Ribbon>
    );
  }

  return badges[0]; // 只显示优先级最高的标签
};
```

### 购物车Hook设计

```typescript
// 购物车状态管理Hook
interface CartItem {
  id: number;
  product: Product;
  quantity: number;
  selected: boolean;
}

interface UseCartReturn {
  items: CartItem[];
  totalItems: number;
  totalPrice: number;
  selectedItems: CartItem[];
  selectedTotalPrice: number;
  loading: boolean;
  addItem: (product: Product, quantity?: number) => Promise<void>;
  removeItem: (itemId: number) => Promise<void>;
  updateQuantity: (itemId: number, quantity: number) => Promise<void>;
  toggleSelect: (itemId: number) => void;
  selectAll: (selected: boolean) => void;
  clearCart: () => Promise<void>;
  refreshCart: () => Promise<void>;
}

function useCart(): UseCartReturn {
  const [items, setItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState(false);

  // 计算总数量
  const totalItems = useMemo(() => {
    return items.reduce((sum, item) => sum + item.quantity, 0);
  }, [items]);

  // 计算总价格
  const totalPrice = useMemo(() => {
    return items.reduce((sum, item) => {
      const price = parseFloat(item.product.discount_price || item.product.price);
      return sum + price * item.quantity;
    }, 0);
  }, [items]);

  // 选中的商品
  const selectedItems = useMemo(() => {
    return items.filter(item => item.selected);
  }, [items]);

  // 选中商品的总价
  const selectedTotalPrice = useMemo(() => {
    return selectedItems.reduce((sum, item) => {
      const price = parseFloat(item.product.discount_price || item.product.price);
      return sum + price * item.quantity;
    }, 0);
  }, [selectedItems]);

  // 添加商品到购物车
  const addItem = useCallback(async (product: Product, quantity = 1) => {
    try {
      setLoading(true);

      // 检查商品是否已存在
      const existingItemIndex = items.findIndex(item => item.product.id === product.id);

      if (existingItemIndex > -1) {
        // 更新数量
        const newItems = [...items];
        newItems[existingItemIndex].quantity += quantity;
        setItems(newItems);

        // 调用API更新
        await api.cart.updateCartItem({
          id: newItems[existingItemIndex].id,
          quantity: newItems[existingItemIndex].quantity,
        });
      } else {
        // 添加新商品
        const response = await api.cart.addToCart({
          product_id: product.id,
          quantity,
        });

        const newItem: CartItem = {
          id: response.data.id,
          product,
          quantity,
          selected: true,
        };

        setItems(prev => [...prev, newItem]);
      }
    } catch (error) {
      console.error('添加到购物车失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [items]);

  // 移除商品
  const removeItem = useCallback(async (itemId: number) => {
    try {
      setLoading(true);
      await api.cart.removeCartItem(itemId);
      setItems(prev => prev.filter(item => item.id !== itemId));
    } catch (error) {
      console.error('移除商品失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // 更新数量
  const updateQuantity = useCallback(async (itemId: number, quantity: number) => {
    if (quantity <= 0) {
      await removeItem(itemId);
      return;
    }

    try {
      setLoading(true);
      await api.cart.updateCartItem({ id: itemId, quantity });
      setItems(prev => prev.map(item =>
        item.id === itemId ? { ...item, quantity } : item
      ));
    } catch (error) {
      console.error('更新数量失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [removeItem]);

  // 切换选中状态
  const toggleSelect = useCallback((itemId: number) => {
    setItems(prev => prev.map(item =>
      item.id === itemId ? { ...item, selected: !item.selected } : item
    ));
  }, []);

  // 全选/取消全选
  const selectAll = useCallback((selected: boolean) => {
    setItems(prev => prev.map(item => ({ ...item, selected })));
  }, []);

  // 清空购物车
  const clearCart = useCallback(async () => {
    try {
      setLoading(true);
      await api.cart.clearCart();
      setItems([]);
    } catch (error) {
      console.error('清空购物车失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // 刷新购物车
  const refreshCart = useCallback(async () => {
    try {
      setLoading(true);
      const response = await api.cart.getCart();
      const cartItems: CartItem[] = response.data.items.map(item => ({
        id: item.id,
        product: item.product,
        quantity: item.quantity,
        selected: item.selected || false,
      }));
      setItems(cartItems);
    } catch (error) {
      console.error('刷新购物车失败:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 初始化时加载购物车
  useEffect(() => {
    refreshCart();
  }, [refreshCart]);

  return {
    items,
    totalItems,
    totalPrice,
    selectedItems,
    selectedTotalPrice,
    loading,
    addItem,
    removeItem,
    updateQuantity,
    toggleSelect,
    selectAll,
    clearCart,
    refreshCart,
  };
}

// 使用购物车Hook的组件
const ShoppingCart: FC = () => {
  const {
    items,
    totalItems,
    selectedTotalPrice,
    loading,
    updateQuantity,
    removeItem,
    toggleSelect,
    selectAll,
  } = useCart();

  const [allSelected, setAllSelected] = useState(false);

  // 检查是否全选
  useEffect(() => {
    setAllSelected(items.length > 0 && items.every(item => item.selected));
  }, [items]);

  const handleSelectAll = (checked: boolean) => {
    selectAll(checked);
  };

  const handleCheckout = () => {
    const selectedItems = items.filter(item => item.selected);
    if (selectedItems.length === 0) {
      message.warning('请选择要结算的商品');
      return;
    }

    // 跳转到结算页面
    router.push('/checkout');
  };

  return (
    <div className="shopping-cart">
      <div className="cart-header">
        <Checkbox
          checked={allSelected}
          onChange={(e) => handleSelectAll(e.target.checked)}
        >
          全选
        </Checkbox>
        <span>共 {totalItems} 件商品</span>
      </div>

      <div className="cart-items">
        {items.map(item => (
          <CartItemCard
            key={item.id}
            item={item}
            onQuantityChange={(quantity) => updateQuantity(item.id, quantity)}
            onRemove={() => removeItem(item.id)}
            onToggleSelect={() => toggleSelect(item.id)}
          />
        ))}
      </div>

      <div className="cart-footer">
        <div className="total-price">
          合计: ¥{selectedTotalPrice.toFixed(2)}
        </div>
        <Button
          type="primary"
          size="large"
          loading={loading}
          onClick={handleCheckout}
          disabled={items.filter(item => item.selected).length === 0}
        >
          去结算
        </Button>
      </div>
    </div>
  );
};
```

### 表单组件设计

````typescript
// 通用表单字段组件
interface FormFieldProps {
  label: string;
  name: string;
  required?: boolean;
  error?: string;
  children: ReactNode;
  tooltip?: string;
  extra?: ReactNode;
}

const FormField: FC<FormFieldProps> = ({
  label,
  name,
  required,
  error,
  children,
  tooltip,
  extra,
}) => {
  return (
    <div className={`form-field ${error ? 'has-error' : ''}`}>
      <label htmlFor={name} className="form-label">
        {required && <span className="required">*</span>}
        {label}
        {tooltip && (
          <Tooltip title={tooltip}>
            <QuestionCircleOutlined style={{ marginLeft: 4 }} />
          </Tooltip>
        )}
      </label>

      <div className="form-control">
        {children}
      </div>

      {error && (
        <div className="form-error">
          <ExclamationCircleOutlined />
          <span>{error}</span>
        </div>
      )}

      {extra && (
        <div className="form-extra">
          {extra}
        </div>
      )}
    </div>
  );
};

// 用户注册表单组件
interface RegisterFormData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  phone: string;
  agreeTerms: boolean;
}

const RegisterForm: FC = () => {
  const form = useForm<RegisterFormData>({
    initialValues: {
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      phone: '',
      agreeTerms: false,
    },
    validationRules: {
      username: [
        validationRules.required('用户名不能为空'),
        validationRules.minLength(3, '用户名至少3个字符'),
        validationRules.maxLength(20, '用户名最多20个字符'),
        validationRules.pattern(/^[a-zA-Z0-9_]+$/, '用户名只能包含字母、数字和下划线'),
      ],
      email: [
        validationRules.required('邮箱不能为空'),
        validationRules.email('请输入有效的邮箱地址'),
      ],
      password: [
        validationRules.required('密码不能为空'),
        validationRules.minLength(6, '密码至少6个字符'),
        (value: string) => {
          if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(value)) {
            return '密码必须包含大小写字母和数字';
          }
          return null;
        },
      ],
      confirmPassword: [
        validationRules.required('请确认密码'),
        (value: string) => {
          if (value !== form.values.password) {
            return '两次输入的密码不一致';
          }
          return null;
        },
      ],
      phone: [
        validationRules.required('手机号不能为空'),
        validationRules.pattern(/^1[3-9]\d{9}$/, '请输入有效的手机号'),
      ],
      agreeTerms: [
        (value: boolean) => {
          if (!value) {
            return '请同意用户协议和隐私政策';
          }
          return null;
        },
      ],
    },
    onSubmit: async (values) => {
      try {
        await api.auth.register({
          username: values.username,
          email: values.email,
          password: values.password,
          phone: values.phone,
        });

        message.success('注册成功！');
        router.push('/login');
      } catch (error: any) {
        message.error(error.message || '注册失败，请重试');
      }
    },
  });

  return (
    <form onSubmit={form.handleSubmit} className="register-form">
      <FormField
        label="用户名"
        name="username"
        required
        error={form.touched.username ? form.errors.username : undefined}
        tooltip="用户名将作为您的唯一标识"
      >
        <Input
          placeholder="请输入用户名"
          value={form.values.username}
          onChange={form.handleChange('username')}
          onBlur={form.handleBlur('username')}
          status={form.touched.username && form.errors.username ? 'error' : ''}
        />
      </FormField>

      <FormField
        label="邮箱"
        name="email"
        required
        error={form.touched.email ? form.errors.email : undefined}
      >
        <Input
          type="email"
          placeholder="请输入邮箱地址"
          value={form.values.email}
          onChange={form.handleChange('email')}
          onBlur={form.handleBlur('email')}
          status={form.touched.email && form.errors.email ? 'error' : ''}
        />
      </FormField>

      <FormField
        label="密码"
        name="password"
        required
        error={form.touched.password ? form.errors.password : undefined}
        extra="密码强度: 包含大小写字母和数字，至少6位"
      >
        <Input.Password
          placeholder="请输入密码"
          value={form.values.password}
          onChange={form.handleChange('password')}
          onBlur={form.handleBlur('password')}
          status={form.touched.password && form.errors.password ? 'error' : ''}
        />
      </FormField>

      <FormField
        label="确认密码"
        name="confirmPassword"
        required
        error={form.touched.confirmPassword ? form.errors.confirmPassword : undefined}
      >
        <Input.Password
          placeholder="请再次输入密码"
          value={form.values.confirmPassword}
          onChange={form.handleChange('confirmPassword')}
          onBlur={form.handleBlur('confirmPassword')}
          status={form.touched.confirmPassword && form.errors.confirmPassword ? 'error' : ''}
        />
      </FormField>

      <FormField
        label="手机号"
        name="phone"
        required
        error={form.touched.phone ? form.errors.phone : undefined}
      >
        <Input
          placeholder="请输入手机号"
          value={form.values.phone}
          onChange={form.handleChange('phone')}
          onBlur={form.handleBlur('phone')}
          status={form.touched.phone && form.errors.phone ? 'error' : ''}
        />
      </FormField>

      <FormField
        label=""
        name="agreeTerms"
        error={form.touched.agreeTerms ? form.errors.agreeTerms : undefined}
      >
        <Checkbox
          checked={form.values.agreeTerms}
          onChange={form.handleChange('agreeTerms')}
        >
          我已阅读并同意
          <a href="/terms" target="_blank">《用户协议》</a>
          和
          <a href="/privacy" target="_blank">《隐私政策》</a>
        </Checkbox>
      </FormField>

      <Button
        type="primary"
        htmlType="submit"
        size="large"
        block
        loading={form.isSubmitting}
        disabled={!form.isValid}
      >
        注册
      </Button>
    </form>
  );
};

---

## 🎯 面试常考知识点

### 1. React组件设计原则

**Q: 如何设计一个高质量的React组件？**

**A: 高质量React组件的设计原则：**

1. **单一职责原则**: 每个组件只负责一个功能
2. **Props接口设计**: 清晰、完整、类型安全的Props定义
3. **状态管理**: 合理区分本地状态和全局状态
4. **性能优化**: 使用memo、useMemo、useCallback避免不必要的重渲染
5. **错误处理**: 完善的错误边界和异常处理
6. **可访问性**: 支持键盘导航、屏幕阅读器等

```typescript
// 良好的组件设计示例
interface ButtonProps {
  // 核心功能
  children: ReactNode;
  onClick?: () => void;

  // 样式控制
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'small' | 'medium' | 'large';

  // 状态控制
  loading?: boolean;
  disabled?: boolean;

  // 扩展性
  className?: string;
  style?: CSSProperties;

  // 可访问性
  'aria-label'?: string;
  'aria-describedby'?: string;
}

const Button = memo<ButtonProps>(({
  children,
  onClick,
  variant = 'primary',
  size = 'medium',
  loading = false,
  disabled = false,
  className,
  style,
  ...ariaProps
}) => {
  const handleClick = useCallback((e: MouseEvent) => {
    if (loading || disabled) return;
    onClick?.(e);
  }, [onClick, loading, disabled]);

  return (
    <button
      className={`btn btn-${variant} btn-${size} ${className || ''}`}
      onClick={handleClick}
      disabled={disabled || loading}
      style={style}
      {...ariaProps}
    >
      {loading && <Spinner />}
      {children}
    </button>
  );
});
````

### 2. Hooks使用最佳实践

**Q: 如何正确使用React Hooks？**

**A: React Hooks最佳实践：**

1. **依赖数组**: 正确设置useEffect、useMemo、useCallback的依赖
2. **自定义Hooks**: 提取可复用的状态逻辑
3. **性能优化**: 避免在每次渲染时创建新的对象和函数
4. **错误处理**: 在异步操作中正确处理错误和清理

```typescript
// 正确的Hooks使用
const UserProfile: FC<{ userId: number }> = ({ userId }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 正确的依赖数组
  const fetchUser = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.getUser(userId);
      setUser(response.data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [userId]); // 依赖userId

  // 清理副作用
  useEffect(() => {
    let isCancelled = false;

    const loadUser = async () => {
      try {
        setLoading(true);
        const response = await api.getUser(userId);
        if (!isCancelled) {
          setUser(response.data);
        }
      } catch (err) {
        if (!isCancelled) {
          setError(err.message);
        }
      } finally {
        if (!isCancelled) {
          setLoading(false);
        }
      }
    };

    loadUser();

    return () => {
      isCancelled = true;
    };
  }, [userId]);

  // 缓存计算结果
  const displayName = useMemo(() => {
    if (!user) return '';
    return `${user.firstName} ${user.lastName}`;
  }, [user]);

  if (loading) return <Loading />;
  if (error) return <Error message={error} />;
  if (!user) return <NotFound />;

  return <div>{displayName}</div>;
};
```

### 3. 组件通信模式

**Q: React组件间有哪些通信方式？各自的适用场景？**

**A: React组件通信的主要方式：**

1. **Props传递**: 父子组件通信，数据向下流动
2. **回调函数**: 子组件向父组件传递数据
3. **Context API**: 跨层级组件通信
4. **状态管理库**: 全局状态管理（Redux、Zustand等）
5. **事件总线**: 松耦合的组件通信
6. **Ref**: 直接访问子组件实例

```typescript
// 1. Props + 回调函数
const Parent: FC = () => {
  const [count, setCount] = useState(0);

  return (
    <Child
      count={count}
      onCountChange={setCount}
    />
  );
};

// 2. Context API
const ThemeContext = createContext<{
  theme: string;
  setTheme: (theme: string) => void;
}>({
  theme: 'light',
  setTheme: () => {},
});

// 3. 自定义Hook + Context
const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
};

// 4. 事件总线
const useEventBus = () => {
  const emit = useCallback((event: string, data: any) => {
    window.dispatchEvent(new CustomEvent(event, { detail: data }));
  }, []);

  const on = useCallback((event: string, handler: (data: any) => void) => {
    const listener = (e: CustomEvent) => handler(e.detail);
    window.addEventListener(event, listener);
    return () => window.removeEventListener(event, listener);
  }, []);

  return { emit, on };
};
```

### 4. 性能优化策略

**Q: React应用有哪些性能优化策略？**

**A: React性能优化的主要策略：**

1. **组件优化**: React.memo、PureComponent
2. **Hook优化**: useMemo、useCallback
3. **代码分割**: React.lazy、动态import
4. **虚拟化**: 大列表的虚拟滚动
5. **预加载**: 资源预加载和预取
6. **缓存策略**: 合理的缓存机制

```typescript
// 组件级优化
const ExpensiveComponent = memo<Props>(({ data, onUpdate }) => {
  // 缓存复杂计算
  const processedData = useMemo(() => {
    return data.map(item => ({
      ...item,
      computed: expensiveCalculation(item),
    }));
  }, [data]);

  // 缓存事件处理函数
  const handleUpdate = useCallback((id: number, updates: any) => {
    onUpdate(id, updates);
  }, [onUpdate]);

  return (
    <div>
      {processedData.map(item => (
        <ItemComponent
          key={item.id}
          item={item}
          onUpdate={handleUpdate}
        />
      ))}
    </div>
  );
}, (prevProps, nextProps) => {
  // 自定义比较函数
  return (
    prevProps.data === nextProps.data &&
    prevProps.onUpdate === nextProps.onUpdate
  );
});

// 代码分割
const LazyComponent = lazy(() =>
  import('./HeavyComponent').then(module => ({
    default: module.HeavyComponent
  }))
);

// 虚拟化大列表
const VirtualizedList: FC<{ items: any[] }> = ({ items }) => {
  return (
    <FixedSizeList
      height={600}
      itemCount={items.length}
      itemSize={50}
      itemData={items}
    >
      {({ index, style, data }) => (
        <div style={style}>
          <ListItem item={data[index]} />
        </div>
      )}
    </FixedSizeList>
  );
};
```

---

## 🏋️ 实战练习

### 练习1: 设计一个通用的数据表格组件

**题目**: 设计一个功能完整的数据表格组件，支持排序、筛选、分页等功能

**要求**:

1. 支持自定义列配置
2. 内置排序和筛选功能
3. 支持分页和虚拟滚动
4. 提供行选择功能
5. 支持自定义渲染器

**解决方案**:

```typescript
// 列配置类型
interface TableColumn<T> {
  key: keyof T;
  title: string;
  width?: number;
  sortable?: boolean;
  filterable?: boolean;
  render?: (value: any, record: T, index: number) => ReactNode;
  align?: 'left' | 'center' | 'right';
  fixed?: 'left' | 'right';
}

// 表格Props类型
interface DataTableProps<T> {
  columns: TableColumn<T>[];
  data: T[];
  loading?: boolean;
  pagination?: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number, pageSize: number) => void;
  };
  rowSelection?: {
    selectedRowKeys: React.Key[];
    onChange: (selectedRowKeys: React.Key[], selectedRows: T[]) => void;
    getCheckboxProps?: (record: T) => { disabled?: boolean };
  };
  onSort?: (key: keyof T, direction: 'asc' | 'desc' | null) => void;
  onFilter?: (key: keyof T, value: any) => void;
  rowKey: keyof T | ((record: T) => React.Key);
  size?: 'small' | 'middle' | 'large';
  bordered?: boolean;
  virtual?: boolean;
  height?: number;
}

function DataTable<T extends Record<string, any>>({
  columns,
  data,
  loading = false,
  pagination,
  rowSelection,
  onSort,
  onFilter,
  rowKey,
  size = 'middle',
  bordered = false,
  virtual = false,
  height = 400,
}: DataTableProps<T>): JSX.Element {
  const [sortConfig, setSortConfig] = useState<{
    key: keyof T | null;
    direction: 'asc' | 'desc' | null;
  }>({ key: null, direction: null });

  const [filters, setFilters] = useState<Record<keyof T, any>>({} as Record<keyof T, any>);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

  // 获取行的key
  const getRowKey = useCallback((record: T, index: number): React.Key => {
    if (typeof rowKey === 'function') {
      return rowKey(record);
    }
    return record[rowKey] as React.Key;
  }, [rowKey]);

  // 处理排序
  const handleSort = useCallback((key: keyof T) => {
    let direction: 'asc' | 'desc' | null = 'asc';

    if (sortConfig.key === key) {
      if (sortConfig.direction === 'asc') {
        direction = 'desc';
      } else if (sortConfig.direction === 'desc') {
        direction = null;
      }
    }

    setSortConfig({ key: direction ? key : null, direction });
    onSort?.(key, direction);
  }, [sortConfig, onSort]);

  // 处理筛选
  const handleFilter = useCallback((key: keyof T, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    onFilter?.(key, value);
  }, [filters, onFilter]);

  // 处理行选择
  const handleRowSelect = useCallback((record: T, selected: boolean) => {
    const key = getRowKey(record, 0);
    let newSelectedKeys: React.Key[];

    if (selected) {
      newSelectedKeys = [...selectedKeys, key];
    } else {
      newSelectedKeys = selectedKeys.filter(k => k !== key);
    }

    setSelectedKeys(newSelectedKeys);

    if (rowSelection) {
      const selectedRows = data.filter(item =>
        newSelectedKeys.includes(getRowKey(item, 0))
      );
      rowSelection.onChange(newSelectedKeys, selectedRows);
    }
  }, [selectedKeys, data, getRowKey, rowSelection]);

  // 全选处理
  const handleSelectAll = useCallback((selected: boolean) => {
    const newSelectedKeys = selected
      ? data.map((record, index) => getRowKey(record, index))
      : [];

    setSelectedKeys(newSelectedKeys);

    if (rowSelection) {
      const selectedRows = selected ? data : [];
      rowSelection.onChange(newSelectedKeys, selectedRows);
    }
  }, [data, getRowKey, rowSelection]);

  // 渲染表头
  const renderHeader = () => (
    <thead>
      <tr>
        {rowSelection && (
          <th style={{ width: 50 }}>
            <Checkbox
              checked={selectedKeys.length === data.length && data.length > 0}
              indeterminate={selectedKeys.length > 0 && selectedKeys.length < data.length}
              onChange={(e) => handleSelectAll(e.target.checked)}
            />
          </th>
        )}
        {columns.map(column => (
          <th
            key={String(column.key)}
            style={{
              width: column.width,
              textAlign: column.align || 'left',
              cursor: column.sortable ? 'pointer' : 'default',
            }}
            onClick={() => column.sortable && handleSort(column.key)}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <span>{column.title}</span>
              {column.sortable && (
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                  <CaretUpOutlined
                    style={{
                      fontSize: 10,
                      color: sortConfig.key === column.key && sortConfig.direction === 'asc'
                        ? '#1890ff' : '#bfbfbf'
                    }}
                  />
                  <CaretDownOutlined
                    style={{
                      fontSize: 10,
                      color: sortConfig.key === column.key && sortConfig.direction === 'desc'
                        ? '#1890ff' : '#bfbfbf'
                    }}
                  />
                </div>
              )}
              {column.filterable && (
                <FilterDropdown
                  column={column}
                  value={filters[column.key]}
                  onChange={(value) => handleFilter(column.key, value)}
                />
              )}
            </div>
          </th>
        ))}
      </tr>
    </thead>
  );

  // 渲染表格行
  const renderRow = (record: T, index: number) => {
    const key = getRowKey(record, index);
    const isSelected = selectedKeys.includes(key);

    return (
      <tr
        key={key}
        style={{
          backgroundColor: isSelected ? '#e6f7ff' : undefined,
        }}
      >
        {rowSelection && (
          <td>
            <Checkbox
              checked={isSelected}
              onChange={(e) => handleRowSelect(record, e.target.checked)}
              {...(rowSelection.getCheckboxProps?.(record) || {})}
            />
          </td>
        )}
        {columns.map(column => (
          <td
            key={String(column.key)}
            style={{
              textAlign: column.align || 'left',
              padding: size === 'small' ? '8px' : size === 'large' ? '16px' : '12px',
            }}
          >
            {column.render
              ? column.render(record[column.key], record, index)
              : String(record[column.key] || '')
            }
          </td>
        ))}
      </tr>
    );
  };

  // 虚拟滚动渲染
  const renderVirtualTable = () => (
    <FixedSizeList
      height={height}
      itemCount={data.length}
      itemSize={size === 'small' ? 40 : size === 'large' ? 60 : 50}
      itemData={data}
    >
      {({ index, style }) => (
        <div style={style}>
          {renderRow(data[index], index)}
        </div>
      )}
    </FixedSizeList>
  );

  // 普通表格渲染
  const renderNormalTable = () => (
    <tbody>
      {data.map((record, index) => renderRow(record, index))}
    </tbody>
  );

  return (
    <div className={`data-table ${bordered ? 'bordered' : ''}`}>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        {renderHeader()}
        {virtual ? renderVirtualTable() : renderNormalTable()}
      </table>

      {pagination && (
        <div style={{ marginTop: 16, textAlign: 'right' }}>
          <Pagination
            current={pagination.current}
            pageSize={pagination.pageSize}
            total={pagination.total}
            onChange={pagination.onChange}
            showSizeChanger
            showQuickJumper
            showTotal={(total, range) =>
              `第 ${range[0]}-${range[1]} 条/共 ${total} 条`
            }
          />
        </div>
      )}

      {loading && (
        <div style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(255, 255, 255, 0.8)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}>
          <Spin size="large" />
        </div>
      )}
    </div>
  );
}

// 筛选下拉组件
interface FilterDropdownProps<T> {
  column: TableColumn<T>;
  value: any;
  onChange: (value: any) => void;
}

function FilterDropdown<T>({ column, value, onChange }: FilterDropdownProps<T>) {
  const [visible, setVisible] = useState(false);
  const [inputValue, setInputValue] = useState(value || '');

  const handleConfirm = () => {
    onChange(inputValue);
    setVisible(false);
  };

  const handleReset = () => {
    setInputValue('');
    onChange(null);
    setVisible(false);
  };

  return (
    <Dropdown
      open={visible}
      onOpenChange={setVisible}
      dropdownRender={() => (
        <div style={{ padding: 8, backgroundColor: 'white', boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }}>
          <Input
            placeholder={`搜索 ${column.title}`}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onPressEnter={handleConfirm}
            style={{ marginBottom: 8 }}
          />
          <div style={{ display: 'flex', gap: 8 }}>
            <Button size="small" onClick={handleReset}>
              重置
            </Button>
            <Button size="small" type="primary" onClick={handleConfirm}>
              确定
            </Button>
          </div>
        </div>
      )}
    >
      <FilterOutlined style={{ color: value ? '#1890ff' : '#bfbfbf' }} />
    </Dropdown>
  );
}

// 使用示例
const UserTable: FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const columns: TableColumn<User>[] = [
    {
      key: 'id',
      title: 'ID',
      width: 80,
      sortable: true,
    },
    {
      key: 'username',
      title: '用户名',
      width: 150,
      sortable: true,
      filterable: true,
    },
    {
      key: 'email',
      title: '邮箱',
      width: 200,
      filterable: true,
    },
    {
      key: 'status',
      title: '状态',
      width: 100,
      render: (value: string) => (
        <Tag color={value === 'active' ? 'green' : 'red'}>
          {value === 'active' ? '活跃' : '禁用'}
        </Tag>
      ),
    },
    {
      key: 'created_at',
      title: '创建时间',
      width: 180,
      sortable: true,
      render: (value: string) => new Date(value).toLocaleString(),
    },
    {
      key: 'actions',
      title: '操作',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button size="small" onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button size="small" danger onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

  const handleEdit = (user: User) => {
    // 编辑用户逻辑
  };

  const handleDelete = (userId: number) => {
    // 删除用户逻辑
  };

  const handlePaginationChange = (page: number, pageSize: number) => {
    setPagination(prev => ({ ...prev, current: page, pageSize }));
    // 重新加载数据
  };

  const handleSort = (key: keyof User, direction: 'asc' | 'desc' | null) => {
    // 处理排序
    console.log('排序:', key, direction);
  };

  const handleFilter = (key: keyof User, value: any) => {
    // 处理筛选
    console.log('筛选:', key, value);
  };

  return (
    <DataTable
      columns={columns}
      data={users}
      loading={loading}
      pagination={{
        ...pagination,
        onChange: handlePaginationChange,
      }}
      rowSelection={{
        selectedRowKeys,
        onChange: setSelectedRowKeys,
      }}
      onSort={handleSort}
      onFilter={handleFilter}
      rowKey="id"
      bordered
    />
  );
};
```

这个数据表格组件展示了：

1. **完整的TypeScript类型定义**
2. **灵活的列配置系统**
3. **内置的排序和筛选功能**
4. **行选择和分页支持**
5. **虚拟滚动优化**
6. **自定义渲染器**
7. **响应式设计**

---

## 📚 本章总结

通过本章学习，我们深入掌握了React组件设计与Hooks应用：

### 🎯 核心收获

1. **组件设计** ⚛️
   - 掌握了TypeScript在React中的类型定义
   - 学会了设计灵活且类型安全的组件Props
   - 理解了组件的生命周期和状态管理

2. **Hooks应用** 🎣
   - 深入理解了useState、useEffect、useCallback等核心Hooks
   - 学会了设计自定义Hooks提取可复用逻辑
   - 掌握了Hooks的性能优化技巧

3. **组件通信** 🔗
   - 掌握了多种组件通信模式
   - 学会了使用Context API进行跨层级通信
   - 理解了事件总线等高级通信方式

4. **性能优化** ⚡
   - 学会了使用React.memo、useMemo优化渲染性能
   - 掌握了虚拟滚动、懒加载等优化技术
   - 理解了代码分割和资源优化策略

5. **实战应用** 💼
   - 分析了Mall-Frontend项目中的组件设计
   - 学会了构建企业级的业务组件
   - 掌握了复杂表单和数据表格的实现

### 🚀 技术进阶

- **下一步学习**: Next.js框架应用与SSR/SSG
- **实践建议**: 在项目中应用学到的组件设计模式
- **深入方向**: React性能调优和架构设计

### 💡 最佳实践

1. **组件设计**: 遵循单一职责原则，保持组件的简洁和可复用
2. **类型安全**: 充分利用TypeScript的类型系统
3. **性能优化**: 合理使用memo和Hooks优化
4. **用户体验**: 注重加载状态、错误处理和可访问性

React + TypeScript的组合为我们提供了强大的类型安全保障和优秀的开发体验！ 🎉

---

_下一章我们将学习《Next.js框架应用与SSR/SSG》，探索现代React应用的服务端渲染技术！_ 🚀

```

```
