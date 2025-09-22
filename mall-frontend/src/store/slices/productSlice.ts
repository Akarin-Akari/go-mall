import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Product, Category, PageResult, PaginationParams } from '@/types';
import { productAPI } from '@/services/api';
import { message } from 'antd';

// 商品状态接口
interface ProductState {
  // 商品列表
  products: Product[];
  total: number;
  loading: boolean;

  // 当前商品详情
  currentProduct: Product | null;
  productLoading: boolean;

  // 分类列表
  categories: Category[];
  categoriesLoading: boolean;

  // 搜索和筛选
  searchParams: {
    keyword?: string;
    category_id?: number;
    status?: string;
    min_price?: number;
    max_price?: number;
    page: number;
    page_size: number;
  };
}

// 初始状态
const initialState: ProductState = {
  products: [],
  total: 0,
  loading: false,

  currentProduct: null,
  productLoading: false,

  categories: [],
  categoriesLoading: false,

  searchParams: {
    page: 1,
    page_size: 10,
  },
};

// 异步actions
export const fetchProductsAsync = createAsyncThunk(
  'product/fetchProducts',
  async (
    params: PaginationParams & {
      category_id?: number;
      status?: string;
      min_price?: number;
      max_price?: number;
      sort_by?: string;
      min_rating?: number;
    },
    { rejectWithValue }
  ) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { getPagedProducts } = await import('@/data/mockProducts');
        const result = getPagedProducts(
          params.page || 1,
          params.page_size || 10,
          {
            keyword: params.keyword,
            categoryId: params.category_id,
            minPrice: params.min_price,
            maxPrice: params.max_price,
            minRating: params.min_rating,
            sortBy: params.sort_by,
          }
        );

        // 模拟网络延迟
        await new Promise(resolve => setTimeout(resolve, 500));

        return {
          list: result.products,
          total: result.total,
          page: result.page,
          page_size: result.pageSize,
        };
      }

      // 生产环境使用真实API
      const response = await productAPI.getProducts(params);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取商品列表失败');
    }
  }
);

export const fetchProductDetailAsync = createAsyncThunk(
  'product/fetchProductDetail',
  async (id: number, { rejectWithValue }) => {
    try {
      // 在开发环境使用模拟数据
      if (process.env.NODE_ENV === 'development') {
        const { mockProducts } = await import('@/data/mockProducts');
        const product = mockProducts.find((p: Product) => p.id === id);

        if (!product) {
          return rejectWithValue('商品不存在');
        }

        // 模拟网络延迟
        await new Promise(resolve => setTimeout(resolve, 800));

        return product;
      }

      // 生产环境使用真实API
      const response = await productAPI.getProductDetail(id);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取商品详情失败');
    }
  }
);

export const fetchCategoriesAsync = createAsyncThunk(
  'product/fetchCategories',
  async (_, { rejectWithValue }) => {
    try {
      const response = await productAPI.getCategories();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '获取分类列表失败');
    }
  }
);

export const createProductAsync = createAsyncThunk(
  'product/createProduct',
  async (
    productData: Omit<Product, 'id' | 'created_at' | 'updated_at'>,
    { rejectWithValue }
  ) => {
    try {
      const response = await productAPI.createProduct(productData);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '创建商品失败');
    }
  }
);

export const updateProductAsync = createAsyncThunk(
  'product/updateProduct',
  async (
    { id, data }: { id: number; data: Partial<Product> },
    { rejectWithValue }
  ) => {
    try {
      const response = await productAPI.updateProduct(id, data);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.message || '更新商品失败');
    }
  }
);

export const deleteProductAsync = createAsyncThunk(
  'product/deleteProduct',
  async (id: number, { rejectWithValue }) => {
    try {
      await productAPI.deleteProduct(id);
      return id;
    } catch (error: any) {
      return rejectWithValue(error.message || '删除商品失败');
    }
  }
);

