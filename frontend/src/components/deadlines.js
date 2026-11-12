// Calendar-only comparisons use Noxorbit's existing Africa/Lagos deadline policy.
// Never parse YYYY-MM-DD with the browser's local timezone for day labels.
export function calendarToday(now = new Date()) {
  return new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Africa/Lagos', year: 'numeric', month: '2-digit', day: '2-digit',
  }).format(now);
}

export function calendarOffset(today, days) {
  const [year, month, date] = today.split('-').map(Number);
  return new Date(Date.UTC(year, month - 1, date + days)).toISOString().slice(0, 10);
}

const PRIORITIES = { urgent: 0, high: 1, medium: 2, low: 3 };

export function deadlineLabel(dueDate, today) {
  if (!dueDate || !today) return null;
  if (dueDate < today) return { label: 'Overdue', kind: 'overdue' };
  if (dueDate === today) return { label: 'Due today', kind: 'today' };
  if (dueDate === calendarOffset(today, 1)) return { label: 'Due tomorrow', kind: 'tomorrow' };
  if (dueDate <= calendarOffset(today, 3)) return { label: 'Due soon', kind: 'soon' };
  return { label: 'Upcoming', kind: 'upcoming' };
}

export function compareDeadlineTasks(a, b) {
  const priority = (PRIORITIES[a.priority] ?? 4) - (PRIORITIES[b.priority] ?? 4);
  if (priority !== 0) return priority;
  return a.title.localeCompare(b.title) || a.id.localeCompare(b.id);
}

export function groupAssignedDeadlines(tasks, today) {
  if (!today) return [];
  const groups = new Map();
  for (const task of tasks) {
    if (task.status === 'done' || !task.due_date) continue;
    if (!groups.has(task.due_date)) groups.set(task.due_date, []);
    groups.get(task.due_date).push(task);
  }
  return [...groups.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, items]) => ({ date, items: items.sort(compareDeadlineTasks), timing: deadlineLabel(date, today) }));
}
