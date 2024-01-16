-- name: ChannelDecoderGet :one
select d.script from channel c left join decoder d on c.decoder_name = d.name where c.name = ?;

-- name: TaskGroupsListPending :many
select * from task_group;

-- name: TasksAppend :many
insert into task (msgs, repo_uri, dest_branch, post_hook_name, status, timestamp)
select
  ?,
  p.repo_uri,
  p.dest_branch,
  ph.name,
  'pending',
  datetime('now')
from pipeline p
  join subscription s on p.name = s.pipeline_name
  join channel c on s.channel_name = c.name
  -- join the post_hook if it exists but don't fail if it doesn't
  left join post_hook ph on p.post_hook_name = ph.name
where c.name = ?
returning id;


-- name: TaskGroupsStatusCompSwap :many
-- set the status of all tasks in a group where the status matches the
-- req_status. returns the ids of the tasks that were updated
update task
set
  task_group_fingerprint = ?,
  status = ?,
  warnings = ?,
  failure_reason = ?
WHERE status = sqlc.arg(req_status)
and id IN (sqlc.slice('ids'))
returning id;
