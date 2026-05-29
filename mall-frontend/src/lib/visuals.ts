import { Category, Product } from '@/types';

// Deterministic visual mapping for categories
export const getCategoryVisual = (id: number) => {
    const visuals = [
        { color: 'from-blue-500/20 to-blue-600/20', image: 'https://images.unsplash.com/photo-1550009158-9ebf69173e03?q=80&w=1000&auto=format&fit=crop' }, // Electronics
        { color: 'from-purple-500/20 to-pink-600/20', image: 'https://images.unsplash.com/photo-1445205170230-053b83016050?q=80&w=1000&auto=format&fit=crop' }, // Fashion
        { color: 'from-orange-500/20 to-red-600/20', image: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?q=80&w=1000&auto=format&fit=crop' }, // Accessories
        { color: 'from-green-500/20 to-emerald-600/20', image: 'https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?q=80&w=1000&auto=format&fit=crop' }, // Home
        { color: 'from-yellow-500/20 to-amber-600/20', image: 'https://images.unsplash.com/photo-1544367563-12123d8966bf?q=80&w=1000&auto=format&fit=crop' }, // Books
    ];
    return visuals[id % visuals.length];
};

export const getBentoSize = (index: number) => {
    const pattern = ['large', 'medium', 'small', 'small']; // 4-item pattern
    return pattern[index % pattern.length] as 'large' | 'medium' | 'small';
};
