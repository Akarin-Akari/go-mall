'use client';

import React, { useEffect, useMemo, useCallback } from 'react';
import { Card, Table, Tag, Button, Space, Typography } from 'antd';
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
import {
  fetchOrdersAsync,
  selectOrders,
  selectOrderLoading,
  selectOrderTotal,
  selectOrderSearchParams,
  setSearchParams,
} from '@/store/slices/orderSlice';

const { Title, Text, Paragraph } = Typography;

// 订单状态配置
const ORDER_STATUS_CONFIG = {
  pending: { label: '待支付', color: 'orange', icon: <ClockCircleOutlined /> },
  paid: { label: '已支付', color: 'blue', icon: <PayCircleOutlined /> },
  shipped: { label: '已发货', color: 'cyan', icon: <CarOutlined /> },
  delivered: { label: '已送达', color: 'green', icon: <CheckCircleOutlined /> },
  completed: { label: '已完成', color: 'green', icon: <CheckCircleOutlined /> },
  cancelled: { label: '已取消', color: 'red', icon: <CloseCircleOutlined /> },
  refunded: { label: '已退款', color: 'purple', icon: <UndoOutlined /> },
} as const;

export default function OrdersPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();

  const orders = useAppSelector(selectOrders);
  const loading = useAppSelector(selectOrderLoading);
  const total = useAppSelector(selectOrderTotal);
  const searchParams = useAppSelector(selectOrderSearchParams);

  const loadOrders = useCallback(() => {
    dispatch(fetchOrdersAsync(searchParams));
  }, [dispatch, searchParams]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const columns = useMemo(
    () => [
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
          const config =
            ORDER_STATUS_CONFIG[
              record.status as keyof typeof ORDER_STATUS_CONFIG
            ];
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
            onClick={() => router.push(ROUTES.ORDER_DETAIL(record.id))}
          >
            查看详情
          </Button>
        ),
      },
    ],
    [router]
  );

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
            current: searchParams.page,
            pageSize: searchParams.page_size,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: t => `共 ${t} 条记录`,
            onChange: (page, pageSize) => {
              dispatch(setSearchParams({ page, page_size: pageSize }));
              dispatch(
                fetchOrdersAsync({ ...searchParams, page, page_size: pageSize })
              );
            },
          }}
        />
      </Card>
    </div>
  );
}
