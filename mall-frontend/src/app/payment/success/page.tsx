'use client';

import React from 'react';
import { Result, Button, Card, Typography, Space } from 'antd';
import { useRouter, useSearchParams } from 'next/navigation';
import { ROUTES } from '@/constants';

const { Title, Paragraph, Text } = Typography;

export default function PaymentSuccessPage() {
  const router = useRouter();
  const search = useSearchParams();
  const orderId = search.get('order_id');

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Result
          status='success'
          title='支付成功'
          subTitle={
            <Space direction='vertical' size={4}>
              <Text>感谢您的购买，订单已完成支付。</Text>
              {orderId && <Text type='secondary'>订单号：{orderId}</Text>}
            </Space>
          }
          extra={
            <Space>
              {orderId ? (
                <Button
                  type='primary'
                  onClick={() =>
                    router.push(ROUTES.ORDER_DETAIL(Number(orderId)))
                  }
                >
                  查看订单详情
                </Button>
              ) : (
                <Button
                  type='primary'
                  onClick={() => router.push(ROUTES.ORDERS)}
                >
                  查看我的订单
                </Button>
              )}
              <Button onClick={() => router.push(ROUTES.HOME)}>返回首页</Button>
            </Space>
          }
        />
      </Card>
    </div>
  );
}
