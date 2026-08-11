export type Operation = {
  id: string;
  type: "insert" | "delete";
  position: number;
  text?: string;
  length?: number;
};

function createOperationId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  // Fallback for browsers that do not implement UUid. It only
  // needs to be unique for retry de-duplication within the running server. 
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export function generateOperation(
  oldContent: string,
  newContent: string,
): Operation | null {
  if (oldContent === newContent) {
    return null;
  }

  let start = 0;

  while (
    start < oldContent.length &&
    start < newContent.length &&
    oldContent[start] === newContent[start]
  ) {
    start++;
  }

  let oldEnd = oldContent.length;
  let newEnd = newContent.length;

  while (
    oldEnd > start &&
    newEnd > start &&
    oldContent[oldEnd - 1] === newContent[newEnd - 1]
  ) {
    oldEnd--;
    newEnd--;
  }

  if (oldEnd === start) {
    return {
      id: createOperationId(),
      type: "insert",
      position: start,
      text: newContent.slice(start, newEnd),
    };
  }

  if (newEnd === start) {
    return {
      id: createOperationId(),
      type: "delete",
      position: start,
      length: oldEnd - start,
    };
  }

  return {
    id: createOperationId(),
    type: "delete",
    position: start,
    length: oldEnd - start,
  };
}

export function applyOperation(
  content: string,
  operation: Operation,
): string {
  if (operation.type === "insert") {
    return (
      content.slice(0, operation.position) +
      (operation.text ?? "") +
      content.slice(operation.position)
    );
  }

  if (operation.type === "delete") {
    return (
      content.slice(0, operation.position) +
      content.slice(
        operation.position + (operation.length ?? 0),
      )
    );
  }

  return content;
}
export function transformCursor(
  position: number,
  operation: Operation,
): number {
  if (operation.type === "insert") {
    if (operation.position <= position) {
      return position + (operation.text?.length ?? 0);
    }
  }

  if (operation.type === "delete") {
    if (operation.position < position) {
      const deletedBefore = Math.min(
        operation.length ?? 0,
        position - operation.position,
      );

      return position - deletedBefore;
    }
  }

  return position;
}