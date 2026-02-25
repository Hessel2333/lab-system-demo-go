<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'

import StatusUpdateDialog from './StatusUpdateDialog.vue'
import MaintenanceDialog from './MaintenanceDialog.vue'
import { Clock, Calendar, FileText, AlertTriangle, CheckCircle, Activity, History, Zap, User, Download, Upload, ShieldCheck, Search, Wrench } from 'lucide-vue-next'
import { fetchAuthorizedUsers, updateInstrumentAdmin, fetchInstrument } from '@/api/instruments'

import { fetchReservations, type Reservation } from '@/api/reservations'
import { fetchUsers } from '@/api/organization'
import type { Instrument } from '@/api/instruments'
import type { User as OrganizationUser } from '@/api/organization'

console.log('InstrumentDetailDialog setup called')

const props = defineProps<{
  open: boolean
  instrumentData?: Instrument | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
}>()

const activeTab = ref('overview')
const instrument = ref<any>({ stats: {} })
const authorizedUsers = ref<OrganizationUser[]>([])
const loadingUsers = ref(false)

// Data States
const allReservations = ref<Reservation[]>([])
const loadingReservations = ref(false)

// Dialog States
const showStatusUpdate = ref(false)
const showMaintenance = ref(false)
const showAdminDialog = ref(false)
const maintenanceMode = ref<'maintenance' | 'repair'>('maintenance')
const pendingStatus = ref('')

// Admin Selection State
const userList = ref<OrganizationUser[]>([])
const selectedAdmin = ref<OrganizationUser | null>(null)
const searchQuery = ref('')
const loadingUsersList = ref(false)

// Watch for data changes
watch(() => props.instrumentData, async (val) => {
    console.log('InstrumentData changed:', val)
    if (val) {
        // Initialize with passed data first (skeleton)
        instrument.value = {
            ...val,
            id: val.ID,
            image: 'https://placehold.co/600x400?text=Instrument',
            stats: {
                runTime: val.run_time,
                health: val.health,
                reservations: val.reservations_count
            }
        }
        
        // Fetch full details from backend to get Department/Supplier relations
        // which are not included in the list view API
        try {
            const fullData = await fetchInstrument(val.ID)
            instrument.value = {
                ...instrument.value,
                ...fullData,
                department: fullData.department,
                supplier: fullData.supplier
            }
        } catch (e) {
            console.error('Failed to fetch full instrument details', e)
        }

        // Mock documents if none exist
        if (!instrument.value.documents || instrument.value.documents.length === 0) {
// ...
            instrument.value.documents = [
                { name: '用户操作手册.pdf', type: 'pdf', url: '#', upload_date: '2024-01-15' },
                { name: '出厂检验证书.jpg', type: 'img', url: '#', upload_date: '2023-12-20' }
            ]
        }

        // Fetch user permissions/authorized users
        // Note: activeTab watcher handles this when tab is clicked, but we can also prefetch or lazy load
        if (activeTab.value === 'users') loadAuthorizedUsers()
        
        // Always fetch reservations/timeline data
        await loadReservationsData()
    }
}, { immediate: true })

watch(activeTab, (val) => {
    if (val === 'users' && props.instrumentData) {
        loadAuthorizedUsers()
    }
})

const loadReservationsData = async () => {
    if (!props.instrumentData) return
    loadingReservations.value = true
    try {
        // Fetch all reservations (no date limit for history)
        const res = await fetchReservations(props.instrumentData.ID)
        allReservations.value = res.sort((a, b) => new Date(b.start_time).getTime() - new Date(a.start_time).getTime())
    } catch(e) {
        console.error('Failed to load reservations', e)
    } finally {
        loadingReservations.value = false
    }
}

const loadAuthorizedUsers = async () => {
    if (!props.instrumentData) return
    loadingUsers.value = true
    try {
        authorizedUsers.value = await fetchAuthorizedUsers(props.instrumentData.ID)
    } catch (e) {
        console.error(e)
    } finally {
        loadingUsers.value = false
    }
}

const handleUpload = () => {
    alert('文件上传功能演示：此处将调用S3/MinIO上传接口')
}

// Action Handlers
const handleStartUse = () => {
    pendingStatus.value = 'in_use'
    showStatusUpdate.value = true
}

const handleRepair = () => {
    maintenanceMode.value = 'repair'
    showMaintenance.value = true
}

const handleScrap = () => {
    pendingStatus.value = 'retired'
    showStatusUpdate.value = true
}

const handleChangeAdmin = async () => {
    showAdminDialog.value = true
    loadingUsersList.value = true
    try {
        userList.value = await fetchUsers()
    } catch (e) {
        console.error(e)
    } finally {
        loadingUsersList.value = false
    }
}

const filteredUsers = computed(() => {
    if (!searchQuery.value) return userList.value
    const q = searchQuery.value.toLowerCase()
    return userList.value.filter(u => 
        u.real_name.toLowerCase().includes(q) || 
        u.username.toLowerCase().includes(q)
    )
})

