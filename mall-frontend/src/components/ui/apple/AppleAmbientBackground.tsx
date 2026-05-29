'use client';

import React from 'react';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';

interface AppleAmbientBackgroundProps extends React.HTMLAttributes<HTMLDivElement> {
    src?: string; // Optional background image to blur
    fallbackColor?: string; // Fallback gradient color
}

export function AppleAmbientBackground({
    className,
    src,
    fallbackColor = '#007AFF', // Apple Blue default
    ...props
}: AppleAmbientBackgroundProps) {
    return (
        <div
            className={cn(
                'absolute inset-0 z-0 overflow-hidden pointer-events-none',
                className
            )}
            {...props}
        >
            {/* Layer 1: Base Gradient / Veil */}
            <div
                className="absolute inset-0 z-[1] bg-white/30 dark:bg-black/30 backdrop-blur-3xl"
                style={{ mixBlendMode: 'overlay' }}
            />

            {/* Layer 2: The Animated Ambient Source */}
            <div className="absolute inset-[-50%] w-[200%] h-[200%] z-[0]">
                <div
                    className="absolute inset-0 w-full h-full opacity-60"
                    style={{
                        backgroundImage: src ? `url(${src})` : `linear-gradient(to bottom right, ${fallbackColor}, #e0e0e0)`,
                        backgroundSize: 'cover',
                        backgroundPosition: 'center',
                        filter: 'blur(80px) saturate(1.8)',
                        animation: 'var(--animate-shift-background)'
                    }}
                />
            </div>

            {/* Layer 3: Fade to Page Background (Bottom Mask) */}
            <div className="absolute inset-0 z-[2] bg-gradient-to-b from-transparent via-transparent to-[var(--background)]" />
        </div>
    );
}
