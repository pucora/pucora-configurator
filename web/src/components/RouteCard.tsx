import type { Route } from '../types/profile'

interface RouteCardProps {
  route: Route
  selected: boolean
  hasError: boolean
  onSelect: () => void
  onDuplicate: () => void
  onDelete: () => void
  dragHandleProps?: React.HTMLAttributes<HTMLButtonElement>
}

export function RouteCard({
  route,
  selected,
  hasError,
  onSelect,
  onDuplicate,
  onDelete,
  dragHandleProps,
}: RouteCardProps) {
  return (
    <div
      className={`rounded-lg border p-3 cursor-pointer transition-colors ${
        selected
          ? 'border-cyan-500 bg-cyan-950/30'
          : hasError
            ? 'border-red-500/50 bg-red-950/20 hover:border-red-400'
            : 'border-slate-700 bg-slate-900/50 hover:border-slate-600'
      }`}
      onClick={onSelect}
    >
      <div className="flex items-start gap-2">
        <button
          type="button"
          className="mt-0.5 text-slate-500 hover:text-slate-300 cursor-grab active:cursor-grabbing"
          onClick={(e) => e.stopPropagation()}
          {...dragHandleProps}
        >
          ⠿
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="px-1.5 py-0.5 text-xs font-bold rounded bg-slate-800 text-cyan-300">
              {route.method}
            </span>
            <span className="font-mono text-sm truncate">{route.path || '/...'}</span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            {route.backend.type} → {route.backend.host || 'no host'}
          </p>
        </div>
        <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
          <button type="button" onClick={onDuplicate} className="text-xs text-slate-500 hover:text-white px-1" title="Duplicate">
            ⧉
          </button>
          <button type="button" onClick={onDelete} className="text-xs text-slate-500 hover:text-red-400 px-1" title="Delete">
            ✕
          </button>
        </div>
      </div>
      {hasError && <p className="text-xs text-red-400 mt-1 ml-6">Has validation errors</p>}
    </div>
  )
}
