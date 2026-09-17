import { createEffect, createSignal } from 'solid-js';
import {
  activeTool,
  maxMinutes,
  origin,
  routeDestination,
  routeOrigin,
  setActiveTool,
  setMaxMinutes,
  setOrigin,
  setRouteDestination,
  setRouteOrigin,
  setTravelMode,
  travelMode,
} from './analysisStore';

export interface WorkspacePayload {
  id?: string;
  title: string;
  center_lat: number;
  center_lng: number;
  zoom: number;
  layers_json: {
    tool: string;
    origin?: { lat: number; lng: number; address?: string } | null;
    travelMode?: string;
    maxMinutes?: number;
    routeOrigin?: { lat: number; lng: number; address?: string } | null;
    routeDestination?: { lat: number; lng: number; address?: string } | null;
  };
}

const [workspaceID, setWorkspaceID] = createSignal<string | null>(null);

let saveTimer: ReturnType<typeof setTimeout> | null = null;
let skipNextSave = false;

// Serializes the current workbench inputs into a workspace snapshot. The map
// focus (mapFocus) and transient results are intentionally excluded — only the
// durable tool inputs are persisted so a reload restores the exact prior state.
function snapshot(): WorkspacePayload {
  return {
    id: workspaceID() ?? undefined,
    title: 'Default workspace',
    center_lat: origin()?.lat ?? routeOrigin()?.lat ?? 0,
    center_lng: origin()?.lng ?? routeOrigin()?.lng ?? 0,
    zoom: 12,
    layers_json: {
      tool: activeTool(),
      origin: origin(),
      travelMode: travelMode(),
      maxMinutes: maxMinutes(),
      routeOrigin: routeOrigin(),
      routeDestination: routeDestination(),
    },
  };
}

async function saveWorkspace() {
  try {
    const res = await fetch('/api/v1/workspaces', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(snapshot()),
    });
    if (!res.ok) return;
    const data = await res.json();
    if (data.id) setWorkspaceID(data.id);
  } catch {
    // Persistence is best-effort; the next change will retry.
  }
}

export function scheduleSave() {
  if (skipNextSave) return;
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    void saveWorkspace();
  }, 800);
}

// Loads the user's most recently created workspace and replays its inputs into
// the analysis store. Returns true when a snapshot was restored.
export async function loadWorkspace(): Promise<boolean> {
  try {
    const res = await fetch('/api/v1/workspaces');
    if (!res.ok) return false;
    const data = await res.json();
    const ws = data.workspaces?.[0];
    if (!ws) return false;

    setWorkspaceID(ws.id);

    skipNextSave = true;
    const state = ws.layers_json ?? {};
    if (state.tool === 'route' || state.tool === 'isochrone') setActiveTool(state.tool);
    if (typeof state.travelMode === 'string') setTravelMode(state.travelMode);
    if (typeof state.maxMinutes === 'number') setMaxMinutes(state.maxMinutes);
    if (state.origin) setOrigin(state.origin);
    if (state.routeOrigin) setRouteOrigin(state.routeOrigin);
    if (state.routeDestination) setRouteDestination(state.routeDestination);
    skipNextSave = false;
    return true;
  } catch {
    return false;
  }
}

// Hooks debounced saves into every durable input signal. Must be called once
// after the workbench mounts. Reading each signal inside the effect registers
// it as a dependency, so a change in any of them schedules a persistence.
export function startWorkspacePersistence() {
  createEffect(() => {
    activeTool();
    origin();
    travelMode();
    maxMinutes();
    routeOrigin();
    routeDestination();
    scheduleSave();
  });
}