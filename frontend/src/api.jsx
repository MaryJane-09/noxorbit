export const API = "http://localhost:8080";

export async function fetchMe() {
  try {
    const response = await fetch(`${API}/me`, { credentials: "include" });
    return response.ok ? await response.json() : null;
  } catch {
    return null;
  }
}