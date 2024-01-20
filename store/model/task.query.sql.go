// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: task.query.sql

package model

import (
	"context"
	"strings"

	store "github.com/bluebrown/kobold/store"
	null "github.com/volatiletech/null/v8"
)

const channelDecoderGet = `-- name: ChannelDecoderGet :one
select d.script from channel c left join decoder d on c.decoder_name = d.name where c.name = ?
`

// ChannelDecoderGet
//
//	select d.script from channel c left join decoder d on c.decoder_name = d.name where c.name = ?
func (q *Queries) ChannelDecoderGet(ctx context.Context, name string) ([]byte, error) {
	row := q.db.QueryRowContext(ctx, channelDecoderGet, name)
	var script []byte
	err := row.Scan(&script)
	return script, err
}

const taskGroupsListPending = `-- name: TaskGroupsListPending :many
select fingerprint, repo_uri, dest_branch, post_hook, task_ids, msgs from task_group
`

// TaskGroupsListPending
//
//	select fingerprint, repo_uri, dest_branch, post_hook, task_ids, msgs from task_group
func (q *Queries) TaskGroupsListPending(ctx context.Context) ([]TaskGroup, error) {
	rows, err := q.db.QueryContext(ctx, taskGroupsListPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TaskGroup{}
	for rows.Next() {
		var i TaskGroup
		if err := rows.Scan(
			&i.Fingerprint,
			&i.RepoUri,
			&i.DestBranch,
			&i.PostHook,
			&i.TaskIds,
			&i.Msgs,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const taskGroupsStatusCompSwap = `-- name: TaskGroupsStatusCompSwap :many
update task
set
  task_group_fingerprint = ?,
  status = ?,
  warnings = ?,
  failure_reason = ?
WHERE status = ?5
and id IN (/*SLICE:ids*/?)
returning id
`

type TaskGroupsStatusCompSwapParams struct {
	TaskGroupFingerprint null.String     `json:"task_group_fingerprint"`
	Status               string          `json:"status"`
	Warnings             store.SliceText `json:"warnings"`
	FailureReason        null.String     `json:"failure_reason"`
	ReqStatus            string          `json:"req_status"`
	Ids                  []string        `json:"ids"`
}

// set the status of all tasks in a group where the status matches the
// req_status. returns the ids of the tasks that were updated
//
//	update task
//	set
//	  task_group_fingerprint = ?,
//	  status = ?,
//	  warnings = ?,
//	  failure_reason = ?
//	WHERE status = ?5
//	and id IN (/*SLICE:ids*/?)
//	returning id
func (q *Queries) TaskGroupsStatusCompSwap(ctx context.Context, arg TaskGroupsStatusCompSwapParams) ([]string, error) {
	query := taskGroupsStatusCompSwap
	var queryParams []interface{}
	queryParams = append(queryParams, arg.TaskGroupFingerprint)
	queryParams = append(queryParams, arg.Status)
	queryParams = append(queryParams, arg.Warnings)
	queryParams = append(queryParams, arg.FailureReason)
	queryParams = append(queryParams, arg.ReqStatus)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const tasksAppend = `-- name: TasksAppend :many
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
returning id
`

type TasksAppendParams struct {
	Msgs store.SliceText `json:"msgs"`
	Name string          `json:"name"`
}

// TasksAppend
//
//	insert into task (msgs, repo_uri, dest_branch, post_hook_name, status, timestamp)
//	select
//	  ?,
//	  p.repo_uri,
//	  p.dest_branch,
//	  ph.name,
//	  'pending',
//	  datetime('now')
//	from pipeline p
//	  join subscription s on p.name = s.pipeline_name
//	  join channel c on s.channel_name = c.name
//	  -- join the post_hook if it exists but don't fail if it doesn't
//	  left join post_hook ph on p.post_hook_name = ph.name
//	where c.name = ?
//	returning id
func (q *Queries) TasksAppend(ctx context.Context, arg TasksAppendParams) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, tasksAppend, arg.Msgs, arg.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}