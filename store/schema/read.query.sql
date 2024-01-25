-- name: ChannelGet :one
select * from channel where name = ?;

-- name: ChannelList :many
select * from channel;

-- name: DecoderGet :one
select * from decoder where name = ?;

-- name: DecoderList :many
select * from decoder;

-- name: PipelineGet :one
select * from pipeline_list_item where name = ?;

-- name: PipelineList :many
select * from pipeline_list_item;

-- name: PostHookGet :one
select * from post_hook where name = ?;

-- name: PostHookList :many
select * from post_hook;

-- name: TaskGet :one
select * from task where id = ?;

-- name: TaskList :many
select * from task
where status in (sqlc.slice('status'))
order by timestamp desc
limit ? offset ?;

-- name: RunGet :one
select * from run
where fingerprint = ?;

-- name: RunList :many
select * from run
where status in (sqlc.slice('status'))
limit ? offset ?;

-- name: PipelineRunList :many
select p.name, r.* from run r
left join pipeline p on r.repo_uri = p.repo_uri and ifnull(r.dest_branch, '') = ifnull(p.dest_branch, '')
where p.name = ?
and r.status in (sqlc.slice('status'))
limit ? offset ?;
