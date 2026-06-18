interface KeyValueEditorProps {
  label: string
  help?: string
  value: Record<string, string>
  onChange: (v: Record<string, string>) => void
  keyPlaceholder?: string
  valuePlaceholder?: string
}

export function KeyValueEditor({
  label,
  help,
  value,
  onChange,
  keyPlaceholder = 'key',
  valuePlaceholder = 'value',
}: KeyValueEditorProps) {
  const entries = Object.entries(value)

  const update = (idx: number, k: string, v: string) => {
    const next: Record<string, string> = {}
    entries.forEach(([ek, ev], i) => {
      if (i === idx) {
        if (k) next[k] = v
      } else {
        next[ek] = ev
      }
    })
    onChange(next)
  }

  const remove = (idx: number) => {
    const next: Record<string, string> = {}
    entries.forEach(([k, v], i) => {
      if (i !== idx) next[k] = v
    })
    onChange(next)
  }

  const add = () => onChange({ ...value, '': '' })

  return (
    <div className="space-y-2">
      <label>{label}</label>
      {help && <p className="text-xs text-slate-500 normal-case">{help}</p>}
      {entries.map(([k, v], i) => (
        <div key={i} className="flex gap-1">
          <input
            className="flex-1 text-xs"
            value={k}
            placeholder={keyPlaceholder}
            onChange={(e) => update(i, e.target.value, v)}
          />
          <input
            className="flex-1 text-xs"
            value={v}
            placeholder={valuePlaceholder}
            onChange={(e) => update(i, k, e.target.value)}
          />
          <button type="button" onClick={() => remove(i)} className="text-slate-500 hover:text-red-400 px-1">×</button>
        </div>
      ))}
      <button
        type="button"
        onClick={add}
        className="text-xs text-cyan-400 hover:text-cyan-300"
      >
        + Add mapping
      </button>
    </div>
  )
}
