import { useState } from "react";

type SearchResult = {
  roomId: string;
  content: string;
};

type RemoteCursor = {
  username: string;
  position: number;
  selectionStart: number;
  selectionEnd: number;
  color: string;
};

type RightPanelProps = {
  aiQuestion: string;
  aiAnswer: string;
  searchResults: SearchResult[];
  isSearching: boolean;
  searchQuery?: string;
  users: string[];
  remoteCursors: Record<string, RemoteCursor>;

  onAIQuestionChange: (value: string) => void;
  onAskAI: () => void;
  onSearchQueryChange?: (value: string) => void;
  onSearch?: () => void;
  onSelectSearchResult: (content: string) => void;

  username: string;
};

const USER_COLORS = [
  "#6366f1", 
  "#ec4899", 
  "#10b981", 
  "#f59e0b", 
  "#06b6d4", 
  "#8b5cf6",
];

function getUserColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return USER_COLORS[Math.abs(hash) % USER_COLORS.length];
}

function RightPanel({
  aiQuestion,
  aiAnswer,
  searchResults,
  isSearching,
  searchQuery = "",
  users,
  remoteCursors,
  onAIQuestionChange,
  onAskAI,
  onSearchQueryChange,
  onSearch,
  onSelectSearchResult,
  username,
}: RightPanelProps) {
  const [copiedAnswer, setCopiedAnswer] = useState(false);

  const handlePromptChip = (text: string) => {
    onAIQuestionChange(text);
  };

  const handleCopyAnswer = () => {
    if (!aiAnswer) return;
    void navigator.clipboard.writeText(aiAnswer);
    setCopiedAnswer(true);
    setTimeout(() => setCopiedAnswer(false), 2000);
  };

  return (
    <aside className="right-panel">
      <section className="side-card ai-card">
        <div className="side-card-header">
          <div className="card-icon ai-icon">✦</div>
          <div>
            <strong>SyncSpace AI</strong>
            <span>Workspace Assistant</span>
          </div>
        </div>

        <p>Ask questions about your workspace notes, architecture, or documents.</p>

        <div className="ai-chips">
          <button type="button" className="ai-chip" onClick={() => handlePromptChip("Summarize this document's main points")}>
            ⚡ Summarize
          </button>
          <button type="button" className="ai-chip" onClick={() => handlePromptChip("Extract key action items and next steps")}>
            ✓ Action items
          </button>
          <button type="button" className="ai-chip" onClick={() => handlePromptChip("Review the architecture and identify risks")}>
            🛡 Architecture
          </button>
        </div>

        <div className="ai-input-wrap">
          <textarea
            value={aiQuestion}
            onChange={(e) => onAIQuestionChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                onAskAI();
              }
            }}
            placeholder="Ask AI or generate content..."
          />
        </div>

        <button className="primary-button ai-submit-button" type="button" onClick={onAskAI}>
          <span>Ask AI Assistant</span>
          <span>→</span>
        </button>

        {aiAnswer && (
          <div className="ai-answer-wrap">
            <div className="ai-answer-header">
              <span>RESPONSE</span>
              <button type="button" className="copy-btn" onClick={handleCopyAnswer}>
                {copiedAnswer ? "✓ Copied" : "Copy"}
              </button>
            </div>
            <div className="ai-answer">{aiAnswer}</div>
          </div>
        )}
      </section>

      <section className="side-card">
        <div className="side-card-header">
          <div className="card-icon search-icon">⌕</div>
          <div>
            <strong>Workspace Search</strong>
            <span>Full-Text Document Retrieval</span>
          </div>
        </div>

        {onSearchQueryChange && onSearch && (
          <form
            className="search-form"
            onSubmit={(e) => {
              e.preventDefault();
              onSearch();
            }}
          >
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => onSearchQueryChange(e.target.value)}
              placeholder="Search across documents…"
            />
            <button type="submit" disabled={isSearching || !searchQuery.trim()}>
              {isSearching ? "…" : "Search"}
            </button>
          </form>
        )}

        {isSearching ? (
          <div className="search-loading">
            <span className="loading-spinner" />
            <p>Searching workspace documents…</p>
          </div>
        ) : searchResults.length > 0 ? (
          <div className="search-results">
            <div className="results-count">{searchResults.length} match(es) found</div>
            {searchResults.map((result) => (
              <button
                className="result"
                key={result.roomId}
                type="button"
                onClick={() => onSelectSearchResult(result.content)}
                title="Click to view or insert content"
              >
                <div className="result-header">
                  <strong>Room: {result.roomId}</strong>
                  <span className="result-badge">MATCH</span>
                </div>
                <span>{result.content.slice(0, 120)}...</span>
              </button>
            ))}
          </div>
        ) : (
          <div className="empty-search">
            <span>⌕</span>
            <p>Type keywords above to search all workspace documents.</p>
          </div>
        )}
      </section>

      <section className="side-card">
        <div className="side-card-header">
          <div className="card-icon collab-icon">♢</div>
          <div>
            <strong>Live Collaboration</strong>
            <span>{users.length || 1} online in room</span>
          </div>
        </div>

        <div className="presence-list">
          <div className="presence-user">
            <span className="presence-avatar" style={{ background: getUserColor(username) }}>
              {username[0]?.toUpperCase() ?? "U"}
            </span>
            <div>
              <strong>{username}</strong>
              <span className="user-sub">Active now (You)</span>
            </div>
            <span className="online-indicator" title="Connected" />
          </div>

          {Object.entries(remoteCursors).map(([userId, cursor]) => (
            <div className="presence-user" key={userId}>
              <span className="presence-avatar remote" style={{ background: getUserColor(cursor.username) }}>
                {cursor.username[0]?.toUpperCase() ?? "C"}
              </span>
              <div>
                <strong>{cursor.username}</strong>
                <span className="user-sub">Editing at pos {cursor.position}</span>
              </div>
              <span className="online-indicator active" title="Active collaborator" />
            </div>
          ))}
        </div>
      </section>
      <section className="side-card system-card">
        <div className="side-card-header">
          <div className="card-icon system-icon">⌁</div>
          <div>
            <strong>System Status</strong>
            <span>Realtime Infrastructure</span>
          </div>
        </div>

        <div className="system-row">
          <span>Engine</span>
          <strong className="badge-highlight">WebSocket + OT</strong>
        </div>

        <div className="system-row">
          <span>Sync State</span>
          <span className="status-badge live">● Authoritative Hub</span>
        </div>

        <div className="system-row">
          <span>Fanout</span>
          <strong>Redis Pub/Sub</strong>
        </div>

        <div className="system-row">
          <span>Persistence</span>
          <strong>PostgreSQL / FTS</strong>
        </div>
      </section>
    </aside>
  );
}

export default RightPanel;