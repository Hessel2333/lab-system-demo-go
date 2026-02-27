import { defineStore } from 'pinia'

export type AppRole = 'researcher' | 'procurement' | 'leader'

interface RoleProfile {
  id: number
  name: string
  title: string
}

const ROLE_STORAGE_KEY = 'lab-system-role'

const ROLE_PROFILE_MAP: Record<AppRole, RoleProfile> = {
  researcher: { id: 1, name: '研发人员A', title: '研发人员' },
  procurement: { id: 2, name: '采购人员', title: '采购人员' },
  leader: { id: 101, name: '团队长', title: '团队负责人' },
}

const parseStoredRole = (): AppRole => {
  if (typeof window === 'undefined') return 'researcher'
  const raw = window.localStorage.getItem(ROLE_STORAGE_KEY)
  if (raw === 'researcher' || raw === 'procurement' || raw === 'leader') return raw
  return 'researcher'
}

export const useSessionStore = defineStore('session', {
  state: () => ({
    currentRole: parseStoredRole() as AppRole,
  }),
  getters: {
    currentUserId: (state) => ROLE_PROFILE_MAP[state.currentRole].id,
    currentUserName: (state) => ROLE_PROFILE_MAP[state.currentRole].name,
    currentUserTitle: (state) => ROLE_PROFILE_MAP[state.currentRole].title,
    currentRoleLabel: (state) => ROLE_PROFILE_MAP[state.currentRole].title,
  },
  actions: {
    setRole(role: AppRole) {
      this.currentRole = role
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(ROLE_STORAGE_KEY, role)
      }
    },
  },
})

