-- Project participation and roles. Apply after migration 011.
-- Workspace members still retain read access to projects.
BEGIN;
CREATE TABLE public.project_memberships (
    project_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('lead','member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id,user_id),
    CONSTRAINT project_memberships_project_fk FOREIGN KEY (project_id,workspace_id)
        REFERENCES public.projects(id,workspace_id) ON DELETE CASCADE,
    CONSTRAINT project_memberships_workspace_member_fk FOREIGN KEY (workspace_id,user_id)
        REFERENCES public.workspace_memberships(workspace_id,user_id) ON DELETE CASCADE
);
CREATE INDEX project_memberships_workspace_user_idx ON public.project_memberships(workspace_id,user_id);
ALTER TABLE public.project_memberships ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.project_memberships FROM anon,authenticated;

-- Existing projects keep their creators, owners and assigned contributors.
INSERT INTO public.project_memberships (project_id,workspace_id,user_id,role)
SELECT p.id,p.workspace_id,m.user_id,
    CASE WHEN m.role='owner' OR p.created_by=m.user_id THEN 'lead' ELSE 'member' END
FROM public.projects p
JOIN public.workspace_memberships m ON m.workspace_id=p.workspace_id
WHERE m.role='owner' OR m.user_id=p.created_by OR EXISTS (
    SELECT 1 FROM public.tasks t
    WHERE t.project_id=p.id AND t.assignee_id=m.user_id AND t.deleted_at IS NULL
)
ON CONFLICT (project_id,user_id) DO UPDATE SET
    role=CASE WHEN EXCLUDED.role='lead' THEN 'lead' ELSE project_memberships.role END;

-- New projects automatically enroll their creator as a lead.
-- Their membership is already required by the project creation handler.
CREATE FUNCTION public.noxorbit_project_creator_membership() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.created_by IS NOT NULL THEN
        INSERT INTO public.project_memberships(project_id,workspace_id,user_id,role)
        SELECT NEW.id,NEW.workspace_id,NEW.created_by,'lead'
        WHERE EXISTS (
          SELECT 1 FROM public.workspace_memberships
          WHERE workspace_id=NEW.workspace_id AND user_id=NEW.created_by
        )
        ON CONFLICT(project_id,user_id) DO NOTHING;
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER noxorbit_project_creator_membership
AFTER INSERT ON public.projects
FOR EACH ROW EXECUTE FUNCTION public.noxorbit_project_creator_membership();
COMMIT;
