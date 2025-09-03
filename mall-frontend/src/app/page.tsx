'use client';

import React, { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Button,
  Typography,
  Carousel,
  Space,
  Tag,
  Spin,
  message,
  Divider,
  Badge,
  Image
} from 'antd';
import {
  ShoppingCartOutlined,
  HeartOutlined,
  EyeOutlined,
  StarFilled,
  FireOutlined,
  ThunderboltOutlined,
  GiftOutlined,
  RightOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import MainLayout from '@/components/layout/MainLayout';
import ProductCard from '@/components/business/ProductCard';
import CategoryGrid from '@/components/business/CategoryGrid';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchProductsAsync, selectProducts, fetchCategoriesAsync, selectCategories, selectProductLoading } from '@/store/slices/productSlice';
import { addToCartAsync } from '@/store/slices/cartSlice';
import { Product, Category } from '@/types';

const { Title, Text, Paragraph } = Typography;

// 轮播图数据
const bannerData = [
  {
    id: 1,
    title: '新年大促销',
    subtitle: '全场商品5折起，限时抢购',
    image: 'https://images.unsplash.com/photo-1607082348824-0a96f2a4b9da?w=1200&h=400&fit=crop',
    link: '/products?sale=true',
    color: '#ff4d4f'
  },
  {
    id: 2,
    title: '电子产品专场',
    subtitle: 'iPhone、MacBook等热门产品',
    image: 'https://images.unsplash.com/photo-1468495244123-6c6c332eeece?w=1200&h=400&fit=crop',
    link: '/categories/1',
    color: '#1890ff'
  },
  {
    id: 3,
    title: '时尚服饰',
    subtitle: '春季新款上市，潮流穿搭',
    image: 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1200&h=400&fit=crop',
    link: '/categories/2',
    color: '#52c41a'
  }
];

