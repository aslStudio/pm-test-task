import type { Contract, Evidence } from '../model/types'
import { ContractField } from './ContractField'
import styles from './ContractTable.module.scss'

interface Props {
  contracts: Contract[]
  asOf: string
  onEvidence: (evidence: Evidence) => void
}

export function ContractTable({ contracts, asOf, onEvidence }: Props) {
  return <section className={styles.section}>
    <div className={styles.heading}>
      <h2>Условия на {asOf}</h2>
      <span>{contracts.length} договора</span>
    </div>
    <div className={styles.scroll}>
      <table className={styles.table}>
        <thead><tr><th>Договор</th><th>Поставщик</th><th>Сумма</th><th>Срок поставки</th><th>Что проверить</th></tr></thead>
        <tbody>{contracts.map((contract) => <tr key={contract.id}>
          <td className={styles.id}>{contract.id}</td>
          <td><ContractField field={contract.supplier} onEvidence={onEvidence} /></td>
          <td><ContractField field={contract.amount} currency onEvidence={onEvidence} /></td>
          <td><ContractField field={contract.deadline} onEvidence={onEvidence} /></td>
          <td>
            {contract.notices.length === 0 && <span className={styles.muted}>—</span>}
            {contract.notices.map((notice, index) => <button className={styles.notice} type="button" key={index} onClick={() => onEvidence(notice.evidence)}>{notice.message}</button>)}
            {contract.review && <strong className={styles.review}>Требуется ручная проверка</strong>}
          </td>
        </tr>)}</tbody>
      </table>
    </div>
  </section>
}
