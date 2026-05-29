'use client';

import React, { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAppSelector } from '@/store';
import { selectAuth } from '@/store/slices/authSlice';
import { selectCartItemCount } from '@/store/slices/cartSlice';
import { useAuth } from '@/hooks/useAuth';
import { ROUTES } from '@/constants';
import { AppleButton } from '@/components/ui/apple/AppleButton';
import { ShoppingBag, User, Search, Menu as MenuIcon, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Badge, Dropdown, Avatar, Menu } from 'antd';
import { motion, AnimatePresence } from 'framer-motion';

interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const router = useRouter();
  const pathname = usePathname();
  const { logout, restoreAuthState } = useAuth();
  const { user, isAuthenticated } = useAppSelector(selectAuth);
  const cartItemCount = useAppSelector(selectCartItemCount);

  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    restoreAuthState();
    const handleScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [restoreAuthState]);

  const navLinks = [
    { name: 'Store', path: '/' },
    { name: 'Mac', path: '/categories/1' },
    { name: 'iPhone', path: '/products' },
    { name: 'Watch', path: '#' },
    { name: 'Accessories', path: '#' },
    { name: 'Support', path: '#' },
  ];

  const userMenu = {
    items: [
      { key: 'profile', label: 'Profile', icon: <User size={14} />, onClick: () => router.push(ROUTES.PROFILE) },
      { key: 'orders', label: 'Orders', onClick: () => router.push(ROUTES.USER_ORDERS) },
      { type: 'divider' as const },
      { key: 'logout', label: 'Logout', onClick: logout },
    ]
  };

  return (
    <div className="min-h-screen flex flex-col bg-[#f5f5f7]">
      {/* Glassmorphism Header */}
      <header
        className={cn(
          "fixed top-0 left-0 right-0 z-50 transition-all duration-300 border-b border-transparent",
          scrolled ? "bg-white/70 backdrop-blur-xl border-gray-200" : "bg-transparent h-[50px] md:h-[60px]"
        )}
      >
        <div className="max-w-[1024px] mx-auto px-4 h-12 flex items-center justify-between text-xs md:text-sm">

          {/* Mobile Menu Button */}
          <button className="md:hidden text-[#1d1d1f]" onClick={() => setMobileMenuOpen(!mobileMenuOpen)}>
            {mobileMenuOpen ? <X /> : <MenuIcon />}
          </button>

          {/* Logo (Apple Icon mockup) */}
          <div className="cursor-pointer" onClick={() => router.push('/')}>
            <svg className="w-5 h-5 fill-current text-[#1d1d1f]" viewBox="0 0 17 19" xmlns="http://www.w3.org/2000/svg">
              <path d="M14.9366 1.77621C14.0754 2.84651 12.8711 3.55113 11.6669 3.47958C11.5165 2.19154 12.043 0.97519 12.8711 0.259685C13.6232 -0.384282 14.9366 -0.169572 15.087 1.11471C15.1623 1.32942 15.087 1.56157 14.9366 1.77621ZM14.4849 13.5855C13.8077 14.5878 12.9464 15.8041 11.8174 15.8041C10.7639 15.8041 10.3877 15.1599 9.18379 15.1599C7.97987 15.1599 7.52835 15.7326 6.55018 15.8041C5.42151 15.8756 4.36802 14.6593 3.69085 13.657C2.26127 11.5818 1.13259 8.36142 2.63749 5.71363C3.39003 4.42559 4.74426 3.56672 6.02347 3.56672C7.07695 3.56672 8.05513 4.28253 8.65715 4.28253C9.25917 4.28253 10.463 3.42366 11.6669 3.42366C12.1183 3.42366 13.8447 3.42366 14.8611 4.78345C14.7858 4.85499 12.6454 6.07147 12.6454 8.71926C12.6454 10.9376 14.5601 12.2257 14.6353 12.2972C14.5601 12.4404 14.8611 13.1563 14.4849 13.5855Z" />
            </svg>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex flex-1 justify-center space-x-8 text-[#1d1d1f]/80">
            {navLinks.map(link => (
              <a
                key={link.name}
                href={link.path}
                className="hover:text-black transition-colors"
              >
                {link.name}
              </a>
            ))}
          </nav>

          {/* Right Icons */}
          <div className="flex items-center space-x-6 text-[#1d1d1f]/80">
            <Search className="w-4 h-4 cursor-pointer hover:text-black transition-colors hidden sm:block" />

            <div onClick={() => router.push('/cart')} className="cursor-pointer hover:text-black relative">
              <ShoppingBag className="w-4 h-4" />
              {cartItemCount > 0 && (
                <span className="absolute -top-2 -right-2 bg-black text-white text-[10px] w-4 h-4 flex items-center justify-center rounded-full">
                  {cartItemCount}
                </span>
              )}
            </div>

            {isAuthenticated ? (
              <Dropdown menu={userMenu} placement="bottomRight">
                <Avatar size="small" icon={<User size={14} />} className="cursor-pointer bg-transparent text-current hover:text-black" />
              </Dropdown>
            ) : (
              <AppleButton size="sm" variant="primary" onClick={() => router.push(ROUTES.LOGIN)}>Sign In</AppleButton>
            )}
          </div>
        </div>
      </header>

      {/* Mobile Menu Overlay */}
      <AnimatePresence>
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="fixed inset-0 top-12 z-40 bg-white/95 backdrop-blur-xl p-6 md:hidden flex flex-col space-y-6"
          >
            {navLinks.map(link => (
              <a key={link.name} href={link.path} className="text-xl font-medium text-[#1d1d1f] py-2 border-b border-gray-100">
                {link.name}
              </a>
            ))}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Content Spacer for fixed header */}
      <main className="flex-1 pt-12">
        {children}
      </main>

      {/* Simple Footer */}
      <footer className="bg-[#f5f5f7] border-t border-gray-200 py-12 px-4 text-xs text-[#86868b]">
        <div className="max-w-[1024px] mx-auto space-y-4">
          <div className="flex flex-col md:flex-row justify-between gap-4">
            <p>Copyright © 2026 Mall Inc. All rights reserved.</p>
            <div className="flex gap-4">
              <span>Privacy Policy</span>
              <span>Terms of Use</span>
              <span>Sales and Refunds</span>
              <span>Legal</span>
              <span>Site Map</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default MainLayout;