export default function Home() {
  const router = useRouter();
  const dispatch = useAppDispatch();

  const products = useAppSelector(selectProducts) || [];
  const categories = useAppSelector(selectCategories) || [];
  const productsLoading = useAppSelector(selectProductLoading);
  const categoriesLoading = useAppSelector((state) => state.product.categoriesLoading);

  const [featuredProducts, setFeaturedProducts] = useState<Product[]>([]);
  const [hotProducts, setHotProducts] = useState<Product[]>([]);
  const [newProducts, setNewProducts] = useState<Product[]>([]);

  // 获取数据
  useEffect(() => {
    // 获取商品列表
    dispatch(fetchProductsAsync({ page: 1, page_size: 20 }));
    // 获取分类列表
    dispatch(fetchCategoriesAsync());
  }, [dispatch]);

  // 处理商品数据
  useEffect(() => {
    if (products && Array.isArray(products) && products.length > 0) {
      // 模拟不同类型的商品分类
      setFeaturedProducts(products.slice(0, 8)); // 精选商品
      setHotProducts(products.slice(8, 16)); // 热销商品
      setNewProducts(products.slice(0, 4)); // 新品推荐
    } else {
      // 如果没有商品数据，设置为空数组
      setFeaturedProducts([]);
      setHotProducts([]);
      setNewProducts([]);
    }
  }, [products]);

  // 添加到购物车
  const handleAddToCart = async (product: Product) => {
    try {
      await dispatch(addToCartAsync({
        product_id: product.id,
        quantity: 1
      }));
      message.success(`${product.name} 已添加到购物车`);
    } catch (error) {
      message.error('添加购物车失败，请重试');
    }
  };

  // 查看商品详情
  const handleViewProduct = (productId: number) => {
    router.push(`/products/${productId}`);
  };

  // 查看分类
  const handleViewCategory = (categoryId: number) => {
    router.push(`/categories/${categoryId}`);
  };

  return (
    <MainLayout>
      <div style={{ background: '#f5f5f5', minHeight: '100vh' }}>
        {/* 轮播图 */}
        <div style={{ marginBottom: 24 }}>
          <Carousel autoplay effect="fade" style={{ borderRadius: 8, overflow: 'hidden' }}>
            {bannerData.map(banner => (
              <div key={banner.id}>
                <div
                  style={{
                    height: 400,
                    background: `linear-gradient(45deg, ${banner.color}20, ${banner.color}40), url(${banner.image})`,
                    backgroundSize: 'cover',
                    backgroundPosition: 'center',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    position: 'relative',
                    cursor: 'pointer'
                  }}
                  onClick={() => router.push(banner.link)}
                >
                  <div style={{ textAlign: 'center', color: 'white', zIndex: 1 }}>
                    <Title level={1} style={{ color: 'white', marginBottom: 16, fontSize: 48 }}>
                      {banner.title}
                    </Title>
                    <Paragraph style={{ color: 'white', fontSize: 18, marginBottom: 24 }}>
                      {banner.subtitle}
                    </Paragraph>
                    <Button
                      type="primary"
                      size="large"
                      style={{
                        background: banner.color,
                        borderColor: banner.color,
                        height: 48,
                        fontSize: 16,
                        paddingLeft: 32,
                        paddingRight: 32
                      }}
                    >
                      立即抢购 <RightOutlined />
                    </Button>
                  </div>
                  <div
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      background: 'rgba(0,0,0,0.3)',
                      zIndex: 0
                    }}
                  />
                </div>
              </div>
            ))}
          </Carousel>
        </div>

        {/* 分类导航 */}
        <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: '24px 24px 16px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <Title level={3} style={{ margin: 0 }}>
              <GiftOutlined style={{ marginRight: 8, color: '#1890ff' }} />
              商品分类
            </Title>
            <Button type="link" onClick={() => router.push('/categories')}>
              查看全部 <RightOutlined />
            </Button>
          </div>

          <Spin spinning={categoriesLoading}>
            <CategoryGrid
              categories={categories && Array.isArray(categories) ? categories.slice(0, 6) : []}
              onCategoryClick={handleViewCategory}
            />
          </Spin>
        </Card>

        {/* 新品推荐 */}
        <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: '24px 24px 16px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <Title level={3} style={{ margin: 0 }}>
              <ThunderboltOutlined style={{ marginRight: 8, color: '#52c41a' }} />
              新品推荐
            </Title>
            <Button type="link" onClick={() => router.push('/products?sort=newest')}>
              查看更多 <RightOutlined />
            </Button>
          </div>

          <Spin spinning={productsLoading}>
            <Row gutter={[16, 16]}>
              {newProducts.map(product => (
                <Col xs={24} sm={12} md={6} key={product.id}>
                  <ProductCard
                    product={product}
                    onAddToCart={() => handleAddToCart(product)}
                    onViewDetail={() => handleViewProduct(product.id)}
                    showBadge="新品"
                    badgeColor="#52c41a"
                  />
                </Col>
              ))}
            </Row>
          </Spin>
        </Card>

        {/* 热销商品 */}
        <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: '24px 24px 16px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <Title level={3} style={{ margin: 0 }}>
              <FireOutlined style={{ marginRight: 8, color: '#ff4d4f' }} />
              热销商品
            </Title>
            <Button type="link" onClick={() => router.push('/products?sort=popular')}>
              查看更多 <RightOutlined />
            </Button>
          </div>

          <Spin spinning={productsLoading}>
            <Row gutter={[16, 16]}>
              {hotProducts.map(product => (
                <Col xs={24} sm={12} md={6} key={product.id}>
                  <ProductCard
                    product={product}
                    onAddToCart={() => handleAddToCart(product)}
                    onViewDetail={() => handleViewProduct(product.id)}
                    showBadge="热销"
                    badgeColor="#ff4d4f"
                  />
                </Col>
              ))}
            </Row>
          </Spin>
        </Card>

        {/* 精选商品 */}
        <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: '24px 24px 16px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <Title level={3} style={{ margin: 0 }}>
              <StarFilled style={{ marginRight: 8, color: '#faad14' }} />
              精选商品
            </Title>
            <Button type="link" onClick={() => router.push('/products')}>
              查看全部 <RightOutlined />
            </Button>
          </div>

          <Spin spinning={productsLoading}>
            <Row gutter={[16, 16]}>
              {featuredProducts.map(product => (
                <Col xs={24} sm={12} md={6} key={product.id}>
                  <ProductCard
                    product={product}
                    onAddToCart={() => handleAddToCart(product)}
                    onViewDetail={() => handleViewProduct(product.id)}
                    showBadge="精选"
                    badgeColor="#faad14"
                  />
                </Col>
              ))}
            </Row>
          </Spin>
        </Card>

        {/* 品牌推荐 */}
        <Card style={{ marginBottom: 24 }}>
          <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>
            品牌推荐
          </Title>
          <Row gutter={[16, 16]} justify="center">
            {[
              { name: 'Apple', logo: '🍎', color: '#000000' },
              { name: 'Nike', logo: '✓', color: '#ff6900' },
              { name: 'Adidas', logo: '⚡', color: '#000000' },
              { name: 'Samsung', logo: '📱', color: '#1428a0' },
              { name: 'Sony', logo: '🎵', color: '#000000' },
              { name: 'LG', logo: '📺', color: '#a50034' },
            ].map((brand, index) => (
              <Col xs={12} sm={8} md={4} key={index}>
                <Card
                  hoverable
                  style={{ textAlign: 'center', border: '1px solid #f0f0f0' }}
                  bodyStyle={{ padding: '16px 8px' }}
                >
                  <div style={{ fontSize: 32, marginBottom: 8 }}>{brand.logo}</div>
                  <Text strong style={{ color: brand.color }}>{brand.name}</Text>
                </Card>
              </Col>
            ))}
          </Row>
        </Card>

        {/* 服务保障 */}
        <Card>
          <Row gutter={[16, 16]} justify="center">
            {[
              { icon: '🚚', title: '免费配送', desc: '满99元免费配送' },
              { icon: '🔒', title: '安全支付', desc: '多种支付方式' },
              { icon: '↩️', title: '7天退换', desc: '7天无理由退换货' },
              { icon: '🎧', title: '24小时客服', desc: '专业客服在线服务' },
            ].map((service, index) => (
              <Col xs={12} sm={6} key={index}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: 32, marginBottom: 8 }}>{service.icon}</div>
                  <Title level={5} style={{ marginBottom: 4 }}>{service.title}</Title>
                  <Text type="secondary">{service.desc}</Text>
                </div>
              </Col>
            ))}
          </Row>
        </Card>
      </div>
    </MainLayout>
  );
}
