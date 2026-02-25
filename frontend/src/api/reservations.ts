
import request from '@/utils/request'

export interface Reservation {
    ID: number // Changed from string to number to match GORM
    instrument_id: number
    user_id: string
    user_name: string
    start_time: string // ISO string
    end_time: string   // ISO string
    type: 'usage' | 'maintenance'
    description: string
    status: string
}

// Helper to type the unwrapped response
// Since our interceptor returns response.data, we need to tell TS that request.* returns T
export const fetchReservations = async (instrumentId: string | number, start?: Date, end?: Date): Promise<Reservation[]> => {
    return request.get('/reservations', {
        params: {
            instrument_id: instrumentId,
            start: start ? start.toISOString() : undefined,
            end: end ? end.toISOString() : undefined
        }
    }) as any as Promise<Reservation[]>
}

export const createReservation = async (data: Omit<Reservation, 'ID' | 'status' | 'user_id' | 'user_name'>): Promise<Reservation> => {
    return request.post('/reservations', data) as any as Promise<Reservation>
}

export const cancelReservation = async (id: number): Promise<void> => {
    return request.delete(`/reservations/${id}`) as any as Promise<void>
}
