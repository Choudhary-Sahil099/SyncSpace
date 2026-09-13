const DB_NAME = "syncspace-offline";
const STORE_NAME = "documents";

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, 1);
    request.onupgradeneeded = () => request.result.createObjectStore(STORE_NAME);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

export async function saveOfflineDocument(roomId: string, content: string, version: number) {
  if (!("indexedDB" in window)) return;
  const db = await openDB();
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readwrite");
    tx.objectStore(STORE_NAME).put({ content, version, savedAt: Date.now() }, roomId);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
  db.close();
}

export async function loadOfflineDocument(roomId: string): Promise<{content: string; version: number} | null> {
  if (!("indexedDB" in window)) return null;
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const request = db.transaction(STORE_NAME, "readonly").objectStore(STORE_NAME).get(roomId);
    request.onsuccess = () => { db.close(); resolve(request.result ?? null); };
    request.onerror = () => { db.close(); reject(request.error); };
  });
}
