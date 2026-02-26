import { ref } from 'vue'
import { toast } from 'vue-sonner'

const getErrorMessage = (e: any, fallback: string) => {
  return e?.response?.data?.error || e?.message || fallback
}

export const useActionFeedback = () => {
  const pendingMap = ref<Record<string, boolean>>({})

  const isPending = (key: string) => !!pendingMap.value[key]

  const runAction = async <T>(
    key: string,
    task: () => Promise<T>,
    options?: { successMessage?: string; errorMessage?: string }
  ): Promise<T> => {
    if (isPending(key)) {
      return Promise.reject(new Error('action already pending'))
    }
    pendingMap.value = { ...pendingMap.value, [key]: true }
    try {
      const result = await task()
      if (options?.successMessage) toast.success(options.successMessage)
      return result
    } catch (e: any) {
      toast.error(getErrorMessage(e, options?.errorMessage || '操作失败'))
      throw e
    } finally {
      const next = { ...pendingMap.value }
      delete next[key]
      pendingMap.value = next
    }
  }

  return { isPending, runAction }
}
