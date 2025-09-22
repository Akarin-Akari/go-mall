'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Card,
  Typography,
  Button,
  Space,
  Checkbox,
  InputNumber,
  Image,
  Divider,
  Empty,
  Spin,
  message,
  Popconfirm,
  Tag,
  Tooltip,
  Badge,
  Alert,
} from 'antd';
import {
  DeleteOutlined,
  ShoppingCartOutlined,
  HeartOutlined,
  GiftOutlined,
  SafetyCertificateOutlined,
  TruckOutlined,
  MinusOutlined,
  PlusOutlined,
  ClearOutlined,
  ShopOutlined,
  HomeOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  fetchCartAsync,
  updateCartItemAsync,
  removeCartItemAsync,
  clearCartAsync,
  toggleSelectItem,
  toggleSelectAll,
  selectCart,
} from '@/store/slices/cartSlice';
import MainLayout from '@/components/layout/MainLayout';
import { ROUTES } from '@/constants';
import { formatPrice } from '@/utils';
import { CartItem } from '@/types';

const { Title, Text, Paragraph } = Typography;

const CartPage: React.FC = () => {
  const [loadingItems, setLoadingItems] = useState<Record<number, boolean>>({});

  const router = useRouter();
  const dispatch = useAppDispatch();
  const { items, loading } = useAppSelector(selectCart);

  // 加载购物车数据
  useEffect(() => {
    dispatch(fetchCartAsync());
  }, [dispatch]);

  // 计算选中商品的统计信息
  const selectedStats = React.useMemo(() => {
    const selected = items.filter(item => item.selected);
    const quantity = selected.reduce((sum, item) => sum + item.quantity, 0);
    const amount = selected.reduce(
      (sum, item) => sum + parseFloat(item.price) * item.quantity,
      0
    );

    return {
      count: selected.length,
      quantity,
      amount: amount.toFixed(2),
    };
  }, [items]);

  // 处理数量变化
  const handleQuantityChange = useCallback(
    async (itemId: number, newQuantity: number) => {
      if (newQuantity < 1) return;

      try {
        setLoadingItems(prev => ({ ...prev, [itemId]: true }));
        await dispatch(
          updateCartItemAsync({ id: itemId, quantity: newQuantity })
        );
      } catch {
        message.error('更新数量失败');
      } finally {
        setLoadingItems(prev => ({ ...prev, [itemId]: false }));
      }
    },
    [dispatch]
  );

  // 处理商品选择
  const handleItemSelect = useCallback(
    (itemId: number, checked: boolean) => {
      dispatch(toggleSelectItem({ item_id: itemId, selected: checked }));
    },
    [dispatch]
  );

  // 处理全选
  const handleSelectAll = useCallback(
    (checked: boolean) => {
      dispatch(toggleSelectAll(checked));
    },
    [dispatch]
  );

  // 处理删除商品
  const handleRemoveItem = useCallback(
    async (itemId: number) => {
      try {
        await dispatch(removeCartItemAsync(itemId));
        message.success('商品已删除');
      } catch {
        message.error('删除失败');
      }
    },
    [dispatch]
  );

  // 处理批量删除
  const handleBatchRemove = useCallback(async () => {
    const selectedItemIds = items
      .filter(item => item.selected)
      .map(item => item.id);

    if (selectedItemIds.length === 0) {
      message.warning('请选择要删除的商品');
      return;
    }

    try {
      for (const itemId of selectedItemIds) {
        await dispatch(removeCartItemAsync(itemId));
      }
      message.success(`已删除 ${selectedItemIds.length} 件商品`);
    } catch {
      message.error('批量删除失败');
    }
  }, [items, dispatch]);

  // 处理清空购物车
  const handleClearCart = useCallback(async () => {
    try {
      await dispatch(clearCartAsync());
      message.success('购物车已清空');
    } catch {
      message.error('清空失败');
    }
  }, [dispatch]);

  // 处理结算
  const handleCheckout = useCallback(() => {
    const selectedItems = items.filter(item => item.selected);

    if (selectedItems.length === 0) {
      message.warning('请选择要结算的商品');
      return;
    }

    // 检查库存
    const outOfStockItems = selectedItems.filter(item => item.quantity > 10); // 假设库存限制
    if (outOfStockItems.length > 0) {
      message.error('部分商品库存不足，请调整数量');
      return;
    }

    router.push(ROUTES.CHECKOUT);
  }, [items, router]);

  // 渲染购物车商品项
  const renderCartItem = (item: CartItem) => (
    <Card
      key={item.id}
      style={{ marginBottom: 16 }}
      bodyStyle={{ padding: 16 }}
    >
      <Row align='middle' gutter={[16, 16]}>
        {/* 选择框 */}
        <Col flex='none'>
          <Checkbox
            checked={item.selected}
            onChange={e => handleItemSelect(item.id, e.target.checked)}
          />
        </Col>

        {/* 商品图片 */}
        <Col flex='none'>
          <div style={{ width: 100, height: 100, position: 'relative' }}>
            <Image
              src={item.image || '/images/product-placeholder.svg'}
              alt={item.product_name}
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
                borderRadius: 8,
              }}
              preview={false}
            />
            {!item.selected && (
              <div
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  right: 0,
                  bottom: 0,
                  backgroundColor: 'rgba(0, 0, 0, 0.3)',
                  borderRadius: 8,
                }}
              />
            )}
          </div>
        </Col>

        {/* 商品信息 */}
        <Col flex='auto'>
          <div>
            <Title level={5} style={{ margin: 0, marginBottom: 4 }}>
              {item.product_name}
            </Title>
            {item.sku_name && (
              <Text type='secondary' style={{ fontSize: 12 }}>
                规格：{item.sku_name}
              </Text>
            )}
            <div style={{ marginTop: 8 }}>
              <Space wrap>
                <Tag color='blue'>正品保证</Tag>
                <Tag color='green'>7天退换</Tag>
              </Space>
            </div>
          </div>
        </Col>

        {/* 单价 */}
        <Col flex='none' style={{ textAlign: 'center', minWidth: 100 }}>
          <Text strong style={{ color: '#ff4d4f', fontSize: 16 }}>
            ¥{formatPrice(parseFloat(item.price))}
          </Text>
        </Col>

        {/* 数量控制 */}
        <Col flex='none' style={{ textAlign: 'center', minWidth: 120 }}>
          <Space>
            <Button
              type='text'
              icon={<MinusOutlined />}
              size='small'
              disabled={item.quantity <= 1 || loadingItems[item.id]}
              onClick={() => handleQuantityChange(item.id, item.quantity - 1)}
            />
            <InputNumber
              min={1}
              max={99}
              value={item.quantity}
              onChange={value => value && handleQuantityChange(item.id, value)}
              style={{ width: 60 }}
              size='small'
              disabled={loadingItems[item.id]}
            />
            <Button
              type='text'
              icon={<PlusOutlined />}
              size='small'
              disabled={item.quantity >= 99 || loadingItems[item.id]}
              onClick={() => handleQuantityChange(item.id, item.quantity + 1)}
            />
          </Space>
        </Col>

        {/* 小计 */}
        <Col flex='none' style={{ textAlign: 'center', minWidth: 100 }}>
          <Text strong style={{ color: '#ff4d4f', fontSize: 16 }}>
            ¥{formatPrice(parseFloat(item.price) * item.quantity)}
          </Text>
        </Col>

        {/* 操作按钮 */}
        <Col flex='none'>
          <Space direction='vertical' size='small'>
            <Tooltip title='移入收藏夹'>
              <Button type='text' icon={<HeartOutlined />} size='small' />
            </Tooltip>
            <Popconfirm
              title='确定要删除这件商品吗？'
              onConfirm={() => handleRemoveItem(item.id)}
              okText='确定'
              cancelText='取消'
            >
              <Tooltip title='删除'>
                <Button
                  type='text'
                  icon={<DeleteOutlined />}
                  size='small'
                  danger
                />
              </Tooltip>
            </Popconfirm>
          </Space>
        </Col>
      </Row>
    </Card>
  );

  // 空购物车状态
  if (!loading && items.length === 0) {
    return (
      <MainLayout>
        <div style={{ padding: '0 24px', maxWidth: 1200, margin: '0 auto' }}>
          <Card
            style={{ marginTop: 24, textAlign: 'center', padding: '60px 0' }}
          >
            <Empty
              image={
                <ShoppingCartOutlined
                  style={{ fontSize: 80, color: '#d9d9d9' }}
                />
              }
              description={
                <div>
                  <Title level={4} type='secondary'>
                    购物车是空的
                  </Title>
                  <Paragraph type='secondary'>快去挑选心仪的商品吧！</Paragraph>
                </div>
              }
            >
              <Space>
                <Button type='primary' onClick={() => router.push(ROUTES.HOME)}>
                  <HomeOutlined />
                  返回首页
                </Button>
                <Button onClick={() => router.push(ROUTES.PRODUCTS)}>
                  <ShopOutlined />
                  去逛逛
                </Button>
              </Space>
            </Empty>
          </Card>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div style={{ padding: '0 24px', maxWidth: 1200, margin: '0 auto' }}>
        {/* 页面标题 */}
        <div style={{ margin: '24px 0' }}>
          <Title level={2}>
            <ShoppingCartOutlined style={{ marginRight: 8 }} />
            购物车
            {items.length > 0 && (
              <Badge count={items.length} style={{ marginLeft: 8 }} />
            )}
          </Title>
        </div>

        <Spin spinning={loading}>
          <Row gutter={[24, 24]}>
            {/* 左侧：购物车商品列表 */}
            <Col xs={24} lg={16}>
              {/* 列表头部 */}
              <Card style={{ marginBottom: 16 }}>
                <Row align='middle' justify='space-between'>
                  <Col>
                    <Space>
                      <Checkbox
                        checked={
                          items.length > 0 && items.every(item => item.selected)
                        }
                        indeterminate={
                          items.some(item => item.selected) &&
                          !items.every(item => item.selected)
                        }
                        onChange={e => handleSelectAll(e.target.checked)}
                      >
                        全选
                      </Checkbox>
                      <Text type='secondary'>共 {items.length} 件商品</Text>
                    </Space>
                  </Col>
                  <Col>
                    <Space>
                      <Popconfirm
                        title='确定要删除选中的商品吗？'
                        onConfirm={handleBatchRemove}
                        disabled={selectedStats.count === 0}
                      >
                        <Button
                          type='text'
                          icon={<DeleteOutlined />}
                          disabled={selectedStats.count === 0}
                        >
                          删除选中
                        </Button>
                      </Popconfirm>
                      <Popconfirm
                        title='确定要清空购物车吗？'
                        onConfirm={handleClearCart}
                        disabled={items.length === 0}
                      >
                        <Button
                          type='text'
                          icon={<ClearOutlined />}
                          disabled={items.length === 0}
                        >
                          清空购物车
                        </Button>
                      </Popconfirm>
                    </Space>
                  </Col>
                </Row>
              </Card>

              {/* 商品列表 */}
              <div>{items.map(renderCartItem)}</div>
            </Col>

            {/* 右侧：结算信息 */}
            <Col xs={24} lg={8}>
              <div style={{ position: 'sticky', top: 80 }}>
                {/* 优惠信息 */}
                <Card style={{ marginBottom: 16 }}>
                  <Title level={5}>
                    <GiftOutlined
                      style={{ marginRight: 8, color: '#ff4d4f' }}
                    />
                    优惠信息
                  </Title>
                  <Alert
                    message='满99免运费'
                    description='再购买 ¥20.00 即可享受免运费'
                    type='info'
                    showIcon
                    style={{ marginBottom: 12 }}
                  />
                  <Space direction='vertical' style={{ width: '100%' }}>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Text>满减优惠：</Text>
                      <Text type='secondary'>暂无可用</Text>
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Text>优惠券：</Text>
                      <Button type='link' size='small'>
                        选择优惠券
                      </Button>
                    </div>
                  </Space>
                </Card>

                {/* 结算信息 */}
                <Card>
                  <Title level={5}>结算信息</Title>
                  <Space
                    direction='vertical'
                    style={{ width: '100%' }}
                    size='middle'
                  >
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Text>商品件数：</Text>
                      <Text>{selectedStats.quantity} 件</Text>
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Text>商品总价：</Text>
                      <Text>¥{selectedStats.amount}</Text>
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                      }}
                    >
                      <Text>运费：</Text>
                      <Text>¥0.00</Text>
                    </div>
                    <Divider style={{ margin: '12px 0' }} />
                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <Text strong style={{ fontSize: 16 }}>
                        应付总额：
                      </Text>
                      <Text strong style={{ fontSize: 20, color: '#ff4d4f' }}>
                        ¥{selectedStats.amount}
                      </Text>
                    </div>

                    <Button
                      type='primary'
                      size='large'
                      block
                      onClick={handleCheckout}
                      disabled={selectedStats.count === 0}
                      style={{ marginTop: 16 }}
                    >
                      去结算 ({selectedStats.count})
                    </Button>

                    {/* 服务保障 */}
                    <div
                      style={{
                        marginTop: 16,
                        padding: 12,
                        backgroundColor: '#f9f9f9',
                        borderRadius: 6,
                      }}
                    >
                      <Space wrap>
                        <Tag icon={<SafetyCertificateOutlined />} color='green'>
                          正品保证
                        </Tag>
                        <Tag icon={<TruckOutlined />} color='blue'>
                          极速发货
                        </Tag>
                        <Tag color='orange'>7天退换</Tag>
                      </Space>
                    </div>
                  </Space>
                </Card>
              </div>
            </Col>
          </Row>
        </Spin>
      </div>
    </MainLayout>
  );
};

export default CartPage;
