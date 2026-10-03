import { useEffect, useId, useRef, type SyntheticEvent } from 'react'

interface ConfirmDialogProps {
  open: boolean
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  destructive?: boolean
  error?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog(props: ConfirmDialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const cancelButtonRef = useRef<HTMLButtonElement>(null)
  const returnFocusRef = useRef<HTMLElement | null>(null)
  const titleId = useId()
  const descriptionId = useId()

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (props.open && !dialog.open) {
      returnFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null
      dialog.showModal()
      cancelButtonRef.current?.focus()
    }
    if (!props.open && dialog.open) {
      dialog.close()
      returnFocusRef.current?.focus()
    }
  }, [props.open])

  function handleCancel(event: SyntheticEvent<HTMLDialogElement>) {
    event.preventDefault()
    if (!props.busy) props.onCancel()
  }

  return (
    <dialog ref={dialogRef} className="confirm-dialog" aria-labelledby={titleId} aria-describedby={descriptionId} onCancel={handleCancel}>
      <h2 id={titleId}>{props.title}</h2>
      <p id={descriptionId}>{props.description}</p>
      {props.error && <p className="dialog-error" role="alert">{props.error}</p>}
      <div className="dialog-actions">
        <button ref={cancelButtonRef} className="button button--quiet" type="button" onClick={props.onCancel} disabled={props.busy}>Cancelar</button>
        <button className={`button${props.destructive ? ' button--danger' : ''}`} type="button" onClick={props.onConfirm} disabled={props.busy}>
          {props.busy ? 'Processando…' : props.confirmLabel}
        </button>
      </div>
    </dialog>
  )
}
