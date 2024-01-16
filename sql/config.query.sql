-- name: ChannelPut :exec
insert into channel(name, decoder_name) values (?, ?)
on conflict(name) do update set decoder_name = excluded.decoder_name;

-- name: DecoderPut :exec
insert into decoder(name, script) values (?, ?)
on conflict(name) do update set script = excluded.script;

-- name: PipelinePut :exec
insert into pipeline(name, repo_uri, dest_branch, post_hook_name) values (?, ?, ?, ?)
on conflict(name) do update set repo_uri = excluded.repo_uri, dest_branch = excluded.dest_branch, post_hook_name = excluded.post_hook_name;

-- name: SubscriptionPut :exec
insert into subscription(pipeline_name, channel_name) values (?, ?)
on conflict(pipeline_name, channel_name) do nothing;

-- name: PostHookPut :exec
insert into post_hook(name, script) values (?, ?)
on conflict(name) do update set script = excluded.script;