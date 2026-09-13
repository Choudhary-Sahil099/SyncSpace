import type { ReactNode } from "react";
import Sidebar from "./Sidebar";
import Topbar from "./Topbar";
import type { Workspace, WorkspaceDocument } from "../../api/api";
import type { Session } from "../../hooks/useCollaboration";

type Props = {
  children: ReactNode;
  session: Session;
  workspace?: Workspace;
  activePage: "home" | "workspace";
  onLogout: () => void;
  onHome: () => void;
  onOpenDocument: (document: WorkspaceDocument, workspace?: Workspace) => void;
  onWorkspaceChange?: (id: number) => void;
  workspaces?: Workspace[];
  connectionStatus?: string;
};
export default function AppShell({
  children,
  session,
  workspace,
  activePage,
  onLogout,
  onHome,
  onOpenDocument,
  onWorkspaceChange,
  workspaces = workspace ? [workspace] : [],
  connectionStatus = "Ready",
}: Props) {
  return (
    <div className="app-shell">
      <Sidebar
        session={session}
        workspace={workspace}
        activePage={activePage}
        workspaces={workspaces}
        onLogout={onLogout}
        onHome={onHome}
        onOpenDocument={onOpenDocument}
        onWorkspaceChange={onWorkspaceChange}
      />
      <div className="main-area">
        <Topbar
          connectionStatus={connectionStatus}
          workspaceName={workspace?.name ?? "My Workspace"}
          page={activePage}
          onHome={onHome}
          session={session}
        />
        <main className="page-content">{children}</main>
      </div>
    </div>
  );
}
