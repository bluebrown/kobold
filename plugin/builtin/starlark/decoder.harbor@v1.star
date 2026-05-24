load("encoding/json.star", "json")

def main(input):
    data = json.decode(input)

    event = data["event_data"]

    output = []
    for resource in event["resources"]:
        parts = resource["resource_url"].split("@")
        name = parts[0].split(":")[0]
        output.append(name + ":" + resource["tag"] + "@" + resource["digest"])

    return output
