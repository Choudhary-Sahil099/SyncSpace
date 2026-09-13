import { useState, type RefObject } from "react";
import MarkdownPreview from "./editor/MarkdownPreview";

type DocumentEditorProps = {
  content: string;
  textareaRef: RefObject<HTMLTextAreaElement | null>;
  documentVersion: number;
  errorMessage: string | null;
  onChange: (event: React.ChangeEvent<HTMLTextAreaElement>) => void;
  onSelect: (event: React.SyntheticEvent<HTMLTextAreaElement>) => void;
  title: string;
  description: string;
  remoteCursors?: Record<string, { username: string; position: number; color?: string }>;
  onUpdateDetails?: (title: string, description: string) => Promise<void>;
  onDeleteDocument?: () => Promise<void>;
  onShare?: () => void;
};

function DocumentEditor({
  content,
  textareaRef,
  documentVersion,
  errorMessage,
  onChange,
  onSelect,
  title,
  description,
  remoteCursors = {},
  onUpdateDetails,
  onDeleteDocument,
  onShare,
}: DocumentEditorProps) {
  const [viewMode, setViewMode] = useState<"edit" | "split" | "preview">("edit");
  const [showMenu, setShowMenu] = useState(false);
  const [showRenameModal, setShowRenameModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [newTitle, setNewTitle] = useState(title);
  const [newDesc, setNewDesc] = useState(description);
  const [isSavingDetails, setIsSavingDetails] = useState(false);
  const [copyToast, setCopyToast] = useState<string | null>(null);

  const wordCount = content.trim() ? content.trim().split(/\s+/).length : 0;
  const readTime = Math.max(1, Math.ceil(wordCount / 200));

  const applyFormatting = (prefix: string, suffix: string = "", placeholder: string = "text") => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = content.slice(start, end);
    const replacement = selectedText ? `${prefix}${selectedText}${suffix}` : `${prefix}${placeholder}${suffix}`;

    const newContent = content.slice(0, start) + replacement + content.slice(end);

    // Update using native property setter so React's synthetic event handler receives the input
    const nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype, "value")?.set;
    nativeSetter?.call(textarea, newContent);

    const event = new Event("input", { bubbles: true });
    textarea.dispatchEvent(event);

    setTimeout(() => {
      textarea.focus();
      const newCursorStart = start + prefix.length;
      const newCursorEnd = selectedText ? newCursorStart + selectedText.length : newCursorStart + placeholder.length;
      textarea.setSelectionRange(newCursorStart, newCursorEnd);
    }, 10);
  };

  const handleShareClick = () => {
    if (onShare) {
      onShare();
    } else {
      void navigator.clipboard.writeText(window.location.href);
      setCopyToast("Link copied to clipboard!");
      setTimeout(() => setCopyToast(null), 2500);
    }
  };

  const handleExport = () => {
    const blob = new Blob([content], { type: "text/markdown;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${title.toLowerCase().replace(/[^a-z0-9]+/g, "-") || "document"}.md`;
    a.click();
    URL.revokeObjectURL(url);
    setShowMenu(false);
  };

  const handleSaveDetails = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim() || !onUpdateDetails) return;
    setIsSavingDetails(true);
    try {
      await onUpdateDetails(newTitle, newDesc);
      setShowRenameModal(false);
      setShowMenu(false);
    } catch {
    
    } finally {
      setIsSavingDetails(false);
    }
  };

  const handleDelete = async () => {
    if (!onDeleteDocument) return;
    try {
      await onDeleteDocument();
    } catch {}
  };

  const collaboratorList = Object.values(remoteCursors);

  return (
    <section className="document-area">
      <div className="document-header">
        <div className="document-heading">
          <div className="document-breadcrumb">
            <span>Workspace</span>
            <span>/</span>
            <span>Documents</span>
            <span>/</span>
            <strong>{title}</strong>
          </div>

          <div className="document-type">
            <span className="document-type-icon">◆</span>
            COLLABORATIVE DOCUMENT
          </div>

          <h1>{title}</h1>

          <p>{description || "Real-time collaborative document."}</p>

          <div className="document-meta">
            <span>
              <span className="meta-icon">◷</span>
              Live Sync
            </span>

            <span className="meta-divider" />

            <span>Version {documentVersion}</span>

            <span className="meta-divider" />

            <span>{wordCount} words · {readTime} min read</span>

            {collaboratorList.length > 0 && (
              <>
                <span className="meta-divider" />
                <div className="collaborator-tags">
                  {collaboratorList.map((c, i) => (
                    <span key={i} className="collab-tag" title={`${c.username} is editing`}>
                      <span className="collab-pulse" />
                      {c.username}
                    </span>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>

        <div className="document-actions">
          <button type="button" className="document-action" onClick={handleShareClick} title="Share document">
            ↗ <span>Share</span>
          </button>

          <div className="dropdown-container">
            <button
              type="button"
              className="document-action"
              onClick={() => setShowMenu((v) => !v)}
              aria-label="More options"
            >
              ⋯
            </button>

            {showMenu && (
              <div className="document-menu-dropdown">
                {onUpdateDetails && (
                  <button
                    type="button"
                    onClick={() => {
                      setNewTitle(title);
                      setNewDesc(description);
                      setShowRenameModal(true);
                      setShowMenu(false);
                    }}
                  >
                    ✏️ Rename / Edit Details
                  </button>
                )}
                <button type="button" onClick={handleExport}>
                  📥 Export as Markdown
                </button>
                {onDeleteDocument && (
                  <button
                    type="button"
                    className="danger-action"
                    onClick={() => {
                      setShowDeleteModal(true);
                      setShowMenu(false);
                    }}
                  >
                    🗑️ Delete Document
                  </button>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {copyToast && <div className="floating-toast">✓ {copyToast}</div>}

      {errorMessage && (
        <div className="notice">
          <span className="notice-icon">⚠</span>
          <span>{errorMessage}</span>
        </div>
      )}

      <div className="editor-wrapper">
        <div className="editor-toolbar">
          <div className="toolbar-group">
            <button
              type="button"
              className="toolbar-button"
              title="Bold (**text**)"
              onClick={() => applyFormatting("**", "**", "bold text")}
            >
              <strong>B</strong>
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Italic (*text*)"
              onClick={() => applyFormatting("*", "*", "italic text")}
            >
              <i>I</i>
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Strikethrough (~~text~~)"
              onClick={() => applyFormatting("~~", "~~", "strikethrough")}
            >
              <s>S</s>
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Inline Code (`code`)"
              onClick={() => applyFormatting("`", "`", "code")}
            >
              {"<>"}
            </button>
          </div>

          <span className="toolbar-divider" />

          <div className="toolbar-group">
            <button
              type="button"
              className="toolbar-button"
              title="Heading 1"
              onClick={() => applyFormatting("# ", "", "Heading 1")}
            >
              H1
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Heading 2"
              onClick={() => applyFormatting("## ", "", "Heading 2")}
            >
              H2
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Heading 3"
              onClick={() => applyFormatting("### ", "", "Heading 3")}
            >
              H3
            </button>
          </div>

          <span className="toolbar-divider" />

          <div className="toolbar-group">
            <button
              type="button"
              className="toolbar-button toolbar-wide"
              title="Bullet List"
              onClick={() => applyFormatting("- ", "", "Item")}
            >
              • List
            </button>

            <button
              type="button"
              className="toolbar-button toolbar-wide"
              title="Task Checkbox"
              onClick={() => applyFormatting("- [ ] ", "", "New task")}
            >
              ☑ Task
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Blockquote"
              onClick={() => applyFormatting("> ", "", "Quote")}
            >
              ❝
            </button>

            <button
              type="button"
              className="toolbar-button"
              title="Code Block"
              onClick={() => applyFormatting("```\n", "\n```", "code here")}
            >
              {`{ }`}
            </button>
          </div>

          <span className="toolbar-divider" />

          {/* View Mode Controls */}
          <div className="view-mode-selector">
            <button
              type="button"
              className={`mode-btn ${viewMode === "edit" ? "active" : ""}`}
              onClick={() => setViewMode("edit")}
              title="Raw Editor"
            >
              Edit
            </button>
            <button
              type="button"
              className={`mode-btn ${viewMode === "split" ? "active" : ""}`}
              onClick={() => setViewMode("split")}
              title="Side-by-side Edit & Preview"
            >
              Split
            </button>
            <button
              type="button"
              className={`mode-btn ${viewMode === "preview" ? "active" : ""}`}
              onClick={() => setViewMode("preview")}
              title="Rendered Markdown Preview"
            >
              Preview
            </button>
          </div>
        </div>

        <div className={`editor-content-container view-${viewMode}`}>
          {(viewMode === "edit" || viewMode === "split") && (
            <div className="editor-pane">
              <textarea
                ref={textareaRef}
                value={content}
                onChange={onChange}
                onSelect={onSelect}
                placeholder="Start writing your document notes, architecture, or tasks in Markdown…"
                spellCheck={false}
              />
            </div>
          )}

          {(viewMode === "preview" || viewMode === "split") && (
            <div className="preview-pane">
              <MarkdownPreview content={content} />
            </div>
          )}
        </div>

        <div className="editor-footer">
          <div className="save-state">
            <span className="save-dot" />
            <span>
              {navigator.onLine ? "WebSocket OT Active" : "Working offline (queued)"}
            </span>
          </div>

          <div className="editor-footer-right">
            <span className="editor-status">
              <span className="status-dot" />
              Synced
            </span>

            <span className="footer-divider" />

            <span>Version {documentVersion}</span>

            <span className="footer-divider" />

            <span>{content.length} chars</span>
          </div>
        </div>
      </div>

      {/* Rename Document Modal */}
      {showRenameModal && (
        <div className="modal-backdrop" onMouseDown={() => setShowRenameModal(false)}>
          <div className="modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setShowRenameModal(false)}>
              ×
            </button>
            <p className="eyebrow">DOCUMENT</p>
            <h2>Edit Document Details</h2>
            <form onSubmit={handleSaveDetails}>
              <label className="form-label">
                Title
                <input
                  autoFocus
                  required
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  placeholder="Document title"
                />
              </label>
              <label className="form-label">
                Description
                <textarea
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  placeholder="Document description"
                />
              </label>
              <div className="modal-actions">
                <button type="button" className="secondary-button" onClick={() => setShowRenameModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="primary-button" disabled={isSavingDetails || !newTitle.trim()}>
                  {isSavingDetails ? "Saving…" : "Save Changes"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete */}
      {showDeleteModal && (
        <div className="modal-backdrop" onMouseDown={() => setShowDeleteModal(false)}>
          <div className="modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setShowDeleteModal(false)}>
              ×
            </button>
            <p className="eyebrow" style={{ color: "#ef4444" }}>DANGER</p>
            <h2>Delete Document</h2>
            <p className="modal-copy">
              Are you sure you want to delete <strong>"{title}"</strong>? This will permanently remove the document from this workspace.
            </p>
            <div className="modal-actions">
              <button type="button" className="secondary-button" onClick={() => setShowDeleteModal(false)}>
                Cancel
              </button>
              <button type="button" className="danger-button" onClick={handleDelete}>
                Delete Permanently
              </button>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

export default DocumentEditor;