const onStatusUpdated = () => {
    showStatusUpdate.value = false
    emit('refresh') 
    if (pendingStatus.value) instrument.value.status = pendingStatus.value
}

const onMaintenanceSubmit = (data: any) => {
    console.log('Maintenance submitted:', data)
    showMaintenance.value = false
    alert(maintenanceMode.value === 'maintenance' ? '保养记录已提交' : '报修申请已提交')
    // Verification: in real app, we would post to API here and then reload data
    // For now we just reload data (mocking the submission effect being instant if we had it)
}

const onAdminSubmit = async () => {
    if (!selectedAdmin.value) return
    try {
        await updateInstrumentAdmin(instrument.value.id, selectedAdmin.value.real_name)
        instrument.value.admin = selectedAdmin.value.real_name
        showAdminDialog.value = false
        alert('管理员变更成功')
    } catch (e) {
        alert('变更失败')
    }
}

// Lifecycle Step Detail Logic
const selectedStep = ref<any>(null)
const handleStepClick = (step: any) => {
    // Only open if step is completed or current or has data
    // For demo, we let them click "Arrival" specifically as it's the requested feature
    selectedStep.value = step
}

const lifecycleSteps = [
  { label: '采购规划', stage: 'planning' },
  { label: '采购实施', stage: 'procurement' },
  { label: '到货验收', stage: 'arrival' },
  { label: '投入使用', stage: 'active' },
  { label: '维护/维修', stage: 'maintenance' },
  { label: '报废处置', stage: 'retired' },
]

const getStepStatus = (stepStage: string, currentStage: string) => {
    const stages = ['planning', 'procurement', 'arrival', 'active', 'maintenance', 'retired']
    const stepIdx = stages.indexOf(stepStage)
    const currentIdx = stages.indexOf(currentStage)
    if (stepIdx < currentIdx) return 'completed'
    if (stepIdx === currentIdx) return 'current'
    return 'upcoming'
}

const timeline = computed(() => {
    const events = []
    
    // 1. Initial Purchase/Arrival
    if (props.instrumentData?.purchase_date && !props.instrumentData.purchase_date.startsWith('0001')) {
        events.push({
            id: 'init-1',
            date: props.instrumentData.purchase_date.split('T')[0], // Simple date
            user: props.instrumentData.admin || 'System',
            action: '到货验收',
            type: 'lifecycle',
            note: '设备到货，完成初步验收'
        })
    }

    // 2. Maintenance from API (Exclude daily usage)
    allReservations.value.forEach(res => {
        if (res.type === 'usage') return // Skip usage for timeline
        
        events.push({
            id: `res-${res.ID}`,
            date: new Date(res.start_time).toLocaleString('zh-CN', { hour12: false }),
            user: res.user_name || '未知用户',
            action: res.type === 'maintenance' ? '维护保养' : '重大事件',
            type: res.type,
            note: res.description || (res.type === 'maintenance' ? '定期维护' : '系统事件')
        })
    })

    // Sort by date desc
    return events.sort((a, b) => new Date(b.date as string).getTime() - new Date(a.date as string).getTime())
})

const reservationHistory = computed(() => {
    return allReservations.value
        .filter(r => r.type === 'usage')
        .map(r => ({
            id: r.ID,
            date: r.start_time.split('T')[0],
            time: `${new Date(r.start_time).getHours()}:00 - ${new Date(r.end_time).getHours()}:00`,
            project: r.description || '常规实验',
            status: new Date(r.end_time) < new Date() ? 'completed' : 'upcoming',
            user: r.user_name // Add user for display
        }))
})

const maintenanceHistory = computed(() => {
    return allReservations.value
        .filter(r => r.type === 'maintenance')
        .map(r => ({
            id: r.ID,
            date: r.start_time.split('T')[0],
            user: r.user_name,
            content: r.description || '例行维护',
            status: 'completed'
        }))
})

