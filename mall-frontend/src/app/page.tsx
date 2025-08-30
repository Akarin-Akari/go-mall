'use client';

import React from 'react';
import { Card, Row, Col, Statistic, Button, Typography } from 'antd';
import { ShoppingCartOutlined, UserOutlined, ShopOutlined, DollarOutlined } from '@ant-design/icons';
import MainLayout from '@/components/layout/MainLayout';

const { Title, Paragraph } = Typography;

export default function Home() {
  return (
    <MainLayout>
      <div>
        <Title level={1}>欢迎来到Go商城 🛒</Title>
        <Paragraph>
          这是一个基于React + Next.js + TypeScript构建的现代化商城前端应用，
          与Go后端API完美集成，支持完整的电商功能。
        </Paragraph>

        <Row gutter={[16, 16]} style={{ marginTop: 32 }}>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="商品总数"
                value={1128}
                prefix={<ShopOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>

          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="用户总数"
                value={9280}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#1890ff' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="订单总数"
                value={5420}
                prefix={<ShoppingCartOutlined />}
                valueStyle={{ color: '#722ed1' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card>
              <Statistic
                title="销售额"
                value={128900}
                prefix={<DollarOutlined />}
                precision={2}
                valueStyle={{ color: '#cf1322' }}
                suffix="元"
              />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]} style={{ marginTop: 32 }}>
          <Col xs={24} md={12}>
            <Card title="🚀 技术特性" bordered={false}>
              <ul style={{ paddingLeft: 20 }}>
                <li>React 18 + Next.js 14 现代化架构</li>
                <li>TypeScript 类型安全开发</li>
                <li>Ant Design 5.0 企业级UI组件</li>
                <li>Redux Toolkit 状态管理</li>
                <li>TanStack Query 数据获取</li>
                <li>完整的认证和权限系统</li>
              </ul>
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card title="📱 跨平台支持" bordered={false}>
              <ul style={{ paddingLeft: 20 }}>
                <li>响应式Web设计</li>
                <li>React Native移动端支持</li>
                <li>PWA渐进式Web应用</li>
                <li>服务端渲染(SSR)优化</li>
                <li>SEO友好的页面结构</li>
                <li>闪购外卖功能扩展</li>
              </ul>
            </Card>
          </Col>
        </Row>

        <Card style={{ marginTop: 32, textAlign: 'center' }}>
          <Title level={3}>开始使用</Title>
          <Paragraph>
            探索我们的商城功能，体验现代化的购物体验
          </Paragraph>
          <div style={{ marginTop: 24 }}>
            <Button type="primary" size="large" style={{ marginRight: 16 }}>
              浏览商品
            </Button>
            <Button size="large">
              查看文档
            </Button>
          </div>
        </Card>
      </div>
    </MainLayout>
  );
}
