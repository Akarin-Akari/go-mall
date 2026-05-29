'use client';

import React from 'react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

interface AppleCardProps extends React.HTMLAttributes<HTMLDivElement> {
    children: React.ReactNode;
    noPadding?: boolean;
    hoverEffect?: boolean;
}

export function AppleCard({
    className,
    children,
    noPadding = false,
    hoverEffect = true,
    ...props
}: AppleCardProps) {
    return (
        <motion.div
            className={cn(
                'relative overflow-hidden bg-white text-black dark:bg-[#1c1c1e] dark:text-white',
                'rounded-3xl shadow-sm border border-black/5 dark:border-white/5',
                'transition-shadow duration-300',
                hoverEffect && 'hover:shadow-md cursor-pointer',
                noPadding ? 'p-0' : 'p-6',
                className
            )}
            whileHover={hoverEffect ? { scale: 1.02, y: -2 } : {}}
            transition={{ type: 'spring', stiffness: 300, damping: 20 }}
            {...props as any}
        >
            {/* Soft Light Hover Overlay */}
            {hoverEffect && (
                <div
                    className="absolute inset-0 bg-black/5 dark:bg-white/10 opacity-0 hover:opacity-100 transition-opacity duration-300 pointer-events-none z-10"
                    style={{ mixBlendMode: 'soft-light' }}
                />
            )}

            <div className="relative z-[1] h-full">
                {children}
            </div>
        </motion.div>
    );
}
