import React from "react";
import "./editor.css";

type Props = {
  content: string;
};

export default function MarkdownPreview({ content }: Props) {
  if (!content.trim()) {
    return (
      <div className="markdown-preview empty">
        <p className="placeholder-text">Nothing to preview yet. Start typing markdown in the editor!</p>
      </div>
    );
  }

  const lines = content.split("\n");
  const renderedElements: React.ReactNode[] = [];
  let inCodeBlock = false;
  let codeBlockContent: string[] = [];
  let codeBlockLang = "";

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Handle code block toggle
    if (line.trim().startsWith("```")) {
      if (inCodeBlock) {
        renderedElements.push(
          <div className="preview-codeblock" key={`code-${i}`}>
            {codeBlockLang && <div className="codeblock-lang">{codeBlockLang}</div>}
            <pre>
              <code>{codeBlockContent.join("\n")}</code>
            </pre>
          </div>
        );
        inCodeBlock = false;
        codeBlockContent = [];
        codeBlockLang = "";
      } else {
        inCodeBlock = true;
        codeBlockLang = line.trim().replace(/^```/, "").trim();
      }
      continue;
    }

    if (inCodeBlock) {
      codeBlockContent.push(line);
      continue;
    }

    // Horizontal Rule
    if (line.trim() === "---" || line.trim() === "***" || line.trim() === "___") {
      renderedElements.push(<hr key={`hr-${i}`} className="preview-hr" />);
      continue;
    }

    // Headings
    if (line.startsWith("# ")) {
      renderedElements.push(
        <h1 key={`h1-${i}`} className="preview-h1">
          {renderInline(line.slice(2))}
        </h1>
      );
      continue;
    }
    if (line.startsWith("## ")) {
      renderedElements.push(
        <h2 key={`h2-${i}`} className="preview-h2">
          {renderInline(line.slice(3))}
        </h2>
      );
      continue;
    }
    if (line.startsWith("### ")) {
      renderedElements.push(
        <h3 key={`h3-${i}`} className="preview-h3">
          {renderInline(line.slice(4))}
        </h3>
      );
      continue;
    }
    if (line.startsWith("#### ")) {
      renderedElements.push(
        <h4 key={`h4-${i}`} className="preview-h4">
          {renderInline(line.slice(5))}
        </h4>
      );
      continue;
    }

    // Blockquote
    if (line.startsWith("> ")) {
      renderedElements.push(
        <blockquote key={`quote-${i}`} className="preview-quote">
          {renderInline(line.slice(2))}
        </blockquote>
      );
      continue;
    }

    // Task list items: - [ ] or - [x]
    if (/^\s*[-*]\s*\[([ xX])\]\s+(.*)/.test(line)) {
      const match = line.match(/^\s*[-*]\s*\[([ xX])\]\s+(.*)/);
      if (match) {
        const checked = match[1].toLowerCase() === "x";
        const taskText = match[2];
        renderedElements.push(
          <div key={`task-${i}`} className={`preview-task ${checked ? "completed" : ""}`}>
            <input type="checkbox" checked={checked} readOnly />
            <span>{renderInline(taskText)}</span>
          </div>
        );
        continue;
      }
    }

    // Bullet List
    if (/^\s*[-*]\s+(.*)/.test(line)) {
      const match = line.match(/^\s*[-*]\s+(.*)/);
      if (match) {
        renderedElements.push(
          <div key={`li-${i}`} className="preview-bullet">
            <span className="bullet-dot">•</span>
            <span>{renderInline(match[1])}</span>
          </div>
        );
        continue;
      }
    }

    // Numbered List
    if (/^\s*(\d+)\.\s+(.*)/.test(line)) {
      const match = line.match(/^\s*(\d+)\.\s+(.*)/);
      if (match) {
        renderedElements.push(
          <div key={`num-${i}`} className="preview-numbered">
            <span className="numbered-prefix">{match[1]}.</span>
            <span>{renderInline(match[2])}</span>
          </div>
        );
        continue;
      }
    }

    // Empty line
    if (!line.trim()) {
      renderedElements.push(<div key={`empty-${i}`} className="preview-spacer" />);
      continue;
    }

    // Regular paragraph
    renderedElements.push(
      <p key={`p-${i}`} className="preview-p">
        {renderInline(line)}
      </p>
    );
  }

  // Close dangling code block if any
  if (inCodeBlock && codeBlockContent.length > 0) {
    renderedElements.push(
      <div className="preview-codeblock" key="code-dangling">
        <pre>
          <code>{codeBlockContent.join("\n")}</code>
        </pre>
      </div>
    );
  }

  return <div className="markdown-preview">{renderedElements}</div>;
}

/**
 * Parses bold, italic, strikethrough, inline code, links
 */
function renderInline(text: string): React.ReactNode[] {
  const parts: React.ReactNode[] = [];
  let remaining = text;
  let keyIndex = 0;

  while (remaining.length > 0) {
    // Inline code: `code`
    const codeMatch = remaining.match(/^`([^`]+)`/);
    if (codeMatch) {
      parts.push(<code key={`inline-code-${keyIndex++}`} className="inline-code">{codeMatch[1]}</code>);
      remaining = remaining.slice(codeMatch[0].length);
      continue;
    }

    // Bold + Italic: ***text***
    const boldItalicMatch = remaining.match(/^\*\*\*([^*]+)\*\*\*/);
    if (boldItalicMatch) {
      parts.push(<strong key={`bi-${keyIndex++}`}><em>{boldItalicMatch[1]}</em></strong>);
      remaining = remaining.slice(boldItalicMatch[0].length);
      continue;
    }

    // Bold: **text**
    const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/);
    if (boldMatch) {
      parts.push(<strong key={`b-${keyIndex++}`}>{boldMatch[1]}</strong>);
      remaining = remaining.slice(boldMatch[0].length);
      continue;
    }

    // Italic: *text* or _text_
    const italicMatch = remaining.match(/^\*([^*]+)\*/) || remaining.match(/^_([^_]+)_/);
    if (italicMatch) {
      parts.push(<em key={`i-${keyIndex++}`}>{italicMatch[1]}</em>);
      remaining = remaining.slice(italicMatch[0].length);
      continue;
    }

    // Strikethrough: ~~text~~
    const strikeMatch = remaining.match(/^~~([^~]+)~~/);
    if (strikeMatch) {
      parts.push(<del key={`s-${keyIndex++}`}>{strikeMatch[1]}</del>);
      remaining = remaining.slice(strikeMatch[0].length);
      continue;
    }

    // Link: [text](url)
    const linkMatch = remaining.match(/^\[([^\]]+)\]\(([^)]+)\)/);
    if (linkMatch) {
      parts.push(
        <a key={`a-${keyIndex++}`} href={linkMatch[2]} target="_blank" rel="noopener noreferrer" className="preview-link">
          {linkMatch[1]}
        </a>
      );
      remaining = remaining.slice(linkMatch[0].length);
      continue;
    }

    // Find next special character
    const nextSpecial = remaining.search(/[`*_~\[]/);
    if (nextSpecial === -1) {
      parts.push(remaining);
      break;
    } else if (nextSpecial === 0) {
      // Character didn't match full pattern, take it as plain text
      parts.push(remaining[0]);
      remaining = remaining.slice(1);
    } else {
      parts.push(remaining.slice(0, nextSpecial));
      remaining = remaining.slice(nextSpecial);
    }
  }

  return parts;
}
