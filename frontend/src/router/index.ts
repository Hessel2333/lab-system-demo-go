import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/views/MainLayout.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            component: MainLayout,
            redirect: '/dashboard',
            children: [
                {
                    path: 'dashboard',
                    // Lazy load is better practices
                    component: () => import('@/views/Dashboard.vue')
                },
                {
                    path: 'instruments',
                    component: () => import('@/views/Instruments.vue')
                },
                {
                    path: 'users',
                    component: () => import('@/views/UserManagement.vue')
                },
                {
                    path: 'suppliers',
                    name: 'suppliers',
                    component: () => import('@/views/SupplierManagement.vue')
                },
                // Placeholders for other modules
                { path: 'experiments', component: () => import('@/views/Dashboard.vue') }, // TODO
                { path: 'materials', component: () => import('@/views/Dashboard.vue') }, // TODO
                {
                    path: 'reagents',
                    component: () => import('@/views/ReagentManagement.vue')
                },
                { path: 'consumables', component: () => import('@/views/Dashboard.vue') }, // TODO
                { path: 'polymer', component: () => import('@/views/Dashboard.vue') }, // TODO
                { path: 'analysis', component: () => import('@/views/Dashboard.vue') }, // TODO
                { path: 'ai-center', component: () => import('@/views/Dashboard.vue') }, // TODO
            ]
        }
    ]
})

export default router
