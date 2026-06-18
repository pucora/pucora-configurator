import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { Route } from '../types/profile'
import { RouteCard } from './RouteCard'
import { useProfileStore } from '../store/profileStore'

function SortableRouteCard({
  route,
  index,
  selected,
  hasError,
}: {
  route: Route
  index: number
  selected: boolean
  hasError: boolean
}) {
  const { setSelectedRoute, updateProfile } = useProfileStore()
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: `route-${index}` })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div ref={setNodeRef} style={style}>
      <RouteCard
        route={route}
        selected={selected}
        hasError={hasError}
        onSelect={() => setSelectedRoute(index)}
        onDuplicate={() =>
          updateProfile((p) => {
            const copy = structuredClone(route)
            p.routes.splice(index + 1, 0, copy)
            return p
          })
        }
        onDelete={() =>
          updateProfile((p) => {
            p.routes.splice(index, 1)
            return p
          })
        }
        dragHandleProps={{ ...attributes, ...listeners }}
      />
    </div>
  )
}

interface RouteListProps {
  routes: Route[]
  allRoutes: Route[]
}

export function RouteList({ routes, allRoutes }: RouteListProps) {
  const { selectedRouteIndex, validationErrors, updateProfile } = useProfileStore()
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  const errorIndices = new Set(
    validationErrors
      .map((e) => e.field.match(/^routes\[(\d+)\]/)?.[1])
      .filter(Boolean)
      .map(Number),
  )

  const onDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = parseInt(String(active.id).replace('route-', ''), 10)
    const newIndex = parseInt(String(over.id).replace('route-', ''), 10)
    updateProfile((p) => {
      const [item] = p.routes.splice(oldIndex, 1)
      p.routes.splice(newIndex, 0, item)
      return p
    })
  }

  if (routes.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-slate-700 p-8 text-center text-slate-500">
        <p className="text-sm">{allRoutes.length > 0 ? 'No routes match search' : 'No routes yet'}</p>
        <p className="text-xs mt-1">Drag a template from the palette below or click to add</p>
      </div>
    )
  }

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext items={routes.map((r) => {
        const i = allRoutes.indexOf(r)
        return `route-${i}`
      })} strategy={verticalListSortingStrategy}>
        <div className="space-y-2">
          {routes.map((route) => {
            const i = allRoutes.indexOf(route)
            return (
            <SortableRouteCard
              key={`route-${i}-${route.path}-${route.method}`}
              route={route}
              index={i}
              selected={selectedRouteIndex === i}
              hasError={errorIndices.has(i)}
            />
            )
          })}
        </div>
      </SortableContext>
    </DndContext>
  )
}
