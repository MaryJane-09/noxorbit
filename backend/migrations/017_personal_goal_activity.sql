-- Account-private goal history, separate from public workspace activity.
-- Apply after 016_personal_goals.sql. Each event is written in the same
-- PostgreSQL transaction as the corresponding goal/milestone change.
BEGIN;

CREATE TABLE public.personal_goal_activity (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    goal_id uuid NOT NULL,
    goal_title text NOT NULL,
    milestone_title text,
    action text NOT NULL CHECK (action IN (
        'goal.created','goal.updated','goal.deleted',
        'milestone.added','milestone.completed',
        'milestone.reopened','milestone.deleted'
    )),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX personal_goal_activity_user_recent
    ON public.personal_goal_activity(user_id,created_at DESC,id DESC);
ALTER TABLE public.personal_goal_activity ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.personal_goal_activity FROM anon,authenticated;

CREATE FUNCTION public.noxorbit_audit_personal_goals()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO public.personal_goal_activity(user_id,goal_id,goal_title,action)
        VALUES (NEW.user_id,NEW.id,NEW.title,'goal.created');
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO public.personal_goal_activity(user_id,goal_id,goal_title,action)
        VALUES (OLD.user_id,OLD.id,OLD.title,'goal.deleted');
        RETURN OLD;
    ELSIF NEW.title IS DISTINCT FROM OLD.title
       OR NEW.description IS DISTINCT FROM OLD.description
       OR NEW.target_date IS DISTINCT FROM OLD.target_date THEN
        INSERT INTO public.personal_goal_activity(user_id,goal_id,goal_title,action)
        VALUES (NEW.user_id,NEW.id,NEW.title,'goal.updated');
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_personal_goals_activity
AFTER INSERT OR UPDATE OR DELETE ON public.personal_goals
FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_personal_goals();

CREATE FUNCTION public.noxorbit_audit_goal_milestones()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    owner_id uuid;
    current_title text;
    event_action text;
BEGIN
    -- During a goal's cascade deletion the parent may already be gone;
    -- the parent deletion event is sufficient and privacy is preserved.
    SELECT g.user_id,g.title INTO owner_id,current_title
    FROM public.personal_goals g
    WHERE g.id=CASE WHEN TG_OP='DELETE' THEN OLD.goal_id ELSE NEW.goal_id END;
    IF owner_id IS NULL THEN
        IF TG_OP='DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
    END IF;

    IF TG_OP='INSERT' THEN
        event_action:='milestone.added';
    ELSIF TG_OP='DELETE' THEN
        event_action:='milestone.deleted';
    ELSIF NEW.completed IS DISTINCT FROM OLD.completed THEN
        event_action:=CASE WHEN NEW.completed THEN 'milestone.completed' ELSE 'milestone.reopened' END;
    ELSE
        RETURN NEW;
    END IF;

    INSERT INTO public.personal_goal_activity(user_id,goal_id,goal_title,milestone_title,action)
    VALUES(owner_id,
           CASE WHEN TG_OP='DELETE' THEN OLD.goal_id ELSE NEW.goal_id END,
           current_title,
           CASE WHEN TG_OP='DELETE' THEN OLD.title ELSE NEW.title END,
           event_action);
    IF TG_OP='DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
END $$;
CREATE TRIGGER noxorbit_goal_milestones_activity
AFTER INSERT OR UPDATE OR DELETE ON public.goal_milestones
FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_goal_milestones();

-- Earlier goals predate this audit trail. Preserve their real creation time
-- rather than pretending they were created when migration 017 was installed.
INSERT INTO public.personal_goal_activity(user_id,goal_id,goal_title,action,created_at)
SELECT user_id,id,title,'goal.created',created_at FROM public.personal_goals;

COMMIT;
