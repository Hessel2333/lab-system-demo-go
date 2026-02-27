import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { watch } from 'vue'
import axios from 'axios'
import App from './App.vue'
import router from './router'
import { useSessionStore } from '@/stores/session'

import './assets/index.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const sessionStore = useSessionStore(pinia)
watch(
  () => sessionStore.currentUserId,
  (userId) => {
    axios.defaults.headers.common['X-User-ID'] = String(userId)
  },
  { immediate: true }
)

app.mount('#app')
