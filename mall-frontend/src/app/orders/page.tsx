'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { Card, Table, Tag, Button, Space, message, Typography } from 'antd';
import {
  ClockCircleOutlined,
  PayCircleOutlined,
  CarOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  UndoOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import { Order } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice, formatDateTime } from '@/utils';

const { Title, Text, Paragraph } = Typography;

// 订单状态配置
const ORDER_STATUS_CONFIG = {
  pending: {
    label: '待支付',
    color: 'orange',
    icon: <ClockCircleOutlined />,
  },
  paid: {
    label: '已支付',
    color: 'blue',
    icon: <PayCircleOutlined />,
  },
  shipped: {
    label: '已发货',
    color: 'cyan',
    icon: <CarOutlined />,
  },
  delivered: {
    label: '已送达',
    color: 'green',
    icon: <CheckCircleOutlined />,
  },
  completed: {
    label: '已完成',
    color: 'green',
    icon: <CheckCircleOutlined />,
  },
  cancelled: {
    label: '已取消',
    color: 'red',
    icon: <CloseCircleOutlined />,
  },
  refunded: {
    label: '已退款',
    color: 'purple',
    icon: <UndoOutlined />,
  },
};

export default function OrdersPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);

  // 模拟订单数据
  const mockOrders: Order[] = [
    {
      id: 1,
      order_no: 'ORD20240101001',
      user_id: 1,
      status: 'pending',
      payment_status: 'pending',
      shipping_status: 'pending',
      total_amount: '299.99',
      payable_amount: '299.99',
      paid_amount: '0.00',
      shipping_address: {
        id: 1,
        user_id: 1,
        name: '张三',
        phone: '13800138000',
        province: '北京市',
        city: '北京市',
        district: '朝阳区',
        detail: '某某街道某某小区',
        postal_code: '100000',
        is_default: true,
      },
      items: [
        {
          id: 1,
          order_id: 1,
          product_id: 1,
          product_name: '商品名称1',
          price: '299.99',
          quantity: 1,
          total_amount: '299.99',
          image: '/images/product1.jpg',
        },
      ],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  ];

  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      setOrders(mockOrders);
    } catch (error) {
      message.error('加载订单失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const columns = [
    {
      title: '订单信息',
      key: 'orderInfo',
      render: (record: Order) => (
        <div>
          <Text strong>{record.order_no}</Text>
          <br />
          <Text type='secondary' style={{ fontSize: 12 }}>
            {formatDateTime(record.created_at)}
          </Text>
        </div>
      ),
    },
    {
      title: '状态',
      key: 'status',
      render: (record: Order) => {
        const config = ORDER_STATUS_CONFIG[record.status];
        return (
          <Tag color={config.color} icon={config.icon}>
            {config.label}
          </Tag>
        );
      },
    },
    {
      title: '金额',
      key: 'amount',
      render: (record: Order) => (
        <Text strong style={{ color: '#ff4d4f', fontSize: 16 }}>
          ¥{formatPrice(record.total_amount)}
        </Text>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (record: Order) => (
        <Button
          type='link'
          size='small'
          icon={<EyeOutlined />}
          onClick={() => router.push(`${ROUTES.ORDERS}/${record.id}`)}
        >
          查看详情
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div style={{ marginBottom: 16 }}>
          <Title level={2}>我的订单</Title>
          <Paragraph type='secondary'>查看和管理您的所有订单</Paragraph>
        </div>

        <Table
          columns={columns}
          dataSource={orders}
          rowKey='id'
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: total => `共 ${total} 条记录`,
          }}
        />
      </Card>
    </div>
  );
}
