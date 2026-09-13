import { useState } from "react";
import type { Workspace, WorkspaceDocument } from "../../api/api";
import type { Session } from "../../hooks/useCollaboration";
import "./styles/sidebar.css";

type Props = {
  session: Session;
  workspace?: Workspace;
  workspaces?: Workspace[];
  activePage: "home" | "workspace";
  onLogout: () => void;
  onHome: () => void;
  onOpenDocument: (document: WorkspaceDocument, workspace?: Workspace) => void;
  onWorkspaceChange?: (id: number) => void;
};
export default function Sidebar({
  session,
  workspace,
  workspaces = workspace ? [workspace] : [],
  activePage,
  onLogout,
  onHome,
  onOpenDocument,
  onWorkspaceChange,
}: Props) {
  const [collapsed, setCollapsed] = useState(false);
  const [panel, setPanel] = useState<"settings" | "help" | null>(null);
  return (
    <aside className={`sidebar ${collapsed ? "collapsed" : ""}`}>
      <div className="sidebar-logo">
        <div className="logo-mark">S</div>
        {!collapsed && <span>SyncSpace</span>}
        <button
          className="collapse-button"
          onClick={() => setCollapsed((v) => !v)}
          aria-label="Toggle sidebar"
        >
          {collapsed ? "→" : "←"}
        </button>
      </div>
      {!collapsed && (
        <>
          <div className="workspace-selector">
            <div className="workspace-icon">
              {workspace?.name?.[0]?.toUpperCase() ?? "S"}
            </div>
            <div className="workspace-info">
              <strong>{workspace?.name ?? "My Workspace"}</strong>
              <span>{workspace?.role === "owner" ? "Owner" : "Workspace"}</span>
            </div>
            {workspaces.length > 1 ? (
              <select
                aria-label="Switch workspace"
                value={workspace?.id ?? ""}
                onChange={(e) => onWorkspaceChange?.(Number(e.target.value))}
              >
                {workspaces.map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.name}
                  </option>
                ))}
              </select>
            ) : (
              <span className="workspace-arrow">⌄</span>
            )}
          </div>
          <nav className="sidebar-nav">
            <button
              className={`nav-item ${activePage === "home" ? "active" : ""}`}
              onClick={onHome}
            >
              <span>⌂</span>
              <span>Home</span>
            </button>
            <button
              className={`nav-item ${activePage === "workspace" ? "active" : ""}`}
              onClick={() =>
                workspace?.documents[0] &&
                onOpenDocument(workspace.documents[0], workspace)
              }
            >
              <span>▤</span>
              <span>Documents</span>
            </button>
            <button
              className="nav-item"
              onClick={() =>
                document
                  .getElementById("tasks-section")
                  ?.scrollIntoView({ behavior: "smooth" })
              }
            >
              <span>✓</span>
              <span>Tasks</span>
            </button>
            <button
              className="nav-item"
              onClick={() =>
                document
                  .getElementById("activity-section")
                  ?.scrollIntoView({ behavior: "smooth" })
              }
            >
              <span>◔</span>
              <span>Activity</span>
            </button>
          </nav>
          <div className="sidebar-section">
            <div className="sidebar-section-header">
              <span>DOCUMENTS</span>
              <span>{workspace?.documents.length ?? 0}</span>
            </div>
            {workspace?.documents.slice(0, 6).map((doc) => (
              <button
                key={doc.id}
                className="document-link"
                onClick={() => onOpenDocument(doc, workspace)}
              >
                <span>📄</span>
                <span>{doc.title}</span>
              </button>
            ))}
            {!workspace?.documents.length && (
              <div className="sidebar-empty">No documents yet</div>
            )}
          </div>
          <div className="sidebar-bottom">
            <button
              className="sidebar-bottom-item"
              onClick={() => setPanel("settings")}
            >
              <span>⚙</span>
              <span>Settings</span>
            </button>
            <button
              className="sidebar-bottom-item"
              onClick={() => setPanel("help")}
            >
              <span>?</span>
              <span>Help & Support</span>
            </button>
            <div className="sidebar-user">
              <div className="user-avatar">
                {session.user.username.slice(0, 2).toUpperCase()}
              </div>
              <div className="sidebar-user-info">
                <strong>{session.user.username}</strong>
                <span>{session.user.email}</span>
              </div>
              <button
                className="logout-button"
                onClick={onLogout}
                title="Logout"
              >
                ↪
              </button>
            </div>
          </div>
        </>
      )}
      <>
        {panel && (
          <div
            className="sidebar-modal-backdrop"
            onMouseDown={() => setPanel(null)}
          >
            <div
              className="sidebar-modal"
              onMouseDown={(e) => e.stopPropagation()}
            >
              <button className="modal-close" onClick={() => setPanel(null)}>
                ×
              </button>
              {panel === "settings" ? (
                <>
                  <p className="eyebrow">ACCOUNT</p>
                  <h3>Settings</h3>
                  <p>
                    Your account is signed in as{" "}
                    <strong>{session.user.email}</strong>. Session and workspace
                    access are managed securely by the SyncSpace server.
                  </p>
                </>
              ) : (
                <>
                  <p className="eyebrow">SUPPORT</p>
                  <h3>Help & Support</h3>
                  <p>
                    Use the Home page to create or join workspaces. Open a
                    document to collaborate in real time. If the connection
                    drops, edits are queued locally and retried.
                  </p>
                </>
              )}
            </div>
          </div>
        )}
      </>
    </aside>
  );
}
