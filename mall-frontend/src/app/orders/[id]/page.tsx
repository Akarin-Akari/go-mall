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
  Spin,
  message,
  Steps,
  Image,
  Descriptions,
  Timeline,
  Breadcrumb,
  Alert
} from 'antd';
import { 
  ArrowLeftOutlined,
  ShoppingCartOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  TruckOutlined,
  CloseCircleOutlined,
  PayCircleOutlined,
  HomeOutlined,
  ShopOutlined,
  CopyOutlined,
  PhoneOutlined,
  EnvironmentOutlined
} from '@ant-design/icons';
import { useRouter, useParams } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import {
  fetchOrderDetailAsync,
  cancelOrderAsync,
  selectOrder
} from '@/store/slices/orderSlice';
import MainLayout from '@/components/layout/MainLayout';
import { Order, OrderStatus } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice, formatDateTime } from '@/utils';

const { Title, Text, Paragraph } = Typography;
const { Step } = Steps;

// 订单状态配置
const ORDER_STATUS_CONFIG = {
  pending: { 
    label: '待支付', 
    color: 'orange', 
    icon: <ClockCircleOutlined />,
    step: 0
  },
  paid: { 
    label: '已支付', 
    color: 'blue', 
    icon: <PayCircleOutlined />,
    step: 1
  },
  shipped: { 
    label: '已发货', 
    color: 'cyan', 
    icon: <TruckOutlined />,
    step: 2
  },
  delivered: { 
    label: '已送达', 
    color: 'green', 
    icon: <CheckCircleOutlined />,
    step: 3
  },
  completed: { 
    label: '已完成', 
    color: 'green', 
    icon: <CheckCircleOutlined />,
    step: 4
  },
  cancelled: { 
    label: '已取消', 
    color: 'red', 
    icon: <CloseCircleOutlined />,
    step: -1
  },
  refunded: { 
    label: '已退款', 
    color: 'purple', 
    icon: <CloseCircleOutlined />,
    step: -1
  }
};

