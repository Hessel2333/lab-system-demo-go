<script setup lang="ts">
import { ref, watch } from 'vue'
import { X, Shield, Database, Lock } from 'lucide-vue-next'
import type { User } from '@/api/organization'
import {
  fetchUserPermissions,
  updateUserPermission,
  fetchUserReagentPermissions,
  updateUserReagentPermissions,
  type InstrumentPermission
} from '@/api/organization'

const props = defineProps<{
  modelValue: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const activeTab = ref('instruments')
const permissions = ref<InstrumentPermission[]>([])
const loading = ref(false)
const reagentPermLoading = ref(false)
const savingReagentPerm = ref(false)
const reagentPerm = ref({
  is_dispense_key_holder_a: false,
  is_dispense_key_holder_b: false,
})

const loadPermissions = async () => {
    if (!props.user) return
    loading.value = true
    try {
        permissions.value = await fetchUserPermissions(props.user.ID)
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const loadReagentPermissions = async () => {
  if (!props.user) return
  reagentPermLoading.value = true
  try {
    const data = await fetchUserReagentPermissions(props.user.ID)
    reagentPerm.value = {
      is_dispense_key_holder_a: !!data.is_dispense_key_holder_a,
      is_dispense_key_holder_b: !!data.is_dispense_key_holder_b,
    }
  } catch (e) {
    console.error(e)
  } finally {
    reagentPermLoading.value = false
  }
}

watch(() => props.modelValue, (val) => {
    if (val && props.user) {
        loadPermissions()
        loadReagentPermissions()
        activeTab.value = 'instruments'
    }
})

const togglePermission = async (perm: InstrumentPermission) => {
    if (!props.user) return
    const newStatus = !perm.has_permission
    try {
        // Optimistic update
        perm.has_permission = newStatus
        await updateUserPermission(props.user.ID, perm.instrument_id, newStatus)
    } catch (e) {
        // Revert on failure
        perm.has_permission = !newStatus
        console.error(e)
    }
}

const saveReagentPermissions = async () => {
  if (!props.user) return
  if (reagentPerm.value.is_dispense_key_holder_a && reagentPerm.value.is_dispense_key_holder_b) {
    alert('同一用户不能同时担任A/B双签持有人')
    return
  }

  savingReagentPerm.value = true
  try {
    await updateUserReagentPermissions(props.user.ID, reagentPerm.value)
  } catch (e) {
    console.error(e)
    alert('保存失败')
  } finally {
    savingReagentPerm.value = false
  }
}
</script>

<template>
  <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" style="z-index: 100;">
    <div class="fixed inset-0 bg-gray-900/30 backdrop-blur-sm transition-opacity" @click="emit('update:modelValue', false)"></div>

    <div class="relative w-full max-w-lg transform overflow-hidden rounded-xl bg-white flex flex-col max-h-[85vh] shadow-2xl transition-all sm:my-8 border border-gray-100">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
          <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center text-blue-600">
                  <Shield class="w-5 h-5" />
              </div>
              <div>
                  <h3 class="text-lg font-bold text-gray-900">权限管理</h3>
                  <p class="text-sm text-gray-500" v-if="user">正在配置 <span class="font-medium text-gray-900">{{ user.real_name }}</span> 的系统权限</p>
              </div>
          </div>
          <button @click="emit('update:modelValue', false)" class="text-gray-400 hover:text-gray-500 transition-colors p-1 rounded-full hover:bg-gray-100">
              <X class="h-5 w-5" />
          </button>
      </div>

      <!-- Tabs -->
      <div class="flex border-b border-gray-100 px-6 gap-6">
          <button 
            @click="activeTab = 'instruments'"
            class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2"
            :class="activeTab === 'instruments' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
          >
              <Shield class="w-4 h-4" /> 仪器预约权限
          </button>
           <button 
            @click="activeTab = 'data'"
            class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2"
            :class="activeTab === 'data' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
          >
              <Database class="w-4 h-4" /> 数据访问权限
          </button>
          <button 
            @click="activeTab = 'role'"
            class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2"
            :class="activeTab === 'role' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
          >
              <Lock class="w-4 h-4" /> 试剂双签角色
          </button>
      </div>
      
      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-0 min-h-[300px]">
          
          <!-- Instruments Tab -->
          <div v-if="activeTab === 'instruments'">
              <div v-if="loading" class="flex items-center justify-center h-40 text-gray-400 text-sm">
                  加载中...
              </div>
              <div v-else class="divide-y divide-gray-100">
                  <div v-for="perm in permissions" :key="perm.instrument_id" class="flex items-center justify-between px-6 py-4 hover:bg-gray-50 transition-colors">
                      <div>
                          <h4 class="text-sm font-medium text-gray-900">{{ perm.instrument_name }}</h4>
                          <p class="text-xs text-gray-500 mt-0.5">ID: {{ perm.instrument_id }}</p>
                      </div>
                      
                      <button 
                        @click="togglePermission(perm)"
                        class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perm.has_permission ? 'bg-green-500' : 'bg-gray-200'"
                      >
                            <span 
                                class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                                :class="perm.has_permission ? 'translate-x-5' : 'translate-x-0'"
                            />
                      </button>
                  </div>
                
                  <div v-if="permissions.length === 0 && !loading" class="text-center py-12 text-gray-400 text-sm">
                      暂无仪器数据
                  </div>
              </div>
          </div>

          <!-- Data Tab (Placeholder) -->
          <div v-else-if="activeTab === 'data'" class="flex flex-col items-center justify-center h-64 text-gray-400 gap-3">
              <div class="w-12 h-12 rounded-full bg-gray-50 flex items-center justify-center">
                  <Lock class="w-6 h-6 opacity-20" />
              </div>
              <p class="text-sm">该模块尚未开放配置</p>
          </div>
          <!-- Reagent Role Tab -->
          <div v-else class="p-6 space-y-4">
              <div class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
                双签持有人为全局唯一角色。开启后将自动替换原持有人。
              </div>

              <div v-if="reagentPermLoading" class="py-8 text-sm text-gray-400 text-center">加载中...</div>

              <div v-else class="space-y-3">
                <div class="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
                  <div>
                    <div class="text-sm font-medium text-gray-900">钥匙持有人 A</div>
                    <div class="text-xs text-gray-500">建议配置为研发团队C团队长</div>
                  </div>
                  <button
                    @click="reagentPerm.is_dispense_key_holder_a = !reagentPerm.is_dispense_key_holder_a; if (reagentPerm.is_dispense_key_holder_a) reagentPerm.is_dispense_key_holder_b = false"
                    class="relative inline-flex h-6 w-11 rounded-full border-2 border-transparent transition-colors"
                    :class="reagentPerm.is_dispense_key_holder_a ? 'bg-green-500' : 'bg-gray-200'"
                  >
                    <span
                      class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition"
                      :class="reagentPerm.is_dispense_key_holder_a ? 'translate-x-5' : 'translate-x-0'"
                    />
                  </button>
                </div>

                <div class="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
                  <div>
                    <div class="text-sm font-medium text-gray-900">钥匙持有人 B</div>
                    <div class="text-xs text-gray-500">建议配置为采购人员</div>
                  </div>
                  <button
                    @click="reagentPerm.is_dispense_key_holder_b = !reagentPerm.is_dispense_key_holder_b; if (reagentPerm.is_dispense_key_holder_b) reagentPerm.is_dispense_key_holder_a = false"
                    class="relative inline-flex h-6 w-11 rounded-full border-2 border-transparent transition-colors"
                    :class="reagentPerm.is_dispense_key_holder_b ? 'bg-green-500' : 'bg-gray-200'"
                  >
                    <span
                      class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition"
                      :class="reagentPerm.is_dispense_key_holder_b ? 'translate-x-5' : 'translate-x-0'"
                    />
                  </button>
                </div>

                <div class="pt-1">
                  <button
                    @click="saveReagentPermissions"
                    :disabled="savingReagentPerm"
                    class="h-9 rounded-lg bg-black px-4 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-60"
                  >
                    {{ savingReagentPerm ? '保存中...' : '保存双签角色' }}
                  </button>
                </div>
              </div>
          </div>

      </div>
      
      <!-- Footer -->
      <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end">
          <button class="bg-white border border-gray-300 text-gray-700 px-4 py-2 rounded-lg font-medium hover:bg-gray-50 transition-all shadow-sm" @click="emit('update:modelValue', false)">
              完成
          </button>
      </div>
    </div>
  </div>
</template>
