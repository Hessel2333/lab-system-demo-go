import axios from 'axios'

const service = axios.create({
    baseURL: '/api', // Proxy is set up in vite.config.ts usually, or use full URL
    timeout: 5000
})

service.interceptors.response.use(
    (response: any) => {
        return response.data
    },
    (error: any) => {
        return Promise.reject(error)
    }
)

export default service