// 创建slice
const productSlice = createSlice({
  name: 'product',
  initialState,
  reducers: {
    // 设置搜索参数
    setSearchParams: (
      state,
      action: PayloadAction<Partial<ProductState['searchParams']>>
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

    // 清除当前商品
    clearCurrentProduct: state => {
      state.currentProduct = null;
    },

    // 更新商品库存（用于购买后更新）
    updateProductStock: (
      state,
      action: PayloadAction<{ id: number; stock: number; sold_count: number }>
    ) => {
      const product = state.products.find(p => p.id === action.payload.id);
      if (product) {
        product.stock = action.payload.stock;
        product.sold_count = action.payload.sold_count;
      }

      if (
        state.currentProduct &&
        state.currentProduct.id === action.payload.id
      ) {
        state.currentProduct.stock = action.payload.stock;
        state.currentProduct.sold_count = action.payload.sold_count;
      }
    },
  },
  extraReducers: builder => {
    // 获取商品列表
    builder
      .addCase(fetchProductsAsync.pending, state => {
        state.loading = true;
      })
      .addCase(fetchProductsAsync.fulfilled, (state, action) => {
        state.loading = false;
        state.products = action.payload?.list || [];
        state.total = action.payload?.total || 0;
      })
      .addCase(fetchProductsAsync.rejected, (state, action) => {
        state.loading = false;
        message.error(action.payload as string);
      });

    // 获取商品详情
    builder
      .addCase(fetchProductDetailAsync.pending, state => {
        state.productLoading = true;
      })
      .addCase(fetchProductDetailAsync.fulfilled, (state, action) => {
        state.productLoading = false;
        state.currentProduct = action.payload;
      })
      .addCase(fetchProductDetailAsync.rejected, (state, action) => {
        state.productLoading = false;
        state.currentProduct = null;
        message.error(action.payload as string);
      });

    // 获取分类列表
    builder
      .addCase(fetchCategoriesAsync.pending, state => {
        state.categoriesLoading = true;
      })
      .addCase(fetchCategoriesAsync.fulfilled, (state, action) => {
        state.categoriesLoading = false;
        state.categories = action.payload || [];
      })
      .addCase(fetchCategoriesAsync.rejected, (state, action) => {
        state.categoriesLoading = false;
        message.error(action.payload as string);
      });

    // 创建商品
    builder
      .addCase(createProductAsync.fulfilled, (state, action) => {
        state.products.unshift(action.payload);
        state.total += 1;
      })
      .addCase(createProductAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });

    // 更新商品
    builder
      .addCase(updateProductAsync.fulfilled, (state, action) => {
        const index = state.products.findIndex(p => p.id === action.payload.id);
        if (index !== -1) {
          state.products[index] = action.payload;
        }

        if (
          state.currentProduct &&
          state.currentProduct.id === action.payload.id
        ) {
          state.currentProduct = action.payload;
        }
      })
      .addCase(updateProductAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });

    // 删除商品
    builder
      .addCase(deleteProductAsync.fulfilled, (state, action) => {
        state.products = state.products.filter(p => p.id !== action.payload);
        state.total -= 1;

        if (
          state.currentProduct &&
          state.currentProduct.id === action.payload
        ) {
          state.currentProduct = null;
        }
      })
      .addCase(deleteProductAsync.rejected, (state, action) => {
        message.error(action.payload as string);
      });
  },
});

// 导出actions
export const {
  setSearchParams,
  resetSearchParams,
  clearCurrentProduct,
  updateProductStock,
} = productSlice.actions;

// 选择器
export const selectProduct = (state: { product: ProductState }) =>
  state.product;
export const selectProducts = (state: { product: ProductState }) =>
  state.product.products;
export const selectCurrentProduct = (state: { product: ProductState }) =>
  state.product.currentProduct;
export const selectCategories = (state: { product: ProductState }) =>
  state.product.categories;
export const selectProductLoading = (state: { product: ProductState }) =>
  state.product.loading;
export const selectProductDetailLoading = (state: { product: ProductState }) =>
  state.product.productLoading;
export const selectSearchParams = (state: { product: ProductState }) =>
  state.product.searchParams;
export const selectProductTotal = (state: { product: ProductState }) =>
  state.product.total;

// 导出reducer
export default productSlice.reducer;
