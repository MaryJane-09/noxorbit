-- 008_task_recycle_bin.sql
-- Apply in Supabase SQL Editor before deploying the task recycle-bin API.
-- Deleted tasks are retained and hidden from all normal task queries.
BEGIN;
ALTER TABLE public.tasks
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN deleted_by UUID REFERENCES public.users(id) ON DELETE SET NULL;
CREATE INDEX tasks_deleted_workspace_idx
    ON public.tasks (workspace_id, deleted_at DESC, id DESC)
    WHERE deleted_at IS NOT NULL;
CREATE INDEX tasks_live_project_idx
    ON public.tasks (project_id, status, created_at DESC)
    WHERE deleted_at IS NULL;
COMMIT;
-- No automated purge: deleted records remain until an authorized owner/creator
-- permanently deletes them through a later explicit feature.
