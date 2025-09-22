'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Row,
  Col,
  Card,
  Typography,
  Button,
  Space,
  Form,
  Input,
  Radio,
  Divider,
  Spin,
  message,
  Image,
  Select,
  Checkbox,
  Modal,
  Breadcrumb,
  Tag,
} from 'antd';
import {
  EnvironmentOutlined,
  PlusOutlined,
  PayCircleOutlined,
  TruckOutlined,
  HomeOutlined,
  ShoppingCartOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useRouter } from 'next/navigation';
import { useAppDispatch, useAppSelector } from '@/store';
import { fetchCartAsync, selectCart } from '@/store/slices/cartSlice';
import { createOrderAsync, selectOrder } from '@/store/slices/orderSlice';
import MainLayout from '@/components/layout/MainLayout';
import { Address } from '@/types';
import { ROUTES } from '@/constants';
import { formatPrice } from '@/utils';

const { Title, Text } = Typography;
const { Option } = Select;
const { TextArea } = Input;

// 模拟地址数据
const mockAddresses: Address[] = [
  {
    id: 1,
    name: '张三',
    phone: '13800138000',
    province: '北京市',
    city: '北京市',
    district: '朝阳区',
    detail: '朝阳路123号',
    is_default: true,
  },
  {
    id: 2,
    name: '李四',
    phone: '13900139000',
    province: '上海市',
    city: '上海市',
    district: '浦东新区',
    detail: '世纪大道456号',
    is_default: false,
  },
];

