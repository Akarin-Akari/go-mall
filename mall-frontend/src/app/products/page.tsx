'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Card,
  Input,
  Select,
  Button,
  Space,
  Spin,
  Empty,
  Breadcrumb,
  Affix,
  Drawer,
  Badge,
  Slider,
  Checkbox,
  Rate,
  Typography,
  Divider,
  message,
} from 'antd';
import {
  SearchOutlined,
  FilterOutlined,
  AppstoreOutlined,
  BarsOutlined,
  SortAscendingOutlined,
  SortDescendingOutlined,
  HomeOutlined,
  ShopOutlined,
  ClearOutlined,
} from '@ant-design/icons';
import { Pagination, Tag } from 'antd';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchProductsAsync, selectProduct } from '@/store/slices/productSlice';
import ProductCard from '@/components/business/ProductCard';
import MainLayout from '@/components/layout/MainLayout';
import { Product, Category } from '@/types';
import { ROUTES, PAGINATION } from '@/constants';
import { debounce } from '@/utils';

const { Search } = Input;
const { Option } = Select;
const { Text, Title } = Typography;

// 排序选项
const SORT_OPTIONS = [
  { label: '综合排序', value: 'default' },
  { label: '价格从低到高', value: 'price_asc' },
  { label: '价格从高到低', value: 'price_desc' },
  { label: '销量从高到低', value: 'sales_desc' },
  { label: '评分从高到低', value: 'rating_desc' },
  { label: '最新发布', value: 'created_desc' },
];

// 每页显示数量选项
const PAGE_SIZE_OPTIONS = [
  { label: '12个/页', value: 12 },
  { label: '24个/页', value: 24 },
  { label: '48个/页', value: 48 },
];

// 价格区间选项
const PRICE_RANGES = [
  { label: '全部价格', value: null },
  { label: '0-50元', value: [0, 50] },
  { label: '50-100元', value: [50, 100] },
  { label: '100-200元', value: [100, 200] },
  { label: '200-500元', value: [200, 500] },
  { label: '500元以上', value: [500, 9999] },
];

