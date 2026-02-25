import request from '@/utils/request'

export interface Department {
    ID: number
    name: string
    type: string // institute, team, department
    children?: Department[]
    parent_id?: number
}

export interface User {
    ID: number
    username: string
    real_name: string
    role: string
    department: Department
    department_id: number
}

export const fetchDepartments = async (): Promise<Department[]> => {
    return request.get('/departments') as any as Promise<Department[]>
}

export const fetchUsers = async (departmentId?: number): Promise<User[]> => {
    return request.get('/users', {
        params: { department_id: departmentId }
    }) as any as Promise<User[]>
}

export const createUser = async (data: Partial<User>) => {
    return request.post<User>('/users', data)
}

export const updateUser = async (id: number, data: Partial<User>) => {
    return request.put<User>(`/users/${id}`, data)
}

export const deleteUser = async (id: number) => {
    return request.delete(`/users/${id}`)
}

export interface InstrumentPermission {
    instrument_id: number
    instrument_name: string
    category: string
    has_permission: boolean
}

export const fetchUserPermissions = async (userId: number): Promise<InstrumentPermission[]> => {
    return request.get(`/users/${userId}/permissions`) as any
}

export const updateUserPermission = async (userId: number, instrumentId: number, status: boolean) => {
    return request.post(`/users/${userId}/permissions`, {
        instrument_id: instrumentId,
        status: status
    })
}
