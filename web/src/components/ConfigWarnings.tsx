import { useProfileStore } from '../store/profileStore'
import { computeAdvisories } from '../lib/advisories'

export function ConfigWarnings() {
  const { profile, preview } = useProfileStore()
  const advisories = preview?.advisories?.length
    ? preview.advisories
    : computeAdvisories(profile)

  if (advisories.length === 0) return null

  return (
    <div className="px-4 py-2 border-b border-amber-900/40 bg-amber-950/20 space-y-1">
      <p className="text-xs font-semibold text-amber-400 uppercase">Advisories</p>
      {advisories.map((a) => (
        <p key={`${a.field}-${a.message}`} className="text-xs text-amber-200/90">
          <span className="text-amber-500">[{a.level}]</span> {a.field}: {a.message}
        </p>
      ))}
    </div>
  )
}
