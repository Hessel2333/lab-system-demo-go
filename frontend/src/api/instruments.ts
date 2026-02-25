import request from '@/utils/request'
import type { User } from './organization'

export interface Department {
    ID: number
    name: string
}

export interface Supplier {
    ID: number
    name: string
    type: string
    contact_person: string
    rating: number
    status: string
}

export interface Instrument {
    ID: number
    name: string
    model: string
    brand: string
    status: string
    location: string
    purchase_date: string
    planning_date?: string
    procurement_date?: string
    admin: string
    run_time: number
    health: number
    reservations_count: number
    lifecycle_stage: string
    budget: number
    application_reason: string
    documents: any[]
    department?: Department
    supplier?: Supplier
}

const API_BASE = '/api'

export const fetchInstruments = async (): Promise<Instrument[]> => {
    return request.get('/instruments') as any
}

export const fetchInstrument = async (id: number): Promise<Instrument> => {
    return request.get<Instrument>(`/instruments/${id}`) as unknown as Promise<Instrument>
}

export const fetchAuthorizedUsers = async (instrumentId: number): Promise<User[]> => {
    return request.get(`/instruments/${instrumentId}/authorized_users`) as any
}

export const createInstrument = async (data: Partial<Instrument>): Promise<Instrument> => {
    const res = await fetch(`${API_BASE}/instruments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
    if (!res.ok) throw new Error('Failed to create instrument')
    return res.json()
}

export const updateInstrumentStatus = async (id: number, status: string): Promise<void> => {
    const res = await fetch(`${API_BASE}/instruments/${id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status })
    })
    if (!res.ok) throw new Error('Failed to update status')
}

export const updateInstrumentAdmin = async (id: number, adminName: string) => {
    return request.put(`/instruments/${id}/admin`, { admin: adminName })
}


