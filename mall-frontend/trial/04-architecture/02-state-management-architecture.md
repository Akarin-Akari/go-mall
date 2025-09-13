# 第2章：状态管理架构设计 🔄

> *"状态管理不是技术问题，而是架构问题！"* 🚀

## 📚 本章导览

状态管理是现代前端应用的核心挑战之一。随着应用复杂度的增加，如何设计一个既能满足业务需求，又能保持代码可维护性的状态管理架构，成为了前端架构师必须掌握的核心技能。本章将从状态管理的基本概念出发，深入对比各种状态管理方案，结合Mall-Frontend项目的实际案例，探讨企业级状态管理架构的设计原则和最佳实践。

### 🎯 学习目标

通过本章学习，你将掌握：

- **状态管理理论** - 理解状态的分类、生命周期和管理原则
- **方案深度对比** - 掌握Redux、Zustand、Context API等方案的优劣
- **架构设计模式** - 学会设计可扩展的状态管理架构
- **数据流设计** - 理解单向数据流和双向绑定的应用场景
- **性能优化策略** - 掌握状态管理的性能优化技巧
- **测试策略** - 学会状态管理的测试方法和最佳实践
- **企业级实践** - 大型项目的状态管理治理经验
- **迁移策略** - 状态管理方案的升级和迁移方法

### 🛠️ 技术栈概览

```typescript
{
  "stateManagement": {
    "global": ["Redux Toolkit", "Zustand", "Valtio", "Jotai"],
    "local": ["useState", "useReducer", "useImmer"],
    "server": ["React Query", "SWR", "Apollo Client"],
    "form": ["React Hook Form", "Formik", "Final Form"]
  },
  "patterns": {
    "flux": ["Redux", "MobX", "Vuex"],
    "atomic": ["Jotai", "Recoil", "Valtio"],
    "proxy": ["MobX", "Valtio", "Vue 3 Reactivity"]
  },
  "tools": {
    "devtools": ["Redux DevTools", "Zustand DevTools"],
    "testing": ["Redux Mock Store", "MSW", "Testing Library"],
    "middleware": ["Redux Thunk", "Redux Saga", "Redux Observable"]
  }
}
```

### 📖 本章目录

