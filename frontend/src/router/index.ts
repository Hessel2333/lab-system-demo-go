import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/views/MainLayout.vue'
import UnderConstruction from '@/views/UnderConstruction.vue'

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
                    path: 'users/permission-settings',
                    redirect: { path: '/users', query: { view: 'policy' } }
                },
                {
                    path: 'master-data',
                    component: () => import('@/views/MasterDataCenter.vue')
                },
                {
                    path: 'suppliers',
                    name: 'suppliers',
                    component: () => import('@/views/SupplierManagement.vue')
                },
                // Placeholders for other modules
                { path: 'experiments', component: UnderConstruction, props: { moduleName: '实验管理' } },
                { path: 'materials', redirect: '/master-data' },
                {
                    path: 'reagents',
                    component: () => import('@/views/ReagentManagement.vue')
                },
                { path: 'consumables', component: UnderConstruction, props: { moduleName: '耗材管理' } },
                { path: 'polymer', component: UnderConstruction, props: { moduleName: '聚合物数据库' } },
                { path: 'analysis', component: UnderConstruction, props: { moduleName: '基因库分析' } },
                { path: 'ai-center', component: UnderConstruction, props: { moduleName: 'AI 智能中心' } },
            ]
        }
    ]
})

export default router
