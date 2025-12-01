import { createRouter, createWebHistory } from 'vue-router'
import ProjectList from '../views/ProjectList.vue'
import ProjectDetail from '../views/ProjectDetail.vue'
import ExecutionHistory from '../views/ExecutionHistory.vue'
import ScheduleManager from '../views/ScheduleManager.vue'


// 定義路由設定
const router = createRouter({
    // 使用 HTML5 History 模式
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: ProjectList // 首頁顯示專案列表
        },
        {
            path: '/projects/:id',
            name: 'project-detail',
            component: ProjectDetail // 專案詳情頁
        },
        {
            path: '/projects/:id/executions',
            name: 'execution-history',
            component: ExecutionHistory // 執行歷史頁
        },
        {
            path: '/schedules',
            name: 'schedules',
            component: ScheduleManager // 全域排程管理
        },

    ]
})

export default router