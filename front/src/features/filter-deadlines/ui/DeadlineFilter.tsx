import { useState } from 'react'
import type { Contract } from '@/entities/contract/model/types'
import styles from './DeadlineFilter.module.scss'

interface Props { contracts: Contract[] }

export function DeadlineFilter({ contracts }: Props) {
  const [before, setBefore] = useState('2026-04-01')
  const matches = contracts.filter((contract) => contract.deadline.value && contract.deadline.value < before)

  return <section className={styles.section}>
    <h2>Поставка до даты</h2>
    <label className={styles.field}>
      Раньше
      <input type="date" value={before} onChange={(event) => setBefore(event.target.value)} />
    </label>
    <div className={styles.results}>
      {matches.length === 0 && <p>Подходящих договоров нет.</p>}
      {matches.map((contract) => <p key={contract.id}>
        {contract.id} · {contract.deadline.value}
        {contract.notices.some((notice) => notice.kind === 'proposal' && notice.message.includes('срок поставки')) && ' · есть несогласованное изменение срока'}
      </p>)}
    </div>
  </section>
}
