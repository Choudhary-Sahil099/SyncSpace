import { useEffect, useRef, useState } from "react";

function App() {
  const [content, setContent] = useState("");
  const [users, setUsers] = useState<string[]>([]);
  const socketRef = useRef<WebSocket | null>(null);
  const versionRef = useRef(0);
  // debounce --> delay messsage so that the message recieved is very corrent to avoid the previous issue of not similarity
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const username = "user-" + Math.floor(Math.random() * 1000);

    const socket = new WebSocket(
      `ws://localhost:8080/ws/room1?username=${username}`,
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

      if (message.type === "edit" || message.type === "document_sync") {
        setContent(message.content);

        if (message.version !== undefined) {
          versionRef.current = message.version;
        }
      }
      if (message.type === "version_conflict") {
        console.log("CONFLICT RECOVERY", message.version);

        setContent(message.content);

        versionRef.current = message.version;
      }
      console.log("DOCUMENT VERSION:", message.version);
    };

    return () => {
      socket.close();
    };
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value;

    // instant local update
    setContent(newContent);
    // clear previous debounce
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    // debounce websocket send
    timeoutRef.current = setTimeout(() => {
      socketRef.current?.send(
        JSON.stringify({
          type: "edit",
          roomId: "room1",
          content: newContent,
          version: versionRef.current, // change to 1 when to test the conflict detection else versionRef.current
        }),
      );
    }, 50);
    console.log("SENDING VERSION:", versionRef.current);
  };

  return (
    <div style={{ padding: "40px" }}>
      <h1>SyncSpace</h1>

      <textarea
        value={content}
        onChange={handleChange}
        style={{
          width: "100%",
          height: "500px",
          fontSize: "18px",
        }}
      />
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
