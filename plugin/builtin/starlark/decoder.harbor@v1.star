load("encoding/json.star", "json")

def main(input):
    data = json.decode(input)

    event = data["event_data"]

    output = []
    for resource in event["resources"]:
        output.append(resource["resource_url"])

    return output
