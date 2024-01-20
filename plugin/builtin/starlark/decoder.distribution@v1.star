load("encoding/json.star", "json")

def main(input):
    data = json.decode(input)

    events = data.get("events", None)
    if events == None:
        events = [data]

    output = []
    for item in events:
        output.append(
            item["request"]["host"] + "/" + item["target"]["repository"] +
            ":" + item["target"]["tag"] +
            "@" + item["target"]["digest"],
        )

    return output