- [状态管理基础理论](#状态管理基础理论)
- [状态管理方案深度对比](#状态管理方案深度对比)
- [Redux生态系统架构](#redux生态系统架构)
- [Zustand轻量级方案](#zustand轻量级方案)
- [Context API企业应用](#context-api企业应用)
- [原子化状态管理](#原子化状态管理)
- [服务端状态管理](#服务端状态管理)
- [状态管理架构设计](#状态管理架构设计)
- [性能优化策略](#性能优化策略)
- [测试策略与实践](#测试策略与实践)
- [企业级治理实践](#企业级治理实践)
- [Mall-Frontend状态架构](#mall-frontend状态架构)
- [面试常考知识点](#面试常考知识点)
- [实战练习](#实战练习)

---

## 🎯 状态管理基础理论

### 状态的分类和特征

在现代前端应用中，状态可以按照不同的维度进行分类：

```typescript
// 状态分类体系
interface StateClassification {
  // 按作用域分类
  scope: {
    local: '组件内部状态';
    shared: '组件间共享状态';
    global: '全局应用状态';
  };
  
  // 按数据源分类
  source: {
    client: '客户端状态';
    server: '服务端状态';
    url: 'URL状态';
    form: '表单状态';
  };
  
  // 按生命周期分类
  lifecycle: {
    ephemeral: '临时状态';
    session: '会话状态';
    persistent: '持久化状态';
  };
  
  // 按变更频率分类
  frequency: {
    static: '静态状态';
    dynamic: '动态状态';
    realtime: '实时状态';
  };
}

// 状态管理原则
const stateManagementPrinciples = {
  // 1. 单一数据源 (Single Source of Truth)
  singleSource: {
    principle: '每个状态都应该有唯一的数据源',
    benefits: ['避免数据不一致', '简化调试', '提高可预测性'],
    implementation: `
      // ❌ 多个数据源
      const [userA, setUserA] = useState(user);
      const [userB, setUserB] = useState(user);
      
      // ✅ 单一数据源
      const user = useUserStore(state => state.user);
    `
  },
  
  // 2. 状态不可变性 (Immutability)
  immutability: {
    principle: '状态应该是不可变的，通过创建新状态来更新',
    benefits: ['时间旅行调试', '性能优化', '避免副作用'],
    implementation: `
      // ❌ 直接修改状态
      state.user.name = 'New Name';
      
      // ✅ 创建新状态
      setState(prevState => ({
        ...prevState,
        user: {
          ...prevState.user,
          name: 'New Name'
        }
      }));
    `
  },
  
  // 3. 可预测性 (Predictability)
  predictability: {
    principle: '相同的输入应该产生相同的输出',
    benefits: ['易于测试', '便于调试', '提高可靠性'],
    implementation: `
      // ✅ 纯函数reducer
      function userReducer(state, action) {
        switch (action.type) {
          case 'UPDATE_NAME':
            return { ...state, name: action.payload };
          default:
            return state;
        }
      }
    `
  },
  
  // 4. 最小化状态 (Minimal State)
  minimalState: {
    principle: '只存储必要的状态，派生状态应该通过计算得出',
    benefits: ['减少复杂性', '避免同步问题', '提高性能'],
    implementation: `
      // ❌ 存储派生状态
      interface State {
        items: Item[];
        totalPrice: number; // 派生状态
        itemCount: number;  // 派生状态
      }
      
      // ✅ 只存储必要状态
      interface State {
        items: Item[];
      }
      
      // 派生状态通过选择器计算
      const totalPrice = useSelector(state => 
        state.items.reduce((sum, item) => sum + item.price, 0)
      );
    `
  }
};
```

### 状态管理模式演进

```typescript
// 状态管理模式的历史演进
const stateManagementEvolution = {
  // 1. 原始时代 - 直接DOM操作
  primitive: {
    period: '2010年以前',
    approach: '直接操作DOM，使用全局变量',
    problems: ['状态散乱', '难以维护', '容易出错'],
    example: `
      // jQuery时代的状态管理
      let currentUser = null;
      
      function updateUser(user) {
        currentUser = user;
        $('#username').text(user.name);
        $('#avatar').attr('src', user.avatar);
      }
    `
  },
  
  // 2. MVC时代 - 模型视图分离
  mvc: {
    period: '2010-2013',
    approach: 'Backbone.js, Angular 1.x的双向绑定',
    problems: ['双向绑定复杂', '难以调试', '性能问题'],
    example: `
      // Backbone.js模式
      const UserModel = Backbone.Model.extend({
        defaults: { name: '', email: '' }
      });
      
      const UserView = Backbone.View.extend({
        render: function() {
          this.$el.html(template(this.model.toJSON()));
        }
      });
    `
  },
  
  // 3. Flux时代 - 单向数据流
  flux: {
    period: '2014-2015',
    approach: 'Facebook Flux架构，单向数据流',
    benefits: ['可预测性', '易于调试', '清晰的数据流'],
    example: `
      // Flux架构
      Action -> Dispatcher -> Store -> View -> Action
      
      // Action
      const UserActions = {
        updateUser: (user) => ({
          type: 'UPDATE_USER',
          payload: user
        })
      };
    `
  },
  
  // 4. Redux时代 - 函数式状态管理
  redux: {
    period: '2015-2019',
    approach: 'Redux的函数式状态管理',
    benefits: ['时间旅行', '中间件支持', '强大的生态'],
    problems: ['样板代码多', '学习曲线陡峭'],
    example: `
      // Redux模式
      const userReducer = (state = initialState, action) => {
        switch (action.type) {
          case 'UPDATE_USER':
            return { ...state, user: action.payload };
          default:
            return state;
        }
      };
    `
  },
  
  // 5. 现代时代 - 多样化方案
  modern: {
    period: '2019至今',
    approach: 'Hooks、Zustand、Jotai等轻量化方案',
    benefits: ['简化API', '更好的TypeScript支持', '更小的包体积'],
    trends: ['原子化状态', '服务端状态分离', '类型安全']
  }
};
```

### 状态管理决策树

```typescript
// 状态管理方案选择决策树
const stateManagementDecisionTree = {
  questions: [
    {
      id: 1,
      question: '状态是否需要在多个组件间共享？',
      no: {
        recommendation: 'useState / useReducer',
        reason: '本地状态足够，无需引入复杂的状态管理'
      },
      yes: { nextQuestion: 2 }
    },
    {
      id: 2,
      question: '应用规模是否较大（>100个组件）？',
      no: {
        recommendation: 'Context API + useReducer',
        reason: '中小型应用，Context API足够应对'
      },
      yes: { nextQuestion: 3 }
    },
    {
      id: 3,
      question: '团队是否熟悉函数式编程？',
      yes: {
        recommendation: 'Redux Toolkit',
        reason: '大型应用，需要强大的状态管理和调试工具'
      },
      no: { nextQuestion: 4 }
    },
    {
      id: 4,
      question: '是否需要复杂的异步逻辑？',
      yes: {
        recommendation: 'Redux Toolkit + RTK Query',
        reason: '复杂异步场景，Redux生态更成熟'
      },
      no: {
        recommendation: 'Zustand',
        reason: '简单易用，学习成本低'
      }
    }
  ],
  
  // 特殊场景推荐
  specialCases: {
    formManagement: {
      recommendation: 'React Hook Form + Zod',
      reason: '专门的表单状态管理，性能更好'
    },
    serverState: {
      recommendation: 'React Query / SWR',
      reason: '服务端状态有特殊需求，需要专门的解决方案'
    },
    realTimeData: {
      recommendation: 'Zustand + WebSocket',
      reason: '实时数据需要响应式更新'
    },
    atomicState: {
      recommendation: 'Jotai / Recoil',
      reason: '细粒度状态管理，避免不必要的重渲染'
    }
  }
};
```

### 状态管理性能考量

```typescript
// 性能优化策略
const performanceConsiderations = {
  // 1. 重渲染优化
  reRenderOptimization: {
    problems: ['不必要的重渲染', '性能瓶颈', '用户体验差'],
    solutions: [
      'React.memo',
      'useMemo',
      'useCallback',
      '状态分割',
      '选择器优化'
    ],
    example: `
      // ❌ 会导致所有组件重渲染
      const globalState = {
        user: { ... },
        products: [ ... ],
        cart: [ ... ],
        ui: { ... }
      };
      
      // ✅ 状态分割，减少重渲染
      const userStore = create((set) => ({ ... }));
      const productStore = create((set) => ({ ... }));
      const cartStore = create((set) => ({ ... }));
    `
  },
  
  // 2. 内存优化
  memoryOptimization: {
    strategies: [
      '状态清理',
      '弱引用',
      '分页加载',
      '虚拟滚动'
    ],
    example: `
      // 组件卸载时清理状态
      useEffect(() => {
        return () => {
          // 清理大型数据
          clearLargeDataSet();
        };
      }, []);
    `
  },
  
  // 3. 网络优化
  networkOptimization: {
    techniques: [
      '请求去重',
      '缓存策略',
      '乐观更新',
      '批量请求'
    ]
  }
};
```

---

## 🔄 状态管理方案深度对比

### 主流方案技术对比

```typescript
// 状态管理方案对比矩阵
interface StateManagementSolution {
  name: string;
  bundleSize: string;
  learningCurve: 'Easy' | 'Medium' | 'Hard';
  typescript: 'Excellent' | 'Good' | 'Basic';
  devtools: boolean;
  middleware: boolean;
  persistence: boolean;
  ssr: boolean;
  performance: 'Excellent' | 'Good' | 'Average';
  ecosystem: 'Rich' | 'Growing' | 'Limited';
  useCase: string[];
}

const stateManagementComparison: StateManagementSolution[] = [
  {
    name: 'Redux Toolkit',
    bundleSize: '~47kb',
    learningCurve: 'Hard',
    typescript: 'Excellent',
    devtools: true,
    middleware: true,
    persistence: true,
    ssr: true,
    performance: 'Good',
    ecosystem: 'Rich',
    useCase: ['大型应用', '复杂状态逻辑', '时间旅行调试', '团队协作']
  },
  {
    name: 'Zustand',
    bundleSize: '~2.5kb',
    learningCurve: 'Easy',
    typescript: 'Excellent',
    devtools: true,
    middleware: true,
    persistence: true,
    ssr: true,
    performance: 'Excellent',
    ecosystem: 'Growing',
    useCase: ['中小型应用', '快速原型', '简单状态管理', '性能敏感']
  },
  {
    name: 'Context API',
    bundleSize: '0kb (内置)',
    learningCurve: 'Medium',
    typescript: 'Good',
    devtools: false,
    middleware: false,
    persistence: false,
    ssr: true,
    performance: 'Average',
    ecosystem: 'Limited',
    useCase: ['主题管理', '用户认证', '简单共享状态', '避免prop drilling']
  },
  {
    name: 'Jotai',
    bundleSize: '~13kb',
    learningCurve: 'Medium',
    typescript: 'Excellent',
    devtools: true,
    middleware: false,
    persistence: true,
    ssr: true,
    performance: 'Excellent',
    ecosystem: 'Growing',
    useCase: ['原子化状态', '细粒度更新', '复杂依赖关系', '性能优化']
  },
  {
    name: 'Valtio',
    bundleSize: '~9kb',
    learningCurve: 'Easy',
    typescript: 'Good',
    devtools: true,
    middleware: false,
    persistence: false,
    ssr: false,
    performance: 'Excellent',
    ecosystem: 'Limited',
    useCase: ['代理状态', '简单API', '快速开发', '原型验证']
  },
  {
    name: 'React Query',
    bundleSize: '~39kb',
    learningCurve: 'Medium',
    typescript: 'Excellent',
    devtools: true,
    middleware: false,
    persistence: true,
    ssr: true,
    performance: 'Excellent',
    ecosystem: 'Rich',
    useCase: ['服务端状态', '缓存管理', '数据同步', 'API状态管理']
  }
];

// 详细功能对比
const featureComparison = {
  // 1. API设计对比
  apiDesign: {
    redux: {
      complexity: 'High',
      boilerplate: 'Much',
      example: `
        // Redux Toolkit
        const userSlice = createSlice({
          name: 'user',
          initialState: { name: '', email: '' },
          reducers: {
            updateUser: (state, action) => {
              state.name = action.payload.name;
              state.email = action.payload.email;
            }
          }
        });

        // 使用
        const dispatch = useDispatch();
        const user = useSelector(state => state.user);
        dispatch(userSlice.actions.updateUser({ name: 'John', email: 'john@example.com' }));
      `
    },
    zustand: {
      complexity: 'Low',
      boilerplate: 'Minimal',
      example: `
        // Zustand
        const useUserStore = create((set) => ({
          name: '',
          email: '',
          updateUser: (user) => set({ name: user.name, email: user.email })
        }));

        // 使用
        const { name, email, updateUser } = useUserStore();
        updateUser({ name: 'John', email: 'john@example.com' });
      `
    },
    context: {
      complexity: 'Medium',
      boilerplate: 'Medium',
      example: `
        // Context API
        const UserContext = createContext();

        function UserProvider({ children }) {
          const [user, setUser] = useState({ name: '', email: '' });

          const updateUser = (newUser) => {
            setUser(prev => ({ ...prev, ...newUser }));
          };

          return (
            <UserContext.Provider value={{ user, updateUser }}>
              {children}
            </UserContext.Provider>
          );
        }

        // 使用
        const { user, updateUser } = useContext(UserContext);
      `
    },
    jotai: {
      complexity: 'Medium',
      boilerplate: 'Low',
      example: `
        // Jotai
        const nameAtom = atom('');
        const emailAtom = atom('');
        const userAtom = atom(
          (get) => ({ name: get(nameAtom), email: get(emailAtom) }),
          (get, set, user) => {
            set(nameAtom, user.name);
            set(emailAtom, user.email);
          }
        );

        // 使用
        const [user, setUser] = useAtom(userAtom);
        setUser({ name: 'John', email: 'john@example.com' });
      `
    }
  }
};
```

---

## 🏗️ Redux生态系统架构

### Redux Toolkit现代化实践

Redux Toolkit (RTK) 是Redux官方推荐的现代化开发方式，大大简化了Redux的使用：

```typescript
// Redux Toolkit 完整架构示例
// store/index.ts - Store配置
import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { persistStore, persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage';

import userSlice from './slices/userSlice';
import productSlice from './slices/productSlice';
import cartSlice from './slices/cartSlice';
import { apiSlice } from './api/apiSlice';

// 持久化配置
const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['user', 'cart'], // 只持久化用户和购物车状态
};

const persistedUserReducer = persistReducer(persistConfig, userSlice.reducer);

// 配置Store
export const store = configureStore({
  reducer: {
    user: persistedUserReducer,
    product: productSlice.reducer,
    cart: cartSlice.reducer,
    api: apiSlice.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }).concat(apiSlice.middleware),
  devTools: process.env.NODE_ENV !== 'production',
});

// 设置RTK Query监听器
setupListeners(store.dispatch);

export const persistor = persistStore(store);
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// store/slices/userSlice.ts - 用户状态切片
import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { User, LoginCredentials } from '@/types/user';
import { authApi } from '@/services/authApi';

interface UserState {
  currentUser: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  preferences: {
    theme: 'light' | 'dark';
    language: string;
    notifications: boolean;
  };
}

const initialState: UserState = {
  currentUser: null,
  isAuthenticated: false,
  loading: false,
  error: null,
  preferences: {
    theme: 'light',
    language: 'zh-CN',
    notifications: true,
  },
};

// 异步Thunk
export const loginUser = createAsyncThunk(
  'user/login',
  async (credentials: LoginCredentials, { rejectWithValue }) => {
    try {
      const response = await authApi.login(credentials);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || 'Login failed');
    }
  }
);

export const fetchUserProfile = createAsyncThunk(
  'user/fetchProfile',
  async (_, { getState, rejectWithValue }) => {
    try {
      const state = getState() as RootState;
      if (!state.user.isAuthenticated) {
        throw new Error('User not authenticated');
      }

      const response = await authApi.getProfile();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || 'Failed to fetch profile');
    }
  }
);

export const updateUserPreferences = createAsyncThunk(
  'user/updatePreferences',
  async (preferences: Partial<UserState['preferences']>, { getState, rejectWithValue }) => {
    try {
      const state = getState() as RootState;
      const updatedPreferences = { ...state.user.preferences, ...preferences };

      await authApi.updatePreferences(updatedPreferences);
      return updatedPreferences;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || 'Failed to update preferences');
    }
  }
);

// Slice定义
const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    // 同步actions
    logout: (state) => {
      state.currentUser = null;
      state.isAuthenticated = false;
      state.error = null;
    },

    clearError: (state) => {
      state.error = null;
    },

    updateLocalPreferences: (state, action: PayloadAction<Partial<UserState['preferences']>>) => {
      state.preferences = { ...state.preferences, ...action.payload };
    },

    setUser: (state, action: PayloadAction<User>) => {
      state.currentUser = action.payload;
      state.isAuthenticated = true;
    },
  },

  // 异步actions的处理
  extraReducers: (builder) => {
    builder
      // 登录
      .addCase(loginUser.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.loading = false;
        state.currentUser = action.payload.user;
        state.isAuthenticated = true;
        state.error = null;
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
        state.isAuthenticated = false;
      })

      // 获取用户资料
      .addCase(fetchUserProfile.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchUserProfile.fulfilled, (state, action) => {
        state.loading = false;
        state.currentUser = action.payload;
      })
      .addCase(fetchUserProfile.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })

      // 更新偏好设置
      .addCase(updateUserPreferences.fulfilled, (state, action) => {
        state.preferences = action.payload;
      });
  },
});

export const { logout, clearError, updateLocalPreferences, setUser } = userSlice.actions;
export default userSlice;

// store/slices/cartSlice.ts - 购物车状态切片
import { createSlice, createSelector, PayloadAction } from '@reduxjs/toolkit';
import { Product } from '@/types/product';

interface CartItem {
  product: Product;
  quantity: number;
  selectedOptions?: Record<string, string>;
}

interface CartState {
  items: CartItem[];
  isOpen: boolean;
  lastUpdated: number;
}

const initialState: CartState = {
  items: [],
  isOpen: false,
  lastUpdated: Date.now(),
};

const cartSlice = createSlice({
  name: 'cart',
  initialState,
  reducers: {
    addToCart: (state, action: PayloadAction<{ product: Product; quantity?: number; options?: Record<string, string> }>) => {
      const { product, quantity = 1, options = {} } = action.payload;

      // 查找是否已存在相同商品和选项的项目
      const existingItemIndex = state.items.findIndex(
        item =>
          item.product.id === product.id &&
          JSON.stringify(item.selectedOptions) === JSON.stringify(options)
      );

      if (existingItemIndex >= 0) {
        // 更新数量
        state.items[existingItemIndex].quantity += quantity;
      } else {
        // 添加新项目
        state.items.push({
          product,
          quantity,
          selectedOptions: options,
        });
      }

      state.lastUpdated = Date.now();
    },

    removeFromCart: (state, action: PayloadAction<{ productId: string; options?: Record<string, string> }>) => {
      const { productId, options = {} } = action.payload;

      state.items = state.items.filter(
        item => !(
          item.product.id === productId &&
          JSON.stringify(item.selectedOptions) === JSON.stringify(options)
        )
      );

      state.lastUpdated = Date.now();
    },

    updateQuantity: (state, action: PayloadAction<{ productId: string; quantity: number; options?: Record<string, string> }>) => {
      const { productId, quantity, options = {} } = action.payload;

      const itemIndex = state.items.findIndex(
        item =>
          item.product.id === productId &&
          JSON.stringify(item.selectedOptions) === JSON.stringify(options)
      );

      if (itemIndex >= 0) {
        if (quantity <= 0) {
          state.items.splice(itemIndex, 1);
        } else {
          state.items[itemIndex].quantity = quantity;
        }
        state.lastUpdated = Date.now();
      }
    },

    clearCart: (state) => {
      state.items = [];
      state.lastUpdated = Date.now();
    },

    toggleCart: (state) => {
      state.isOpen = !state.isOpen;
    },

    setCartOpen: (state, action: PayloadAction<boolean>) => {
      state.isOpen = action.payload;
    },
  },
});

export const {
  addToCart,
  removeFromCart,
  updateQuantity,
  clearCart,
  toggleCart,
  setCartOpen
} = cartSlice.actions;

export default cartSlice;

// 选择器 (Selectors)
export const selectCartItems = (state: RootState) => state.cart.items;
export const selectCartIsOpen = (state: RootState) => state.cart.isOpen;

// 记忆化选择器
export const selectCartTotal = createSelector(
  [selectCartItems],
  (items) => items.reduce((total, item) => total + (parseFloat(item.product.price) * item.quantity), 0)
);

export const selectCartItemCount = createSelector(
  [selectCartItems],
  (items) => items.reduce((count, item) => count + item.quantity, 0)
);

export const selectCartItemsByCategory = createSelector(
  [selectCartItems],
  (items) => {
    const grouped: Record<string, CartItem[]> = {};
    items.forEach(item => {
      const category = item.product.categoryId;
      if (!grouped[category]) {
        grouped[category] = [];
      }
      grouped[category].push(item);
    });
    return grouped;
  }
);
```

---

## ⚡ Zustand轻量级方案

### Zustand核心架构

Zustand是一个轻量级的状态管理库，提供了简洁的API和优秀的TypeScript支持：

```typescript
// stores/userStore.ts - Zustand用户状态管理
import { create } from 'zustand';
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { User, LoginCredentials } from '@/types/user';
import { authApi } from '@/services/authApi';

interface UserState {
  // 状态
  currentUser: User | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  preferences: {
    theme: 'light' | 'dark';
    language: string;
    notifications: boolean;
  };

  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  fetchProfile: () => Promise<void>;
  updatePreferences: (preferences: Partial<UserState['preferences']>) => Promise<void>;
  clearError: () => void;
  setUser: (user: User) => void;
}

export const useUserStore = create<UserState>()(
  devtools(
    persist(
      subscribeWithSelector(
        immer((set, get) => ({
          // 初始状态
          currentUser: null,
          isAuthenticated: false,
          loading: false,
          error: null,
          preferences: {
            theme: 'light',
            language: 'zh-CN',
            notifications: true,
          },

          // Actions
          login: async (credentials) => {
            set((state) => {
              state.loading = true;
              state.error = null;
            });

            try {
              const response = await authApi.login(credentials);
              set((state) => {
                state.currentUser = response.data.user;
                state.isAuthenticated = true;
                state.loading = false;
              });
            } catch (error: any) {
              set((state) => {
                state.error = error.response?.data?.message || 'Login failed';
                state.loading = false;
                state.isAuthenticated = false;
              });
            }
          },

          logout: () => {
            set((state) => {
              state.currentUser = null;
              state.isAuthenticated = false;
              state.error = null;
            });
          },

          fetchProfile: async () => {
            const { isAuthenticated } = get();
            if (!isAuthenticated) {
              throw new Error('User not authenticated');
            }

            set((state) => {
              state.loading = true;
            });

            try {
              const response = await authApi.getProfile();
              set((state) => {
                state.currentUser = response.data;
                state.loading = false;
              });
            } catch (error: any) {
              set((state) => {
                state.error = error.response?.data?.message || 'Failed to fetch profile';
                state.loading = false;
              });
            }
          },

          updatePreferences: async (newPreferences) => {
            const { preferences } = get();
            const updatedPreferences = { ...preferences, ...newPreferences };

            try {
              await authApi.updatePreferences(updatedPreferences);
              set((state) => {
                state.preferences = updatedPreferences;
              });
            } catch (error: any) {
              set((state) => {
                state.error = error.response?.data?.message || 'Failed to update preferences';
              });
            }
          },

          clearError: () => {
            set((state) => {
              state.error = null;
            });
          },

          setUser: (user) => {
            set((state) => {
              state.currentUser = user;
              state.isAuthenticated = true;
            });
          },
        }))
      ),
      {
        name: 'user-store',
        partialize: (state) => ({
          currentUser: state.currentUser,
          isAuthenticated: state.isAuthenticated,
          preferences: state.preferences,
        }),
      }
    ),
    { name: 'user-store' }
  )
);

// 选择器
export const useUser = () => useUserStore((state) => state.currentUser);
export const useIsAuthenticated = () => useUserStore((state) => state.isAuthenticated);
export const useUserLoading = () => useUserStore((state) => state.loading);
export const useUserError = () => useUserStore((state) => state.error);
export const useUserPreferences = () => useUserStore((state) => state.preferences);

// stores/cartStore.ts - Zustand购物车状态管理
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { Product } from '@/types/product';

interface CartItem {
  product: Product;
  quantity: number;
  selectedOptions?: Record<string, string>;
}

interface CartState {
  items: CartItem[];
  isOpen: boolean;
  lastUpdated: number;

  // Actions
  addToCart: (product: Product, quantity?: number, options?: Record<string, string>) => void;
  removeFromCart: (productId: string, options?: Record<string, string>) => void;
  updateQuantity: (productId: string, quantity: number, options?: Record<string, string>) => void;
  clearCart: () => void;
  toggleCart: () => void;
  setCartOpen: (isOpen: boolean) => void;

  // Computed
  getTotalPrice: () => number;
  getItemCount: () => number;
  getItemsByCategory: () => Record<string, CartItem[]>;
}

export const useCartStore = create<CartState>()(
  devtools(
    persist(
      immer((set, get) => ({
        items: [],
        isOpen: false,
        lastUpdated: Date.now(),

        addToCart: (product, quantity = 1, options = {}) => {
          set((state) => {
            const existingItemIndex = state.items.findIndex(
              item =>
                item.product.id === product.id &&
                JSON.stringify(item.selectedOptions) === JSON.stringify(options)
            );

            if (existingItemIndex >= 0) {
              state.items[existingItemIndex].quantity += quantity;
            } else {
              state.items.push({
                product,
                quantity,
                selectedOptions: options,
              });
            }

            state.lastUpdated = Date.now();
          });
        },

        removeFromCart: (productId, options = {}) => {
          set((state) => {
            state.items = state.items.filter(
              item => !(
                item.product.id === productId &&
                JSON.stringify(item.selectedOptions) === JSON.stringify(options)
              )
            );
            state.lastUpdated = Date.now();
          });
        },

        updateQuantity: (productId, quantity, options = {}) => {
          set((state) => {
            const itemIndex = state.items.findIndex(
              item =>
                item.product.id === productId &&
                JSON.stringify(item.selectedOptions) === JSON.stringify(options)
            );

            if (itemIndex >= 0) {
              if (quantity <= 0) {
                state.items.splice(itemIndex, 1);
              } else {
                state.items[itemIndex].quantity = quantity;
              }
              state.lastUpdated = Date.now();
            }
          });
        },

        clearCart: () => {
          set((state) => {
            state.items = [];
            state.lastUpdated = Date.now();
          });
        },

        toggleCart: () => {
          set((state) => {
            state.isOpen = !state.isOpen;
          });
        },

        setCartOpen: (isOpen) => {
          set((state) => {
            state.isOpen = isOpen;
          });
        },

        // Computed values
        getTotalPrice: () => {
          const { items } = get();
          return items.reduce((total, item) => total + (parseFloat(item.product.price) * item.quantity), 0);
        },

        getItemCount: () => {
          const { items } = get();
          return items.reduce((count, item) => count + item.quantity, 0);
        },

        getItemsByCategory: () => {
          const { items } = get();
          const grouped: Record<string, CartItem[]> = {};
          items.forEach(item => {
            const category = item.product.categoryId;
            if (!grouped[category]) {
              grouped[category] = [];
            }
            grouped[category].push(item);
          });
          return grouped;
        },
      })),
      {
        name: 'cart-store',
        partialize: (state) => ({
          items: state.items,
          lastUpdated: state.lastUpdated,
        }),
      }
    ),
    { name: 'cart-store' }
  )
);

// 选择器Hooks
export const useCartItems = () => useCartStore((state) => state.items);
export const useCartIsOpen = () => useCartStore((state) => state.isOpen);
export const useCartTotal = () => useCartStore((state) => state.getTotalPrice());
export const useCartItemCount = () => useCartStore((state) => state.getItemCount());
```

### Zustand高级模式

```typescript
// stores/storeFactory.ts - Store工厂模式
import { create, StateCreator } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// 通用Store接口
interface BaseState {
  loading: boolean;
  error: string | null;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

// 基础Store创建器
const createBaseSlice: StateCreator<BaseState> = (set) => ({
  loading: false,
  error: null,

  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),
});

// 异步操作Mixin
interface AsyncMixin {
  executeAsync: <T>(asyncFn: () => Promise<T>) => Promise<T>;
}

const createAsyncMixin = <T extends BaseState>(): StateCreator<T & AsyncMixin, [], [], AsyncMixin> =>
  (set, get) => ({
    executeAsync: async (asyncFn) => {
      set({ loading: true, error: null } as Partial<T & AsyncMixin>);

      try {
        const result = await asyncFn();
        set({ loading: false } as Partial<T & AsyncMixin>);
        return result;
      } catch (error: any) {
        set({
          loading: false,
          error: error.message || 'An error occurred'
        } as Partial<T & AsyncMixin>);
        throw error;
      }
    },
  });

// 使用工厂创建特定Store
export function createEntityStore<T>(
  name: string,
  initialState: T,
  actions: (set: any, get: any) => Record<string, any>
) {
  return create(
    devtools(
      persist(
        (set, get) => ({
          ...initialState,
          ...createBaseSlice(set, get),
          ...createAsyncMixin<BaseState>()(set, get, {} as any),
          ...actions(set, get),
        }),
        { name }
      ),
      { name }
    )
  );
}

// stores/productStore.ts - 使用工厂创建产品Store
interface ProductState extends BaseState {
  products: Product[];
  selectedProduct: Product | null;
  filters: ProductFilters;

  fetchProducts: () => Promise<void>;
  fetchProduct: (id: string) => Promise<void>;
  setFilters: (filters: Partial<ProductFilters>) => void;
  clearFilters: () => void;
}

export const useProductStore = createEntityStore<Omit<ProductState, keyof BaseState | keyof AsyncMixin>>(
  'product-store',
  {
    products: [],
    selectedProduct: null,
    filters: {},
  },
  (set, get) => ({
    fetchProducts: async () => {
      const { executeAsync } = get();
      return executeAsync(async () => {
        const response = await productApi.getProducts(get().filters);
        set({ products: response.data });
      });
    },

    fetchProduct: async (id: string) => {
      const { executeAsync } = get();
      return executeAsync(async () => {
        const response = await productApi.getProduct(id);
        set({ selectedProduct: response.data });
      });
    },

    setFilters: (newFilters) => {
      set((state) => ({
        filters: { ...state.filters, ...newFilters }
      }));
    },

    clearFilters: () => {
      set({ filters: {} });
    },
  })
);
```

---

## 🎯 面试常考知识点

### 1. 状态管理方案选择

**Q: 如何为项目选择合适的状态管理方案？**

**A: 状态管理方案选择决策框架：**

```typescript
// 状态管理选择决策矩阵
const stateManagementDecisionMatrix = {
  // 项目规模维度
  projectScale: {
    small: {
      components: '<50个组件',
      developers: '1-3人',
      recommendation: 'useState + useContext',
      reason: '简单直接，无需引入额外复杂性'
    },
    medium: {
      components: '50-200个组件',
      developers: '3-10人',
      recommendation: 'Zustand + React Query',
      reason: '轻量级，易于学习，性能优秀'
    },
    large: {
      components: '>200个组件',
      developers: '>10人',
      recommendation: 'Redux Toolkit + RTK Query',
      reason: '标准化，强大的调试工具，团队协作友好'
    }
  },

  // 状态复杂度维度
  stateComplexity: {
    simple: {
      description: '简单CRUD，基础状态',
      recommendation: 'useState + useReducer',
      features: ['本地状态', '简单更新', '无复杂逻辑']
    },
    moderate: {
      description: '中等复杂度，跨组件状态',
      recommendation: 'Context API + useReducer',
      features: ['共享状态', '中等复杂逻辑', '有限的异步操作']
    },
    complex: {
      description: '复杂业务逻辑，大量异步操作',
      recommendation: 'Redux Toolkit',
      features: ['复杂状态逻辑', '大量异步操作', '时间旅行调试']
    }
  },

  // 性能要求维度
  performanceRequirements: {
    standard: {
      description: '标准性能要求',
      recommendation: 'Zustand',
      optimizations: ['选择性订阅', '状态分割']
    },
    high: {
      description: '高性能要求',
      recommendation: 'Jotai',
      optimizations: ['原子化更新', '细粒度控制', '避免不必要重渲染']
    },
    extreme: {
      description: '极致性能要求',
      recommendation: 'Valtio + 手动优化',
      optimizations: ['代理状态', '精确更新', '自定义优化']
    }
  },

  // 团队技能维度
  teamSkills: {
    beginner: {
      description: '团队React经验较少',
      recommendation: 'Zustand',
      reason: 'API简单，学习曲线平缓'
    },
    intermediate: {
      description: '团队有一定React经验',
      recommendation: 'Redux Toolkit',
      reason: '标准化实践，丰富的学习资源'
    },
    advanced: {
      description: '团队React经验丰富',
      recommendation: '根据具体需求选择',
      reason: '可以根据项目特点选择最适合的方案'
    }
  }
};

// 常见面试问题和答案
const commonInterviewQuestions = {
  q1: {
    question: 'Redux和Zustand的主要区别是什么？',
    answer: {
      redux: {
        pros: ['强大的调试工具', '丰富的生态', '标准化实践', '时间旅行'],
        cons: ['样板代码多', '学习曲线陡峭', '包体积大'],
        useCase: '大型应用，复杂状态逻辑'
      },
      zustand: {
        pros: ['API简单', '包体积小', '性能优秀', 'TypeScript友好'],
        cons: ['生态相对较小', '调试工具有限'],
        useCase: '中小型应用，快速开发'
      }
    }
  },

  q2: {
    question: '什么时候应该使用Context API？',
    answer: {
      适合场景: [
        '主题切换',
        '用户认证状态',
        '语言设置',
        '避免prop drilling'
      ],
      不适合场景: [
        '频繁更新的状态',
        '复杂的状态逻辑',
        '需要性能优化的场景'
      ],
      原因: 'Context会导致所有消费者重新渲染，不适合频繁变化的状态'
    }
  },

  q3: {
    question: '如何避免状态管理中的性能问题？',
    answer: {
      strategies: [
        '状态分割：将大的状态对象拆分成小的独立状态',
        '选择性订阅：只订阅需要的状态片段',
        '记忆化：使用useMemo和useCallback避免不必要的计算',
        '原子化：使用Jotai等原子化状态管理',
        '虚拟化：对大列表使用虚拟滚动'
      ],
      example: `
        // ❌ 会导致所有组件重渲染
        const globalState = {
          user: { ... },
          products: [ ... ],
          ui: { ... }
        };

        // ✅ 状态分割
        const useUserStore = create(...);
        const useProductStore = create(...);
        const useUIStore = create(...);
      `
    }
  },

  q4: {
    question: '如何处理异步状态管理？',
    answer: {
      approaches: [
        'Redux Thunk: 简单的异步action',
        'Redux Saga: 复杂的异步流程控制',
        'RTK Query: 专门的数据获取',
        'React Query: 服务端状态管理',
        'Zustand: 内置异步支持'
      ],
      bestPractices: [
        '分离客户端状态和服务端状态',
        '使用专门的数据获取库',
        '实现乐观更新',
        '处理加载和错误状态'
      ]
    }
  }
};
```

### 2. 状态管理架构设计

**Q: 如何设计大型应用的状态管理架构？**

**A: 企业级状态管理架构设计：**

```typescript
// 企业级状态管理架构
const enterpriseStateArchitecture = {
  // 1. 分层架构
  layeredArchitecture: {
    presentation: {
      layer: '表现层',
      responsibility: 'UI组件，用户交互',
      stateTypes: ['UI状态', '表单状态', '临时状态'],
      tools: ['useState', 'useReducer', 'React Hook Form']
    },

    business: {
      layer: '业务层',
      responsibility: '业务逻辑，状态管理',
      stateTypes: ['业务状态', '应用状态', '用户状态'],
      tools: ['Redux Toolkit', 'Zustand', 'Context API']
    },

    data: {
      layer: '数据层',
      responsibility: '数据获取，缓存管理',
      stateTypes: ['服务端状态', '缓存状态', 'API状态'],
      tools: ['React Query', 'SWR', 'Apollo Client']
    },

    infrastructure: {
      layer: '基础设施层',
      responsibility: '持久化，同步，监控',
      stateTypes: ['持久化状态', '同步状态', '监控状态'],
      tools: ['Redux Persist', 'LocalStorage', 'WebSocket']
    }
  },

  // 2. 模块化设计
  modularDesign: {
    featureModules: {
      structure: 'features/[feature]/store/',
      example: `
        features/
          user/
            store/
              userSlice.ts
              userSelectors.ts
              userThunks.ts
          product/
            store/
              productSlice.ts
              productSelectors.ts
          cart/
            store/
              cartSlice.ts
      `,
      benefits: ['独立开发', '易于测试', '代码复用']
    },

    sharedModules: {
      structure: 'shared/store/',
      example: `
        shared/
          store/
            rootReducer.ts
            store.ts
            middleware/
            types/
      `,
      purpose: '共享状态，通用逻辑'
    }
  },

  // 3. 状态规范化
  stateNormalization: {
    principle: '扁平化状态结构，避免嵌套',
    example: `
      // ❌ 嵌套结构
      {
        users: {
          1: {
            id: 1,
            name: 'John',
            posts: [
              { id: 1, title: 'Post 1', author: 1 },
              { id: 2, title: 'Post 2', author: 1 }
            ]
          }
        }
      }

      // ✅ 规范化结构
      {
        users: {
          byId: {
            1: { id: 1, name: 'John' }
          },
          allIds: [1]
        },
        posts: {
          byId: {
            1: { id: 1, title: 'Post 1', authorId: 1 },
            2: { id: 2, title: 'Post 2', authorId: 1 }
          },
          allIds: [1, 2]
        }
      }
    `,
    benefits: ['避免数据重复', '更新效率高', '查询性能好']
  },

  // 4. 状态同步策略
  stateSynchronization: {
    clientToServer: {
      strategies: ['乐观更新', '悲观更新', '混合策略'],
      implementation: `
        // 乐观更新示例
        const updateUser = async (userId, updates) => {
          // 立即更新本地状态
          dispatch(updateUserOptimistic({ userId, updates }));

          try {
            // 发送服务器请求
            const result = await api.updateUser(userId, updates);
            dispatch(updateUserSuccess(result));
          } catch (error) {
            // 回滚本地状态
            dispatch(updateUserFailure({ userId, error }));
          }
        };
      `
    },

    serverToClient: {
      strategies: ['轮询', 'WebSocket', 'SSE', 'GraphQL订阅'],
      implementation: `
        // WebSocket实时同步
        const useRealtimeSync = () => {
          useEffect(() => {
            const ws = new WebSocket('ws://localhost:3001');

            ws.onmessage = (event) => {
              const update = JSON.parse(event.data);
              dispatch(applyServerUpdate(update));
            };

            return () => ws.close();
          }, []);
        };
      `
    }
  }
};
```

---

## 📚 实战练习

### 练习1：设计电商状态管理架构

**任务**: 为Mall-Frontend设计完整的状态管理架构，包括用户、商品、购物车、订单等模块。

**要求**:
- 选择合适的状态管理方案
- 设计模块化的状态结构
- 实现状态持久化
- 添加性能优化策略

### 练习2：实现状态管理迁移

**任务**: 将一个使用Context API的应用迁移到Zustand。

**要求**:
- 分析现有Context结构
- 设计Zustand store架构
- 实现渐进式迁移
- 保证功能完整性

### 练习3：构建状态管理测试套件

**任务**: 为状态管理模块编写完整的测试套件。

**要求**:
- 单元测试：测试reducers和actions
- 集成测试：测试组件与状态的集成
- 端到端测试：测试完整的用户流程
- 性能测试：测试状态更新性能

---

## 📚 本章总结

通过本章学习，我们全面掌握了状态管理架构设计的核心技术：

### 🎯 核心收获

1. **状态管理理论精通** 🎯
   - 掌握了状态分类和管理原则
   - 理解了状态管理模式的演进历程
   - 学会了状态管理方案的选择策略

2. **方案深度对比** 🔄
   - 深入对比了Redux、Zustand、Context API等方案
   - 掌握了各方案的优劣和适用场景
   - 理解了企业级选择策略

3. **架构设计能力** 🏗️
   - 掌握了分层架构和模块化设计
   - 学会了状态规范化和同步策略
   - 理解了大型应用的状态治理

4. **性能优化技巧** ⚡
   - 掌握了状态管理的性能优化方法
   - 学会了避免不必要的重渲染
   - 理解了内存和网络优化策略

5. **企业级实践** 🚀
   - 学会了大型团队的状态管理规范
   - 掌握了状态管理的测试策略
   - 理解了迁移和升级的方法

### 🚀 技术进阶

- **下一步学习**: 组件库设计与开发
- **实践建议**: 在项目中应用分层状态管理架构
- **深入方向**: 微前端状态共享和跨应用状态同步

状态管理是现代前端应用的核心，选择合适的方案和架构设计是成功的关键！ 🎉

---

*下一章我们将学习《组件库设计与开发》，探索可复用组件系统的构建！* 🚀
```
```
```
```
