import { useEffect, useRef, useState } from "react";
import {
  generateOperation,
  applyOperation,
  transformCursor,
  type Operation,
} from "./operation";
function App() {
  const [content, setContent] = useState("");
  const [users, setUsers] = useState<string[]>([]);
  const socketRef = useRef<WebSocket | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const previousContentRef = useRef("");
  type RemoteCursor = {
    username: string;
    position: number;
    selectionStart: number;
    selectionEnd: number;
    color: string;
  };

  const [remoteCursors, setRemoteCursors] = useState<
    Record<string, RemoteCursor>
  >({});
  const versionRef = useRef(0);
  const pendingOperationsRef = useRef<Operation[]>([]);
  const operationInFlightRef = useRef(false);

  const sendNextOperation = () => {
    if (operationInFlightRef.current) {
      return;
    }

    const operation = pendingOperationsRef.current[0];

    if (!operation) {
      return;
    }

    operationInFlightRef.current = true;

    socketRef.current?.send(
      JSON.stringify({
        type: "edit",
        roomId: "room1",
        operation: {
          ...operation,
          baseVersion: versionRef.current,
          timestamp: Date.now(),
        },
        version: versionRef.current,
      }),
    );

    console.log(
      "SENDING QUEUED OPERATION:",
      operation,
      "BASE VERSION:",
      versionRef.current,
    );
  };
  useEffect(() => {
    const username = "user-" + Math.floor(Math.random() * 1000);

    const socket = new WebSocket(
      `ws://${window.location.hostname}:8080/ws/room1?username=${username}`,
    );

    socketRef.current = socket;

    socket.onopen = () => {
      console.log("CONNECTED");
    };

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);

      console.log("MESSAGE:", message);
      if (message.type === "users_list") {
        setUsers(message.users);
      }
      if (message.type === "cursor_move") {
        setRemoteCursors((prev) => ({
          ...prev,
          [message.userId]: {
            username: message.username,
            position: message.cursor.position,
            selectionStart: message.cursor.selectionStart,
            selectionEnd: message.cursor.selectionEnd,
            color: "blue",
          },
        }));

        console.log(
          "REMOTE CURSOR:",
          message.username,
          message.cursor.position,
        );
      }
      if (message.type === "document_sync") {
        const syncedContent = message.content ?? "";

        setContent(syncedContent);

        previousContentRef.current = syncedContent;

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }
      }
      if (message.type === "edit_ack") {
        console.log("EDIT ACK:", message.version);

        versionRef.current = message.version;
        pendingOperationsRef.current.shift();

        operationInFlightRef.current = false;

        console.log(
          "OPERATION ACKNOWLEDGED",
          "REMAINING QUEUE:",
          pendingOperationsRef.current.length,
        );
        sendNextOperation();
      }
      if (message.type === "edit") {
        if (message.operation) {
          const selectionStart = textareaRef.current?.selectionStart ?? 0;

          const selectionEnd = textareaRef.current?.selectionEnd ?? 0;
          console.log("REMOTE EDIT DEBUG:", {
            currentContent: previousContentRef.current,
            cursorStart: selectionStart,
            cursorEnd: selectionEnd,
            operation: message.operation,
          });
          const newSelectionStart = transformCursor(
            selectionStart,
            message.operation,
          );

          const newSelectionEnd = transformCursor(
            selectionEnd,
            message.operation,
          );
          console.log("CURSOR BEFORE:", selectionStart, "->", selectionEnd);

          console.log(
            "CURSOR AFTER TRANSFORM:",
            newSelectionStart,
            "->",
            newSelectionEnd,
          );
          const updatedContent = applyOperation(
            previousContentRef.current,
            message.operation,
          );

          setContent(updatedContent);
          previousContentRef.current = updatedContent;
          requestAnimationFrame(() => {
            console.log(
              "RESTORING CURSOR:",
              newSelectionStart,
              "->",
              newSelectionEnd,
            );
            if (textareaRef.current) {
              textareaRef.current.selectionStart = newSelectionStart;
              textareaRef.current.selectionEnd = newSelectionEnd;
            }

            console.log(
              "CURSOR AFTER RESTORE:",
              textareaRef.current?.selectionStart,
              "->",
              textareaRef.current?.selectionEnd,
            );
          });
        }

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }
      }

      if (message.type === "version_conflict") {
        console.log("CONFLICT RECOVERY", message.version);

        const recoveredContent = message.content ?? "";

        setContent(recoveredContent);

        previousContentRef.current = recoveredContent;

        versionRef.current = message.version;
      }
      if (message.version !== undefined) {
        console.log("DOCUMENT VERSION:", message.version);
      }
    };

    return () => {
      socket.close();
    };
  }, []);
  const handleCursorMove = (e: React.SyntheticEvent<HTMLTextAreaElement>) => {
    const selectionStart = e.currentTarget.selectionStart;
    const selectionEnd = e.currentTarget.selectionEnd;

    const position = selectionEnd;

    socketRef.current?.send(
      JSON.stringify({
        type: "cursor_move",
        roomId: "room1",
        cursor: {
          position,
          selectionStart,
          selectionEnd,
        },
      }),
    );

    console.log(
      "CURSOR SENT:",
      position,
      "Selection:",
      selectionStart,
      "->",
      selectionEnd,
    );
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;

    const oldContent = previousContentRef.current ?? "";

    const operation = generateOperation(oldContent, newContent);
    console.log("OLD:", JSON.stringify(oldContent));
    console.log("NEW:", JSON.stringify(newContent));
    console.log("GENERATED OP:", operation);

    setContent(newContent);

    previousContentRef.current = newContent;

    if (!operation) {
      return;
    }

    pendingOperationsRef.current.push(operation);

    console.log(
      "OPERATION QUEUED:",
      operation,
      "QUEUE SIZE:",
      pendingOperationsRef.current.length,
    );

    sendNextOperation();
  };

  return (
    <div style={{ padding: "40px" }}>
      <h1>SyncSpace</h1>

      <textarea
        ref={textareaRef}
        value={content}
        onChange={handleChange}
        onSelect={handleCursorMove}
        style={{
          width: "100%",
          height: "500px",
          fontSize: "18px",
        }}
      />
      <h3>Remote Cursors</h3>

      {Object.entries(remoteCursors).map(([userId, cursor]) => (
        <div
          key={userId}
          style={{
            color: cursor.color,
            fontWeight: "bold",
            marginBottom: "12px",
          }}
        >
          <div>● {cursor.username}</div>

          <div>Cursor: {cursor.position}</div>

          <div>
            Selection: {cursor.selectionStart} → {cursor.selectionEnd}
          </div>
        </div>
      ))}
      <div>
        <h3>Collaborators</h3>

        {users.map((user) => (
          <div key={user}>{user}</div>
        ))}
      </div>
    </div>
  );
}

export default App;
