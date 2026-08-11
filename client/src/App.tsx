import { useEffect, useLayoutEffect, useRef, useState } from "react";
import {
  generateOperation,
  applyOperation,
  transformCursor,
  type Operation,
} from "./operation";

const ROOM_ID = "room1";
const RECONNECT_BASE_DELAY_MS = 500;
const RECONNECT_MAX_DELAY_MS = 10_000;

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
  const pendingCursorRef = useRef<{
    start: number;
    end: number;
  } | null>(null);
  const pendingOperationsRef = useRef<Operation[]>([]);
  const operationInFlightRef = useRef(false);
  const waitingForSyncRef = useRef(true);
  const hasSynchronizedRef = useRef(false);
  const recoveringRef = useRef(false);
  const reconnectTimerRef = useRef<number | null>(null);
  const reconnectAttemptRef = useRef(0);
  const [connectionStatus, setConnectionStatus] = useState("Connecting…");

  const sendNextOperation = () => {
    const socket = socketRef.current;
    if (
      operationInFlightRef.current ||
      waitingForSyncRef.current ||
      socket?.readyState !== WebSocket.OPEN
    ) {
      return;
    }

    const operation = pendingOperationsRef.current[0];

    if (!operation) {
      return;
    }

    operationInFlightRef.current = true;

    try {
      socket.send(
        JSON.stringify({
          type: "edit",
          roomId: ROOM_ID,
          operation: {
            ...operation,
            baseVersion: versionRef.current,
            timestamp: Date.now(),
          },
          version: versionRef.current,
        }),
      );
    } catch {
      operationInFlightRef.current = false;
    }

    console.log(
      "SENDING QUEUED OPERATION:",
      operation,
      "BASE VERSION:",
      versionRef.current,
    );
  };
  useLayoutEffect(() => {
    const cursor = pendingCursorRef.current;

    if (!cursor || !textareaRef.current) {
      return;
    }

    textareaRef.current.selectionStart = cursor.start;
    textareaRef.current.selectionEnd = cursor.end;

    console.log("LAYOUT CURSOR RESTORE:", cursor.start, "->", cursor.end);

    pendingCursorRef.current = null;
  }, [content]);
  useEffect(() => {
    const usernameKey = "syncspace-username";
    const username =
      sessionStorage.getItem(usernameKey) ??
      `user-${Math.floor(Math.random() * 1000)}`;
    sessionStorage.setItem(usernameKey, username);
    let disposed = false;

    const reapplyPendingOperations = (serverContent: string) => {
      const recoveredContent = pendingOperationsRef.current.reduce(
        (currentContent, operation) => applyOperation(currentContent, operation),
        serverContent,
      );

      setContent(recoveredContent);
      previousContentRef.current = recoveredContent;
    };

    const connect = () => {
      if (disposed) {
        return;
      }

      if (!navigator.onLine) {
        setConnectionStatus("Waiting for network…");
        return;
      }

      const currentSocket = socketRef.current;
      if (
        currentSocket?.readyState === WebSocket.OPEN ||
        currentSocket?.readyState === WebSocket.CONNECTING
      ) {
        return;
      }

      waitingForSyncRef.current = true;
      recoveringRef.current = hasSynchronizedRef.current;
      const protocol = window.location.protocol === "https:" ? "wss" : "ws";
      const socket = new WebSocket(
        `${protocol}://${window.location.hostname}:8080/ws/${ROOM_ID}?username=${encodeURIComponent(username)}`,
      );
      socketRef.current = socket;

      socket.onopen = () => {
        if (socketRef.current !== socket) return;
        reconnectAttemptRef.current = 0;
        setConnectionStatus("Connected — synchronizing…");
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

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }

        // snapshot sending
        reapplyPendingOperations(syncedContent);
        waitingForSyncRef.current = false;
        hasSynchronizedRef.current = true;
        setConnectionStatus("Connected");
        sendNextOperation();
      }
      if (message.type === "edit_ack") {
        console.log("EDIT ACK:", message.version);

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }

        const acknowledgedOperation = pendingOperationsRef.current[0];
        if (
          acknowledgedOperation &&
          (!message.operation || message.operation.id === acknowledgedOperation.id)
        ) {
          pendingOperationsRef.current.shift();
        }

        operationInFlightRef.current = false;

        if (recoveringRef.current && message.content !== undefined) {
          reapplyPendingOperations(message.content);
        }

        if (pendingOperationsRef.current.length === 0) {
          recoveringRef.current = false;
        }

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
          pendingCursorRef.current = {
            start: newSelectionStart,
            end: newSelectionEnd,
          };
          setContent(updatedContent);
          previousContentRef.current = updatedContent;
        }

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }
      }

      if (message.type === "version_conflict") {
        console.log("CONFLICT RECOVERY", message.version);

        const recoveredContent = message.content ?? "";

        versionRef.current = message.version;
        reapplyPendingOperations(recoveredContent);
      }
      if (message.version !== undefined) {
        console.log("DOCUMENT VERSION:", message.version);
      }
      };

      const scheduleReconnect = () => {
        if (disposed || reconnectTimerRef.current !== null) return;

        operationInFlightRef.current = false;
        waitingForSyncRef.current = true;

        if (!navigator.onLine) {
          setConnectionStatus("Waiting for network…");
          return;
        }

        const attempt = reconnectAttemptRef.current++;
        const delay = Math.min(
          RECONNECT_BASE_DELAY_MS * 2 ** attempt,
          RECONNECT_MAX_DELAY_MS,
        );
        setConnectionStatus(`Reconnecting in ${Math.ceil(delay / 1000)}s…`);
        reconnectTimerRef.current = window.setTimeout(() => {
          reconnectTimerRef.current = null;
          connect();
        }, delay);
      };

      socket.onerror = () => socket.close();
      socket.onclose = () => {
        if (socketRef.current === socket) socketRef.current = null;
        scheduleReconnect();
      };
    };

    connect();

    const handleOffline = () => {
      operationInFlightRef.current = false;
      waitingForSyncRef.current = true;
      setConnectionStatus("Waiting for network…");
      socketRef.current?.close();
    };

    const handleOnline = () => {
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      setConnectionStatus("Reconnecting…");
      connect();
    };

    window.addEventListener("offline", handleOffline);
    window.addEventListener("online", handleOnline);

    return () => {
      disposed = true;
      window.removeEventListener("offline", handleOffline);
      window.removeEventListener("online", handleOnline);
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current);
      }
      socketRef.current?.close();
    };
  }, []);
  const handleCursorMove = (e: React.SyntheticEvent<HTMLTextAreaElement>) => {
    const selectionStart = e.currentTarget.selectionStart;
    const selectionEnd = e.currentTarget.selectionEnd;

    const position = selectionEnd;

    if (socketRef.current?.readyState !== WebSocket.OPEN) return;

    socketRef.current.send(
      JSON.stringify({
        type: "cursor_move",
        roomId: ROOM_ID,
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
      <p>{connectionStatus}</p>

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