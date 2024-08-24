load("http.star", "http")

def main(repo, src_branch, dest_branch, title, body, changes, warnings):
    parts = repo.split("/")
    name = parts[-1].removesuffix(".git")
    owner = parts[-2].split(":")[-1]

    url = host_env["GITEA_HOST"] + "/api/v1/repos/" + owner + "/" + name + "/pulls"

    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": host_env["GITEA_AUTH_HEADER"],
    }

    data = {"title": title, "body": body, "head": dest_branch, "base": src_branch}

    res = http.post(url, headers = headers, json_body = data)
    if res.status_code != 201:
        print("hook: pr failed: " + url)
        return res.body()

    print("gitea pr created: " + res.json()["url"])
    return None
