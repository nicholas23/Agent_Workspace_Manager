import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

// 建立 Vue 應用程式實例
const app = createApp(App)

// 使用 Pinia 狀態管理
app.use(createPinia())
// 使用 Vue Router
app.use(router)
// 使用 Element Plus UI 庫
app.use(ElementPlus)

// 掛載應用程式到 DOM
app.mount('#app')
