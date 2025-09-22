'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Card,
  Typography,
  Button,
  Space,
  Rate,
  Tag,
  Divider,
  Breadcrumb,
  Spin,
  message,
  Badge,
  Tooltip,
  InputNumber,
  Radio,
  Tabs,
  Avatar,
  Progress,
  Empty,
} from 'antd';
import {
  ArrowLeftOutlined,
  ShoppingCartOutlined,
  HeartOutlined,
  HeartFilled,
  ShareAltOutlined,
  StarFilled,
  CheckCircleOutlined,
  TruckOutlined,
  SafetyCertificateOutlined,
  UserOutlined,
  CustomerServiceOutlined,
  HomeOutlined,
  ShopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useRouter, useParams } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  fetchProductDetailAsync,
  selectProduct,
} from '@/store/slices/productSlice';
import MainLayout from '@/components/layout/MainLayout';
import ProductImageGallery from '@/components/business/ProductImageGallery';
import { Product } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice, formatNumber } from '@/utils';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

interface ProductSpec {
  name: string;
  options: string[];
}

const ProductDetailPage: React.FC = () => {
  const [selectedSpecs, setSelectedSpecs] = useState<Record<string, string>>(
    {}
  );
  const [quantity, setQuantity] = useState(1);
  const [isFavorited, setIsFavorited] = useState(false);
  const [addingToCart, setAddingToCart] = useState(false);
  const [buyingNow, setBuyingNow] = useState(false);

  const router = useRouter();
  const params = useParams();
  const dispatch = useAppDispatch();
  const { currentProduct, productLoading } = useAppSelector(selectProduct);

  const productId = parseInt(params?.id as string);

  // 加载商品详情
  useEffect(() => {
    if (productId) {
      dispatch(fetchProductDetailAsync(productId));
    }
  }, [dispatch, productId]);

  // 模拟商品规格数据
  const productSpecs: ProductSpec[] = [
    {
      name: '颜色',
      options: ['深空黑', '银色', '金色', '深紫色'],
    },
    {
      name: '容量',
      options: ['128GB', '256GB', '512GB', '1TB'],
    },
    {
      name: '尺寸',
      options: ['6.1英寸', '6.7英寸'],
    },
  ];

  // 处理规格选择
  const handleSpecChange = useCallback((specName: string, value: string) => {
    setSelectedSpecs(prev => ({
      ...prev,
      [specName]: value,
    }));
  }, []);

  // 处理数量变化
  const handleQuantityChange = useCallback((value: number | null) => {
    if (value && value > 0) {
      setQuantity(value);
    }
  }, []);

  // 处理收藏切换
  const handleToggleFavorite = useCallback(async () => {
    try {
      setIsFavorited(!isFavorited);
      message.success(isFavorited ? '已取消收藏' : '已添加到收藏');
    } catch (error) {
      setIsFavorited(isFavorited); // 回滚状态
      message.error('操作失败，请重试');
    }
  }, [isFavorited]);

  // 处理分享
  const handleShare = useCallback(() => {
    if (navigator.share && currentProduct) {
      navigator.share({
        title: currentProduct.name,
        text: currentProduct.description,
        url: window.location.href,
      });
    } else {
      // 复制链接到剪贴板
      navigator.clipboard.writeText(window.location.href);
      message.success('链接已复制到剪贴板');
    }
  }, [currentProduct]);

  // 处理添加到购物车
  const handleAddToCart = useCallback(async () => {
    if (!currentProduct) return;

    try {
      setAddingToCart(true);

      // 检查规格选择
      const requiredSpecs = productSpecs.filter(
        spec => spec.options.length > 1
      );
      const missingSpecs = requiredSpecs.filter(
        spec => !selectedSpecs[spec.name]
      );

      if (missingSpecs.length > 0) {
        message.warning(
          `请选择${missingSpecs.map(spec => spec.name).join('、')}`
        );
        return;
      }

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      message.success('已添加到购物车');
    } catch (error) {
      message.error('添加失败，请重试');
    } finally {
      setAddingToCart(false);
    }
  }, [currentProduct, selectedSpecs, productSpecs, quantity]);

  // 处理立即购买
  const handleBuyNow = useCallback(async () => {
    if (!currentProduct) return;

    try {
      setBuyingNow(true);

      // 检查规格选择
      const requiredSpecs = productSpecs.filter(
        spec => spec.options.length > 1
      );
      const missingSpecs = requiredSpecs.filter(
        spec => !selectedSpecs[spec.name]
      );

      if (missingSpecs.length > 0) {
        message.warning(
          `请选择${missingSpecs.map(spec => spec.name).join('、')}`
        );
        return;
      }

      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 1000));

      // 跳转到结算页面
      router.push(ROUTES.CHECKOUT);
    } catch (error) {
      message.error('购买失败，请重试');
    } finally {
      setBuyingNow(false);
    }
  }, [currentProduct, selectedSpecs, productSpecs, quantity, router]);

  // 返回上一页
  const handleGoBack = useCallback(() => {
    router.back();
  }, [router]);

  if (productLoading) {
    return (
      <MainLayout>
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '60vh',
          }}
        >
          <Spin size='large' />
        </div>
      </MainLayout>
    );
  }

  if (!currentProduct) {
    return (
      <MainLayout>
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '60vh',
          }}
        >
          <Empty description='商品不存在' />
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div style={{ padding: '0 24px', maxWidth: 1200, margin: '0 auto' }}>
        {/* 面包屑导航 */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            margin: '16px 0',
          }}
        >
          <Breadcrumb>
            <Breadcrumb.Item href={ROUTES.HOME}>
              <HomeOutlined />
              <span>首页</span>
            </Breadcrumb.Item>
            <Breadcrumb.Item href={ROUTES.PRODUCTS}>
              <ShopOutlined />
              <span>商品列表</span>
            </Breadcrumb.Item>
            <Breadcrumb.Item>
              {currentProduct.category_name || '商品详情'}
            </Breadcrumb.Item>
            <Breadcrumb.Item>{currentProduct.name}</Breadcrumb.Item>
          </Breadcrumb>

          <Button
            icon={<ArrowLeftOutlined />}
            onClick={handleGoBack}
            type='text'
          >
            返回
          </Button>
        </div>

        {/* 主要内容区域 */}
        <Row gutter={[32, 32]}>
          {/* 左侧：商品图片 */}
          <Col xs={24} md={12} lg={10}>
            <ProductImageGallery
              images={
                currentProduct.images || ['/images/product-placeholder.svg']
              }
              productName={currentProduct.name}
            />
          </Col>

          {/* 右侧：商品信息 */}
          <Col xs={24} md={12} lg={14}>
            <div style={{ position: 'sticky', top: 80 }}>
              {/* 商品标题和标签 */}
              <div style={{ marginBottom: 16 }}>
                <Space wrap style={{ marginBottom: 8 }}>
                  {currentProduct.discount_price && (
                    <Tag color='orange'>限时优惠</Tag>
                  )}
                </Space>

                <Title level={2} style={{ margin: 0, lineHeight: 1.3 }}>
                  {currentProduct.name}
                </Title>

                <Paragraph
                  type='secondary'
                  style={{ fontSize: 16, margin: '8px 0 16px' }}
                >
                  {currentProduct.description}
                </Paragraph>
              </div>

              {/* 评分和销量 */}
              <div style={{ marginBottom: 24 }}>
                <Space split={<Divider type='vertical' />}>
                  <Space>
                    <Rate
                      disabled
                      allowHalf
                      value={currentProduct.rating || 0}
                      style={{ fontSize: 16 }}
                    />
                    <Text strong>
                      {currentProduct.rating?.toFixed(1) || '暂无评分'}
                    </Text>
                  </Space>
                  <Text type='secondary'>
                    已售 {formatNumber(currentProduct.sales_count || 0)}
                  </Text>
                  <Text type='secondary'>库存 {currentProduct.stock}</Text>
                </Space>
              </div>

              {/* 价格信息 */}
              <Card style={{ marginBottom: 24, backgroundColor: '#fafafa' }}>
                <Space
                  direction='vertical'
                  size='small'
                  style={{ width: '100%' }}
                >
                  <div
                    style={{ display: 'flex', alignItems: 'baseline', gap: 16 }}
                  >
                    <Text
                      style={{
                        fontSize: 32,
                        color: '#ff4d4f',
                        fontWeight: 'bold',
                      }}
                    >
                      ¥
                      {formatPrice(
                        currentProduct.discount_price || currentProduct.price
                      )}
                    </Text>
                    {currentProduct.discount_price &&
                      parseFloat(currentProduct.discount_price) <
                        parseFloat(currentProduct.price) && (
                        <Text delete type='secondary' style={{ fontSize: 18 }}>
                          ¥{formatPrice(currentProduct.price)}
                        </Text>
                      )}
                  </div>

                  {currentProduct.discount_price &&
                    parseFloat(currentProduct.discount_price) <
                      parseFloat(currentProduct.price) && (
                      <div>
                        <Tag color='red' style={{ fontSize: 14 }}>
                          立省 ¥
                          {formatPrice(
                            (
                              parseFloat(currentProduct.price) -
                              parseFloat(currentProduct.discount_price)
                            ).toString()
                          )}
                        </Tag>
                        <Text type='secondary' style={{ marginLeft: 8 }}>
                          折扣{' '}
                          {(
                            (parseFloat(currentProduct.discount_price) /
                              parseFloat(currentProduct.price)) *
                            100
                          ).toFixed(0)}
                          %
                        </Text>
                      </div>
                    )}
                </Space>
              </Card>

              {/* 商品规格选择 */}
              <div style={{ marginBottom: 24 }}>
                {productSpecs.map(spec => (
                  <div key={spec.name} style={{ marginBottom: 16 }}>
                    <Text strong style={{ display: 'block', marginBottom: 8 }}>
                      {spec.name}：
                      {selectedSpecs[spec.name] && (
                        <Text type='secondary' style={{ fontWeight: 'normal' }}>
                          {selectedSpecs[spec.name]}
                        </Text>
                      )}
                    </Text>
                    <Radio.Group
                      value={selectedSpecs[spec.name]}
                      onChange={e =>
                        handleSpecChange(spec.name, e.target.value)
                      }
                    >
                      <Space wrap>
                        {spec.options.map(option => (
                          <Radio.Button
                            key={option}
                            value={option}
                            style={{ marginBottom: 8 }}
                          >
                            {option}
                          </Radio.Button>
                        ))}
                      </Space>
                    </Radio.Group>
                  </div>
                ))}
              </div>

              {/* 数量选择 */}
              <div style={{ marginBottom: 24 }}>
                <Text strong style={{ display: 'block', marginBottom: 8 }}>
                  数量：
                </Text>
                <InputNumber
                  min={1}
                  max={currentProduct.stock}
                  value={quantity}
                  onChange={handleQuantityChange}
                  size='large'
                  style={{ width: 120 }}
                />
                <Text type='secondary' style={{ marginLeft: 16 }}>
                  库存 {currentProduct.stock} 件
                </Text>
              </div>

              {/* 操作按钮 */}
              <div style={{ marginBottom: 24 }}>
                <Space
                  direction='vertical'
                  style={{ width: '100%' }}
                  size='middle'
                >
                  <Row gutter={[12, 12]}>
                    <Col span={12}>
                      <Button
                        type='primary'
                        size='large'
                        block
                        icon={<ShoppingCartOutlined />}
                        loading={addingToCart}
                        onClick={handleAddToCart}
                        disabled={currentProduct.stock === 0}
                      >
                        加入购物车
                      </Button>
                    </Col>
                    <Col span={12}>
                      <Button
                        type='primary'
                        size='large'
                        block
                        icon={<ThunderboltOutlined />}
                        loading={buyingNow}
                        onClick={handleBuyNow}
                        disabled={currentProduct.stock === 0}
                        style={{
                          background:
                            'linear-gradient(135deg, #ff6b6b, #ff8e8e)',
                          borderColor: '#ff6b6b',
                        }}
                      >
                        立即购买
                      </Button>
                    </Col>
                  </Row>

                  <Row gutter={[12, 12]}>
                    <Col span={12}>
                      <Button
                        size='large'
                        block
                        icon={
                          isFavorited ? (
                            <HeartFilled style={{ color: '#ff4d4f' }} />
                          ) : (
                            <HeartOutlined />
                          )
                        }
                        onClick={handleToggleFavorite}
                      >
                        {isFavorited ? '已收藏' : '收藏'}
                      </Button>
                    </Col>
                    <Col span={12}>
                      <Button
                        size='large'
                        block
                        icon={<ShareAltOutlined />}
                        onClick={handleShare}
                      >
                        分享
                      </Button>
                    </Col>
                  </Row>
                </Space>
              </div>

              {/* 服务保障 */}
              <Card size='small' style={{ backgroundColor: '#f9f9f9' }}>
                <Space direction='vertical' style={{ width: '100%' }}>
                  <Text strong>服务保障</Text>
                  <Space wrap>
                    <Tag icon={<TruckOutlined />} color='blue'>
                      免费配送
                    </Tag>
                    <Tag icon={<SafetyCertificateOutlined />} color='green'>
                      正品保证
                    </Tag>
                    <Tag icon={<CustomerServiceOutlined />} color='orange'>
                      7天退换
                    </Tag>
                    <Tag icon={<CheckCircleOutlined />} color='purple'>
                      售后保修
                    </Tag>
                  </Space>
                </Space>
              </Card>
            </div>
          </Col>
        </Row>

        {/* 商品详情标签页 */}
        <Card style={{ marginTop: 32 }}>
          <Tabs defaultActiveKey='details' size='large'>
            <TabPane tab='商品详情' key='details'>
              <div style={{ padding: '24px 0' }}>
                <Title level={4}>商品介绍</Title>
                <Paragraph style={{ fontSize: 16, lineHeight: 1.8 }}>
                  {currentProduct.description}
                </Paragraph>

                <Divider />

                <Title level={4}>产品特色</Title>
                <ul style={{ fontSize: 16, lineHeight: 1.8 }}>
                  <li>高品质材料，精工制作</li>
                  <li>人性化设计，使用便捷</li>
                  <li>严格质检，品质保证</li>
                  <li>完善售后，无忧购买</li>
                </ul>

                <Divider />

                <Title level={4}>规格参数</Title>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Text strong>商品名称：</Text>
                    <Text>{currentProduct.name}</Text>
                  </Col>
                  <Col span={12}>
                    <Text strong>商品编号：</Text>
                    <Text>
                      #{currentProduct.id.toString().padStart(8, '0')}
                    </Text>
                  </Col>
                  <Col span={12}>
                    <Text strong>商品分类：</Text>
                    <Text>{currentProduct.category_name || '未分类'}</Text>
                  </Col>
                  <Col span={12}>
                    <Text strong>商品状态：</Text>
                    <Tag
                      color={
                        currentProduct.status === 'active' ? 'green' : 'red'
                      }
                    >
                      {currentProduct.status === 'active' ? '在售' : '下架'}
                    </Tag>
                  </Col>
                </Row>
              </div>
            </TabPane>

            <TabPane tab='用户评价' key='reviews'>
              <div style={{ padding: '24px 0' }}>
                {/* 评价统计 */}
                <Row gutter={[32, 32]} style={{ marginBottom: 32 }}>
                  <Col xs={24} md={8}>
                    <Card
                      style={{
                        textAlign: 'center',
                        backgroundColor: '#fafafa',
                      }}
                    >
                      <div
                        style={{
                          fontSize: 48,
                          color: '#ff4d4f',
                          fontWeight: 'bold',
                        }}
                      >
                        {currentProduct.rating?.toFixed(1) || '0.0'}
                      </div>
                      <Rate
                        disabled
                        allowHalf
                        value={currentProduct.rating || 0}
                        style={{ fontSize: 20, margin: '8px 0' }}
                      />
                      <div style={{ color: '#666' }}>
                        基于 {formatNumber(currentProduct.sales_count || 0)}{' '}
                        条评价
                      </div>
                    </Card>
                  </Col>
                  <Col xs={24} md={16}>
                    <Space
                      direction='vertical'
                      style={{ width: '100%' }}
                      size='middle'
                    >
                      {[5, 4, 3, 2, 1].map(star => (
                        <div
                          key={star}
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 16,
                          }}
                        >
                          <div style={{ width: 60 }}>
                            <Rate
                              disabled
                              value={star}
                              style={{ fontSize: 14 }}
                            />
                          </div>
                          <Progress
                            percent={
                              star === 5
                                ? 80
                                : star === 4
                                  ? 15
                                  : star === 3
                                    ? 3
                                    : star === 2
                                      ? 1
                                      : 1
                            }
                            strokeColor='#1890ff'
                            style={{ flex: 1 }}
                          />
                          <Text style={{ width: 40, textAlign: 'right' }}>
                            {star === 5
                              ? '80%'
                              : star === 4
                                ? '15%'
                                : star === 3
                                  ? '3%'
                                  : star === 2
                                    ? '1%'
                                    : '1%'}
                          </Text>
                        </div>
                      ))}
                    </Space>
                  </Col>
                </Row>

                {/* 评价列表 */}
                <div>
                  <Title level={4}>用户评价</Title>
                  <Space
                    direction='vertical'
                    style={{ width: '100%' }}
                    size='large'
                  >
                    {/* 模拟评价数据 */}
                    {[
                      {
                        id: 1,
                        user: { name: '张***', avatar: null },
                        rating: 5,
                        content:
                          '商品质量很好，物流也很快，非常满意的一次购物体验！',
                        images: [],
                        date: '2024-01-20',
                        specs: '颜色：深空黑 容量：256GB',
                      },
                      {
                        id: 2,
                        user: { name: '李***', avatar: null },
                        rating: 4,
                        content:
                          '整体不错，性价比很高，推荐购买。包装也很精美。',
                        images: [],
                        date: '2024-01-18',
                        specs: '颜色：银色 容量：128GB',
                      },
                      {
                        id: 3,
                        user: { name: '王***', avatar: null },
                        rating: 5,
                        content: '超出预期的好，用料扎实，做工精细，值得推荐！',
                        images: [],
                        date: '2024-01-15',
                        specs: '颜色：金色 容量：512GB',
                      },
                    ].map(review => (
                      <Card key={review.id} size='small'>
                        <div style={{ display: 'flex', gap: 16 }}>
                          <Avatar icon={<UserOutlined />} size='large' />
                          <div style={{ flex: 1 }}>
                            <div
                              style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                                alignItems: 'center',
                                marginBottom: 8,
                              }}
                            >
                              <div>
                                <Text strong>{review.user.name}</Text>
                                <Text
                                  type='secondary'
                                  style={{ marginLeft: 16 }}
                                >
                                  {review.date}
                                </Text>
                              </div>
                              <Rate
                                disabled
                                value={review.rating}
                                style={{ fontSize: 14 }}
                              />
                            </div>
                            <Paragraph style={{ margin: '8px 0' }}>
                              {review.content}
                            </Paragraph>
                            <Text type='secondary' style={{ fontSize: 12 }}>
                              {review.specs}
                            </Text>
                          </div>
                        </div>
                      </Card>
                    ))}
                  </Space>
                </div>
              </div>
            </TabPane>

            <TabPane tab='售后保障' key='service'>
              <div style={{ padding: '24px 0' }}>
                <Row gutter={[32, 32]}>
                  <Col xs={24} md={12}>
                    <Card>
                      <Space direction='vertical' style={{ width: '100%' }}>
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                          }}
                        >
                          <TruckOutlined
                            style={{ fontSize: 24, color: '#1890ff' }}
                          />
                          <div>
                            <Title level={5} style={{ margin: 0 }}>
                              免费配送
                            </Title>
                            <Text type='secondary'>全国包邮，48小时内发货</Text>
                          </div>
                        </div>
                        <Divider />
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                          }}
                        >
                          <SafetyCertificateOutlined
                            style={{ fontSize: 24, color: '#52c41a' }}
                          />
                          <div>
                            <Title level={5} style={{ margin: 0 }}>
                              正品保证
                            </Title>
                            <Text type='secondary'>官方授权，假一赔十</Text>
                          </div>
                        </div>
                        <Divider />
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                          }}
                        >
                          <CustomerServiceOutlined
                            style={{ fontSize: 24, color: '#fa8c16' }}
                          />
                          <div>
                            <Title level={5} style={{ margin: 0 }}>
                              7天退换
                            </Title>
                            <Text type='secondary'>7天无理由退换货</Text>
                          </div>
                        </div>
                      </Space>
                    </Card>
                  </Col>
                  <Col xs={24} md={12}>
                    <Card>
                      <Space direction='vertical' style={{ width: '100%' }}>
                        <Title level={5}>售后服务</Title>
                        <ul style={{ paddingLeft: 20 }}>
                          <li>全国联保，享受三包服务</li>
                          <li>专业客服团队，7×24小时在线</li>
                          <li>免费上门安装调试服务</li>
                          <li>定期回访，跟踪使用情况</li>
                        </ul>
                        <Divider />
                        <Title level={5}>联系方式</Title>
                        <Space direction='vertical'>
                          <Text>客服热线：400-888-8888</Text>
                          <Text>在线客服：9:00-22:00</Text>
                          <Text>邮箱：service@mall.com</Text>
                        </Space>
                      </Space>
                    </Card>
                  </Col>
                </Row>
              </div>
            </TabPane>
          </Tabs>
        </Card>
      </div>
    </MainLayout>
  );
};

export default ProductDetailPage;
