import { useEffect, useRef } from 'react'

interface ConfirmDialogProps {
  open: boolean
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  destructive?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog(props: ConfirmDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (props.open && !dialog.open) dialog.showModal()
    if (!props.open && dialog.open) dialog.close()
  }, [props.open])

  return (
    <dialog ref={dialogRef} className="confirm-dialog" onCancel={props.onCancel}>
      <h2>{props.title}</h2>
      <p>{props.description}</p>
      <div className="dialog-actions">
        <button className="button button--quiet" type="button" onClick={props.onCancel} disabled={props.busy}>Cancelar</button>
        <button className={`button${props.destructive ? ' button--danger' : ''}`} type="button" onClick={props.onConfirm} disabled={props.busy}>
          {props.busy ? 'Processando…' : props.confirmLabel}
        </button>
      </div>
    </dialog>
  )
}
