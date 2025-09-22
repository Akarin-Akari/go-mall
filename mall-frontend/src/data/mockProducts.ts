import { Product } from '@/types';

// 模拟商品数据
export const mockProducts: Product[] = [
  {
    id: 1,
    name: 'iPhone 15 Pro Max 256GB 深空黑色',
    description:
      '搭载A17 Pro芯片，支持5G网络，拥有强大的摄影系统和超长续航能力。',
    price: '9999.00',
    discount_price: '8999.00',
    stock: 50,
    category_id: 1,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=400',
      'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?w=400',
    ],
    rating: 4.8,
    sales_count: 1250,
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:00:00Z',
  },
  {
    id: 2,
    name: 'MacBook Pro 14英寸 M3芯片',
    description:
      '全新M3芯片，14英寸Liquid Retina XDR显示屏，专业级性能，适合创意工作者。',
    price: '14999.00',
    discount_price: '13999.00',
    stock: 30,
    category_id: 1,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400',
      'https://images.unsplash.com/photo-1541807084-5c52b6b3adef?w=400',
    ],
    rating: 4.9,
    sales_count: 856,
    created_at: '2024-01-10T10:00:00Z',
    updated_at: '2024-01-10T10:00:00Z',
  },
  {
    id: 3,
    name: 'Nike Air Max 270 运动鞋',
    description: '经典Air Max气垫设计，舒适透气，适合日常运动和休闲穿着。',
    price: '899.00',
    discount_price: '699.00',
    stock: 120,
    category_id: 2,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400',
      'https://images.unsplash.com/photo-1549298916-b41d501d3772?w=400',
    ],
    rating: 4.6,
    sales_count: 2340,
    created_at: '2024-01-20T10:00:00Z',
    updated_at: '2024-01-20T10:00:00Z',
  },
  {
    id: 4,
    name: 'Adidas Ultraboost 22 跑步鞋',
    description: '采用Boost中底技术，提供卓越的能量回馈，专为跑步爱好者设计。',
    price: '1299.00',
    discount_price: '999.00',
    stock: 80,
    category_id: 2,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1606107557195-0e29a4b5b4aa?w=400',
      'https://images.unsplash.com/photo-1595950653106-6c9ebd614d3a?w=400',
    ],
    rating: 4.7,
    sales_count: 1890,
    created_at: '2024-01-18T10:00:00Z',
    updated_at: '2024-01-18T10:00:00Z',
  },
  {
    id: 5,
    name: '北欧简约实木餐桌',
    description: '精选优质橡木制作，简约现代设计，适合4-6人使用，环保健康。',
    price: '2999.00',
    discount_price: '2499.00',
    stock: 25,
    category_id: 3,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=400',
      'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?w=400',
    ],
    rating: 4.5,
    sales_count: 456,
    created_at: '2024-01-12T10:00:00Z',
    updated_at: '2024-01-12T10:00:00Z',
  },
  {
    id: 6,
    name: '智能扫地机器人',
    description: '激光导航，智能规划清扫路径，支持APP远程控制，自动回充。',
    price: '1899.00',
    discount_price: '1599.00',
    stock: 60,
    category_id: 3,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=400',
      'https://images.unsplash.com/photo-1574269909862-7e1d70bb8078?w=400',
    ],
    rating: 4.4,
    sales_count: 1123,
    created_at: '2024-01-25T10:00:00Z',
    updated_at: '2024-01-25T10:00:00Z',
  },
  {
    id: 7,
    name: '《深度学习》- Ian Goodfellow',
    description: '深度学习领域的经典教材，适合机器学习研究者和工程师阅读。',
    price: '128.00',
    discount_price: '99.00',
    stock: 200,
    category_id: 4,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c?w=400',
      'https://images.unsplash.com/photo-1481627834876-b7833e8f5570?w=400',
    ],
    rating: 4.8,
    sales_count: 567,
    created_at: '2024-01-08T10:00:00Z',
    updated_at: '2024-01-08T10:00:00Z',
  },
  {
    id: 8,
    name: '户外登山背包 50L',
    description: '大容量设计，多功能分隔，防水透气，适合长途徒步和登山活动。',
    price: '599.00',
    discount_price: '459.00',
    stock: 90,
    category_id: 5,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400',
      'https://images.unsplash.com/photo-1622260614153-03223fb72052?w=400',
    ],
    rating: 4.6,
    sales_count: 789,
    created_at: '2024-01-22T10:00:00Z',
    updated_at: '2024-01-22T10:00:00Z',
  },
  {
    id: 9,
    name: 'Sony WH-1000XM5 无线降噪耳机',
    description: '业界领先的降噪技术，30小时续航，高解析度音质，舒适佩戴。',
    price: '2399.00',
    discount_price: '1999.00',
    stock: 45,
    category_id: 1,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400',
      'https://images.unsplash.com/photo-1484704849700-f032a568e944?w=400',
    ],
    rating: 4.9,
    sales_count: 1456,
    created_at: '2024-01-28T10:00:00Z',
    updated_at: '2024-01-28T10:00:00Z',
  },
  {
    id: 10,
    name: "Levi's 501 经典牛仔裤",
    description: '经典直筒版型，优质丹宁面料，百搭时尚，经久耐穿。',
    price: '699.00',
    discount_price: '549.00',
    stock: 150,
    category_id: 2,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1542272604-787c3835535d?w=400',
      'https://images.unsplash.com/photo-1541099649105-f69ad21f3246?w=400',
    ],
    rating: 4.5,
    sales_count: 2890,
    created_at: '2024-01-05T10:00:00Z',
    updated_at: '2024-01-05T10:00:00Z',
  },
  {
    id: 11,
    name: '小米13 Ultra 摄影旗舰',
    description: '徕卡专业光学镜头，1英寸大底主摄，专业摄影体验。',
    price: '5999.00',
    discount_price: '5499.00',
    stock: 35,
    category_id: 1,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?w=400',
      'https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=400',
    ],
    rating: 4.7,
    sales_count: 678,
    created_at: '2024-01-30T10:00:00Z',
    updated_at: '2024-01-30T10:00:00Z',
  },
  {
    id: 12,
    name: '宜家 MALM 床架',
    description: '简约现代设计，坚固耐用，多种颜色可选，适合各种卧室风格。',
    price: '899.00',
    discount_price: '699.00',
    stock: 40,
    category_id: 3,
    status: 'active',
    images: [
      'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?w=400',
      'https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=400',
    ],
    rating: 4.3,
    sales_count: 1234,
    created_at: '2024-01-14T10:00:00Z',
    updated_at: '2024-01-14T10:00:00Z',
  },
];

