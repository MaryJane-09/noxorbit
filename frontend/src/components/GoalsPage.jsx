import { useEffect, useRef, useState } from 'react'
import { ArrowRight, CalendarDays, Check, Circle, Flag, Plus, RefreshCw, Target, Trash2, Pencil, X, Clock3 } from 'lucide-react'
import {
  fetchGoals, fetchGoalActivity, createGoal, updateGoal, deleteGoal,
  createGoalMilestone, setGoalMilestone, deleteGoalMilestone,
} from '../api'
import { formatCalendarDate, formatInstant } from './deadlines'
import './GoalsPage.css'

function goalProgress(goal) {
  const total = goal.milestones?.length || 0
  const completed = goal.milestones?.filter((item) => item.completed).length || 0
  return { total, completed, percent: total ? Math.round(completed / total * 100) : 0 }
}

export default function GoalsPage({ preferences, compact = false, onOpenGoals }) {
  const [goals, setGoals] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [revision, setRevision] = useState(0)
  const [busy, setBusy] = useState(false)
  const [editorID, setEditorID] = useState(null)
  const [formOpen, setFormOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [targetDate, setTargetDate] = useState('')
  const [milestoneDrafts, setMilestoneDrafts] = useState({})
  const titleRef = useRef(null)
  const confirmRef = useRef(null)
  const destructiveRef = useRef(null)
  const lastTriggerRef = useRef(null)
  const [pendingDelete, setPendingDelete] = useState(null)
  const [history, setHistory] = useState({ items: [], loading: true, error: '' })

  useEffect(() => {
    const controller = new AbortController()
    fetchGoals({ signal: controller.signal })
      .then((items) => { if (!controller.signal.aborted) { setGoals(items); setError('') } })
      .catch((err) => { if (!controller.signal.aborted) setError(err.message || 'Could not load goals.') })
      .finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [revision])

  useEffect(() => {
    if (compact) return undefined
    const controller = new AbortController()
    fetchGoalActivity({ signal: controller.signal })
      .then((items) => {
        if (!controller.signal.aborted) setHistory({ items, loading: false, error: '' })
      })
      .catch((err) => {
        if (!controller.signal.aborted) {
          setHistory((current) => ({ ...current, loading: false, error: err.message || 'Could not load goal history.' }))
        }
      })
    return () => controller.abort()
  }, [compact, revision])

  // Refresh when users return from another tab or goals are changed elsewhere.
  useEffect(() => {
    const update = () => setRevision((value) => value + 1)
    const onFocus = () => { if (!document.hidden) update() }
    window.addEventListener('noxorbit:goals-changed', update)
    window.addEventListener('focus', onFocus)
    return () => {
      window.removeEventListener('noxorbit:goals-changed', update)
      window.removeEventListener('focus', onFocus)
    }
  }, [])
  useEffect(() => {
    if (formOpen && !compact) titleRef.current?.focus()
  }, [formOpen, editorID, compact])

  function reload() {
    setLoading(true)
    setRevision((value) => value + 1)
  }
  function clearForm() {
    setTitle('')
    setDescription('')
    setTargetDate('')
    setEditorID(null)
    setFormOpen(false)
  }
  function edit(goal) {
    setEditorID(goal.id)
    setTitle(goal.title)
    setDescription(goal.description || '')
    setTargetDate(goal.target_date || '')
    setFormOpen(true)
    setError('')
  }
  async function action(callback, successMessage = '') {
    if (busy) return
    setBusy(true)
    setError('')
    setNotice('')
    try {
      await callback()
      setNotice(successMessage)
      // Every mounted Goals view (including Home) refreshes from one event.
      // Keep the existing cards visible while the new data is fetched.
      window.dispatchEvent(new Event('noxorbit:goals-changed'))
    } catch (err) {
      setError(err.message || 'Could not update your goals.')
    } finally { setBusy(false) }
  }
  async function save(event) {
    event.preventDefault()
    const cleanTitle = title.trim()
    if (!cleanTitle) { setError('Enter a goal title.'); return }
    await action(async () => {
      const data = { title: cleanTitle, description: description.trim(), target_date: targetDate || null }
      if (editorID) await updateGoal(editorID, data)
      else await createGoal(data)
      clearForm()
    }, editorID ? 'Goal updated.' : 'Goal created. Add a milestone to track progress.')
  }
  async function addMilestone(event, goalID) {
    event.preventDefault()
    const name = (milestoneDrafts[goalID] || '').trim()
    if (!name) return
    await action(async () => {
      await createGoalMilestone(goalID, name)
      setMilestoneDrafts((current) => ({ ...current, [goalID]: '' }))
    }, 'Milestone added.')
  }
  useEffect(() => {
    if (!pendingDelete) return
    confirmRef.current?.focus()
    function onKeyDown(event) {
      if (event.key === 'Escape') {
        event.preventDefault()
        setPendingDelete(null)
      } else if (event.key === 'Tab') {
        // Keep keyboard navigation inside the small confirmation dialog.
        if (event.shiftKey && document.activeElement === confirmRef.current) {
          event.preventDefault()
          destructiveRef.current?.focus()
        } else if (!event.shiftKey && document.activeElement === destructiveRef.current) {
          event.preventDefault()
          confirmRef.current?.focus()
        }
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [pendingDelete])

  function closeConfirmation() {
    setPendingDelete(null)
    window.setTimeout(() => lastTriggerRef.current?.focus(), 0)
  }
  function confirmDeletion() {
    if (!pendingDelete || busy) return
    const target = pendingDelete
    setPendingDelete(null)
    if (target.type === 'goal') {
      void action(() => deleteGoal(target.goalID), 'Goal deleted.')
    } else {
      void action(() => deleteGoalMilestone(target.goalID, target.milestoneID), 'Milestone deleted.')
    }
  }
  // The Home preview highlights goals with real progress first, rather than
  // filling its limited space with zero-milestone goals created most recently.
  const visibleGoals = compact
    ? [...goals].sort((a, b) => {
        const aScore = goalProgress(a)
        const bScore = goalProgress(b)
        const rank = (p) => p.total === 0 ? 2 : p.completed === p.total ? 1 : 0
        return rank(aScore) - rank(bScore) || bScore.percent - aScore.percent
      }).slice(0, 2)
    : goals
  return <section className={`goals-page${compact ? ' goals-page--compact' : ''}`} aria-label={compact ? 'Your personal goals' : undefined}>
    {!compact && <header className="goals-heading">
      <div>
        <p className="dashboard-eyebrow"><Target size={16} aria-hidden="true" /> PERSONAL · ALL WORKSPACES</p>
        <h1>My goals<span>.</span></h1>
        <p>Choose the outcomes that matter to you. Track real progress, one milestone at a time.</p>
      </div>
      <button type="button" className="goals-refresh" onClick={reload} disabled={loading || busy}>
        <RefreshCw size={16} aria-hidden="true" /> Refresh
      </button>
    </header>}

    {compact ? <div className="goals-compact-heading">
      <div><p className="dashboard-eyebrow">WHERE YOU'RE HEADED</p><h2>My goals</h2></div>
      <button type="button" className="goals-inline-action" onClick={onOpenGoals}>View goals <ArrowRight size={16} /></button>
    </div> : <div className="goals-page-toolbar">
      <span>{loading ? 'Loading goals…' : `${goals.length} personal ${goals.length === 1 ? 'goal' : 'goals'}`}</span>
      <button type="button" className="goals-primary" onClick={() => { clearForm(); setFormOpen(true) }} disabled={busy || formOpen}>
        <Plus size={17} /> New goal
      </button>
    </div>}

    {error && <div className="goals-alert" role="alert">{error} <button type="button" onClick={reload}>Retry</button></div>}
    {!compact && notice && <p className="goals-notice" role="status">{notice}</p>}

    {!compact && formOpen && <form className="goals-form" onSubmit={save} aria-label={editorID ? 'Edit your goal' : 'Create a personal goal'}>
      <div className="goals-form-heading"><h2>{editorID ? 'Edit goal' : 'Create a goal'}</h2>
        <button type="button" className="goals-icon-button" onClick={clearForm} aria-label="Close goal form"><X size={17} /></button>
      </div>
      <label htmlFor="goal-name">Goal title</label>
      <input id="goal-name" ref={titleRef} value={title} maxLength={120} required
        onChange={(e) => setTitle(e.target.value)} placeholder="e.g. Complete my AI learning foundations" />
      <label htmlFor="goal-description">Description (optional)</label>
      <textarea id="goal-description" value={description} maxLength={1000}
        onChange={(e) => setDescription(e.target.value)} rows={3} placeholder="Why does this matter?" />
      <label htmlFor="goal-target">Target date (optional)</label>
      <input id="goal-target" type="date" value={targetDate} onChange={(e) => setTargetDate(e.target.value)} />
      <div className="goals-form-actions">
        <button type="button" className="goals-secondary" onClick={clearForm}>Cancel</button>
        <button type="submit" className="goals-primary" disabled={busy}>{busy ? 'Saving…' : editorID ? 'Save changes' : 'Create goal'}</button>
      </div>
    </form>}

    {loading ? <p className="goals-loading" role="status">Loading your goals…</p>
      : error && goals.length === 0 ? null
      : visibleGoals.length === 0 && (compact || !formOpen) ? <div className="goals-empty">
        <span className="goals-symbol"><Flag size={22} aria-hidden="true" /></span>
        <h3>{error ? 'Goals are unavailable' : 'A goal starts with a direction.'}</h3>
        <p>{error ? 'Check the connection and try again.' : 'Create a personal goal, then add milestones. Your progress will update as you finish them.'}</p>
        {compact ? <button type="button" className="goals-primary" onClick={onOpenGoals}>Explore goals <ArrowRight size={16} /></button>
          : !error && <button type="button" className="goals-primary" onClick={() => setFormOpen(true)}><Plus size={16} /> Create your first goal</button>}
      </div> : <div className="goals-list">
        {visibleGoals.map((goal) => {
          const progress = goalProgress(goal)
          const status = progress.total === 0 ? 'Planning' : progress.completed === progress.total ? 'Achieved' : 'In progress'
          const statusClass = status === 'Achieved' ? 'achieved' : status === 'Planning' ? 'planning' : 'active'
          return <article className="goals-card" key={goal.id}>
            <div className="goals-card-top">
              <div className="goals-card-identity"><span className="goals-symbol"><Target size={20} aria-hidden="true" /></span><span className={`goals-status goals-status--${statusClass}`}>{status}</span></div>
              {!compact && <div className="goals-card-actions">
                <button type="button" className="goals-icon-button" onClick={() => edit(goal)} disabled={busy} aria-label={`Edit ${goal.title}`}><Pencil size={16} /></button>
                <button type="button" className="goals-icon-button goals-danger" disabled={busy}
                  onClick={(event) => { lastTriggerRef.current = event.currentTarget; setPendingDelete({ type: 'goal', goalID: goal.id, label: goal.title }) }}
                  aria-label={`Delete ${goal.title}`}><Trash2 size={16} /></button>
              </div>}
            </div>
            <h3>{goal.title}</h3>
            {goal.description && !compact && <p className="goals-description">{goal.description}</p>}
            {goal.target_date && <small className="goals-target"><CalendarDays size={14} /> Target {formatCalendarDate(goal.target_date, preferences?.date_format)}</small>}
            <div className="goals-progress-head"><span>{progress.completed} of {progress.total} milestones</span><strong>{progress.percent}%</strong></div>
            <div className="goals-progress-track" role="progressbar" aria-label={`Progress for ${goal.title}`}
              aria-valuemin={0} aria-valuemax={100} aria-valuenow={progress.percent}>
              <span style={{ width: `${progress.percent}%` }} />
            </div>
            {compact ? <button type="button" className="goals-inline-action" onClick={onOpenGoals}>Manage milestones <ArrowRight size={15} /></button>
              : <>
                <div className="goals-milestones">
                  {goal.milestones.length === 0 && <p className="goals-muted">Add your first milestone to begin measuring progress.</p>}
                  {goal.milestones.map((item) => <div className="goals-milestone" key={item.id}>
                    <button type="button" className="goals-milestone-check" disabled={busy}
                      onClick={() => action(() => setGoalMilestone(goal.id,item.id,!item.completed))}
                      aria-label={`${item.completed ? 'Mark incomplete' : 'Mark complete'}: ${item.title}`}>
                      {item.completed ? <Check size={17} /> : <Circle size={17} />}
                    </button>
                    <span className={item.completed ? 'goals-complete' : ''}>{item.title}</span>
                    <button type="button" className="goals-icon-button goals-danger" disabled={busy}
                      onClick={(event) => { lastTriggerRef.current = event.currentTarget; setPendingDelete({ type: 'milestone', goalID: goal.id, milestoneID: item.id, label: item.title }) }}
                      aria-label={`Delete milestone ${item.title}`}><X size={15} /></button>
                  </div>)}
                </div>
                <form className="goals-add-milestone" onSubmit={(e) => addMilestone(e,goal.id)}>
                  <input id={`goal-milestone-${goal.id}`} aria-label={`New milestone for ${goal.title}`} className="goals-milestone-input" placeholder="Add a milestone" maxLength={160}
                    value={milestoneDrafts[goal.id] || ''}
                    onChange={(e) => setMilestoneDrafts((current) => ({ ...current, [goal.id]: e.target.value }))} />
                  <button type="submit" disabled={busy || !(milestoneDrafts[goal.id] || '').trim()}><Plus size={16} /> Add</button>
                </form>
              </>}
          </article>
        })}
      </div>}

    {!compact && <section className="goals-history" aria-labelledby="goals-history-heading">
      <div className="goals-history-heading">
        <span className="goals-symbol"><Clock3 size={19} aria-hidden="true" /></span>
        <div><h2 id="goals-history-heading">Recent goal activity</h2>
          <p>Private to your account. Workspace members cannot see this history.</p>
        </div>
      </div>
      {history.loading ? <p role="status" className="goals-muted">Loading your goal history…</p>
        : history.error ? <p role="alert" className="goals-history-error">{history.error}. Make sure migration 017 has been applied.</p>
        : history.items.length === 0 ? <p className="goals-muted">Goal and milestone changes will appear here.</p>
        : <ol className="goals-history-list">{history.items.slice(0, 6).map((item) => {
          const verbs = {
            'goal.created': 'Created goal', 'goal.updated': 'Updated goal',
            'goal.deleted': 'Deleted goal', 'milestone.added': 'Added milestone',
            'milestone.completed': 'Completed milestone', 'milestone.reopened': 'Reopened milestone',
            'milestone.deleted': 'Deleted milestone',
          }
          return <li key={item.id}>
            <span className="goals-history-dot" aria-hidden="true" />
            <div className="goals-history-detail">
              <strong>{verbs[item.action] || 'Updated goal'}</strong>
              <span>{item.milestone_title ? `${item.milestone_title} · ${item.goal_title}` : item.goal_title}</span>
            </div>
            <time dateTime={item.created_at}>{formatInstant(item.created_at, preferences)}</time>
          </li>
        })}</ol>}
    </section>}

    {pendingDelete && <div className="goals-confirm-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) closeConfirmation() }}>
      <div className="goals-confirm-dialog" role="alertdialog" aria-modal="true" aria-labelledby="goals-confirm-title" aria-describedby="goals-confirm-description">
        <span className="goals-confirm-icon"><Trash2 size={21} aria-hidden="true" /></span>
        <h2 id="goals-confirm-title">{pendingDelete.type === 'goal' ? 'Delete this goal?' : 'Delete this milestone?'}</h2>
        <p id="goals-confirm-description">
          {pendingDelete.type === 'goal'
            ? <>This will permanently delete <strong>{pendingDelete.label}</strong> and its milestones.</>
            : <>Remove <strong>{pendingDelete.label}</strong> from this goal? Its progress will be recalculated.</>}
        </p>
        <div className="goals-confirm-actions">
          <button type="button" ref={confirmRef} className="goals-secondary" disabled={busy} onClick={closeConfirmation}>Cancel</button>
          <button type="button" className="goals-confirm-delete" ref={destructiveRef} disabled={busy} onClick={confirmDeletion}>
            <Trash2 size={16} aria-hidden="true" /> {pendingDelete.type === 'goal' ? 'Delete goal' : 'Delete milestone'}
          </button>
        </div>
      </div>
    </div>}
  </section>
}