const ProductsPage: React.FC = () => {
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [filterVisible, setFilterVisible] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState('default');
  const [pageSize, setPageSize] = useState(24);
  const [selectedCategories, setSelectedCategories] = useState<number[]>([]);
  const [priceRange, setPriceRange] = useState<[number, number] | null>(null);
  const [selectedRating, setSelectedRating] = useState<number | null>(null);
  const [customPriceRange, setCustomPriceRange] = useState<[number, number]>([
    0, 1000,
  ]);

  const router = useRouter();
  const searchParams = useSearchParams();
  const dispatch = useAppDispatch();
  const {
    products,
    total,
    loading,
    searchParams: storeSearchParams,
  } = useAppSelector(selectProduct);

  // 从URL参数初始化状态
  useEffect(() => {
    const keyword = searchParams?.get('keyword') || '';
    const category = searchParams?.get('category');
    const sort = searchParams?.get('sort') || 'default';
    const page = parseInt(searchParams?.get('page') || '1');
    const size = parseInt(searchParams?.get('size') || '24');

    setSearchKeyword(keyword);
    setSortBy(sort);
    setPageSize(size);

    if (category) {
      setSelectedCategories([parseInt(category)]);
    }

    // 加载商品数据
    loadProducts({
      keyword,
      category_id: category ? parseInt(category) : undefined,
      page,
      page_size: size,
      sort_by: sort,
    });
  }, [searchParams]);

  // 加载商品数据
  const loadProducts = useCallback(
    (params: any) => {
      dispatch(
        fetchProductsAsync({
          page: 1,
          page_size: 24,
          ...params,
        })
      );
    },
    [dispatch]
  );

  // 防抖搜索
  const debouncedSearch = useCallback(
    debounce((keyword: string) => {
      updateUrlAndLoad({ keyword, page: 1 });
    }, 500),
    []
  );

  // 更新URL和加载数据
  const updateUrlAndLoad = useCallback(
    (params: any) => {
      const url = new URL(window.location.href);

      // 更新URL参数
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          url.searchParams.set(key, String(value));
        } else {
          url.searchParams.delete(key);
        }
      });

      // 更新浏览器URL
      window.history.pushState({}, '', url.toString());

      // 加载数据
      loadProducts({
        keyword: searchKeyword,
        category_id: selectedCategories[0],
        sort_by: sortBy,
        page_size: pageSize,
        min_price: priceRange?.[0],
        max_price: priceRange?.[1],
        min_rating: selectedRating,
        ...params,
      });
    },
    [
      searchKeyword,
      selectedCategories,
      sortBy,
      pageSize,
      priceRange,
      selectedRating,
      loadProducts,
    ]
  );

  // 处理搜索
  const handleSearch = useCallback(
    (value: string) => {
      setSearchKeyword(value);
      debouncedSearch(value);
    },
    [debouncedSearch]
  );

  // 处理排序变化
  const handleSortChange = useCallback(
    (value: string) => {
      setSortBy(value);
      updateUrlAndLoad({ sort: value, page: 1 });
    },
    [updateUrlAndLoad]
  );

  // 处理分类筛选
  const handleCategoryChange = useCallback(
    (categoryId: number, checked: boolean) => {
      let newCategories: number[];
      if (checked) {
        newCategories = [...selectedCategories, categoryId];
      } else {
        newCategories = selectedCategories.filter(id => id !== categoryId);
      }

      setSelectedCategories(newCategories);
      updateUrlAndLoad({
        category: newCategories[0] || undefined,
        page: 1,
      });
    },
    [selectedCategories, updateUrlAndLoad]
  );

  // 处理价格筛选
  const handlePriceRangeChange = useCallback(
    (range: [number, number] | null) => {
      setPriceRange(range);
      updateUrlAndLoad({
        min_price: range?.[0],
        max_price: range?.[1],
        page: 1,
      });
    },
    [updateUrlAndLoad]
  );

  // 清除所有筛选
  const handleClearFilters = useCallback(() => {
    setSelectedCategories([]);
    setPriceRange(null);
    setSelectedRating(null);
    setSearchKeyword('');
    updateUrlAndLoad({
      keyword: undefined,
      category: undefined,
      min_price: undefined,
      max_price: undefined,
      min_rating: undefined,
      page: 1,
    });
  }, [updateUrlAndLoad]);

  // 处理添加到购物车
  const handleAddToCart = useCallback(async (product: Product) => {
    // 这里应该调用购物车API
    console.log('Add to cart:', product);
    // 模拟异步操作
    await new Promise(resolve => setTimeout(resolve, 500));
  }, []);

  // 处理收藏切换
  const handleToggleFavorite = useCallback(async (product: Product) => {
    // 这里应该调用收藏API
    console.log('Toggle favorite:', product);
    // 模拟异步操作
    await new Promise(resolve => setTimeout(resolve, 300));
  }, []);

  // 渲染筛选面板
  const renderFilterPanel = () => (
    <Card title='商品筛选' size='small'>
      <Space direction='vertical' style={{ width: '100%' }} size='large'>
        {/* 分类筛选 */}
        <div>
          <Title level={5}>商品分类</Title>
          <Space wrap>
            {/* 这里应该从store获取分类数据 */}
            {[
              { id: 1, name: '电子产品' },
              { id: 2, name: '服装鞋帽' },
              { id: 3, name: '家居用品' },
              { id: 4, name: '图书音像' },
              { id: 5, name: '运动户外' },
            ].map(category => (
              <Tag.CheckableTag
                key={category.id}
                checked={selectedCategories.includes(category.id)}
                onChange={checked => handleCategoryChange(category.id, checked)}
              >
                {category.name}
              </Tag.CheckableTag>
            ))}
          </Space>
        </div>

        {/* 价格筛选 */}
        <div>
          <Title level={5}>价格区间</Title>
          <Space direction='vertical' style={{ width: '100%' }}>
            {PRICE_RANGES.map((range, index) => (
              <Checkbox
                key={index}
                checked={
                  range.value === null
                    ? priceRange === null
                    : priceRange?.[0] === range.value[0] &&
                      priceRange?.[1] === range.value[1]
                }
                onChange={e => {
                  if (e.target.checked) {
                    handlePriceRangeChange(
                      range.value as [number, number] | null
                    );
                  }
                }}
              >
                {range.label}
              </Checkbox>
            ))}

            <div style={{ marginTop: 16 }}>
              <Text>自定义价格区间：</Text>
              <Slider
                range
                min={0}
                max={1000}
                value={customPriceRange}
                onChange={value =>
                  setCustomPriceRange(value as [number, number])
                }
                onAfterChange={value =>
                  handlePriceRangeChange(value as [number, number])
                }
                tooltip={{ formatter: value => `¥${value}` }}
              />
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  fontSize: 12,
                  color: '#999',
                }}
              >
                <span>¥{customPriceRange[0]}</span>
                <span>¥{customPriceRange[1]}</span>
              </div>
            </div>
          </Space>
        </div>

        {/* 评分筛选 */}
        <div>
          <Title level={5}>用户评分</Title>
          <Space direction='vertical'>
            {[5, 4, 3, 2, 1].map(rating => (
              <Checkbox
                key={rating}
                checked={selectedRating === rating}
                onChange={e => {
                  setSelectedRating(e.target.checked ? rating : null);
                  updateUrlAndLoad({
                    min_rating: e.target.checked ? rating : undefined,
                    page: 1,
                  });
                }}
              >
                <Rate disabled value={rating} style={{ fontSize: 14 }} />
                <Text style={{ marginLeft: 8 }}>{rating}星及以上</Text>
              </Checkbox>
            ))}
          </Space>
        </div>

        {/* 清除筛选 */}
        <Button block icon={<ClearOutlined />} onClick={handleClearFilters}>
          清除筛选
        </Button>
      </Space>
    </Card>
  );

  return (
    <MainLayout>
      <div style={{ padding: '0 24px' }}>
        {/* 面包屑导航 */}
        <Breadcrumb style={{ margin: '16px 0' }}>
          <Breadcrumb.Item href={ROUTES.HOME}>
            <HomeOutlined />
            <span>首页</span>
          </Breadcrumb.Item>
          <Breadcrumb.Item>
            <ShopOutlined />
            <span>商品列表</span>
          </Breadcrumb.Item>
        </Breadcrumb>

        <Row gutter={[24, 24]}>
          {/* 左侧筛选面板 */}
          <Col xs={0} sm={0} md={6} lg={5} xl={4}>
            <Affix offsetTop={80}>{renderFilterPanel()}</Affix>
          </Col>

          {/* 右侧商品列表 */}
          <Col xs={24} sm={24} md={18} lg={19} xl={20}>
            {/* 搜索和工具栏 */}
            <Card style={{ marginBottom: 16 }}>
              <Row gutter={[16, 16]} align='middle'>
                <Col xs={24} sm={12} md={8}>
                  <Search
                    placeholder='搜索商品名称、品牌、型号...'
                    value={searchKeyword}
                    onChange={e => setSearchKeyword(e.target.value)}
                    onSearch={handleSearch}
                    enterButton={<SearchOutlined />}
                    size='large'
                  />
                </Col>

                <Col xs={24} sm={12} md={16}>
                  <Row gutter={[8, 8]} justify='space-between' align='middle'>
                    <Col>
                      <Space>
                        <Text type='secondary'>
                          共找到 <Text strong>{total}</Text> 件商品
                        </Text>
                      </Space>
                    </Col>

                    <Col>
                      <Space>
                        {/* 移动端筛选按钮 */}
                        <Button
                          icon={<FilterOutlined />}
                          onClick={() => setFilterVisible(true)}
                          className='md:hidden'
                        >
                          筛选
                        </Button>

                        {/* 视图切换 */}
                        <Button.Group>
                          <Button
                            icon={<AppstoreOutlined />}
                            type={viewMode === 'grid' ? 'primary' : 'default'}
                            onClick={() => setViewMode('grid')}
                          />
                          <Button
                            icon={<BarsOutlined />}
                            type={viewMode === 'list' ? 'primary' : 'default'}
                            onClick={() => setViewMode('list')}
                          />
                        </Button.Group>

                        {/* 排序选择 */}
                        <Select
                          value={sortBy}
                          onChange={handleSortChange}
                          style={{ width: 140 }}
                          suffixIcon={<SortAscendingOutlined />}
                        >
                          {SORT_OPTIONS.map(option => (
                            <Option key={option.value} value={option.value}>
                              {option.label}
                            </Option>
                          ))}
                        </Select>

                        {/* 每页显示数量 */}
                        <Select
                          value={pageSize}
                          onChange={value => {
                            setPageSize(value);
                            updateUrlAndLoad({ size: value, page: 1 });
                          }}
                          style={{ width: 100 }}
                        >
                          {PAGE_SIZE_OPTIONS.map(option => (
                            <Option key={option.value} value={option.value}>
                              {option.label}
                            </Option>
                          ))}
                        </Select>
                      </Space>
                    </Col>
                  </Row>
                </Col>
              </Row>
            </Card>

            {/* 商品列表 */}
            <Spin spinning={loading}>
              {products.length > 0 ? (
                <>
                  <Row gutter={[16, 16]}>
                    {products.map(product => (
                      <Col
                        key={product.id}
                        xs={viewMode === 'grid' ? 12 : 24}
                        sm={viewMode === 'grid' ? 8 : 24}
                        md={viewMode === 'grid' ? 8 : 24}
                        lg={viewMode === 'grid' ? 6 : 24}
                        xl={viewMode === 'grid' ? 6 : 24}
                      >
                        <ProductCard
                          product={product}
                          size={viewMode === 'grid' ? 'default' : 'large'}
                          onAddToCart={handleAddToCart}
                          onToggleFavorite={handleToggleFavorite}
                        />
                      </Col>
                    ))}
                  </Row>

                  {/* 分页组件 */}
                  <div
                    style={{
                      display: 'flex',
                      justifyContent: 'center',
                      alignItems: 'center',
                      marginTop: 32,
                      padding: '24px 0',
                    }}
                  >
                    <Pagination
                      current={storeSearchParams.page}
                      total={total}
                      pageSize={storeSearchParams.page_size}
                      showSizeChanger
                      showQuickJumper
                      showTotal={(total, range) =>
                        `第 ${range[0]}-${range[1]} 条，共 ${total} 条商品`
                      }
                      pageSizeOptions={['12', '24', '48', '96']}
                      onChange={(page, size) => {
                        updateUrlAndLoad({ page, size });
                      }}
                      onShowSizeChange={(current, size) => {
                        setPageSize(size);
                        updateUrlAndLoad({ page: 1, size });
                      }}
                      size='default'
                      style={{ textAlign: 'center' }}
                    />
                  </div>
                </>
              ) : (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description='暂无商品'
                  style={{ margin: '60px 0' }}
                >
                  <Button type='primary' onClick={handleClearFilters}>
                    清除筛选条件
                  </Button>
                </Empty>
              )}
            </Spin>
          </Col>
        </Row>

        {/* 移动端筛选抽屉 */}
        <Drawer
          title='商品筛选'
          placement='left'
          onClose={() => setFilterVisible(false)}
          open={filterVisible}
          width={300}
        >
          {renderFilterPanel()}
        </Drawer>
      </div>
    </MainLayout>
  );
};

export default ProductsPage;
