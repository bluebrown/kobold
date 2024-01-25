load("encoding/json.star", "json")

def main(input):
    data = json.decode(input)
    repo = "docker.io" + "/" + data["repository"]["repo_name"]
    tag = data["push_data"]["tag"]
    return [repo + ":" + tag]
