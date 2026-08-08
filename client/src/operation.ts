export type Operation = {
  type: "insert" | "delete";
  position: number;
  text?: string;
  length?: number;
};

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
      type: "insert",
      position: start,
      text: newContent.slice(start, newEnd),
    };
  }

  if (newEnd === start) {
    return {
      type: "delete",
      position: start,
      length: oldEnd - start,
    };
  }

  return {
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