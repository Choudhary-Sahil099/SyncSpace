import type { Session } from "../hooks/useCollaboration";

const API_BASE =
  import.meta.env.VITE_API_URL ??
  `${window.location.protocol}//${window.location.hostname}:8080`;

export type WorkspaceDocument = {
  id: number;
  roomId: string;
  title: string;
  description: string;
  updatedAt: string;
};

export type WorkspaceMember = {
  userId: number;
  username: string;
  email: string;
  role: string;
};

export type WorkspaceTask = {
  id: number;
  workspaceId: number;
  title: string;
  completed: boolean;
  createdBy: number;
  createdAt: string;
};

export type Workspace = {
  id: number;
  name: string;
  description: string;
  role: string;
  inviteCode?: string;
  members: WorkspaceMember[];
  documents: WorkspaceDocument[];
  tasks?: WorkspaceTask[];
};

async function request<T>(
  session: Session,
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${session.accessToken}`,
      ...(init.headers ?? {}),
    },
  });
  const body = (await response.json().catch(() => ({}))) as T & {
    error?: string;
  };
  if (!response.ok)
    throw new Error(body.error ?? `Request failed (${response.status})`);
  return body;
}

export const api = {
  listWorkspaces: (session: Session) =>
    request<{ workspaces: Workspace[] }>(session, "/api/workspaces"),
  getWorkspace: (session: Session, id: number) =>
    request<Workspace>(session, `/api/workspaces/${id}`),
  createWorkspace: (session: Session, name: string, description: string) =>
    request<Workspace>(session, "/api/workspaces", {
      method: "POST",
      body: JSON.stringify({ name, description }),
    }),
  createDocument: (
    session: Session,
    workspaceId: number,
    title: string,
    description: string,
  ) =>
    request<WorkspaceDocument>(
      session,
      `/api/workspaces/${workspaceId}/documents`,
      { method: "POST", body: JSON.stringify({ title, description }) },
    ),
  updateDocument: (
    session: Session,
    workspaceId: number,
    docId: number,
    title: string,
    description: string,
  ) =>
    request<WorkspaceDocument>(
      session,
      `/api/workspaces/${workspaceId}/documents/${docId}`,
      { method: "PATCH", body: JSON.stringify({ title, description }) },
    ),
  deleteDocument: (session: Session, workspaceId: number, docId: number) =>
    request<{ success: boolean }>(
      session,
      `/api/workspaces/${workspaceId}/documents/${docId}`,
      { method: "DELETE" },
    ),
  listTasks: (session: Session, workspaceId: number) =>
    request<{ tasks: WorkspaceTask[] }>(
      session,
      `/api/workspaces/${workspaceId}/tasks`,
    ),
  createTask: (session: Session, workspaceId: number, title: string) =>
    request<WorkspaceTask>(session, `/api/workspaces/${workspaceId}/tasks`, {
      method: "POST",
      body: JSON.stringify({ title }),
    }),
  updateTask: (
    session: Session,
    workspaceId: number,
    taskId: number,
    completed: boolean,
  ) =>
    request<WorkspaceTask>(
      session,
      `/api/workspaces/${workspaceId}/tasks/${taskId}`,
      { method: "PATCH", body: JSON.stringify({ completed }) },
    ),
  deleteTask: (session: Session, workspaceId: number, taskId: number) =>
    request<{ success: boolean }>(
      session,
      `/api/workspaces/${workspaceId}/tasks/${taskId}`,
      { method: "DELETE" },
    ),
  createInvite: (session: Session, workspaceId: number) =>
    request<{ code: string }>(
      session,
      `/api/workspaces/${workspaceId}/invites`,
      { method: "POST" },
    ),
  joinWorkspace: (session: Session, code: string) =>
    request<Workspace>(session, "/api/workspaces/join", {
      method: "POST",
      body: JSON.stringify({ code }),
    }),
  search: async (session: Session, query: string) => {
    return request<{ results: { roomId: string; content: string }[] }>(
      session,
      `/api/search?q=${encodeURIComponent(query)}`,
    );
  },
  askAI: async (session: Session, question: string) => {
    return request<{ answer?: string; error?: string }>(
      session,
      "/api/ai/ask",
      { method: "POST", body: JSON.stringify({ question }) },
    );
  },
};

export function sessionUser(session: Session) {
  return session.user;
}
