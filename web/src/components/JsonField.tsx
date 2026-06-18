interface JsonFieldProps {
  label: string
  help?: string
  value: Record<string, unknown> | undefined
  onChange: (v: Record<string, unknown> | undefined) => void
}

export function JsonField({ label, help, value, onChange }: JsonFieldProps) {
  const text = value ? JSON.stringify(value, null, 2) : ''

  return (
    <div className="space-y-1.5">
      <label>{label}</label>
      {help && <p className="text-xs text-slate-500 normal-case">{help}</p>}
      <textarea
        className="font-mono text-xs min-h-[100px]"
        value={text}
        placeholder={'{\n  "episode": "{episode}"\n}'}
        onChange={(e) => {
          const raw = e.target.value.trim()
          if (!raw) {
            onChange(undefined)
            return
          }
          try {
            onChange(JSON.parse(raw) as Record<string, unknown>)
          } catch {
            // keep typing
          }
        }}
      />
    </div>
  )
}
