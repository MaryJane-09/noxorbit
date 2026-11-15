// Date-only task deadlines never represent an instant. Determine the
// current calendar day in the account's saved IANA time zone.
export const DEFAULT_PREFERENCES = { time_zone: 'Africa/Lagos', date_format: 'YYYY-MM-DD' };

export function calendarToday(now = new Date(), zone = DEFAULT_PREFERENCES.time_zone) {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: zone, year: 'numeric', month: '2-digit', day: '2-digit',
  }).formatToParts(now);
  const fields = Object.fromEntries(parts.map(({ type, value }) => [type, value]));
  return `${fields.year}-${fields.month}-${fields.day}`;
}

// Format a YYYY-MM-DD DATE directly, without UTC/local midnight conversion.
export function formatCalendarDate(value, format = DEFAULT_PREFERENCES.date_format) {
  if (!/^[0-9]{4}-[0-9]{2}-[0-9]{2}$/.test(value || '')) return value || '';
  const [year, month, day] = value.split('-');
  if (format === 'DD/MM/YYYY') return `${day}/${month}/${year}`;
  if (format === 'MM/DD/YYYY') return `${month}/${day}/${year}`;
  return value;
}

// UTC timestamps represent real instants: show date AND time in the chosen zone.
export function formatInstant(value, preferences = DEFAULT_PREFERENCES) {
  if (!value) return '';
  const instant = new Date(value);
  if (Number.isNaN(instant.getTime())) return '';
  const zone = preferences?.time_zone || DEFAULT_PREFERENCES.time_zone;
  const date = formatCalendarDate(calendarToday(instant, zone), preferences?.date_format);
  const time = new Intl.DateTimeFormat('en-US', {
    timeZone: zone, hour: 'numeric', minute: '2-digit',
  }).format(instant);
  return `${date} · ${time}`;
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