const OrderDetailPage: React.FC = () => {
  const [cancelling, setCancelling] = useState(false);
  
  const router = useRouter();
  const params = useParams();
  const dispatch = useAppDispatch();
  const { currentOrder, orderLoading } = useAppSelector(selectOrder);

  const orderId = parseInt(params?.id as string);

  // 加载订单详情
  useEffect(() => {
    if (orderId) {
      dispatch(fetchOrderDetailAsync(orderId));
    }
  }, [dispatch, orderId]);

  // 处理取消订单
  const handleCancelOrder = useCallback(async () => {
    if (!currentOrder) return;

    try {
      setCancelling(true);
      await dispatch(cancelOrderAsync({ 
        id: currentOrder.id, 
        reason: '用户主动取消' 
      }));
      message.success('订单已取消');
      // 重新加载订单详情
      dispatch(fetchOrderDetailAsync(orderId));
    } catch (error) {
      message.error('取消订单失败');
    } finally {
      setCancelling(false);
    }
  }, [dispatch, currentOrder, orderId]);

  // 处理继续支付
  const handlePayOrder = useCallback(() => {
    if (!currentOrder) return;
    router.push(`/payment?order_id=${currentOrder.id}`);
  }, [router, currentOrder]);

  // 复制订单号
  const handleCopyOrderNo = useCallback(() => {
    if (!currentOrder) return;
    navigator.clipboard.writeText(currentOrder.order_no);
    message.success('订单号已复制到剪贴板');
  }, [currentOrder]);

  // 返回上一页
  const handleGoBack = useCallback(() => {
    router.back();
  }, [router]);

  // 渲染订单状态步骤
  const renderOrderSteps = () => {
    if (!currentOrder) return null;

    const config = ORDER_STATUS_CONFIG[currentOrder.status];
    const currentStep = config.step;

    if (currentStep === -1) {
      // 取消或退款状态
      return (
        <Alert
          message={config.label}
          description={`订单已于 ${formatDateTime(currentOrder.updated_at)} ${config.label}`}
          type="warning"
          showIcon
          icon={config.icon}
          style={{ marginBottom: 24 }}
        />
      );
    }

    return (
      <Card title="订单状态" style={{ marginBottom: 24 }}>
        <Steps current={currentStep} status={currentStep === 4 ? 'finish' : 'process'}>
          <Step title="提交订单" description="订单已提交" />
          <Step title="支付订单" description="完成支付" />
          <Step title="商家发货" description="等待收货" />
          <Step title="确认收货" description="订单完成" />
        </Steps>
      </Card>
    );
  };

  // 渲染订单商品
  const renderOrderItems = () => {
    if (!currentOrder?.items) return null;

    return (
      <Card title="商品信息" style={{ marginBottom: 24 }}>
        {currentOrder.items.map((item, index) => (
          <div key={index}>
            <Row gutter={16} align="middle">
              <Col span={3}>
                <Image
                  src={item.product_image || '/images/placeholder-product.jpg'}
                  alt={item.product_name}
                  width={80}
                  height={80}
                  style={{ borderRadius: 4, objectFit: 'cover' }}
                  preview={false}
                />
              </Col>
              <Col span={12}>
                <div>
                  <Title level={5} style={{ margin: 0, marginBottom: 4 }}>
                    {item.product_name}
                  </Title>
                  {item.sku_name && (
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      规格: {item.sku_name}
                    </Text>
                  )}
                </div>
              </Col>
              <Col span={3}>
                <Text>¥{formatPrice(item.price)}</Text>
              </Col>
              <Col span={3}>
                <Text>×{item.quantity}</Text>
              </Col>
              <Col span={3}>
                <Text strong>¥{formatPrice((parseFloat(item.price) * item.quantity).toString())}</Text>
              </Col>
            </Row>
            {index < currentOrder.items.length - 1 && <Divider />}
          </div>
        ))}
        
        <Divider />
        
        <Row justify="end">
          <Col>
            <Space direction="vertical" align="end">
              <Text>商品总价: ¥{formatPrice(currentOrder.subtotal || currentOrder.total_amount)}</Text>
              {currentOrder.shipping_fee && parseFloat(currentOrder.shipping_fee) > 0 && (
                <Text>运费: ¥{formatPrice(currentOrder.shipping_fee)}</Text>
              )}
              {currentOrder.discount_amount && parseFloat(currentOrder.discount_amount) > 0 && (
                <Text style={{ color: '#52c41a' }}>
                  优惠: -¥{formatPrice(currentOrder.discount_amount)}
                </Text>
              )}
              <Title level={4} style={{ margin: 0, color: '#ff4d4f' }}>
                实付款: ¥{formatPrice(currentOrder.total_amount)}
              </Title>
            </Space>
          </Col>
        </Row>
      </Card>
    );
  };

  // 渲染订单信息
  const renderOrderInfo = () => {
    if (!currentOrder) return null;

    return (
      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card title="订单信息" size="small">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="订单号">
                <Space>
                  <Text code>{currentOrder.order_no}</Text>
                  <Button 
                    type="link" 
                    size="small" 
                    icon={<CopyOutlined />}
                    onClick={handleCopyOrderNo}
                  >
                    复制
                  </Button>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="订单状态">
                <Tag color={ORDER_STATUS_CONFIG[currentOrder.status].color}>
                  {ORDER_STATUS_CONFIG[currentOrder.status].label}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="下单时间">
                {formatDateTime(currentOrder.created_at)}
              </Descriptions.Item>
              <Descriptions.Item label="支付方式">
                {currentOrder.payment_method || '在线支付'}
              </Descriptions.Item>
              {currentOrder.buyer_message && (
                <Descriptions.Item label="买家留言">
                  {currentOrder.buyer_message}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>
        </Col>
        
        <Col xs={24} lg={12}>
          <Card title="收货信息" size="small">
            <Descriptions column={1} size="small">
              <Descriptions.Item label="收货人">
                <Space>
                  <Text>{currentOrder.receiver_name}</Text>
                  <Button 
                    type="link" 
                    size="small" 
                    icon={<PhoneOutlined />}
                    href={`tel:${currentOrder.receiver_phone}`}
                  >
                    {currentOrder.receiver_phone}
                  </Button>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="收货地址">
                <Space>
                  <EnvironmentOutlined />
                  <Text>
                    {currentOrder.province} {currentOrder.city} {currentOrder.district} {currentOrder.receiver_address}
                  </Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="配送方式">
                {currentOrder.shipping_method === 'standard' ? '标准配送' : '快速配送'}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>
      </Row>
    );
  };

  // 渲染操作按钮
  const renderActions = () => {
    if (!currentOrder) return null;

    const actions = [];

    switch (currentOrder.status) {
      case 'pending':
        actions.push(
          <Button
            key="pay"
            type="primary"
            size="large"
            icon={<PayCircleOutlined />}
            onClick={handlePayOrder}
          >
            立即支付
          </Button>
        );
        actions.push(
          <Button
            key="cancel"
            danger
            size="large"
            icon={<CloseCircleOutlined />}
            loading={cancelling}
            onClick={handleCancelOrder}
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
            size="large"
            icon={<CheckCircleOutlined />}
          >
            确认收货
          </Button>
        );
        break;
    }

    if (actions.length === 0) return null;

    return (
      <Card style={{ marginTop: 24 }}>
        <div style={{ textAlign: 'center' }}>
          <Space size="large">
            {actions}
          </Space>
        </div>
      </Card>
    );
  };

  if (orderLoading) {
    return (
      <MainLayout>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          minHeight: 400 
        }}>
          <Spin size="large" />
        </div>
      </MainLayout>
    );
  }

  if (!currentOrder) {
    return (
      <MainLayout>
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <Title level={3}>订单不存在</Title>
          <Button type="primary" onClick={() => router.push(ROUTES.ORDERS)}>
            返回订单列表
          </Button>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div style={{ padding: '0 24px' }}>
        {/* 面包屑导航 */}
        <Breadcrumb style={{ margin: '16px 0' }}>
          <Breadcrumb.Item href={ROUTES.HOME}>
            <HomeOutlined />
            <span>首页</span>
          </Breadcrumb.Item>
          <Breadcrumb.Item href={ROUTES.ORDERS}>
            <ShopOutlined />
            <span>我的订单</span>
          </Breadcrumb.Item>
          <Breadcrumb.Item>订单详情</Breadcrumb.Item>
        </Breadcrumb>

        {/* 返回按钮 */}
        <div style={{ marginBottom: 16 }}>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={handleGoBack}
          >
            返回
          </Button>
        </div>

        {/* 订单状态步骤 */}
        {renderOrderSteps()}

        {/* 订单商品 */}
        {renderOrderItems()}

        {/* 订单信息 */}
        {renderOrderInfo()}

        {/* 操作按钮 */}
        {renderActions()}
      </div>
    </MainLayout>
  );
};

export default OrderDetailPage;
