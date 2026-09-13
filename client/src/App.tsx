import { useEffect, useState } from "react";
import AuthPage from "./pages/AuthPage";
import HomePage from "./pages/HomePage";
import WorkspacePage from "./pages/WorkspacePage";
import { type Workspace, type WorkspaceDocument } from "./api/api";
import type { Session } from "./hooks/useCollaboration";

const SESSION_STORAGE_KEY = "syncspace-session";
const WORKSPACE_STORAGE_KEY = "syncspace-workspace";
const API_BASE = import.meta.env.VITE_API_URL ?? `${window.location.protocol}//${window.location.hostname}:8080`;

type Route = { page: "home" } | { page: "workspace"; workspace: Workspace; document: WorkspaceDocument };

function getStoredSession(): Session | null {
  const saved = localStorage.getItem(SESSION_STORAGE_KEY);
  if (!saved) return null;
  try { return JSON.parse(saved) as Session; } catch { localStorage.removeItem(SESSION_STORAGE_KEY); return null; }
}

function App() {
  const [session, setSession] = useState<Session | null>(getStoredSession);
  const [route, setRoute] = useState<Route>({ page: "home" });
  const [isRegistering, setIsRegistering] = useState(false);
  const [authError, setAuthError] = useState<string | null>(null);
  const [isSubmittingAuth, setIsSubmittingAuth] = useState(false);

  useEffect(() => {
    if (!session) return;
    const raw = localStorage.getItem(WORKSPACE_STORAGE_KEY);
    if (!raw) return;
    try {
      const saved = JSON.parse(raw) as { workspace: Workspace; document: WorkspaceDocument };
      setRoute({ page: "workspace", workspace: saved.workspace, document: saved.document });
    } catch { localStorage.removeItem(WORKSPACE_STORAGE_KEY); }
  }, [session]);

  const handleAuthSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "");
    const password = String(form.get("password") ?? "");
    const username = String(form.get("username") ?? "");
    setAuthError(null); setIsSubmittingAuth(true);
    try {
      const response = await fetch(`${API_BASE}/api/auth/${isRegistering ? "register" : "login"}`, {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify(isRegistering ? { email, username, password } : { email, password }),
      });
      const body = await response.json() as Session | { error?: string };
      if (!response.ok) throw new Error("error" in body ? body.error ?? "Unable to sign in." : "Unable to sign in.");
      const next = body as Session;
      localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(next));
      setSession(next); setRoute({ page: "home" });
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : "Unable to sign in. Please try again.");
    } finally { setIsSubmittingAuth(false); }
  };

  const logout = () => {
    localStorage.removeItem(SESSION_STORAGE_KEY); localStorage.removeItem(WORKSPACE_STORAGE_KEY);
    setSession(null); setRoute({ page: "home" });
  };

  if (!session) return <AuthPage isRegistering={isRegistering} isSubmittingAuth={isSubmittingAuth} authError={authError} onSubmit={handleAuthSubmit} onToggleMode={() => { setIsRegistering(v => !v); setAuthError(null); }} />;

  if (route.page === "workspace") {
    return (
      <WorkspacePage
        session={session}
        workspace={route.workspace}
        document={route.document}
        onLogout={logout}
        onHome={() => {
          localStorage.removeItem(WORKSPACE_STORAGE_KEY);
          setRoute({ page: "home" });
        }}
        onOpenDocument={(document, workspace) => {
          localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify({ workspace, document }));
          setRoute({ page: "workspace", workspace, document });
        }}
        onWorkspaceUpdated={(updated) => {
          setRoute((r) => (r.page === "workspace" ? { ...r, workspace: updated } : r));
        }}
      />
    );
  }

  return <HomePage session={session} onLogout={logout} onOpenDocument={(document, workspace) => { localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify({ workspace, document })); setRoute({ page: "workspace", workspace, document }); }} />;
}

export default App;
