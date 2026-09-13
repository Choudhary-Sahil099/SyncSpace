import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

import {
  loadOfflineDocument,
  saveOfflineDocument,
} from "../offlineStore";

import {
  generateOperations,
  applyOperation,
  transformCursor,
  type Operation,
} from "../operation";

const RECONNECT_BASE_DELAY_MS = 500;
const RECONNECT_MAX_DELAY_MS = 10_000;

const API_BASE =
  import.meta.env.VITE_API_URL ??
  `${window.location.protocol}//${window.location.hostname}:8080`;

const WS_BASE = API_BASE.replace(/^http/, "ws");

export type Session = {
  accessToken: string;
  user: {
    id: number;
    email: string;
    username: string;
  };
};

export type RemoteCursor = {
  username: string;
  position: number;
  selectionStart: number;
  selectionEnd: number;
  color: string;
};

export type SearchResult = {
  roomId: string;
  content: string;
};

function useCollaboration(session: Session, roomId: string) {
  const [content, setContent] = useState("");

  const [users, setUsers] = useState<string[]>([]);

  const [remoteCursors, setRemoteCursors] =
    useState<Record<string, RemoteCursor>>({});

  const [connectionStatus, setConnectionStatus] =
    useState("Connecting…");

  const [errorMessage, setErrorMessage] =
    useState<string | null>(null);

  const [documentVersion, setDocumentVersion] =
    useState(0);

  const [searchQuery, setSearchQuery] =
    useState("");

  const [searchResults, setSearchResults] =
    useState<SearchResult[]>([]);

  const [isSearching, setIsSearching] =
    useState(false);

  const [aiQuestion, setAiQuestion] =
    useState("");

  const [aiAnswer, setAiAnswer] =
    useState("");

  const socketRef =
    useRef<WebSocket | null>(null);

  const textareaRef =
    useRef<HTMLTextAreaElement | null>(null);

  const previousContentRef =
    useRef("");

  const versionRef =
    useRef(0);

  const pendingOperationsRef =
    useRef<Operation[]>([]);

  const operationInFlightRef =
    useRef(false);

  const waitingForSyncRef =
    useRef(true);

  const hasSynchronizedRef =
    useRef(false);

  const recoveringRef =
    useRef(false);

  const pendingCursorRef =
    useRef<{
      start: number;
      end: number;
    } | null>(null);

  const reconnectTimerRef =
    useRef<number | null>(null);

  const reconnectAttemptRef =
    useRef(0);

  /*
   * Send the next operation in the local queue.
   */
  const sendNextOperation = () => {
    const socket = socketRef.current;

    if (
      operationInFlightRef.current ||
      waitingForSyncRef.current ||
      socket?.readyState !== WebSocket.OPEN
    ) {
      return;
    }

    const operation =
      pendingOperationsRef.current[0];

    if (!operation) {
      return;
    }

    operationInFlightRef.current = true;

    try {
      socket.send(
        JSON.stringify({
          type: "edit",
          roomId: roomId,
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

      setErrorMessage(
        "Your edit is saved locally and will retry when the connection returns.",
      );
    }
  };

  /*
   * Reapply edits that were made locally but
   * haven't received an acknowledgement yet.
   */
  const reapplyPendingOperations = (
    serverContent: string,
  ) => {
    const recoveredContent =
      pendingOperationsRef.current.reduce(
        (currentContent, operation) =>
          applyOperation(
            currentContent,
            operation,
          ),
        serverContent,
      );

    setContent(recoveredContent);

    previousContentRef.current =
      recoveredContent;
  };

  /*
   * Restore textarea cursor after React
   * applies remote changes.
   */
  useLayoutEffect(() => {
    const cursor =
      pendingCursorRef.current;

    if (!cursor || !textareaRef.current) {
      return;
    }

    textareaRef.current.selectionStart =
      cursor.start;

    textareaRef.current.selectionEnd =
      cursor.end;

    pendingCursorRef.current = null;
  }, [content]);

  /*
   * WebSocket + offline synchronization.
   */
  useEffect(() => {
    let disposed = false;

    void loadOfflineDocument(roomId)
      .then((offlineDocument) => {
        if (
          disposed ||
          !offlineDocument
        ) {
          return;
        }

        setContent(
          offlineDocument.content,
        );

        previousContentRef.current =
          offlineDocument.content;

        versionRef.current =
          offlineDocument.version;

        setDocumentVersion(
          offlineDocument.version,
        );
      })
      .catch(() => {
        // IndexedDB is optional.
      });

    const connect = () => {
      if (disposed) {
        return;
      }

      if (!navigator.onLine) {
        setConnectionStatus(
          "Waiting for network…",
        );

        setErrorMessage(
          "You are offline. New edits will be sent when you reconnect.",
        );

        return;
      }

      const currentSocket =
        socketRef.current;

      if (
        currentSocket?.readyState ===
          WebSocket.OPEN ||
        currentSocket?.readyState ===
          WebSocket.CONNECTING
      ) {
        return;
      }

      waitingForSyncRef.current = true;

      recoveringRef.current =
        hasSynchronizedRef.current;

      const socket =
        new WebSocket(
          `${WS_BASE}/ws/${roomId}?access_token=${encodeURIComponent(
            session.accessToken,
          )}`,
        );

      socketRef.current = socket;

      socket.onopen = () => {
        if (
          socketRef.current !== socket
        ) {
          return;
        }

        reconnectAttemptRef.current = 0;

        setConnectionStatus(
          "Connected — synchronizing…",
        );

        console.log("CONNECTED");
      };

      socket.onmessage = (event) => {
        let message: Record<
          string,
          any
        >;

        try {
          message = JSON.parse(
            event.data,
          );
        } catch {
          setErrorMessage(
            "Received an invalid update from the collaboration server.",
          );

          socket.close();

          return;
        }

        if (
          !message ||
          typeof message.type !==
            "string"
        ) {
          socket.close();
          return;
        }

        /*
         * Online users
         */
        if (
          message.type ===
          "users_list"
        ) {
          setUsers(
            message.users ?? [],
          );
        }

        /*
         * Remote cursor
         */
        if (
          message.type ===
          "cursor_move"
        ) {
          setRemoteCursors(
            (previous) => ({
              ...previous,
              [message.userId]: {
                username:
                  message.username,

                position:
                  message.cursor
                    .position,

                selectionStart:
                  message.cursor
                    .selectionStart,

                selectionEnd:
                  message.cursor
                    .selectionEnd,

                color: "blue",
              },
            }),
          );
        }

        /*
         * Remove remote cursor.
         */
        if (
          message.type ===
            "cursor_remove" ||
          message.type ===
            "user_left"
        ) {
          setRemoteCursors(
            (previous) => {
              const updated = {
                ...previous,
              };

              delete updated[
                message.userId
              ];

              return updated;
            },
          );
        }

        /*
         * Authoritative server snapshot.
         */
        if (
          message.type ===
          "document_sync"
        ) {
          const syncedContent =
            message.content ?? "";

          if (
            message.version !==
            undefined
          ) {
            versionRef.current =
              message.version;

            setDocumentVersion(
              message.version,
            );
          }

          reapplyPendingOperations(
            syncedContent,
          );

          void saveOfflineDocument(
            roomId,
            syncedContent,
            message.version ??
              versionRef.current,
          );

          waitingForSyncRef.current =
            false;

          hasSynchronizedRef.current =
            true;

          setConnectionStatus(
            "Connected",
          );

          setErrorMessage(null);

          sendNextOperation();
        }

        /*
         * Our operation was accepted.
         */
        if (
          message.type ===
          "edit_ack"
        ) {
          if (
            message.version !==
            undefined
          ) {
            versionRef.current =
              message.version;

            setDocumentVersion(
              message.version,
            );
          }

          const operation =
            pendingOperationsRef.current[0];

          if (
            operation &&
            (!message.operation ||
              message.operation.id ===
                operation.id)
          ) {
            pendingOperationsRef.current.shift();
          }

          operationInFlightRef.current =
            false;

          if (
            recoveringRef.current &&
            message.content !==
              undefined
          ) {
            reapplyPendingOperations(
              message.content,
            );
          }

          if (
            pendingOperationsRef.current
              .length === 0
          ) {
            recoveringRef.current =
              false;
          }

          sendNextOperation();
        }

        /*
         * Remote operation.
         */
        if (
          message.type === "edit"
        ) {
          if (message.operation) {
            const selectionStart =
              textareaRef.current
                ?.selectionStart ?? 0;

            const selectionEnd =
              textareaRef.current
                ?.selectionEnd ?? 0;

            const newSelectionStart =
              transformCursor(
                selectionStart,
                message.operation,
              );

            const newSelectionEnd =
              transformCursor(
                selectionEnd,
                message.operation,
              );

            const updatedContent =
              applyOperation(
                previousContentRef.current,
                message.operation,
              );

            pendingCursorRef.current =
              {
                start:
                  newSelectionStart,
                end:
                  newSelectionEnd,
              };

            setContent(
              updatedContent,
            );

            previousContentRef.current =
              updatedContent;

            void saveOfflineDocument(
              roomId,
              updatedContent,
              versionRef.current,
            );
          }

          if (
            message.version !==
            undefined
          ) {
            versionRef.current =
              message.version;

            setDocumentVersion(
              message.version,
            );
          }
        }

        /*
         * Version conflict recovery.
         */
        if (
          message.type ===
          "version_conflict"
        ) {
          const recoveredContent =
            message.content ?? "";

          versionRef.current =
            message.version;

          setDocumentVersion(
            message.version,
          );

          reapplyPendingOperations(
            recoveredContent,
          );

          setErrorMessage(
            "The document changed while you were editing. Your local edits are being recovered.",
          );
        }
      };

      const scheduleReconnect = () => {
        if (
          disposed ||
          reconnectTimerRef.current !==
            null
        ) {
          return;
        }

        operationInFlightRef.current =
          false;

        waitingForSyncRef.current =
          true;

        if (!navigator.onLine) {
          setConnectionStatus(
            "Waiting for network…",
          );

          return;
        }

        const attempt =
          reconnectAttemptRef.current++;

        const delay = Math.min(
          RECONNECT_BASE_DELAY_MS *
            2 ** attempt,
          RECONNECT_MAX_DELAY_MS,
        );

        setConnectionStatus(
          `Reconnecting in ${Math.ceil(
            delay / 1000,
          )}s…`,
        );

        reconnectTimerRef.current =
          window.setTimeout(() => {
            reconnectTimerRef.current =
              null;

            connect();
          }, delay);
      };

      socket.onerror = () => {
        setErrorMessage(
          "Unable to reach the collaboration server. Retrying automatically.",
        );

        socket.close();
      };

      socket.onclose = () => {
        if (
          socketRef.current === socket
        ) {
          socketRef.current = null;
        }

        scheduleReconnect();
      };
    };

    connect();

    const handleOffline = () => {
      operationInFlightRef.current =
        false;

      waitingForSyncRef.current =
        true;

      setConnectionStatus(
        "Waiting for network…",
      );

      setErrorMessage(
        "You are offline. New edits will be sent when you reconnect.",
      );

      socketRef.current?.close();
    };

    const handleOnline = () => {
      if (
        reconnectTimerRef.current !==
        null
      ) {
        window.clearTimeout(
          reconnectTimerRef.current,
        );

        reconnectTimerRef.current =
          null;
      }

      setConnectionStatus(
        "Reconnecting…",
      );

      connect();
    };

    window.addEventListener(
      "offline",
      handleOffline,
    );

    window.addEventListener(
      "online",
      handleOnline,
    );

    return () => {
      disposed = true;

      window.removeEventListener(
        "offline",
        handleOffline,
      );

      window.removeEventListener(
        "online",
        handleOnline,
      );

      if (
        reconnectTimerRef.current !==
        null
      ) {
        window.clearTimeout(
          reconnectTimerRef.current,
        );
      }

      socketRef.current?.close();
    };
  }, [session, roomId]);

  /*
   * User edits document.
   */
  const handleChange = (
    event: React.ChangeEvent<HTMLTextAreaElement>,
  ) => {
    const newContent =
      event.target.value;

    const oldContent =
      previousContentRef.current;

    const operations =
      generateOperations(
        oldContent,
        newContent,
      );

    setContent(newContent);

    previousContentRef.current =
      newContent;

    void saveOfflineDocument(
      roomId,
      newContent,
      versionRef.current,
    );

    if (operations.length === 0) {
      return;
    }

    pendingOperationsRef.current.push(
      ...operations,
    );

    sendNextOperation();
  };

  /*
   * Send cursor position to other users.
   */
  const handleCursorMove = (
    event: React.SyntheticEvent<HTMLTextAreaElement>,
  ) => {
    const selectionStart =
      event.currentTarget.selectionStart;

    const selectionEnd =
      event.currentTarget.selectionEnd;

    const socket =
      socketRef.current;

    if (
      socket?.readyState !==
      WebSocket.OPEN
    ) {
      return;
    }

    socket.send(
      JSON.stringify({
        type: "cursor_move",
        roomId: roomId,
        cursor: {
          position: selectionEnd,
          selectionStart,
          selectionEnd,
        },
      }),
    );
  };

  /*
   * Workspace search.
   */
  const handleSearch = async () => {
    if (
      searchQuery.trim().length < 2
    ) {
      return;
    }

    setIsSearching(true);

    try {
      const response =
        await fetch(
          `${API_BASE}/api/search?q=${encodeURIComponent(
            searchQuery,
          )}`,
          {
            headers: {
              Authorization: `Bearer ${session.accessToken}`,
            },
          },
        );

      const body =
        await response.json();

      setSearchResults(
        body.results ?? [],
      );
    } catch {
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  /*
   * AI assistant.
   */
  const askAI = async () => {
    if (!aiQuestion.trim()) {
      return;
    }

    setAiAnswer("Thinking…");

    try {
      const response =
        await fetch(
          `${API_BASE}/api/ai/ask`,
          {
            method: "POST",
            headers: {
              "Content-Type":
                "application/json",
              Authorization: `Bearer ${session.accessToken}`,
            },
            body: JSON.stringify({
              question: aiQuestion,
            }),
          },
        );

      const body =
        await response.json();

      setAiAnswer(
        body.answer ??
          body.error ??
          "No answer available.",
      );
    } catch {
      setAiAnswer(
        "AI service is unavailable. Set OPENAI_API_KEY on the server to enable it.",
      );
    }
  };

  return {
    content,
    users,
    remoteCursors,

    connectionStatus,
    errorMessage,
    documentVersion,

    textareaRef,

    searchQuery,
    setSearchQuery,

    searchResults,
    isSearching,

    aiQuestion,
    setAiQuestion,

    aiAnswer,

    handleSearch,
    askAI,

    handleChange,
    handleCursorMove,

    setContent,
  };
}

export default useCollaboration;