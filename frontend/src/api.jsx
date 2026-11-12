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


// Projects are always scoped by a workspace UUID and validated by the Go API.
function projectsURL(workspaceID, projectID = "") {
  const base = `${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/projects`;
  return projectID ? `${base}/${encodeURIComponent(projectID)}` : base;
}

async function projectRequest(url, { method = "GET", body, signal } = {}) {
  let response;
  try {
    response = await fetch(url, {
      method,
      credentials: "include",
      ...(body === undefined ? {} : {
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }),
      signal,
    });
  } catch (error) {
    if (error.name === "AbortError") throw error;
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  return workspaceResponse(response);
}

export async function fetchProjects(workspaceID, { signal } = {}) {
  const projects = await projectRequest(projectsURL(workspaceID), { signal });
  if (!Array.isArray(projects)) {
    throw new Error("The server returned an invalid project list.");
  }
  return projects;
}

export function createProject(workspaceID, { name, description }) {
  return projectRequest(projectsURL(workspaceID), {
    method: "POST",
    body: { name, description },
  });
}

export function updateProject(workspaceID, projectID, { name, description }) {
  return projectRequest(projectsURL(workspaceID, projectID), {
    method: "PATCH",
    body: { name, description },
  });
}

export function changeProjectStatus(workspaceID, projectID, status) {
  return projectRequest(`${projectsURL(workspaceID, projectID)}/status`, {
    method: "PATCH",
    body: { status },
  });
}


// Tasks are always nested under their parent project and workspace.
function tasksURL(workspaceID, projectID, taskID = "") {
  const base = `${projectsURL(workspaceID, projectID)}/tasks`;
  return taskID ? `${base}/${encodeURIComponent(taskID)}` : base;
}

export async function fetchTasks(workspaceID, projectID, { signal } = {}) {
  const tasks = await projectRequest(tasksURL(workspaceID, projectID), { signal });
  if (!Array.isArray(tasks)) throw new Error("The server returned an invalid task list.");
  return tasks;
}

export function createTask(workspaceID, projectID, details) {
  return projectRequest(tasksURL(workspaceID, projectID), {
    method: "POST", body: details,
  });
}

export function updateTask(workspaceID, projectID, taskID, details) {
  return projectRequest(tasksURL(workspaceID, projectID, taskID), {
    method: "PATCH", body: details,
  });
}

export function changeTaskStatus(workspaceID, projectID, taskID, status) {
  return projectRequest(`${tasksURL(workspaceID, projectID, taskID)}/status`, {
    method: "PATCH", body: { status },
  });
}

export async function deleteTask(workspaceID, projectID, taskID) {
  // DELETE returns 204 No Content, so it must not be parsed as JSON.
  let response;
  try {
    response = await fetch(tasksURL(workspaceID, projectID, taskID), {
      method: "DELETE", credentials: "include",
    });
  } catch {
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  if (response.status === 204) return;
  return workspaceResponse(response);
}


// Only members of the selected workspace can read this directory.
export async function fetchWorkspaceMembers(workspaceID, { signal } = {}) {
  const url = `${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/members`;
  const members = await projectRequest(url, { signal });
  if (!Array.isArray(members)) throw new Error("The server returned an invalid member list.");
  return members;
}


export function createWorkspaceInvite(workspaceID) {
  return projectRequest(`${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/invitations`, {
    method: "POST", body: {},
  });
}
export function acceptWorkspaceInvite(token) {
  return projectRequest(`${API}/api/v1/invitations/accept`, {
    method: "POST", body: { token },
  });
}


// Owner-only member and invitation administration.
export function fetchWorkspaceInvites(workspaceID, { signal } = {}) {
  return projectRequest(`${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/invitations/manage`, { signal });
}

async function deleteWorkspaceResource(url) {
  let response;
  try {
    response = await fetch(url, { method: "DELETE", credentials: "include" });
  } catch {
    throw new Error("Cannot reach the NoxOrbit server. Check that your Go backend is running.");
  }
  if (response.status === 204) return;
  return workspaceResponse(response);
}

export function revokeWorkspaceInvite(workspaceID, invitationID) {
  return deleteWorkspaceResource(`${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/invitations/${encodeURIComponent(invitationID)}`);
}

export function removeWorkspaceMember(workspaceID, userID) {
  return deleteWorkspaceResource(`${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/members/${encodeURIComponent(userID)}`);
}


export function sendWorkspaceEmailInvite(workspaceID, email) {
  return projectRequest(`${API}/api/v1/workspaces/${encodeURIComponent(workspaceID)}/invitations/email`, {
    method: "POST", body: { email },
  });
}


// Across all workspaces the current account belongs to (never another user).
export async function fetchMyTasks({ signal } = {}) {
  const items = await projectRequest(`${API}/api/v1/my-tasks`, { signal });
  if (!Array.isArray(items)) throw new Error("The server returned an invalid assigned task list.");
  return items;
}


// Recoverable task deletion: list only records the current member can restore.
export async function fetchRecycledTasks({ signal } = {}) {
  const items = await projectRequest(`${API}/api/v1/recycle-bin/tasks`, { signal });
  if (!Array.isArray(items)) throw new Error("The server returned an invalid recycle bin list.");
  return items;
}

export async function restoreRecycledTask(taskID) {
  const response = await fetch(`${API}/api/v1/recycle-bin/tasks/${encodeURIComponent(taskID)}/restore`, {
    method: "POST", credentials: "include",
  }).catch(() => { throw new Error("Cannot reach the NoxOrbit server."); });
  if (response.status === 204) return;
  return workspaceResponse(response);
}
