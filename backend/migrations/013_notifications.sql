-- Settings Phase 2 (part 1): persisted in-app alerts and user notification choices.
-- Apply once after migration 012. No email delivery or background jobs.
BEGIN;
CREATE TABLE public.notification_preferences (
 user_id uuid PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
 task_assigned boolean NOT NULL DEFAULT true,
 due_tomorrow boolean NOT NULL DEFAULT true,
 due_today boolean NOT NULL DEFAULT true,
 overdue boolean NOT NULL DEFAULT true,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE public.notifications (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
 workspace_id uuid REFERENCES public.workspaces(id) ON DELETE CASCADE,
 project_id uuid,
 task_id uuid,
 kind text NOT NULL CHECK (kind IN ('task_assigned','due_tomorrow','due_today','overdue')),
 title text NOT NULL CHECK (length(title) BETWEEN 1 AND 220),
 body text NOT NULL DEFAULT '',
 dedupe_key text NOT NULL,
 read_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 CONSTRAINT notifications_unique_per_user UNIQUE(user_id,dedupe_key)
);
CREATE INDEX notifications_user_recent_idx ON public.notifications(user_id,created_at DESC,id DESC);
CREATE INDEX notifications_user_unread_idx ON public.notifications(user_id,created_at DESC) WHERE read_at IS NULL;
ALTER TABLE public.notification_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.notifications ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.notification_preferences,public.notifications FROM anon,authenticated;

-- Trigger runs in the same transaction as assignment. It does not announce
-- self-assignments, removed assignments, or updates unrelated to assignee.
CREATE FUNCTION public.noxorbit_notify_task_assignment() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE enabled boolean;
BEGIN
 IF NEW.assignee_id IS NULL OR NEW.deleted_at IS NOT NULL THEN RETURN NEW; END IF;
 IF TG_OP='UPDATE' AND NEW.assignee_id IS NOT DISTINCT FROM OLD.assignee_id THEN RETURN NEW; END IF;
 IF NEW.assignee_id IS NOT DISTINCT FROM COALESCE(NEW.activity_actor_id,NEW.created_by) THEN RETURN NEW; END IF;
 SELECT COALESCE(p.task_assigned,true) INTO enabled
 FROM public.notification_preferences p WHERE p.user_id=NEW.assignee_id;
 IF COALESCE(enabled,true) THEN
   INSERT INTO public.notifications (user_id,workspace_id,project_id,task_id,kind,title,body,dedupe_key)
   VALUES (NEW.assignee_id,NEW.workspace_id,NEW.project_id,NEW.id,'task_assigned',
           'Task assigned to you',left(NEW.title,200),
           'assigned:'||NEW.id::text||':'||NEW.assignee_id::text||':'||NEW.updated_at::text)
   ON CONFLICT(user_id,dedupe_key) DO NOTHING;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_task_assignment_notification
AFTER INSERT OR UPDATE OF assignee_id ON public.tasks
FOR EACH ROW EXECUTE FUNCTION public.noxorbit_notify_task_assignment();
COMMIT;
