-- 009_activity_history.sql
-- Apply after 008_task_recycle_bin before deploying the Activity API.
-- Activity events are written by database triggers in the SAME transaction as
-- the associated change. They cannot be edited or deleted through the Go API.
BEGIN;

CREATE TABLE public.activity_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES public.workspaces(id) ON DELETE CASCADE,
    actor_id uuid REFERENCES public.users(id) ON DELETE SET NULL,
    action text NOT NULL CHECK (action IN (
        'workspace.created','workspace.updated','project.created','project.updated',
        'project.archived','project.restored','task.created','task.updated',
        'task.assigned','task.status_changed','task.deleted','task.restored',
        'task.deleted_forever','member.joined','member.removed'
    )),
    entity_type text NOT NULL CHECK (entity_type IN ('workspace','project','task','member')),
    entity_id uuid NOT NULL,
    entity_name text NOT NULL,
    details jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX activity_events_workspace_recent_idx
    ON public.activity_events (workspace_id, created_at DESC, id DESC);
ALTER TABLE public.activity_events ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.activity_events FROM anon, authenticated;

-- These private columns are written by the Go API from authenticated session
-- IDs. They are never accepted from browser request bodies.
ALTER TABLE public.workspaces ADD COLUMN activity_actor_id uuid;
ALTER TABLE public.projects ADD COLUMN activity_actor_id uuid;
ALTER TABLE public.tasks ADD COLUMN activity_actor_id uuid;

CREATE FUNCTION public.noxorbit_audit_workspaces()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.id,NEW.activity_actor_id,'workspace.created','workspace',NEW.id,NEW.name);
    ELSIF NEW.name IS DISTINCT FROM OLD.name OR NEW.purpose IS DISTINCT FROM OLD.purpose THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name,details)
        VALUES (NEW.id,NEW.activity_actor_id,'workspace.updated','workspace',NEW.id,NEW.name,
                jsonb_build_object('previous_name',OLD.name));
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_workspaces_activity AFTER INSERT OR UPDATE ON public.workspaces
    FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_workspaces();

CREATE FUNCTION public.noxorbit_audit_projects()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.workspace_id,NEW.created_by,'project.created','project',NEW.id,NEW.name);
    ELSIF NEW.status IS DISTINCT FROM OLD.status THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.workspace_id,NEW.activity_actor_id,
            CASE WHEN NEW.status='archived' THEN 'project.archived' ELSE 'project.restored' END,
            'project',NEW.id,NEW.name);
    ELSIF NEW.name IS DISTINCT FROM OLD.name OR NEW.description IS DISTINCT FROM OLD.description THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name,details)
        VALUES (NEW.workspace_id,NEW.activity_actor_id,'project.updated','project',NEW.id,NEW.name,
                jsonb_build_object('previous_name',OLD.name));
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_projects_activity AFTER INSERT OR UPDATE ON public.projects
    FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_projects();

CREATE FUNCTION public.noxorbit_audit_tasks()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.workspace_id,NEW.created_by,'task.created','task',NEW.id,NEW.title);
        IF NEW.assignee_id IS NOT NULL THEN
            INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
            VALUES (NEW.workspace_id,NEW.created_by,'task.assigned','task',NEW.id,NEW.title);
        END IF;
        RETURN NEW;
    END IF;
    -- Deletion/restoration is distinct from ordinary editing and takes priority.
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.workspace_id,NEW.deleted_by,'task.deleted','task',NEW.id,NEW.title);
    ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
        INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
        VALUES (NEW.workspace_id,NEW.activity_actor_id,'task.restored','task',NEW.id,NEW.title);
    ELSIF NEW.deleted_at IS NULL AND OLD.deleted_at IS NULL THEN
        IF NEW.status IS DISTINCT FROM OLD.status THEN
            INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name,details)
            VALUES (NEW.workspace_id,NEW.activity_actor_id,'task.status_changed','task',NEW.id,NEW.title,
                    jsonb_build_object('from',OLD.status,'to',NEW.status));
        END IF;
        -- Referential action on member removal can clear the assignee without
        -- coming from an actor-authenticated task update. Do not misattribute it.
        IF NEW.assignee_id IS DISTINCT FROM OLD.assignee_id AND NEW.updated_at IS DISTINCT FROM OLD.updated_at THEN
            INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
            VALUES (NEW.workspace_id,NEW.activity_actor_id,'task.assigned','task',NEW.id,NEW.title);
        END IF;
        IF NEW.title IS DISTINCT FROM OLD.title
           OR NEW.description IS DISTINCT FROM OLD.description
           OR NEW.priority IS DISTINCT FROM OLD.priority
           OR NEW.due_date IS DISTINCT FROM OLD.due_date THEN
            INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name,details)
            VALUES (NEW.workspace_id,NEW.activity_actor_id,'task.updated','task',NEW.id,NEW.title,
                    jsonb_build_object('previous_title',OLD.title));
        END IF;
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_tasks_activity AFTER INSERT OR UPDATE ON public.tasks
    FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_tasks();

-- Joining a workspace is a first-party, member-initiated event; its own user
-- ID is authoritative for accepted invitations and initial owner membership.
CREATE FUNCTION public.noxorbit_audit_membership_join()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO public.activity_events (workspace_id,actor_id,action,entity_type,entity_id,entity_name)
    SELECT NEW.workspace_id, NEW.user_id, 'member.joined', 'member', NEW.user_id,
           COALESCE(NULLIF(u.name,''),'Team member')
    FROM public.users u WHERE u.id=NEW.user_id;
    RETURN NEW;
END $$;
CREATE TRIGGER noxorbit_membership_join_activity AFTER INSERT ON public.workspace_memberships
    FOR EACH ROW EXECUTE FUNCTION public.noxorbit_audit_membership_join();

COMMIT;