</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="" class="p-0 overflow-hidden" maxWidth="max-w-4xl">
         <!-- Hero Header -->
         <div class="bg-gradient-to-r from-gray-50 to-white border-b p-6 pb-8">
             <div class="flex gap-8 items-start">
                 <!-- Image -->
                 <div class="w-40 h-32 bg-white rounded-xl shadow-sm border border-gray-100 p-1 shrink-0">
                     <img :src="instrument.image" class="w-full h-full object-cover rounded-lg" />
                 </div>

                 <!-- Info -->
                 <div class="flex-1 pt-1">
                     <div class="flex justify-between items-start">
                         <div>
                             <h2 class="text-2xl font-bold text-gray-900 flex items-center gap-3">
                                 {{ instrument.name }}
                                 <Badge variant="success" class="text-xs px-2 py-0.5" v-if="instrument.status === 'active'">正常运行</Badge>
                                 <Badge variant="warning" class="text-xs px-2 py-0.5" v-else-if="instrument.status === 'maintenance'">维护中</Badge>
                                 <Badge variant="secondary" class="text-xs px-2 py-0.5" v-else>{{ instrument.status }}</Badge>
                             </h2>
                             <div class="text-gray-500 mt-2 flex items-center gap-4 text-sm">
                                 <span class="bg-gray-100 px-2 py-1 rounded text-gray-600 font-mono text-xs">ID: {{ instrument.id }}</span>
                                 <span>{{ instrument.brand }} · {{ instrument.model }}</span>
                                 <span class="flex items-center gap-1"><User class="w-3 h-3" /> {{ instrument.admin || '未分配' }}</span>
                             </div>
                         </div>
                         <div class="flex gap-3">
                             <Button class="bg-blue-600 hover:bg-blue-700 shadow-sm shadow-blue-200" @click="handleStartUse">开始使用</Button>
                             <Button variant="outline" class="text-gray-600" @click="handleRepair">报修</Button>
                         </div>
                     </div>
                     
                     <!-- Interactive Lifecycle Stepper -->
                     <div class="mt-8 relative">
                         <div class="absolute top-4 left-0 w-full h-0.5 bg-gray-100 -translate-y-1/2"></div>
                         <div class="relative flex justify-between z-10 w-[95%] mx-auto">
                             <div v-for="(step, idx) in lifecycleSteps" :key="idx" 
                                 class="flex flex-col items-center group cursor-pointer transition-transform hover:-translate-y-1"
                                 @click="handleStepClick(step)">
                                 <div class="w-8 h-8 rounded-full flex items-center justify-center border-2 transition-all duration-300 bg-white relative z-20"
                                     :class="{
                                         'border-blue-500 text-blue-500': getStepStatus(step.stage, instrument.lifecycle_stage) === 'current',
                                         'border-blue-500 bg-blue-500 text-white': getStepStatus(step.stage, instrument.lifecycle_stage) === 'completed',
                                         'border-gray-200 text-gray-300': getStepStatus(step.stage, instrument.lifecycle_stage) === 'upcoming'
                                     }">
                                     <CheckCircle class="w-4 h-4" v-if="getStepStatus(step.stage, instrument.lifecycle_stage) === 'completed'" />
                                     <div class="w-2.5 h-2.5 rounded-full bg-blue-500" v-else-if="getStepStatus(step.stage, instrument.lifecycle_stage) === 'current'"></div>
                                     <div class="w-2 h-2 rounded-full bg-gray-200" v-else></div>
                                 </div>
                                 <span class="mt-2 text-xs font-medium transition-colors border-b border-transparent group-hover:border-gray-300 pb-0.5"
                                     :class="{
                                         'text-blue-600': getStepStatus(step.stage, instrument.lifecycle_stage) === 'current',
                                         'text-gray-900': getStepStatus(step.stage, instrument.lifecycle_stage) === 'completed',
                                         'text-gray-400': getStepStatus(step.stage, instrument.lifecycle_stage) === 'upcoming'
                                     }">
                                     {{ step.label }}
                                 </span>
                                 <span class="text-[10px] text-gray-400 mt-0.5 font-mono" v-if="step.stage === 'arrival' && instrument.purchase_date && !instrument.purchase_date.startsWith('0001')">
                                     {{ instrument.purchase_date.split('T')[0] }}
                                 </span>
                                 <span class="text-[10px] text-gray-400 mt-0.5 font-mono" v-else-if="step.stage === 'planning' && instrument.planning_date && !instrument.planning_date.startsWith('0001')">
                                     {{ instrument.planning_date.split('T')[0] }}
                                 </span>
                                 <span class="text-[10px] text-gray-400 mt-0.5 font-mono" v-else-if="step.stage === 'procurement' && instrument.procurement_date && !instrument.procurement_date.startsWith('0001')">
                                     {{ instrument.procurement_date.split('T')[0] }}
                                 </span>
                             </div>
                         </div>
                     </div>
                 </div>
             </div>
         </div>

         <!-- Content Area -->
         <div class="p-6 bg-white min-h-[400px]">
             <!-- Tabs Navigation -->
             <div class="flex gap-2 mb-6 border-b pb-1">
                  <button v-for="tab in ['overview', 'maintenance', 'reservations', 'users', 'docs']" :key="tab"
                      @click="activeTab = tab"
                      class="px-4 py-2 rounded-t-lg text-sm font-medium transition-all relative top-[1px]"
                      :class="activeTab === tab 
                          ? 'text-blue-600 border-b-2 border-blue-500 bg-blue-50/50' 
                          : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'">
                      {{ { overview: '概览', maintenance: '维护记录', reservations: '预约记录', users: '授权人员', docs: '文档' }[tab] }}
                  </button>
             </div>

             <!-- Overview Tab -->
             <div v-if="activeTab === 'overview'" class="animate-in fade-in slide-in-from-bottom-4 duration-500">
                 <div class="grid grid-cols-4 gap-4 mb-8">
                     <div class="p-5 rounded-2xl bg-gradient-to-br from-blue-50 to-white border border-blue-100 shadow-sm relative overflow-hidden group hover:shadow-md transition-shadow">
                         <div class="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"><Clock class="w-16 h-16 text-blue-500" /></div>
                         <div class="text-blue-600 font-medium mb-1 flex items-center gap-2"><Clock class="w-4 h-4" /> 累计运行</div>
                         <div class="text-3xl font-bold text-gray-900">{{ instrument.stats.runTime }} <span class="text-base font-normal text-gray-500">小时</span></div>
                     </div>

                     <div class="p-5 rounded-2xl bg-gradient-to-br from-indigo-50 to-white border border-indigo-100 shadow-sm relative overflow-hidden group hover:shadow-md transition-shadow">
                         <div class="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"><History class="w-16 h-16 text-indigo-500" /></div>
                         <div class="text-indigo-600 font-medium mb-1 flex items-center gap-2"><History class="w-4 h-4" /> 采购日期</div>
                         <div class="text-xl font-bold text-gray-900 mt-2">{{ instrument.purchase_date && !instrument.purchase_date.startsWith('0001') ? instrument.purchase_date.split('T')[0] : 'N/A' }}</div>
                     </div>

                     <div class="p-5 rounded-2xl bg-gradient-to-br from-emerald-50 to-white border border-emerald-100 shadow-sm relative overflow-hidden group hover:shadow-md transition-shadow">
                         <div class="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"><Activity class="w-16 h-16 text-emerald-500" /></div>
                         <div class="text-emerald-600 font-medium mb-1 flex items-center gap-2"><Activity class="w-4 h-4" /> 健康度</div>
                         <div class="text-3xl font-bold text-gray-900">{{ instrument.stats.health }} <span class="text-base font-normal text-gray-500">%</span></div>
                     </div>

                     <div class="p-5 rounded-2xl bg-gradient-to-br from-purple-50 to-white border border-purple-100 shadow-sm relative overflow-hidden group hover:shadow-md transition-shadow">
                         <div class="absolute right-0 top-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity"><Calendar class="w-16 h-16 text-purple-500" /></div>
                         <div class="text-purple-600 font-medium mb-1 flex items-center gap-2"><Calendar class="w-4 h-4" /> 预约次数</div>
                         <div class="text-3xl font-bold text-gray-900">{{ instrument.stats.reservations }} <span class="text-base font-normal text-gray-500">次</span></div>
                     </div>
                 </div>

                 <div class="grid grid-cols-3 gap-8">
                     <div class="col-span-2">
                         <h3 class="font-bold text-gray-900 mb-4 flex items-center gap-2 text-lg">
                             <History class="w-5 h-5 text-gray-500" /> 最近动态
                         </h3>
                         <div class="space-y-4">
                             <div v-for="item in timeline.slice(0, 3)" :key="item.id" class="flex gap-4 group">
                                 <div class="flex flex-col items-center">
                                     <div class="w-2 h-2 rounded-full bg-gray-300 ring-4 ring-white group-hover:bg-blue-500 transition-colors"></div>
                                     <div class="w-px h-full bg-gray-100 my-1 group-last:hidden"></div>
                                 </div>
                                 <div class="pb-6 w-full">
                                     <div class="bg-white border rounded-xl p-4 shadow-sm group-hover:border-blue-200 transition-colors">
                                         <div class="flex justify-between items-start mb-1">
                                             <span class="font-bold text-gray-900">{{ item.action }}</span>
                                             <span class="text-xs text-gray-400 font-mono bg-gray-50 px-2 py-1 rounded">{{ item.date }}</span>
                                         </div>
                                         <p class="text-sm text-gray-600">{{ item.note }}</p>
                                         <div class="mt-2 flex items-center gap-2 text-xs text-gray-400">
                                             <User class="w-3 h-3" /> {{ item.user }}
                                         </div>
                                     </div>
                                 </div>
                             </div>
                         </div>
                     </div>

                     <div class="col-span-1">
                         <div class="bg-red-50 rounded-xl p-5 border border-red-100">
                             <h3 class="font-bold text-red-700 mb-3 flex items-center gap-2"><AlertTriangle class="w-4 h-4" /> 危险操作</h3>
                             <p class="text-xs text-red-600/80 mb-4 leading-relaxed">
                                 报废申请或管理员变更属于敏感操作，请确保您有相应的权限。所有操作将被记录。
                             </p>
                             <div class="space-y-3">
                                 <Button variant="outline" class="w-full justify-start text-red-600 hover:text-red-700 hover:bg-red-100 border-red-200" @click="handleScrap">
                                     <AlertTriangle class="w-4 h-4 mr-2" /> 申请报废
                                 </Button>
                                 <Button variant="outline" class="w-full justify-start text-gray-600 hover:bg-white text-left" @click="handleChangeAdmin">
                                     <User class="w-4 h-4 mr-2" /> 变更管理员
                                 </Button>
                             </div>
                         </div>
                     </div>
                 </div>
             </div>
             
             <!-- Other Tabs Placeholders -->
             <div v-else-if="activeTab ==='timeline'" class="py-10 text-center text-gray-500 animate-in fade-in">
                 <div class="max-w-md mx-auto relative border-l-2 border-gray-200 text-left pl-8 space-y-8">
                     <div v-for="item in timeline " :key="item.id" class="relative">
                          <div class="absolute -left-[39px] top-0 w-5 h-5 rounded-full bg-blue-100 border-4 border-white flex items-center justify-center">
                              <div class="w-2 h-2 rounded-full bg-blue-500"></div>
                          </div>
                          <div>
                              <div class="text-sm font-bold text-gray-900">{{ item.action }}</div>
                              <div class="text-xs text-gray-500 mb-2">{{ item.date }}</div>
                              <div class="bg-gray-50 rounded-lg p-3 text-sm text-gray-600 border">{{ item.note }}</div>
                          </div>
                     </div>
                 </div>
             </div>



             <div v-else-if="activeTab === 'users'" class="animate-in fade-in slide-in-from-bottom-2">
                   <div class="flex justify-between items-center mb-6">
                      <h3 class="font-semibold text-gray-900">已授权操作人员</h3>
                      <div class="text-xs text-gray-500">
                          仅显示持有有效上机证的人员
                      </div>
                  </div>
                  
                  <div v-if="loadingUsers" class="text-center py-10 text-gray-400">加载中...</div>
                  
                  <div v-else-if="authorizedUsers.length > 0" class="grid grid-cols-2 md:grid-cols-3 gap-4">
                      <div v-for="user in authorizedUsers" :key="user.ID" class="group flex items-center gap-3 p-3 rounded-xl border border-gray-100 hover:border-blue-200 hover:bg-blue-50/30 transition-all">
                          <div class="w-10 h-10 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-sm font-bold text-gray-600 shadow-sm group-hover:from-blue-100 group-hover:to-blue-200 group-hover:text-blue-700">
                              {{ user.real_name[0] }}
                          </div>
                          <div>
                              <div class="font-medium text-gray-900 group-hover:text-blue-700">{{ user.real_name }}</div>
                              <div class="text-xs text-gray-500 flex items-center gap-1">
                                  <ShieldCheck class="w-3 h-3 text-green-500" /> {{ user.username }}
                              </div>
                          </div>
                      </div>
                  </div>

                  <div v-else class="text-center py-16 bg-gray-50 rounded-xl border border-dashed">
                      <ShieldCheck class="w-10 h-10 mx-auto mb-2 text-gray-300" />
                      <p class="text-gray-400">暂无授权人员</p>
                  </div>
              </div>

             <div v-else-if="activeTab === 'docs'" class="animate-in fade-in slide-in-from-bottom-2 px-6 pb-6">
                 <div class="flex justify-between items-center mb-6 pt-4">
                     <h3 class="font-semibold text-gray-900">文档资料</h3>
                     <Button size="sm" @click="handleUpload">
                        <Upload class="w-4 h-4 mr-2" />
                        上传文档
                     </Button>
                 </div>
                 
                 <div v-if="instrument.documents && instrument.documents.length > 0" class="grid grid-cols-1 gap-3">
                     <div v-for="(doc, idx) in instrument.documents" :key="idx" class="flex items-center justify-between p-4 bg-gray-50 rounded-xl border border-gray-100 hover:border-blue-200 transition-colors group">
                         <div class="flex items-center gap-4">
                             <div class="w-10 h-10 rounded-lg bg-white border flex items-center justify-center shadow-sm">
                                 <FileText class="w-5 h-5 text-red-500" v-if="doc.type === 'pdf'" />
                                 <FileText class="w-5 h-5 text-blue-500" v-else-if="doc.type === 'word'" />
                                 <FileText class="w-5 h-5 text-purple-500" v-else-if="doc.type === 'img'" />
                                 <FileText class="w-5 h-5 text-gray-400" v-else />
                             </div>
                             <div>
                                 <div class="font-medium text-gray-900">{{ doc.name }}</div>
                                 <div class="text-xs text-gray-500">{{ doc.upload_date }}</div>
                             </div>
                         </div>
                         <Button variant="ghost" size="sm" class="text-gray-400 hover:text-blue-600">
                             <Download class="w-4 h-4" />
                         </Button>
                     </div>
                 </div>
                 <div v-else class="text-center py-10 text-gray-400 bg-gray-50 rounded-xl border border-dashed">
                     <FileText class="w-10 h-10 mx-auto mb-2 opacity-20" />
                     <p>暂无文档</p>
                 </div>
             </div>

             <div v-else-if="activeTab === 'reservations'" class="animate-in fade-in slide-in-from-bottom-2">
                 <div class="flex justify-between items-center mb-6">
                     <h3 class="font-semibold text-gray-900">预约历史记录</h3>
                     <Button size="sm" variant="outline">
                         <Download class="w-4 h-4 mr-2" /> 导出记录
                     </Button>
                 </div>

                 <div class="rounded-xl border bg-gray-50/50 overflow-hidden">
                     <table class="w-full text-sm text-left">
                         <thead class="bg-gray-100/50 border-b text-gray-500">
                             <tr>
                                 <th class="px-4 py-3 font-medium">日期</th>
                                 <th class="px-4 py-3 font-medium">时间段</th>
                                 <th class="px-4 py-3 font-medium">实验项目</th>
                                 <th class="px-4 py-3 font-medium">状态</th>
                             </tr>
                         </thead>
                         <tbody class="divide-y divide-gray-100">
                             <tr v-for="res in reservationHistory" :key="res.id" class="hover:bg-white transition-colors">
                                 <td class="px-4 py-3 text-gray-900">{{ res.date }}</td>
                                 <td class="px-4 py-3 font-mono text-gray-600">{{ res.time }}</td>
                                 <td class="px-4 py-3 text-gray-900">{{ res.project }}</td>
                                 <td class="px-4 py-3">
                                     <Badge v-if="res.status === 'completed'" variant="secondary" class="bg-green-100 text-green-700 hover:bg-green-100 border-green-200">已完成</Badge>
                                     <Badge v-else-if="res.status === 'upcoming'" variant="secondary" class="bg-blue-100 text-blue-700 hover:bg-blue-100 border-blue-200">待开始</Badge>
                                     <Badge v-else variant="secondary">已取消</Badge>
                                 </td>
                             </tr>
                         </tbody>
                     </table>
                     <div class="p-4 text-center text-xs text-gray-400 border-t">
                         仅显示最近 10 条记录
                     </div>
                 </div>
             </div>

             <div v-else-if="activeTab === 'maintenance'" class="animate-in fade-in slide-in-from-bottom-2">
                 <div class="flex justify-between items-center mb-6">
                     <h3 class="font-semibold text-gray-900">维护记录</h3>
                     <Button size="sm" variant="outline" @click="handleRepair">
                        <Wrench class="w-4 h-4 mr-2" /> 新增记录
                     </Button>
                 </div>

                 <div v-if="loadingReservations" class="text-center py-8 text-gray-400">加载中...</div>
                 <div v-else-if="maintenanceHistory.length > 0" class="space-y-4">
                     <div v-for="item in maintenanceHistory" :key="item.id" class="p-4 bg-gray-50 rounded-xl border border-gray-100 flex gap-4">
                         <div class="w-10 h-10 rounded-full bg-orange-100 flex items-center justify-center shrink-0">
                             <Wrench class="w-5 h-5 text-orange-600" />
                         </div>
                         <div class="flex-1">
                             <div class="flex justify-between items-start mb-1">
                                 <h4 class="font-bold text-gray-900">{{ item.content }}</h4>
                                 <span class="text-xs text-gray-500 font-mono">{{ item.date }}</span>
                             </div>
                             <p class="text-sm text-gray-600 mb-2">执行人: {{ item.user }}</p>
                         </div>
                     </div>
                 </div>
                 <div v-else class="text-center py-16 bg-gray-50 rounded-xl border border-dashed">
                     <Wrench class="w-10 h-10 mx-auto mb-2 text-gray-300" />
                     <p class="text-gray-400">暂无维护记录</p>
                 </div>
             </div>

             <div v-else class="h-64 flex items-center justify-center text-gray-400 flex-col gap-3 animate-in fade-in">
                 <Zap class="w-10 h-10 opacity-20" />
                 <span>该模块功能开发中...</span>
             </div>
         </div>

      <!-- Action Dialogs (Nested, using Teleport to body) -->
      <StatusUpdateDialog 
          :open="showStatusUpdate" 
          :instrument-id="instrument.id"
          :current-status="instrument.status"
          @close="showStatusUpdate = false"
          @update="onStatusUpdated"
      />
      
      <MaintenanceDialog 
          :open="showMaintenance" 
          :mode="maintenanceMode"
          @close="showMaintenance = false"
          @submit="onMaintenanceSubmit"
      />



      <Dialog :open="showAdminDialog" @close="showAdminDialog = false" title="变更设备管理员" maxWidth="max-w-md">
          <div class="h-96 flex flex-col px-6 py-4">
              <div class="mb-4 relative">
                  <Search class="w-4 h-4 absolute left-3 top-2.5 text-gray-400" />
                  <input 
                    v-model="searchQuery"
                    type="text" 
                    placeholder="搜索姓名或工号..." 
                    class="w-full pl-9 pr-4 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
                  />
              </div>

              <div class="flex-1 overflow-y-auto border border-gray-100 rounded-lg">
                   <div v-if="loadingUsersList" class="py-10 text-center text-gray-400 text-sm">加载人员中...</div>
                   <div v-else class="divide-y divide-gray-50">
                       <div 
                        v-for="user in filteredUsers" 
                        :key="user.ID" 
                        class="p-3 flex items-center justify-between hover:bg-blue-50 cursor-pointer transition-colors"
                        :class="{'bg-blue-50': selectedAdmin?.ID === user.ID}"
                        @click="selectedAdmin = user"
                       >
                           <div class="flex items-center gap-3">
                               <div class="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-600">
                                   {{ user.real_name[0] }}
                               </div>
                               <div>
                                   <div class="text-sm font-medium text-gray-900">{{ user.real_name }}</div>
                                   <div class="text-xs text-gray-500">{{ user.username }} · {{ user.role }}</div>
                               </div>
                           </div>
                           <div v-if="selectedAdmin?.ID === user.ID">
                               <CheckCircle class="w-5 h-5 text-blue-600" />
                           </div>
                       </div>
                   </div>
              </div>
          </div>
          <template #footer>
                <div class="flex items-center justify-between w-full">
                    <span class="text-xs text-gray-500" v-if="selectedAdmin">
                        已选择: <span class="font-bold text-gray-900">{{ selectedAdmin.real_name }}</span>
                    </span>
                    <span class="text-xs text-gray-500" v-else>请选择一位人员</span>
                    
                    <div class="flex gap-2">
                        <Button variant="ghost" @click="showAdminDialog = false">取消</Button>
                        <Button @click="onAdminSubmit" :disabled="!selectedAdmin">确认变更</Button>
                    </div>
                </div>
          </template>
      </Dialog>
    </Dialog>

    <!-- Lifecycle Record Detail Dialog (Nested) -->
    <Dialog :open="!!selectedStep" @close="selectedStep = null" maxWidth="max-w-xl">
        <div v-if="selectedStep" class="p-6">
             <div class="flex justify-between items-start mb-6 border-b border-gray-100 pb-4">
                 <div>
                     <span class="text-xs font-bold text-gray-400 uppercase tracking-wider">生命周期记录</span>
                     <h3 class="text-xl font-bold text-gray-900 mt-1">{{ selectedStep.label }}详情单</h3>
                 </div>
                 <Badge variant="outline" class="uppercase">{{ getStepStatus(selectedStep.stage, instrument.lifecycle_stage) === 'completed' ? '已归档' : '进行中' }}</Badge>
             </div>

             <!-- Mock Form Content based on Stage -->
             <div v-if="selectedStep.stage === 'planning'" class="space-y-6">
                 <div class="grid grid-cols-2 gap-6">
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">立项日期</label>
                         <p class="font-mono text-sm font-semibold">{{ instrument.planning_date?.split('T')[0] || 'N/A' }}</p>
                     </div>
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">申请部门</label>
                         <p class="text-sm font-semibold">{{ instrument.department?.name || '未知部门' }}</p>
                     </div>
                 </div>
                 <div class="bg-gray-50 p-4 rounded-lg border border-gray-100">
                     <h4 class="text-sm font-bold text-gray-900 mb-2">采购必要性说明</h4>
                     <p class="text-xs text-gray-600 leading-relaxed">{{ instrument.application_reason || '随着科研任务增加，现有设备无法满足高通量筛选需求，急需引进该设备以提升效率。' }}</p>
                 </div>
                 <div>
                     <h4 class="text-sm font-bold text-gray-900 mb-3">关联单据</h4>
                     <div class="flex items-center gap-2 px-3 py-2 bg-blue-50 text-blue-700 rounded-lg text-sm border border-blue-100 cursor-pointer hover:bg-blue-100 transition-colors w-fit">
                         <FileText class="w-4 h-4" />
                         设备采购申请表.pdf
                     </div>
                 </div>
             </div>

             <div v-else-if="selectedStep.stage === 'procurement'" class="space-y-6">
                 <div class="grid grid-cols-2 gap-6">
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">采购实施日期</label>
                         <p class="font-mono text-sm font-semibold">{{ instrument.procurement_date?.split('T')[0] || 'N/A' }}</p>
                     </div>
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">预算金额</label>
                         <p class="text-sm font-semibold font-mono">¥ {{ (instrument.budget || 250000).toLocaleString() }}</p>
                     </div>
                 </div>
                 <div class="space-y-3">
                     <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-100">
                         <span class="text-sm text-gray-600">招标方式</span>
                         <span class="text-sm font-bold text-gray-900">公开招标</span>
                     </div>
                     <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-100">
                         <span class="text-sm text-gray-600">中标供应商</span>
                         <span class="text-sm font-bold text-gray-900">{{ instrument.supplier?.name || '待定' }}</span>
                     </div>
                 </div>
                 <div>
                     <h4 class="text-sm font-bold text-gray-900 mb-3">关联单据</h4>
                     <div class="flex gap-3">
                         <div class="flex items-center gap-2 px-3 py-2 bg-blue-50 text-blue-700 rounded-lg text-sm border border-blue-100 cursor-pointer hover:bg-blue-100 transition-colors">
                             <FileText class="w-4 h-4" />
                             中标通知书.pdf
                         </div>
                         <div class="flex items-center gap-2 px-3 py-2 bg-blue-50 text-blue-700 rounded-lg text-sm border border-blue-100 cursor-pointer hover:bg-blue-100 transition-colors">
                            <FileText class="w-4 h-4" />
                            采购合同.pdf
                        </div>
                     </div>
                 </div>
             </div>

             <div v-else-if="selectedStep.stage === 'arrival'" class="space-y-6">
                 <div class="grid grid-cols-2 gap-6">
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">验收日期</label>
                         <p class="font-mono text-sm font-semibold">{{ instrument.purchase_date?.split('T')[0] || 'N/A' }}</p>
                     </div>
                     <div>
                         <label class="block text-xs font-medium text-gray-500 mb-1">验收负责人</label>
                         <div class="flex items-center gap-2">
                            <div class="w-5 h-5 rounded-full bg-gray-200 flex items-center justify-center text-[10px]">
                                {{ instrument.admin?.[0] || 'S' }}
                            </div>
                            <span class="text-sm font-medium">{{ instrument.admin || 'System' }}</span>
                         </div>
                     </div>
                 </div>

                 <div class="bg-gray-50 p-4 rounded-lg border border-gray-100">
                     <h4 class="text-sm font-bold text-gray-900 mb-3 flex items-center gap-2">
                         <CheckCircle class="w-4 h-4 text-green-600" /> 验收检查项
                     </h4>
                     <ul class="space-y-2">
                         <li class="flex items-center gap-2 text-sm text-gray-600">
                             <input type="checkbox" checked readonly class="rounded text-blue-600 focus:ring-0 cursor-default" /> 
                             外观包装完好无损
                         </li>
                         <li class="flex items-center gap-2 text-sm text-gray-600">
                             <input type="checkbox" checked readonly class="rounded text-blue-600 focus:ring-0 cursor-default" /> 
                             配件/说明书齐全
                         </li>
                         <li class="flex items-center gap-2 text-sm text-gray-600">
                             <input type="checkbox" checked readonly class="rounded text-blue-600 focus:ring-0 cursor-default" /> 
                             通电自检正常
                         </li>
                     </ul>
                 </div>

                 <div>
                     <h4 class="text-sm font-bold text-gray-900 mb-3">关联单据</h4>
                     <div class="flex gap-3">
                         <div class="flex items-center gap-2 px-3 py-2 bg-blue-50 text-blue-700 rounded-lg text-sm border border-blue-100 cursor-pointer hover:bg-blue-100 transition-colors">
                             <FileText class="w-4 h-4" />
                             出厂检验证书.jpg
                         </div>
                         <div class="flex items-center gap-2 px-3 py-2 bg-gray-50 text-gray-700 rounded-lg text-sm border border-gray-200 cursor-pointer hover:bg-gray-100 transition-colors">
                             <FileText class="w-4 h-4" />
                             验收确认单.pdf
                         </div>
                     </div>
                 </div>
             </div>
             
             <div v-else-if="selectedStep.stage === 'active'" class="space-y-4">
                 <div class="bg-green-50 p-4 rounded-xl border border-green-100 text-center">
                     <Activity class="w-8 h-8 text-green-600 mx-auto mb-2" />
                     <h4 class="font-bold text-green-800">设备运行中</h4>
                     <p class="text-xs text-green-600 mt-1">当前状态正常，可接受预约</p>
                 </div>
                 <div class="grid grid-cols-2 gap-4">
                     <div class="p-3 bg-gray-50 rounded-lg">
                         <div class="text-gray-500 text-xs">累计运行时长</div>
                         <div class="font-mono font-bold text-lg">{{ instrument.stats.runTime }}h</div>
                     </div>
                     <div class="p-3 bg-gray-50 rounded-lg">
                         <div class="text-gray-500 text-xs">服务次数</div>
                         <div class="font-mono font-bold text-lg">{{ instrument.stats.reservations }}次</div>
                     </div>
                 </div>
             </div>

             <div v-else class="py-12 text-center text-gray-400">
                 <FileText class="w-12 h-12 mx-auto mb-3 opacity-20" />
                 <p>该阶段暂无电子化详细记录</p>
             </div>
        </div>
    </Dialog>
</template>
```
