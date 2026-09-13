import { useState } from "react";
import AppShell from "../components/layout/AppShell";
import DocumentEditor from "../components/DocumentEditor";
import RightPanel from "../components/RightPanel";
import useCollaboration, { type Session } from "../hooks/useCollaboration";
import { api, type Workspace, type WorkspaceDocument } from "../api/api";
import "../components/workspace/workspace.css";

type Props = {
  session: Session;
  workspace: Workspace;
  document: WorkspaceDocument;
  onLogout: () => void;
  onHome: () => void;
  onOpenDocument: (document: WorkspaceDocument, workspace: Workspace) => void;
  onWorkspaceUpdated?: (workspace: Workspace) => void;
};

export default function WorkspacePage({
  session,
  workspace,
  document,
  onLogout,
  onHome,
  onOpenDocument,
  onWorkspaceUpdated,
}: Props) {
  const c = useCollaboration(session, document.roomId);
  const [docState, setDocState] = useState<WorkspaceDocument>(document);
  const [toast, setToast] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const handleUpdateDetails = async (title: string, description: string) => {
    const updated = await api.updateDocument(session, workspace.id, docState.id, title, description);
    setDocState(updated);
    showToast("Document details updated successfully.");
    if (onWorkspaceUpdated) {
      const refreshed = await api.getWorkspace(session, workspace.id);
      onWorkspaceUpdated(refreshed);
    }
  };

  const handleDeleteDocument = async () => {
    await api.deleteDocument(session, workspace.id, docState.id);
    if (onWorkspaceUpdated) {
      const refreshed = await api.getWorkspace(session, workspace.id);
      onWorkspaceUpdated(refreshed);
    }
    onHome();
  };

  const handleShare = async () => {
    try {
      const invite = await api.createInvite(session, workspace.id);
      void navigator.clipboard.writeText(
        `Join my SyncSpace workspace to collaborate on "${docState.title}". Invite code: ${invite.code}`
      );
      showToast("Workspace invite code copied to clipboard!");
    } catch {
      void navigator.clipboard.writeText(window.location.href);
      showToast("Document link copied to clipboard!");
    }
  };

  return (
    <AppShell
      session={session}
      workspace={workspace}
      activePage="workspace"
      onLogout={onLogout}
      onHome={onHome}
      onOpenDocument={(d, w) => onOpenDocument(d, w ?? workspace)}
      workspaces={[workspace]}
      connectionStatus={c.connectionStatus}
    >
      <div className="workspace-page">
        <DocumentEditor
          content={c.content}
          textareaRef={c.textareaRef}
          documentVersion={c.documentVersion}
          errorMessage={c.errorMessage}
          onChange={c.handleChange}
          onSelect={c.handleCursorMove}
          title={docState.title}
          description={docState.description}
          remoteCursors={c.remoteCursors}
          onUpdateDetails={handleUpdateDetails}
          onDeleteDocument={handleDeleteDocument}
          onShare={handleShare}
        />

        <RightPanel
          aiQuestion={c.aiQuestion}
          aiAnswer={c.aiAnswer}
          searchResults={c.searchResults}
          isSearching={c.isSearching}
          searchQuery={c.searchQuery}
          users={c.users}
          remoteCursors={c.remoteCursors}
          onAIQuestionChange={c.setAiQuestion}
          onAskAI={c.askAI}
          onSearchQueryChange={c.setSearchQuery}
          onSearch={c.handleSearch}
          onSelectSearchResult={c.setContent}
          username={session.user.username}
        />
      </div>

      {toast && <button className="toast" onClick={() => setToast(null)}>{toast}</button>}
    </AppShell>
  );
}
