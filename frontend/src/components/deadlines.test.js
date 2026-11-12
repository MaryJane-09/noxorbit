import test from 'node:test'
import assert from 'node:assert/strict'
import { calendarOffset, deadlineLabel, groupAssignedDeadlines } from './deadlines.js'

const TODAY = '2026-11-12'
const task = (id, due_date, priority = 'medium', status = 'todo') => ({
  id, title: id, due_date, priority, status,
})

test('groups all four tomorrow tasks into one reminder and ranks urgent first', () => {
  const tasks = [
    task('low', '2026-11-13', 'low'),
    task('high', '2026-11-13', 'high'),
    task('urgent', '2026-11-13', 'urgent'),
    task('medium', '2026-11-13', 'medium'),
  ]
  const groups = groupAssignedDeadlines(tasks, TODAY)
  assert.equal(groups.length, 1)
  assert.equal(groups[0].timing.label, 'Due tomorrow')
  assert.equal(groups[0].items.length, 4)
  assert.deepEqual(groups[0].items.map((item) => item.id),
    ['urgent', 'high', 'medium', 'low'])
})

test('excludes completed and undated tasks without mutating their priority', () => {
  const tasks = [
    task('done', '2026-11-13', 'urgent', 'done'),
    task('no-due', null),
    task('open', '2026-11-13', 'low'),
  ]
  const groups = groupAssignedDeadlines(tasks, TODAY)
  assert.deepEqual(groups.map((group) => group.items.map((item) => item.id)), [['open']])
  assert.equal(tasks[0].priority, 'urgent')
})

test('marks overdue, today, tomorrow, soon and later correctly', () => {
  assert.equal(deadlineLabel('2026-11-11', TODAY).kind, 'overdue')
  assert.equal(deadlineLabel('2026-11-12', TODAY).kind, 'today')
  assert.equal(deadlineLabel('2026-11-13', TODAY).kind, 'tomorrow')
  assert.equal(deadlineLabel('2026-11-15', TODAY).kind, 'soon')
  assert.equal(deadlineLabel('2026-11-20', TODAY).kind, 'upcoming')
  assert.equal(calendarOffset('2026-12-31', 1), '2027-01-01')
})

test('orders date groups chronologically and leaves source array untouched', () => {
  const tasks = [task('later', '2026-11-20'), task('first', '2026-11-11')]
  const groups = groupAssignedDeadlines(tasks, TODAY)
  assert.deepEqual(groups.map((group) => group.date), ['2026-11-11', '2026-11-20'])
  assert.equal(tasks[0].id, 'later')
})
