import { useCallback, useEffect, useMemo, useState } from "react";
import AppShell from "../components/layout/AppShell";
import Home from "../components/home/Home";
import { api, type Workspace, type WorkspaceDocument } from "../api/api";
import type { Session } from "../hooks/useCollaboration";

type Props = { session: Session; onLogout: () => void; onOpenDocument: (document: WorkspaceDocument, workspace: Workspace) => void };

export default function HomePage({ session, onLogout, onOpenDocument }: Props) {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [activeId, setActiveId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => {
    setLoading(true); setError(null);
    try { const result = await api.listWorkspaces(session); setWorkspaces(result.workspaces); if (result.workspaces[0] && activeId === null) setActiveId(result.workspaces[0].id); }
    catch (e) { setError(e instanceof Error ? e.message : "Unable to load workspace"); }
    finally { setLoading(false); }
  }, [session, activeId]);
  useEffect(() => { void load(); }, [load]);
  const workspace = useMemo(() => workspaces.find(w => w.id === activeId) ?? workspaces[0], [workspaces, activeId]);
  return <AppShell session={session} workspace={workspace} onLogout={onLogout} activePage="home" onHome={() => undefined} onOpenDocument={(d) => workspace && onOpenDocument(d, workspace)} onWorkspaceChange={(id) => setActiveId(id)} workspaces={workspaces}>
    <Home session={session} workspace={workspace} workspaces={workspaces} loading={loading} error={error} onRefresh={load} onOpenDocument={(d) => workspace && onOpenDocument(d, workspace)} onWorkspaceCreated={(w) => { setWorkspaces(prev => [...prev, w]); setActiveId(w.id); }} onWorkspaceJoined={(w) => { setWorkspaces(prev => prev.some(x => x.id === w.id) ? prev.map(x => x.id === w.id ? w : x) : [...prev, w]); setActiveId(w.id); }} />
  </AppShell>;
}