const CheckoutPage: React.FC = () => {
  const [form] = Form.useForm();
  const [selectedAddress, setSelectedAddress] = useState<Address | null>(null);
  const [addresses, setAddresses] = useState<Address[]>(mockAddresses);
  const [addressModalVisible, setAddressModalVisible] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState('alipay');
  const [shippingMethod, setShippingMethod] = useState('standard');
  const [buyerMessage, setBuyerMessage] = useState('');
  const [agreementChecked, setAgreementChecked] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const router = useRouter();

  const dispatch = useAppDispatch();
  const { items, loading } = useAppSelector(selectCart);
  const { orderLoading } = useAppSelector(selectOrder);

  // 获取选中的商品
  const selectedItems = items.filter(item => item.selected);

  // 初始化
  useEffect(() => {
    // 加载购物车数据
    dispatch(fetchCartAsync());

    // 设置默认地址
    const defaultAddress = addresses.find(addr => addr.is_default);
    if (defaultAddress) {
      setSelectedAddress(defaultAddress);
    }
  }, [dispatch, addresses]);

  // 计算订单金额
  const orderSummary = React.useMemo(() => {
    const subtotal = selectedItems.reduce(
      (sum, item) => sum + parseFloat(item.price) * item.quantity,
      0
    );
    const shippingFee = subtotal >= 99 ? 0 : 10; // 满99免运费
    const total = subtotal + shippingFee;

    return {
      subtotal: subtotal.toFixed(2),
      shippingFee: shippingFee.toFixed(2),
      total: total.toFixed(2),
      quantity: selectedItems.reduce((sum, item) => sum + item.quantity, 0),
    };
  }, [selectedItems]);

  // 处理地址选择
  const handleAddressSelect = useCallback((address: Address) => {
    setSelectedAddress(address);
  }, []);

  // 处理新增地址
  const handleAddAddress = useCallback((values: any) => {
    const newAddress: Address = {
      id: Date.now(),
      name: values.name,
      phone: values.phone,
      province: values.province,
      city: values.city,
      district: values.district,
      detail: values.address,
      is_default: values.is_default || false,
    };

    setAddresses(prev => [...prev, newAddress]);
    setAddressModalVisible(false);

    if (newAddress.is_default) {
      setSelectedAddress(newAddress);
    }

    message.success('地址添加成功');
  }, []);

  // 处理提交订单
  const handleSubmitOrder = useCallback(async () => {
    if (!selectedAddress) {
      message.error('请选择收货地址');
      return;
    }

    if (!agreementChecked) {
      message.error('请同意服务协议');
      return;
    }

    if (selectedItems.length === 0) {
      message.error('请选择要购买的商品');
      return;
    }

    try {
      setSubmitting(true);

      const orderData = {
        items: selectedItems.map(item => ({
          product_id: item.product_id,
          sku_id: item.sku_id,
          quantity: item.quantity,
          price: item.price,
        })),
        shipping_address: selectedAddress,
        remark: buyerMessage,
      };

      const result = await dispatch(createOrderAsync(orderData));

      if (result.meta.requestStatus === 'fulfilled') {
        message.success('订单创建成功');
        // 跳转到支付页面或订单详情页面
        const orderId = (result.payload as any).id;
        router.push(`/payment?order_id=${orderId}`);
      }
    } catch (error) {
      message.error('订单创建失败，请重试');
    } finally {
      setSubmitting(false);
    }
  }, [
    selectedAddress,
    agreementChecked,
    selectedItems,
    shippingMethod,
    paymentMethod,
    buyerMessage,
    dispatch,
    router,
  ]);

  // 返回购物车
  const handleGoBack = useCallback(() => {
    router.push(ROUTES.CART);
  }, [router]);

  // 渲染地址选择
  const renderAddressSelection = () => {
    return (
      <Card
        title={
          <Space>
            <EnvironmentOutlined />
            <span>收货地址</span>
          </Space>
        }
        extra={
          <Button
            type='link'
            icon={<PlusOutlined />}
            onClick={() => setAddressModalVisible(true)}
          >
            新增地址
          </Button>
        }
        style={{ marginBottom: 16 }}
      >
        {addresses.length > 0 ? (
          <Radio.Group
            value={selectedAddress?.id}
            onChange={e => {
              const address = addresses.find(
                addr => addr.id === e.target.value
              );
              if (address) handleAddressSelect(address);
            }}
            style={{ width: '100%' }}
          >
            <Space direction='vertical' style={{ width: '100%' }}>
              {addresses.map(address => (
                <Radio
                  key={address.id}
                  value={address.id}
                  style={{ width: '100%' }}
                >
                  <div style={{ marginLeft: 8 }}>
                    <Space>
                      <Text strong>{address.name}</Text>
                      <Text>{address.phone}</Text>
                      {address.is_default && <Tag color='blue'>默认</Tag>}
                    </Space>
                    <br />
                    <Text type='secondary'>
                      {address.province} {address.city} {address.district}{' '}
                      {address.detail}
                    </Text>
                  </div>
                </Radio>
              ))}
            </Space>
          </Radio.Group>
        ) : (
          <div style={{ textAlign: 'center', padding: '20px 0' }}>
            <Text type='secondary'>暂无收货地址</Text>
            <br />
            <Button
              type='primary'
              icon={<PlusOutlined />}
              onClick={() => setAddressModalVisible(true)}
              style={{ marginTop: 8 }}
            >
              添加收货地址
            </Button>
          </div>
        )}
      </Card>
    );
  };

  // 渲染商品列表
  const renderOrderItems = () => {
    return (
      <Card title='确认订单' style={{ marginBottom: 16 }}>
        {selectedItems.map(item => (
          <div key={item.id} style={{ marginBottom: 16 }}>
            <Row gutter={16} align='middle'>
              <Col span={3}>
                <Image
                  src={item.image || '/images/placeholder-product.jpg'}
                  alt={item.product_name}
                  width={60}
                  height={60}
                  style={{ borderRadius: 4, objectFit: 'cover' }}
                  preview={false}
                />
              </Col>
              <Col span={12}>
                <div>
                  <Text strong>{item.product_name}</Text>
                  <br />
                  {item.sku_name && (
                    <Text type='secondary' style={{ fontSize: 12 }}>
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
                <Text strong>
                  ¥
                  {formatPrice(
                    (parseFloat(item.price) * item.quantity).toString()
                  )}
                </Text>
              </Col>
            </Row>
            <Divider />
          </div>
        ))}
      </Card>
    );
  };

  // 渲染配送和支付方式
  const renderShippingAndPayment = () => {
    return (
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card title='配送方式' size='small'>
            <Radio.Group
              value={shippingMethod}
              onChange={e => setShippingMethod(e.target.value)}
            >
              <Space direction='vertical'>
                <Radio value='standard'>
                  <Space>
                    <TruckOutlined />
                    <div>
                      <Text>标准配送</Text>
                      <br />
                      <Text type='secondary' style={{ fontSize: 12 }}>
                        3-5个工作日送达，满99元免运费
                      </Text>
                    </div>
                  </Space>
                </Radio>
                <Radio value='express'>
                  <Space>
                    <TruckOutlined />
                    <div>
                      <Text>快速配送 (+¥5)</Text>
                      <br />
                      <Text type='secondary' style={{ fontSize: 12 }}>
                        1-2个工作日送达
                      </Text>
                    </div>
                  </Space>
                </Radio>
              </Space>
            </Radio.Group>
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card title='支付方式' size='small'>
            <Radio.Group
              value={paymentMethod}
              onChange={e => setPaymentMethod(e.target.value)}
            >
              <Space direction='vertical'>
                <Radio value='alipay'>
                  <Space>
                    <PayCircleOutlined />
                    <Text>支付宝</Text>
                  </Space>
                </Radio>
                <Radio value='wechat'>
                  <Space>
                    <PayCircleOutlined />
                    <Text>微信支付</Text>
                  </Space>
                </Radio>
                <Radio value='unionpay'>
                  <Space>
                    <PayCircleOutlined />
                    <Text>银联支付</Text>
                  </Space>
                </Radio>
              </Space>
            </Radio.Group>
          </Card>
        </Col>
      </Row>
    );
  };

  // 渲染订单摘要
  const renderOrderSummary = () => {
    return (
      <Card title='订单摘要' style={{ marginTop: 16 }}>
        <Space direction='vertical' style={{ width: '100%' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>商品总价 ({orderSummary.quantity}件)</Text>
            <Text>¥{orderSummary.subtotal}</Text>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>运费</Text>
            <Text>
              {parseFloat(orderSummary.shippingFee) === 0
                ? '免运费'
                : `¥${orderSummary.shippingFee}`}
            </Text>
          </div>
          <Divider />
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Title level={4}>实付款</Title>
            <Title level={4} style={{ color: '#ff4d4f' }}>
              ¥{orderSummary.total}
            </Title>
          </div>
        </Space>

        <Divider />

        {/* 买家留言 */}
        <div style={{ marginBottom: 16 }}>
          <Text>买家留言：</Text>
          <TextArea
            placeholder='选填，请先和商家协商一致'
            value={buyerMessage}
            onChange={e => setBuyerMessage(e.target.value)}
            rows={2}
            maxLength={200}
            style={{ marginTop: 8 }}
          />
        </div>

        {/* 服务协议 */}
        <div style={{ marginBottom: 16 }}>
          <Checkbox
            checked={agreementChecked}
            onChange={e => setAgreementChecked(e.target.checked)}
          >
            我已阅读并同意{' '}
            <a href='/agreement' target='_blank'>
              《服务协议》
            </a>{' '}
            和{' '}
            <a href='/privacy' target='_blank'>
              《隐私政策》
            </a>
          </Checkbox>
        </div>

        {/* 提交按钮 */}
        <Button
          type='primary'
          size='large'
          block
          loading={submitting || orderLoading}
          onClick={handleSubmitOrder}
          disabled={
            !selectedAddress || !agreementChecked || selectedItems.length === 0
          }
        >
          提交订单
        </Button>
      </Card>
    );
  };

  if (loading) {
    return (
      <MainLayout>
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: 400,
          }}
        >
          <Spin size='large' />
        </div>
      </MainLayout>
    );
  }

  if (selectedItems.length === 0) {
    return (
      <MainLayout>
        <div style={{ textAlign: 'center', padding: '60px 0' }}>
          <Title level={3}>请选择要购买的商品</Title>
          <Button type='primary' onClick={() => router.push(ROUTES.CART)}>
            返回购物车
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
          <Breadcrumb.Item href={ROUTES.CART}>
            <ShoppingCartOutlined />
            <span>购物车</span>
          </Breadcrumb.Item>
          <Breadcrumb.Item>确认订单</Breadcrumb.Item>
        </Breadcrumb>

        {/* 返回按钮 */}
        <div style={{ marginBottom: 16 }}>
          <Button icon={<ArrowLeftOutlined />} onClick={handleGoBack}>
            返回购物车
          </Button>
        </div>

        <Row gutter={[24, 24]}>
          <Col xs={24} lg={16}>
            {/* 收货地址 */}
            {renderAddressSelection()}

            {/* 确认订单 */}
            {renderOrderItems()}

            {/* 配送和支付方式 */}
            {renderShippingAndPayment()}
          </Col>

          <Col xs={24} lg={8}>
            {/* 订单摘要 */}
            {renderOrderSummary()}
          </Col>
        </Row>

        {/* 新增地址弹窗 */}
        <Modal
          title='新增收货地址'
          open={addressModalVisible}
          onCancel={() => setAddressModalVisible(false)}
          footer={null}
          width={600}
        >
          <Form form={form} layout='vertical' onFinish={handleAddAddress}>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name='name'
                  label='收货人'
                  rules={[{ required: true, message: '请输入收货人姓名' }]}
                >
                  <Input placeholder='请输入收货人姓名' />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name='phone'
                  label='手机号'
                  rules={[
                    { required: true, message: '请输入手机号' },
                    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
                  ]}
                >
                  <Input placeholder='请输入手机号' />
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={8}>
                <Form.Item
                  name='province'
                  label='省份'
                  rules={[{ required: true, message: '请选择省份' }]}
                >
                  <Select placeholder='请选择省份'>
                    <Option value='北京市'>北京市</Option>
                    <Option value='上海市'>上海市</Option>
                    <Option value='广东省'>广东省</Option>
                    <Option value='浙江省'>浙江省</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name='city'
                  label='城市'
                  rules={[{ required: true, message: '请选择城市' }]}
                >
                  <Select placeholder='请选择城市'>
                    <Option value='北京市'>北京市</Option>
                    <Option value='上海市'>上海市</Option>
                    <Option value='深圳市'>深圳市</Option>
                    <Option value='杭州市'>杭州市</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item
                  name='district'
                  label='区县'
                  rules={[{ required: true, message: '请选择区县' }]}
                >
                  <Select placeholder='请选择区县'>
                    <Option value='朝阳区'>朝阳区</Option>
                    <Option value='海淀区'>海淀区</Option>
                    <Option value='浦东新区'>浦东新区</Option>
                    <Option value='西湖区'>西湖区</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>

            <Form.Item
              name='address'
              label='详细地址'
              rules={[{ required: true, message: '请输入详细地址' }]}
            >
              <TextArea
                placeholder='请输入详细地址，如街道、门牌号等'
                rows={3}
              />
            </Form.Item>

            <Form.Item name='is_default' valuePropName='checked'>
              <Checkbox>设为默认地址</Checkbox>
            </Form.Item>

            <Form.Item>
              <Space>
                <Button type='primary' htmlType='submit'>
                  保存地址
                </Button>
                <Button onClick={() => setAddressModalVisible(false)}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </MainLayout>
  );
};

export default CheckoutPage;
