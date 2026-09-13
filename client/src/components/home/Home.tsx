import { useEffect, useMemo, useRef, useState } from "react";
import { api, type Workspace, type WorkspaceDocument, type WorkspaceTask } from "../../api/api";
import type { Session } from "../../hooks/useCollaboration";
import DocumentCard from "./DocumentCard";
import "./home.css";

type Props = {
  session: Session;
  workspace?: Workspace;
  workspaces: Workspace[];
  loading: boolean;
  error: string | null;
  onRefresh: () => void;
  onOpenDocument: (d: WorkspaceDocument) => void;
  onWorkspaceCreated: (w: Workspace) => void;
  onWorkspaceJoined: (w: Workspace) => void;
};

export default function Home({
  session,
  workspace,
  loading,
  error,
  onRefresh,
  onOpenDocument,
  onWorkspaceCreated,
  onWorkspaceJoined,
}: Props) {
  const [modal, setModal] = useState<"document" | "workspace" | "invite" | "join" | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [invite, setInvite] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  // Document filter
  const docs = workspace?.documents ?? [];
  const [searchText, setSearchText] = useState("");
  const filteredDocs = docs.filter((d) =>
    `${d.title} ${d.description}`.toLowerCase().includes(searchText.toLowerCase())
  );

  // Delete document modal state
  const [deleteDocTarget, setDeleteDocTarget] = useState<WorkspaceDocument | null>(null);

  // Tasks state
  const [tasks, setTasks] = useState<WorkspaceTask[]>(workspace?.tasks ?? []);
  const [taskInput, setTaskInput] = useState("");
  const [taskFilter, setTaskFilter] = useState<"all" | "active" | "completed">("all");
  const [isAddingTask, setIsAddingTask] = useState(false);
  const taskInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (workspace?.tasks) {
      setTasks(workspace.tasks);
    } else if (workspace?.id) {
      api.listTasks(session, workspace.id)
        .then((res) => setTasks(res.tasks))
        .catch(() => undefined);
    }
  }, [workspace?.id, workspace?.tasks, session]);

  const initials = session.user.username.slice(0, 2).toUpperCase();

  const greeting = useMemo(() => {
    const h = new Date().getHours();
    return h < 12 ? "Good morning" : h < 18 ? "Good afternoon" : "Good evening";
  }, []);

  const completedTasksCount = tasks.filter((t) => t.completed).length;
  const activeTasksCount = tasks.length - completedTasksCount;
  const taskPercent = tasks.length > 0 ? Math.round((completedTasksCount / tasks.length) * 100) : 0;

  const filteredTasks = useMemo(() => {
    if (taskFilter === "active") return tasks.filter((t) => !t.completed);
    if (taskFilter === "completed") return tasks.filter((t) => t.completed);
    return tasks;
  }, [tasks, taskFilter]);

  async function createDocument() {
    if (!workspace || !title.trim()) return;
    setBusy(true);
    try {
      const d = await api.createDocument(session, workspace.id, title, description);
      setModal(null);
      setTitle("");
      setDescription("");
      onRefresh();
      onOpenDocument(d);
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Could not create document");
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteDocumentConfirm() {
    if (!workspace || !deleteDocTarget) return;
    setBusy(true);
    try {
      await api.deleteDocument(session, workspace.id, deleteDocTarget.id);
      setDeleteDocTarget(null);
      setNotice(`"${deleteDocTarget.title}" deleted.`);
      onRefresh();
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Could not delete document");
    } finally {
      setBusy(false);
    }
  }

  async function createWorkspace() {
    if (!name.trim()) return;
    setBusy(true);
    try {
      const w = await api.createWorkspace(session, name, description);
      setModal(null);
      setName("");
      setDescription("");
      onWorkspaceCreated(w);
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Could not create workspace");
    } finally {
      setBusy(false);
    }
  }

  async function makeInvite() {
    if (!workspace) return;
    setBusy(true);
    try {
      const r = await api.createInvite(session, workspace.id);
      setInvite(r.code);
      setNotice("Invite code created. Share it with your teammate.");
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Could not create invite");
    } finally {
      setBusy(false);
    }
  }

  async function join() {
    if (!code.trim()) return;
    setBusy(true);
    try {
      const w = await api.joinWorkspace(session, code);
      setModal(null);
      setCode("");
      onWorkspaceJoined(w);
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Invalid invite code");
    } finally {
      setBusy(false);
    }
  }

  // Task actions
  async function handleAddTask(e: React.FormEvent) {
    e.preventDefault();
    if (!workspace || !taskInput.trim() || isAddingTask) return;
    setIsAddingTask(true);
    try {
      const newTask = await api.createTask(session, workspace.id, taskInput.trim());
      setTasks((prev) => [newTask, ...prev]);
      setTaskInput("");
    } catch (err) {
      setNotice(err instanceof Error ? err.message : "Could not create task");
    } finally {
      setIsAddingTask(false);
    }
  }

  async function handleToggleTask(taskId: number, currentStatus: boolean) {
    if (!workspace) return;
    // Optimistic update
    setTasks((prev) =>
      prev.map((t) => (t.id === taskId ? { ...t, completed: !currentStatus } : t))
    );
    try {
      await api.updateTask(session, workspace.id, taskId, !currentStatus);
    } catch {
      // Revert if failed
      setTasks((prev) =>
        prev.map((t) => (t.id === taskId ? { ...t, completed: currentStatus } : t))
      );
      setNotice("Could not update task status");
    }
  }

  async function handleDeleteTask(taskId: number) {
    if (!workspace) return;
    const prevTasks = tasks;
    setTasks((prev) => prev.filter((t) => t.id !== taskId));
    try {
      await api.deleteTask(session, workspace.id, taskId);
    } catch {
      setTasks(prevTasks);
      setNotice("Could not delete task");
    }
  }

  return (
    <div className="home-page">
      {/* Hero Header */}
      <section className="home-hero">
        <div>
          <p className="eyebrow">SYNCSPACE PLATFORM</p>
          <h1>
            {greeting}, {session.user.username} <span>👋</span>
          </h1>
          <p>Everything you and your team are working on, in one synchronized space.</p>
        </div>
        <div className="hero-actions">
          <button className="secondary-button" onClick={() => setModal("join")}>
            Join workspace
          </button>
          <button className="secondary-button" onClick={() => setModal("workspace")}>
            + Workspace
          </button>
          <button className="primary-button" onClick={() => setModal("document")}>
            + New Document
          </button>
        </div>
      </section>

      {error && (
        <div className="home-error">
          <span>⚠ {error}</span>
          <button onClick={onRefresh}>Retry</button>
        </div>
      )}

      {/* Workspace Overview Banner */}
      <section className="workspace-overview">
        <div className="workspace-title">
          <div className="workspace-large-icon">{workspace?.name?.[0]?.toUpperCase() ?? "S"}</div>
          <div>
            <span>ACTIVE WORKSPACE</span>
            <h2>{workspace?.name ?? "Loading workspace…"}</h2>
            <p>{workspace?.description || "Collaborative workspace for your team."}</p>
          </div>
        </div>

        <div className="workspace-stats">
          <div>
            <strong>{docs.length}</strong>
            <span>Documents</span>
          </div>
          <div>
            <strong>{workspace?.members.length ?? 1}</strong>
            <span>Members</span>
          </div>
          <div>
            <strong>{completedTasksCount}/{tasks.length}</strong>
            <span>Tasks Done</span>
          </div>
          <div>
            <strong className="role-tag">{workspace?.role ?? "member"}</strong>
            <span>Your role</span>
          </div>
        </div>
      </section>

      {/* Quick Action Cards */}
      <section className="quick-actions">
        <button onClick={() => setModal("document")}>
          <span className="action-icon">📄</span>
          <div>
            <strong>New document</strong>
            <small>Create notes or specs</small>
          </div>
        </button>

        <button
          onClick={() => {
            const el = document.getElementById("tasks-section");
            el?.scrollIntoView({ behavior: "smooth" });
            taskInputRef.current?.focus();
          }}
        >
          <span className="action-icon">✓</span>
          <div>
            <strong>Workspace task</strong>
            <small>Plan sprint tasks</small>
          </div>
        </button>

        <button onClick={() => setModal("invite")}>
          <span className="action-icon">👥</span>
          <div>
            <strong>Invite people</strong>
            <small>Share invite code</small>
          </div>
        </button>
      </section>

      {/* Documents Section */}
      <section className="documents-section">
        <div className="section-heading">
          <div>
            <h2>Documents</h2>
            <p>Collaborative documents in {workspace?.name ?? "your workspace"}</p>
          </div>
          <div className="section-tools">
            <div className="search-input-wrap">
              <span className="search-icon-symbol">⌕</span>
              <input
                id="workspace-search"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                placeholder="Search documents…"
              />
            </div>
            {docs.length > 0 && (
              <button className="view-all" onClick={() => onOpenDocument(docs[0])}>
                Open latest →
              </button>
            )}
          </div>
        </div>

        {loading ? (
          <div className="loading-card">
            <span className="loading-spinner" />
            <p>Loading documents…</p>
          </div>
        ) : filteredDocs.length ? (
          <div className="document-grid">
            {filteredDocs.map((d) => (
              <DocumentCard
                key={d.id}
                document={d}
                onClick={() => onOpenDocument(d)}
                onDelete={() => setDeleteDocTarget(d)}
              />
            ))}
          </div>
        ) : (
          <div className="empty-card">
            <div className="empty-icon">📄</div>
            <h3>{searchText ? "No matching documents" : "No documents yet"}</h3>
            <p>
              {searchText
                ? "Try searching for a different keyword."
                : "Create your first document to collaborate with real-time OT and presence."}
            </p>
            {!searchText && (
              <button className="primary-button" onClick={() => setModal("document")}>
                + Create Document
              </button>
            )}
          </div>
        )}
      </section>

      {/* Interactive Workspace Tasks Section */}
      <section id="tasks-section" className="tasks-section">
        <div className="section-heading">
          <div>
            <h2>Workspace Tasks</h2>
            <p>Track team action items and collaborative milestones</p>
          </div>
          <div className="task-progress-wrap">
            <div className="progress-bar-container">
              <div className="progress-bar-fill" style={{ width: `${taskPercent}%` }} />
            </div>
            <span className="progress-text">{taskPercent}% completed</span>
          </div>
        </div>

        <div className="tasks-container">
          {/* Add Task Input Form */}
          <form className="add-task-form" onSubmit={handleAddTask}>
            <input
              ref={taskInputRef}
              type="text"
              value={taskInput}
              onChange={(e) => setTaskInput(e.target.value)}
              placeholder="Add a new task for this workspace… (Press Enter)"
            />
            <button
              type="submit"
              className="primary-button"
              disabled={isAddingTask || !taskInput.trim()}
            >
              {isAddingTask ? "Adding…" : "+ Add Task"}
            </button>
          </form>

          {/* Filter Tabs */}
          <div className="task-filters">
            <button
              type="button"
              className={`filter-tab ${taskFilter === "all" ? "active" : ""}`}
              onClick={() => setTaskFilter("all")}
            >
              All ({tasks.length})
            </button>
            <button
              type="button"
              className={`filter-tab ${taskFilter === "active" ? "active" : ""}`}
              onClick={() => setTaskFilter("active")}
            >
              Active ({activeTasksCount})
            </button>
            <button
              type="button"
              className={`filter-tab ${taskFilter === "completed" ? "active" : ""}`}
              onClick={() => setTaskFilter("completed")}
            >
              Completed ({completedTasksCount})
            </button>
          </div>

          {/* Tasks List */}
          {filteredTasks.length > 0 ? (
            <div className="tasks-list">
              {filteredTasks.map((task) => (
                <div key={task.id} className={`task-row ${task.completed ? "is-completed" : ""}`}>
                  <label className="task-check-wrap">
                    <input
                      type="checkbox"
                      checked={task.completed}
                      onChange={() => handleToggleTask(task.id, task.completed)}
                    />
                    <span className="task-checkmark" />
                  </label>

                  <div className="task-title-wrap">
                    <span className="task-title">{task.title}</span>
                    <span className="task-meta">
                      {new Date(task.createdAt).toLocaleDateString(undefined, {
                        month: "short",
                        day: "numeric",
                      })}
                    </span>
                  </div>

                  <button
                    type="button"
                    className="task-delete-btn"
                    title="Delete task"
                    onClick={() => handleDeleteTask(task.id)}
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <div className="empty-tasks-placeholder">
              <span className="empty-tasks-icon">✓</span>
              <p>
                {taskFilter === "completed"
                  ? "No completed tasks yet. Check off items above!"
                  : taskFilter === "active"
                  ? "All caught up! Great job."
                  : "No tasks created yet. Add one above to get organized."}
              </p>
            </div>
          )}
        </div>
      </section>

      {/* Activity Section */}
      <section className="activity-section" id="activity-section">
        <div className="section-heading">
          <div>
            <h2>Recent Activity</h2>
            <p>Live status of your workspace</p>
          </div>
        </div>
        <div className="activity-list">
          <div className="activity-item">
            <div className="activity-avatar">{initials}</div>
            <div>
              <strong>You signed in to {workspace?.name ?? "SyncSpace"}</strong>
              <span>Active session · Authenticated via JWT</span>
            </div>
          </div>
          <div className="activity-item">
            <div className="activity-avatar">WS</div>
            <div>
              <strong>{workspace?.members.length ?? 1} member(s) collaborated</strong>
              <span>Real-time Operational Transformation active</span>
            </div>
          </div>
        </div>
      </section>

      {/* Modals */}
      {modal && (
        <div className="modal-backdrop" onMouseDown={() => setModal(null)}>
          <div className="modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setModal(null)}>
              ×
            </button>

            {modal === "document" && (
              <>
                <p className="eyebrow">NEW DOCUMENT</p>
                <h2>Create a document</h2>
                <input
                  autoFocus
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Document title"
                />
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Short description (optional)"
                />
                <button
                  className="primary-button"
                  disabled={busy || !title.trim()}
                  onClick={createDocument}
                >
                  {busy ? "Creating…" : "Create document"}
                </button>
              </>
            )}

            {modal === "workspace" && (
              <>
                <p className="eyebrow">NEW WORKSPACE</p>
                <h2>Create a workspace</h2>
                <input
                  autoFocus
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Workspace name"
                />
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="What is this workspace for?"
                />
                <button
                  className="primary-button"
                  disabled={busy || !name.trim()}
                  onClick={createWorkspace}
                >
                  {busy ? "Creating…" : "Create workspace"}
                </button>
              </>
            )}

            {modal === "invite" && (
              <>
                <p className="eyebrow">COLLABORATION</p>
                <h2>Invite Teammates</h2>
                <p className="modal-copy">
                  Generate a secure workspace invite code and share it with your teammate to collaborate in real time.
                </p>
                <button className="primary-button" disabled={busy} onClick={makeInvite}>
                  {busy ? "Generating…" : "Generate invite code"}
                </button>
                {invite && (
                  <div className="invite-code">
                    <span>{invite}</span>
                    <button
                      onClick={() => {
                        void navigator.clipboard?.writeText(invite);
                        setNotice("Invite code copied to clipboard!");
                      }}
                    >
                      Copy
                    </button>
                  </div>
                )}
              </>
            )}

            {modal === "join" && (
              <>
                <p className="eyebrow">JOIN</p>
                <h2>Join a workspace</h2>
                <input
                  autoFocus
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="Paste invite code"
                />
                <button className="primary-button" disabled={busy || !code.trim()} onClick={join}>
                  {busy ? "Joining…" : "Join workspace"}
                </button>
              </>
            )}
          </div>
        </div>
      )}

      {/* Delete Document Confirmation Modal */}
      {deleteDocTarget && (
        <div className="modal-backdrop" onMouseDown={() => setDeleteDocTarget(null)}>
          <div className="modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setDeleteDocTarget(null)}>
              ×
            </button>
            <p className="eyebrow" style={{ color: "#ef4444" }}>DANGER</p>
            <h2>Delete Document</h2>
            <p className="modal-copy">
              Are you sure you want to delete <strong>"{deleteDocTarget.title}"</strong>? This cannot be undone.
            </p>
            <div className="modal-actions">
              <button
                type="button"
                className="secondary-button"
                onClick={() => setDeleteDocTarget(null)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="danger-button"
                disabled={busy}
                onClick={handleDeleteDocumentConfirm}
              >
                {busy ? "Deleting…" : "Delete Permanently"}
              </button>
            </div>
          </div>
        </div>
      )}

      {notice && (
        <button className="toast" onClick={() => setNotice(null)}>
          {notice}
        </button>
      )}
    </div>
  );
}
