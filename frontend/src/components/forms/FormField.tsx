import { cloneElement, type ReactElement } from 'react'

interface FormFieldProps {
  id: string
  label: string
  hint?: string
  error?: string
  required?: boolean
  children: ReactElement<{ id?: string; 'aria-describedby'?: string; 'aria-invalid'?: boolean }>
}

export function FormField({ id, label, hint, error, required, children }: FormFieldProps) {
  const descriptionIds = [hint ? `${id}-hint` : '', error ? `${id}-error` : ''].filter(Boolean).join(' ') || undefined
  const control = cloneElement(children, {
    id,
    'aria-describedby': descriptionIds,
    'aria-invalid': error ? true : undefined,
  })
  return (
    <div className={`form-field${error ? ' form-field--invalid' : ''}`}>
      <label htmlFor={id}>
        {label} {required && <span aria-hidden="true">*</span>}
      </label>
      {hint && <span className="field-hint" id={`${id}-hint`}>{hint}</span>}
      {control}
      {error && <span className="field-error" id={`${id}-error`} role="alert">{error}</span>}
    </div>
  )
}
