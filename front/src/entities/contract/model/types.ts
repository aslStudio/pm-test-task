import type { Evidence } from '@/shared/model/evidence'

export type { Evidence } from '@/shared/model/evidence'

export interface ContractField {
  value: string
  evidence?: Evidence
  changed: boolean
}

export interface Notice {
  kind: 'proposal' | 'pending' | 'review'
  message: string
  evidence: Evidence
}

export interface Contract {
  id: string
  supplier: ContractField
  amount: ContractField
  deadline: ContractField
  notices: Notice[]
  review: boolean
}
