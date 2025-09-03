'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { 
  Row, 
  Col, 
  Card, 
  Typography, 
  Button, 
  Space, 
  Tag, 
  Divider,
  Empty,
  Spin,
  message,
  Tabs,
  Image,
  Badge,
  Tooltip,
  Breadcrumb,
  Pagination
} from 'antd';
import { 
  ShoppingCartOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  TruckOutlined,
  CloseCircleOutlined,
  EyeOutlined,
  DeleteOutlined,
  PayCircleOutlined,
  HomeOutlined,
  ShopOutlined,
  ReloadOutlined
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  fetchOrdersAsync,
  cancelOrderAsync,
  selectOrder
} from '@/store/slices/orderSlice';
import MainLayout from '@/components/layout/MainLayout';
import { Order, OrderStatus } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice, formatDateTime } from '@/utils';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

// 订单状态配置
const ORDER_STATUS_CONFIG = {
  pending: { 
    label: '待支付', 
    color: 'orange', 
    icon: <ClockCircleOutlined /> 
  },
  paid: { 
    label: '已支付', 
    color: 'blue', 
    icon: <PayCircleOutlined /> 
  },
  shipped: { 
    label: '已发货', 
    color: 'cyan', 
    icon: <TruckOutlined /> 
  },
  delivered: { 
    label: '已送达', 
    color: 'green', 
    icon: <CheckCircleOutlined /> 
  },
  completed: { 
    label: '已完成', 
    color: 'green', 
    icon: <CheckCircleOutlined /> 
  },
  cancelled: { 
    label: '已取消', 
    color: 'red', 
    icon: <CloseCircleOutlined /> 
  },
  refunded: { 
    label: '已退款', 
    color: 'purple', 
    icon: <CloseCircleOutlined /> 
  }
};

const OrdersPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('all');
  const [cancellingOrders, setCancellingOrders] = useState<Set<number>>(new Set());
  
  const router = useRouter();
  const dispatch = useAppDispatch();
  const { orders, total, loading, searchParams } = useAppSelector(selectOrder);

  // 加载订单数据
  useEffect(() => {
    loadOrders();
  }, [activeTab]);

  const loadOrders = useCallback((page = 1) => {
    const status = activeTab === 'all' ? undefined : activeTab as OrderStatus;
    dispatch(fetchOrdersAsync({
      page,
      page_size: 10,
      status
    }));
  }, [dispatch, activeTab]);

  // 处理取消订单
  const handleCancelOrder = useCallback(async (orderId: number) => {
    try {
      setCancellingOrders(prev => new Set(prev).add(orderId));
      await dispatch(cancelOrderAsync({ 
        id: orderId, 
        reason: '用户主动取消' 
      }));
      message.success('订单已取消');
      loadOrders(); // 重新加载数据
    } catch (error) {
      message.error('取消订单失败');
    } finally {
      setCancellingOrders(prev => {
        const newSet = new Set(prev);
        newSet.delete(orderId);
        return newSet;
      });
    }
  }, [dispatch, loadOrders]);

  // 处理查看订单详情
  const handleViewOrder = useCallback((orderId: number) => {
    router.push(ROUTES.ORDER_DETAIL(orderId));
  }, [router]);

  // 处理继续支付
  const handlePayOrder = useCallback((orderId: number) => {
    router.push(`/payment?order_id=${orderId}`);
  }, [router]);

  // 渲染订单状态标签
  const renderOrderStatus = (status: OrderStatus) => {
    const config = ORDER_STATUS_CONFIG[status];
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.label}
      </Tag>
    );
  };

  // 渲染订单操作按钮
  const renderOrderActions = (order: Order) => {
    const actions = [];

    // 查看详情
    actions.push(
      <Button
        key="view"
        type="link"
        size="small"
        icon={<EyeOutlined />}
        onClick={() => handleViewOrder(order.id)}
      >
        查看详情
      </Button>
    );

    // 根据订单状态显示不同操作
    switch (order.status) {
      case 'pending':
        actions.push(
          <Button
            key="pay"
            type="primary"
            size="small"
            icon={<PayCircleOutlined />}
            onClick={() => handlePayOrder(order.id)}
          >
            立即支付
          </Button>
        );
        actions.push(
          <Button
            key="cancel"
            danger
            size="small"
            icon={<CloseCircleOutlined />}
            loading={cancellingOrders.has(order.id)}
            onClick={() => handleCancelOrder(order.id)}
          >
            取消订单
          </Button>
        );
        break;
      
      case 'delivered':
        actions.push(
          <Button
            key="confirm"
            type="primary"
            size="small"
            icon={<CheckCircleOutlined />}
          >
            确认收货
          </Button>
        );
        break;
    }

    return actions;
  };

  // 渲染订单商品列表
  const renderOrderItems = (order: Order) => {
    return (
      <div style={{ marginTop: 16 }}>
        {order.items?.map((item, index) => (
          <div key={index} style={{ display: 'flex', alignItems: 'center', marginBottom: 12 }}>
            <Image
              src={item.product_image || '/images/placeholder-product.jpg'}
              alt={item.product_name}
              width={60}
              height={60}
              style={{ borderRadius: 4, objectFit: 'cover' }}
              preview={false}
            />
            <div style={{ marginLeft: 12, flex: 1 }}>
              <Text strong>{item.product_name}</Text>
              <br />
              <Text type="secondary" style={{ fontSize: 12 }}>
                {item.sku_name && `规格: ${item.sku_name}`}
              </Text>
              <br />
              <Space>
                <Text type="secondary">¥{formatPrice(item.price)}</Text>
                <Text type="secondary">×{item.quantity}</Text>
              </Space>
            </div>
          </div>
        ))}
      </div>
    );
  };

  // 渲染订单卡片
  const renderOrderCard = (order: Order) => {
    return (
      <Card
        key={order.id}
        style={{ marginBottom: 16 }}
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Space>
              <Text strong>订单号: {order.order_no}</Text>
              {renderOrderStatus(order.status)}
            </Space>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {formatDateTime(order.created_at)}
            </Text>
          </div>
        }
        extra={
          <Space>
            {renderOrderActions(order)}
          </Space>
        }
      >
        {renderOrderItems(order)}
        
        <Divider />
        
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Text type="secondary">收货人: {order.receiver_name}</Text>
            <Text type="secondary">电话: {order.receiver_phone}</Text>
          </Space>
          <div style={{ textAlign: 'right' }}>
            <Text type="secondary">共 {order.total_quantity} 件商品</Text>
            <br />
            <Text strong style={{ fontSize: 16, color: '#ff4d4f' }}>
              合计: ¥{formatPrice(order.total_amount)}
            </Text>
          </div>
        </div>
      </Card>
    );
  };

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
            <span>我的订单</span>
          </Breadcrumb.Item>
        </Breadcrumb>

        <Row gutter={[24, 24]}>
          <Col span={24}>
            <Card>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                <Title level={3} style={{ margin: 0 }}>
                  我的订单
                </Title>
                <Button 
                  icon={<ReloadOutlined />} 
                  onClick={() => loadOrders()}
                  loading={loading}
                >
                  刷新
                </Button>
              </div>

              {/* 订单状态标签页 */}
              <Tabs 
                activeKey={activeTab} 
                onChange={setActiveTab}
                size="large"
              >
                <TabPane tab="全部订单" key="all" />
                <TabPane tab={
                  <Badge count={0} size="small">
                    <span>待支付</span>
                  </Badge>
                } key="pending" />
                <TabPane tab="已支付" key="paid" />
                <TabPane tab="已发货" key="shipped" />
                <TabPane tab="已完成" key="completed" />
                <TabPane tab="已取消" key="cancelled" />
              </Tabs>

              {/* 订单列表 */}
              <Spin spinning={loading}>
                {orders.length > 0 ? (
                  <>
                    {orders.map(renderOrderCard)}
                    
                    {/* 分页 */}
                    <div style={{ textAlign: 'center', marginTop: 24 }}>
                      <Pagination
                        current={searchParams.page}
                        total={total}
                        pageSize={searchParams.page_size}
                        showSizeChanger
                        showQuickJumper
                        showTotal={(total, range) =>
                          `第 ${range[0]}-${range[1]} 条，共 ${total} 条订单`
                        }
                        onChange={(page, size) => loadOrders(page)}
                      />
                    </div>
                  </>
                ) : (
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description="暂无订单"
                    style={{ margin: '60px 0' }}
                  >
                    <Button type="primary" onClick={() => router.push(ROUTES.HOME)}>
                      去购物
                    </Button>
                  </Empty>
                )}
              </Spin>
            </Card>
          </Col>
        </Row>
      </div>
    </MainLayout>
  );
};

export default OrdersPage;
