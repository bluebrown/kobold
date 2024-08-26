load("http.star", "http")

def main(repo, src_branch, dest_branch, title, body, changes, warnings):
    parts = repo.split("/")
    name = parts[-1].removesuffix(".git")
    owner = parts[-2].split(":")[-1]

    url = "https://api.github.com/repos/" + owner + "/" + name + "/pulls"

    headers = {
        "Accept": "application/vnd.github+json",
        "Authorization": "Bearer " + host_env["GITHUB_TOKEN"],
        "X-GitHub-Api-Version": "2022-11-28",
    }

    data = {"title": title, "body": body, "head": dest_branch, "base": src_branch}

    res = http.post(url, headers = headers, json_body = data)
    if res.status_code != 201:
        print("hook: pr failed: " + url)
        return res.body()

    print("pull request created: " + res.json()["url"])
    return None
