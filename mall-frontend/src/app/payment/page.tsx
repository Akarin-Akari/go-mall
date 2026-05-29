'use client';

import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import {
  Card,
  Space,
  Typography,
  Button,
  Result,
  Alert,
  Radio,
  Spin,
  QRCode,
  message,
} from 'antd';
import { useRouter, useSearchParams } from 'next/navigation';
import { ROUTES, PAYMENT_METHODS } from '@/constants';
import { paymentAPI, orderAPI } from '@/services/api';

const { Title, Paragraph, Text } = Typography;

type PayStage =
  | 'init'
  | 'creating'
  | 'pending'
  | 'success'
  | 'failed'
  | 'error';

export default function PaymentPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const [orderId, setOrderId] = useState<number | null>(null);
  const [method, setMethod] = useState<string>(PAYMENT_METHODS.ALIPAY);
  const [paymentId, setPaymentId] = useState<number | null>(null);
  const [paymentUrl, setPaymentUrl] = useState<string | undefined>();
  const [qrCode, setQrCode] = useState<string | undefined>();
  const [stage, setStage] = useState<PayStage>('init');
  const [error, setError] = useState<string | null>(null);
  const [orderAmount, setOrderAmount] = useState<string | null>(null);
  const [orderSubject, setOrderSubject] = useState<string>('');

  const pollingTimer = useRef<number | null>(null);

  // 读取URL中的 order_id 参数
  useEffect(() => {
    const idParam = searchParams.get('order_id');
    if (!idParam) {
      setError('缺少必要的订单参数 order_id');
      setStage('error');
      return;
    }
    const idNum = Number(idParam);
    if (Number.isNaN(idNum) || idNum <= 0) {
      setError('订单参数不合法');
      setStage('error');
      return;
    }
    setOrderId(idNum);
  }, [searchParams]);

  // 载入订单金额与主题
  useEffect(() => {
    if (!orderId) return;
    (async () => {
      try {
        const res = await orderAPI.getOrderDetail(orderId);
        const amt = res.data?.payable_amount || res.data?.total_amount;
        if (amt) setOrderAmount(String(amt));
        setOrderSubject(`订单支付 #${orderId}`);
      } catch (err: any) {
        // 金额拉取失败不阻塞创建支付，后续按需再次获取
      }
    })();
  }, [orderId]);

  const clearPolling = () => {
    if (pollingTimer.current) {
      window.clearInterval(pollingTimer.current);
      pollingTimer.current = null;
    }
  };

  useEffect(() => () => clearPolling(), []);

  const handleQueryPaymentOnce = useCallback(
    async (pid: number) => {
      try {
        const res = await paymentAPI.queryPayment(pid);
        const status = String(
          (res.data as any).status ?? (res.data as any).payment_status ?? ''
        ).toLowerCase();
        if (status === 'success' || status === 'paid') {
          setStage('success');
          clearPolling();
          router.replace(
            `${ROUTES.PAYMENT_SUCCESS}?order_id=${orderId}&payment_id=${pid}`
          );
          return true;
        }
        if (status === 'failed') {
          setStage('failed');
          clearPolling();
          router.replace(
            `${ROUTES.PAYMENT_FAILED}?order_id=${orderId}&payment_id=${pid}`
          );
          return true;
        }
        return false; // 继续轮询
      } catch (err: any) {
        // 查询失败不立刻终止，给到轻提示并继续下一轮
        message.warning(err?.message || '支付状态查询失败，稍后将重试');
        return false;
      }
    },
    [orderId, router]
  );

  const startPolling = useCallback(
    (pid: number) => {
      clearPolling();
      const startAt = Date.now();
      pollingTimer.current = window.setInterval(async () => {
        const elapsed = Date.now() - startAt;
        if (elapsed > 2 * 60 * 1000) {
          // 超过2分钟视为失败
          clearPolling();
          setStage('failed');
          router.replace(
            `${ROUTES.PAYMENT_FAILED}?order_id=${orderId}&payment_id=${pid}`
          );
          return;
        }
        await handleQueryPaymentOnce(pid);
      }, 2000) as unknown as number;
    },
    [handleQueryPaymentOnce, orderId, router]
  );

  const createPayment = useCallback(async () => {
    if (!orderId) return;
    setStage('creating');
    setError(null);
    try {
      const returnUrl = `${window.location.origin}${ROUTES.PAYMENT_SUCCESS}`;
      // 若未拿到金额，则兜底拉取一次
      let amount = orderAmount;
      if (!amount) {
        try {
          const ord = await orderAPI.getOrderDetail(orderId);
          amount = String(
            ord.data?.payable_amount || ord.data?.total_amount || '0.01'
          );
          setOrderAmount(amount);
        } catch {}
      }
      const subject = orderSubject || `订单支付 #${orderId}`;
      const notifyUrl = `http://localhost:8081/api/v1/payments/callback/${method === PAYMENT_METHODS.WECHAT ? 'wechat' : 'alipay'}`;
      const res = await paymentAPI.createPayment({
        order_id: orderId,
        payment_method: method,
        amount: String(amount || '0.01'),
        subject,
        return_url: returnUrl,
        notify_url: notifyUrl,
        expired_minutes: 15,
      });
      const { payment_id, payment_url, qr_code } = res.data || {};
      if (!payment_id) throw new Error('创建支付失败：缺少支付单ID');
      setPaymentId(payment_id);
      setPaymentUrl(payment_url);
      setQrCode(qr_code);
      setStage('pending');
      startPolling(payment_id);
      // 有直开链接时，默认新开标签页
      if (payment_url) {
        window.open(payment_url, '_blank');
      }
    } catch (err: any) {
      setStage('error');
      setError(err?.message || '创建支付失败');
    }
  }, [method, orderId, startPolling]);

  // orderId ready 后自动创建一次支付
  useEffect(() => {
    if (orderId) {
      createPayment();
    }
  }, [orderId, createPayment]);

  const content = useMemo(() => {
    if (stage === 'error') {
      return (
        <Result
          status='error'
          title='创建支付失败'
          subTitle={error || '请稍后重试，或更换支付方式再次尝试'}
          extra={
            <Space>
              <Button onClick={() => router.push(ROUTES.ORDERS)}>
                返回订单
              </Button>
              <Button type='primary' onClick={createPayment} icon={undefined}>
                重试创建
              </Button>
            </Space>
          }
        />
      );
    }

    if (stage === 'success') {
      return (
        <Result
          status='success'
          title='支付成功'
          subTitle={<Text>订单 {orderId} 已完成支付，正在为您跳转...</Text>}
        />
      );
    }

    if (stage === 'failed') {
      return (
        <Result
          status='error'
          title='支付未完成'
          subTitle='若已支付，请稍后在订单页刷新查看；可点击下方按钮重新发起'
          extra={
            <Space>
              <Button onClick={() => router.push(ROUTES.ORDERS)}>
                返回订单
              </Button>
              <Button type='primary' onClick={createPayment}>
                重新发起支付
              </Button>
            </Space>
          }
        />
      );
    }

    // creating / pending 展示支付信息与操作
    return (
      <Space direction='vertical' size={16} style={{ width: '100%' }}>
        <Alert
          type='info'
          message='请在新打开的页面完成支付，或使用二维码扫码支付'
          showIcon
        />

        <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
          <Card style={{ flex: 1, minWidth: 320 }}>
            <Title level={4}>选择支付方式</Title>
            <Radio.Group
              value={method}
              onChange={e => setMethod(e.target.value)}
              optionType='button'
              buttonStyle='solid'
            >
              <Radio.Button value={PAYMENT_METHODS.ALIPAY}>支付宝</Radio.Button>
              <Radio.Button value={PAYMENT_METHODS.WECHAT}>
                微信支付
              </Radio.Button>
              <Radio.Button value={PAYMENT_METHODS.BALANCE}>余额</Radio.Button>
            </Radio.Group>
            <div style={{ marginTop: 12 }}>
              <Button
                type='primary'
                onClick={createPayment}
                disabled={!orderId}
              >
                重新生成支付链接
              </Button>
            </div>
          </Card>

          <Card style={{ flex: 1, minWidth: 320 }}>
            <Title level={4}>支付二维码/链接</Title>
            <div
              style={{
                minHeight: 220,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              {stage === 'creating' ? (
                <Spin tip='创建中...' />
              ) : qrCode ? (
                <img
                  src={qrCode}
                  alt='支付二维码'
                  style={{ width: 220, height: 220 }}
                />
              ) : paymentUrl ? (
                <QRCode value={paymentUrl} size={220} />
              ) : (
                <Text type='secondary'>暂无支付信息，请点击上方按钮生成</Text>
              )}
            </div>
            {paymentUrl && (
              <Space>
                <Button
                  type='link'
                  onClick={() => window.open(paymentUrl!, '_blank')}
                >
                  在新标签页打开支付
                </Button>
                <Button
                  onClick={() => handleQueryPaymentOnce(paymentId!)}
                  disabled={!paymentId}
                >
                  我已完成支付，立即验证
                </Button>
              </Space>
            )}
          </Card>
        </div>
      </Space>
    );
  }, [
    stage,
    error,
    router,
    createPayment,
    method,
    qrCode,
    paymentUrl,
    paymentId,
    handleQueryPaymentOnce,
    orderId,
  ]);

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Title level={2}>订单支付</Title>
        <Paragraph type='secondary'>订单号：{orderId ?? '-'}</Paragraph>
        {content}
      </Card>
    </div>
  );
}
