load("http.star", "http")

def main(repo, src_branch, dest_branch, title, body, changes, warnings):
    org, proj, repo, err = get_org_proj_repo(repo)
    if err != None:
        return err

    repo = repo.removesuffix(".git")

    url = "https://dev.azure.com/" + org + "/" + proj + "/_apis/git/repositories/" + repo + "/pullrequests?api-version=7.0"

    headers = {"Content-Type": "application/json"}

    data = {
        "sourceRefName": "refs/heads/" + dest_branch,
        "targetRefName": "refs/heads/" + src_branch,
        "title": title,
        "description": body,
    }

    res = http.post(url, headers = headers, json_body = data, auth = (host_env["ADO_USR"], host_env["ADO_PAT"]))
    if res.status_code != 201:
        print("hook: pr failed: base=" + src_branch + " head=" + dest_branch + " repo=" + repo)
        return res.body()

    print("pull request created: " + res.json()["url"])
    return None

def get_org_proj_repo(url):
    parts = url.split("/")
    if url.startswith("git@ssh"):
        if len(parts) != 4:
            return None, None, None, "invalid url"
        return parts[1], parts[2], parts[3], None
    else:
        if len(parts) != 7:
            return None, None, None, "invalid url"
        return parts[3], parts[4], parts[6], None
