interface TagInputProps {
  label: string
  help?: string
  value: string[]
  onChange: (v: string[]) => void
  suggestions?: string[]
  placeholder?: string
}

export function TagInput({ label, help, value, onChange, suggestions = [], placeholder }: TagInputProps) {
  const add = (tag: string) => {
    const t = tag.trim()
    if (t && !value.includes(t)) onChange([...value, t])
  }

  const remove = (tag: string) => onChange(value.filter((v) => v !== tag))

  return (
    <div className="space-y-1.5">
      <label>{label}</label>
      {help && <p className="text-xs text-slate-500 normal-case">{help}</p>}
      <div className="flex flex-wrap gap-1.5 min-h-[36px] p-2 bg-slate-900 border border-slate-700 rounded-md">
        {value.map((tag) => (
          <span key={tag} className="inline-flex items-center gap-1 px-2 py-0.5 bg-cyan-900/40 text-cyan-200 rounded text-xs">
            {tag}
            <button type="button" onClick={() => remove(tag)} className="text-cyan-400 hover:text-white">×</button>
          </span>
        ))}
        <input
          className="flex-1 min-w-[120px] border-0 bg-transparent p-0 focus:ring-0"
          placeholder={placeholder || 'Type and press Enter'}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              add((e.target as HTMLInputElement).value)
              ;(e.target as HTMLInputElement).value = ''
            }
          }}
        />
      </div>
      {suggestions.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {suggestions.filter((s) => !value.includes(s)).map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => add(s)}
              className="px-2 py-0.5 text-xs rounded border border-slate-700 text-slate-400 hover:border-cyan-600 hover:text-cyan-300"
            >
              + {s}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
