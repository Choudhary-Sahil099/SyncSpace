import { useEffect, useState } from "react";
import type { Session } from "../../hooks/useCollaboration";
import "./styles/topbar.css";

type Props = {
  connectionStatus: string;
  workspaceName: string;
  page: "home" | "workspace";
  onHome: () => void;
  session?: Session;
};

export default function Topbar({
  connectionStatus,
  workspaceName,
  page,
  onHome,
  session,
}: Props) {
  const [notifications, setNotifications] = useState(false);
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    return (localStorage.getItem("syncspace-theme") as "light" | "dark") || "light";
  });

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    localStorage.setItem("syncspace-theme", theme);
  }, [theme]);

  // Global Ctrl+K shortcut to focus search
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        const searchEl =
          document.getElementById("workspace-search") ||
          (document.querySelector(".search-form input") as HTMLElement);
        searchEl?.focus();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const toggleTheme = () => {
    setTheme((prev) => (prev === "light" ? "dark" : "light"));
  };

  const connected = connectionStatus.toLowerCase().includes("connected");
  const initial = session?.user.username.slice(0, 2).toUpperCase() ?? "U";

  const handleSearchClick = () => {
    const searchEl =
      document.getElementById("workspace-search") ||
      (document.querySelector(".search-form input") as HTMLElement);
    searchEl?.focus();
  };

  return (
    <header className="topbar">
      <div className="breadcrumb">
        <button type="button" onClick={onHome} className="breadcrumb-link">
          {workspaceName}
        </button>
        <span className="breadcrumb-separator">/</span>
        <strong>{page === "home" ? "Home Overview" : "Document Canvas"}</strong>
      </div>

      <div className="topbar-actions">
        <div className={`connection-status ${connected ? "connected" : "disconnected"}`}>
          <span className="connection-dot" />
          <span>{connectionStatus}</span>
        </div>

        <button type="button" className="search-button" onClick={handleSearchClick}>
          <span>⌕</span>
          <span>Search</span>
          <kbd>Ctrl K</kbd>
        </button>

        {/* Theme Toggle Button */}
        <button
          type="button"
          className="theme-toggle-button"
          onClick={toggleTheme}
          title={`Switch to ${theme === "light" ? "Dark" : "Light"} mode`}
          aria-label="Toggle theme"
        >
          {theme === "light" ? "🌙" : "☀️"}
        </button>

        <div className="notification-wrap">
          <button
            type="button"
            className="icon-button"
            aria-label="Notifications"
            onClick={() => setNotifications((v) => !v)}
          >
            🔔
          </button>
          {notifications && (
            <div className="notification-popover">
              <div className="popover-header">
                <strong>Notifications</strong>
                <span className="badge-new">NEW</span>
              </div>
              <p>
                {page === "workspace"
                  ? "Real-time collaboration and Operational Transformation are active."
                  : "Welcome to SyncSpace. Your workspace is synced."}
              </p>
            </div>
          )}
        </div>

        <div className="user-avatar" title={`Signed in as ${session?.user.email ?? "User"}`}>
          {initial}
        </div>
      </div>
    </header>
  );
}
