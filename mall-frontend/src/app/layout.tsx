import type { Metadata } from "next";
import { AntdRegistry } from '@ant-design/nextjs-registry';
import AppProviders from '@/components/providers/AppProviders';
import "./globals.css";

export const metadata: Metadata = {
  title: "Mall Frontend - Go商城前端应用",
  description: "基于React + Next.js + TypeScript构建的现代化商城前端应用",
  keywords: "商城,电商,React,Next.js,TypeScript",
  authors: [{ name: "Mall Team" }],
  viewport: "width=device-width, initial-scale=1",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>
        <AntdRegistry>
          <AppProviders>
            {children}
          </AppProviders>
        </AntdRegistry>
      </body>
    </html>
  );
}
