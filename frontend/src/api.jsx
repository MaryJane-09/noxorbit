export const API = "http://localhost:8080";

export async function fetchMe() {
  try {
    const response = await fetch(`${API}/me`, { credentials: "include" });
    return response.ok ? await response.json() : null;
  } catch {
    return null;
  }
}

// Workspace data is fetched from the authenticated Go backend, not demo state.
async function workspaceResponse(response) {
  let body;
  try {
    body = await response.json();
  } catch {
    throw new Error("The server returned an unexpected response.");
  }
  if (!response.ok) {
    if (response.status === 401) {
      throw new Error("Your session has expired. Log out and sign in again.");
    }
    throw new Error(body?.error || "The workspace request failed. Please try again.");
  }
  return body;
}

export async function fetchWorkspaces({ signal } = {}) {
  let response;
  try {
    response = await fetch(`${API}/api/v1/workspaces`, {
      credentials: "include",
      signal,
    });
  } catch (error) {
    if (error.name === "AbortError") throw error;
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  const workspaces = await workspaceResponse(response);
  if (!Array.isArray(workspaces)) {
    throw new Error("The server returned an invalid workspace list.");
  }
  return workspaces;
}

export async function createWorkspace({ name, purpose }) {
  let response;
  try {
    response = await fetch(`${API}/api/v1/workspaces`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, purpose }),
    });
  } catch {
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  return workspaceResponse(response);
}


export async function updateWorkspace(id, { name, purpose }) {
  let response;
  try {
    response = await fetch(`${API}/api/v1/workspaces/${encodeURIComponent(id)}`, {
      method: "PATCH",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, purpose }),
    });
  } catch {
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  return workspaceResponse(response);
}
