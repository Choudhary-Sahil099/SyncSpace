import type { WorkspaceDocument } from "../../api/api";

type Props = {
  document: WorkspaceDocument;
  onClick: () => void;
  onDelete?: () => void;
};

export default function DocumentCard({ document, onClick, onDelete }: Props) {
  const formattedDate = document.updatedAt
    ? new Date(document.updatedAt).toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "Recently updated";

  return (
    <div className="document-card" onClick={onClick} role="button" tabIndex={0} onKeyDown={(e) => e.key === "Enter" && onClick()}>
      <div className="document-card-top">
        <div className="document-card-icon">📄</div>
        {onDelete && (
          <button
            type="button"
            className="card-delete-btn"
            title="Delete document"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
          >
            ×
          </button>
        )}
      </div>

      <div className="document-card-content">
        <h3>{document.title}</h3>
        <p>{document.description || "Collaborative document."}</p>
      </div>

      <div className="document-card-footer">
        <span className="card-time">{formattedDate}</span>
        <span className="card-arrow">Open →</span>
      </div>
    </div>
  );
}
