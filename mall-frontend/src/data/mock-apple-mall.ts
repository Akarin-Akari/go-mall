import { Product, Category } from '@/types';

// Using more reliable placeholder images that look professional or high-quality Unsplash ones
export const HERO_PRODUCTS = [
    {
        id: 1,
        name: 'iPhone 15 Pro',
        subtitle: 'Titanium. So strong. So light. So Pro.',
        price: 999,
        description: 'The first iPhone to feature an aerospace-grade titanium design.',
        // iPhone 15 Pro Titanium Black
        image: 'https://images.unsplash.com/photo-1696446701796-da61225697cc?q=80&w=2070&auto=format&fit=crop',
        tag: 'New'
    },
    {
        id: 2,
        name: 'MacBook Air 15"',
        subtitle: 'Impressively big. Impossibly thin.',
        price: 1299,
        description: 'Supercharged by M2 chip.',
        // MacBook Air nice shot
        image: 'https://images.unsplash.com/photo-1611186871348-b1ce696e52c9?q=80&w=2070&auto=format&fit=crop',
        tag: 'Trending'
    }
];

export const BENTO_CATEGORIES: (Category & { size: 'large' | 'medium' | 'small', image: string, color: string })[] = [
    {
        id: 1,
        name: 'Electronics',
        sort: 1,
        image: 'https://images.unsplash.com/photo-1550009158-9ebf69173e03?q=80&w=1000&auto=format&fit=crop', // Abstract Electronics
        size: 'large',
        color: 'from-blue-500/20 to-blue-600/20'
    },
    {
        id: 2,
        name: 'Fashion',
        sort: 2,
        image: 'https://images.unsplash.com/photo-1445205170230-053b83016050?q=80&w=1000&auto=format&fit=crop', // Fashion Coat
        size: 'medium',
        color: 'from-purple-500/20 to-pink-600/20'
    },
    {
        id: 3,
        name: 'Accessories',
        sort: 3,
        image: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?q=80&w=1000&auto=format&fit=crop', // Watch/Accessories
        size: 'small',
        color: 'from-orange-500/20 to-red-600/20'
    },
    {
        id: 4,
        name: 'Home',
        sort: 4,
        image: 'https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?q=80&w=1000&auto=format&fit=crop', // Living Room
        size: 'small',
        color: 'from-green-500/20 to-emerald-600/20'
    },
];

export const NEW_ARRIVALS = [
    {
        id: 101,
        name: 'AirPods Max',
        price: 549,
        image: 'https://images.unsplash.com/photo-1628202926206-c63a34b1618f?q=80&w=1000&auto=format&fit=crop',
        category: 'Audio'
    },
    {
        id: 102,
        name: 'Apple Watch Ultra',
        price: 799,
        // REPLACED BROKEN IMAGE with a reliable one (or generic smart watch)
        image: 'https://images.unsplash.com/photo-1434493789847-2f02dc6ca35d?q=80&w=1000&auto=format&fit=crop',
        category: 'Wearable'
    },
    {
        id: 103,
        name: 'Sony WH-1000XM5',
        price: 399,
        image: 'https://images.unsplash.com/photo-1618366712010-f4ae9c647dcb?q=80&w=1000&auto=format&fit=crop',
        category: 'Audio'
    },
    {
        id: 104,
        name: 'iPad Pro',
        price: 799,
        image: 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?q=80&w=1000&auto=format&fit=crop',
        category: 'Tablet'
    },
    {
        id: 105,
        name: 'HomePod mini',
        price: 99,
        image: 'https://images.unsplash.com/photo-1543512214-318c77a799d0?q=80&w=1000&auto=format&fit=crop',
        category: 'Home'
    }
];
