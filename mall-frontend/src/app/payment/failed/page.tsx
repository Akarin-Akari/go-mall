'use client';

import React from 'react';
import { Result, Button, Card, Typography, Space } from 'antd';
import { useRouter, useSearchParams } from 'next/navigation';
import { ROUTES } from '@/constants';

const { Text } = Typography;

export default function PaymentFailedPage() {
  const router = useRouter();
  const search = useSearchParams();
  const orderId = search.get('order_id');

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Result
          status='error'
          title='支付未完成'
          subTitle={
            <Space direction='vertical' size={4}>
              <Text>如果您确认已完成支付，请稍后在订单页刷新查看。</Text>
              {orderId && <Text type='secondary'>订单号：{orderId}</Text>}
            </Space>
          }
          extra={
            <Space>
              {orderId && (
                <Button
                  type='primary'
                  onClick={() =>
                    router.push(`${ROUTES.PAYMENT}?order_id=${orderId}`)
                  }
                >
                  重新发起支付
                </Button>
              )}
              <Button onClick={() => router.push(ROUTES.ORDERS)}>
                返回订单
              </Button>
              <Button onClick={() => router.push(ROUTES.HOME)}>返回首页</Button>
            </Space>
          }
        />
      </Card>
    </div>
  );
}
