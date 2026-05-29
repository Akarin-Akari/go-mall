'use client';

import React from 'react';
import { cn } from '@/lib/utils';
import { motion, HTMLMotionProps } from 'framer-motion';

interface AppleButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'ghost' | 'glass';
    size?: 'sm' | 'md' | 'lg' | 'icon';
    isLoading?: boolean;
}

export const AppleButton = React.forwardRef<HTMLButtonElement, AppleButtonProps>(
    ({ className, variant = 'primary', size = 'md', isLoading, children, disabled, ...props }, ref) => {

        const variants = {
            primary: 'bg-[#007AFF] text-white hover:bg-[#0071e3] shadow-sm',
            secondary: 'bg-[#e8e8ed] text-[#1d1d1f] hover:bg-[#d2d2d7] dark:bg-[#3a3a3c] dark:text-white',
            ghost: 'bg-transparent text-[#007AFF] hover:bg-[#007AFF]/10',
            glass: 'bg-white/70 backdrop-blur-md text-[#1d1d1f] border border-white/20 hover:bg-white/90',
        };

        const sizes = {
            sm: 'h-8 px-4 text-xs',
            md: 'h-10 px-6 text-sm',
            lg: 'h-12 px-8 text-base font-medium',
            icon: 'h-10 w-10 p-2 flex items-center justify-center',
        };

        return (
            <motion.button
                ref={ref as any}
                className={cn(
                    'inline-flex items-center justify-center rounded-full font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#007AFF] disabled:opacity-50 disabled:pointer-events-none',
                    variants[variant],
                    sizes[size],
                    className
                )}
                whileTap={{ scale: 0.96 }}
                disabled={disabled || isLoading}
                {...props as any}
            >
                {isLoading ? (
                    <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2" />
                ) : null}
                {children}
            </motion.button>
        );
    }
);

AppleButton.displayName = 'AppleButton';