// 模拟分类数据
export const mockCategories = [
  { id: 1, name: '电子产品', description: '手机、电脑、数码配件等' },
  { id: 2, name: '服装鞋帽', description: '男装、女装、鞋子、配饰等' },
  { id: 3, name: '家居用品', description: '家具、家电、装饰用品等' },
  { id: 4, name: '图书音像', description: '图书、音乐、影视等' },
  { id: 5, name: '运动户外', description: '运动装备、户外用品等' },
];

// 获取分页商品数据
export const getPagedProducts = (
  page: number = 1,
  pageSize: number = 24,
  filters: {
    keyword?: string;
    categoryId?: number;
    minPrice?: number;
    maxPrice?: number;
    minRating?: number;
    sortBy?: string;
  } = {}
) => {
  let filteredProducts = [...mockProducts];

  // 关键词搜索
  if (filters.keyword) {
    const keyword = filters.keyword.toLowerCase();
    filteredProducts = filteredProducts.filter(
      product =>
        product.name.toLowerCase().includes(keyword) ||
        product.description.toLowerCase().includes(keyword)
    );
  }

  // 分类筛选
  if (filters.categoryId) {
    filteredProducts = filteredProducts.filter(
      product => product.category_id === filters.categoryId
    );
  }

  // 价格筛选
  if (filters.minPrice !== undefined) {
    filteredProducts = filteredProducts.filter(
      product =>
        parseFloat(product.discount_price || product.price) >= filters.minPrice!
    );
  }
  if (filters.maxPrice !== undefined) {
    filteredProducts = filteredProducts.filter(
      product =>
        parseFloat(product.discount_price || product.price) <= filters.maxPrice!
    );
  }

  // 评分筛选
  if (filters.minRating) {
    filteredProducts = filteredProducts.filter(
      product => (product.rating || 0) >= filters.minRating!
    );
  }

  // 排序
  switch (filters.sortBy) {
    case 'price_asc':
      filteredProducts.sort(
        (a, b) =>
          parseFloat(a.discount_price || a.price) -
          parseFloat(b.discount_price || b.price)
      );
      break;
    case 'price_desc':
      filteredProducts.sort(
        (a, b) =>
          parseFloat(b.discount_price || b.price) -
          parseFloat(a.discount_price || a.price)
      );
      break;
    case 'sales_desc':
      filteredProducts.sort(
        (a, b) => (b.sales_count || 0) - (a.sales_count || 0)
      );
      break;
    case 'rating_desc':
      filteredProducts.sort((a, b) => (b.rating || 0) - (a.rating || 0));
      break;
    case 'created_desc':
      filteredProducts.sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      break;
    default:
      // 综合排序：销量 * 0.6 + 评分 * 0.4
      filteredProducts.sort((a, b) => {
        const scoreA = (a.sales_count || 0) * 0.0006 + (a.rating || 0) * 0.4;
        const scoreB = (b.sales_count || 0) * 0.0006 + (b.rating || 0) * 0.4;
        return scoreB - scoreA;
      });
  }

  // 分页
  const total = filteredProducts.length;
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const products = filteredProducts.slice(startIndex, endIndex);

  return {
    products,
    total,
    page,
    pageSize,
    totalPages: Math.ceil(total / pageSize),
  };
};
