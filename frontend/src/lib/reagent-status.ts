export type UiStatusVariant = 'default' | 'secondary' | 'warning' | 'info' | 'success' | 'destructive' | 'outline'

export const isArrivedStatus = (status?: string) => status === '已到货' || status === 'Arrived'

export const isInStorageStatus = (status?: string) => status === '在库' || status === 'InStorage'

export const isUsedStatus = (status?: string) => status === '已耗尽' || status === 'Used'

export const getInventoryDisplayStatus = (status?: string) => {
  if (isArrivedStatus(status)) return '已到货'
  if (isInStorageStatus(status)) return '已入库'
  if (isUsedStatus(status)) return '已耗尽'
  return status || '未知'
}

export const getInventoryStatusVariant = (status?: string): UiStatusVariant => {
  const display = getInventoryDisplayStatus(status)
  if (display === '已到货') return 'warning'
  if (display === '已入库') return 'success'
  if (display === '已耗尽') return 'outline'
  return 'default'
}

export const getProcurementReceiveDisplayStatus = (status?: string) => {
  if (status === '待收货') return '待到货'
  if (status === '部分收货') return '待到货'
  if (status === '已收货') return '已到货'
  return status || '待到货'
}

export const getProcurementReceiveStatusVariant = (status?: string): UiStatusVariant => {
  const display = getProcurementReceiveDisplayStatus(status)
  if (display === '待到货') return 'warning'
  if (display === '已到货') return 'success'
  return 'secondary'
}
