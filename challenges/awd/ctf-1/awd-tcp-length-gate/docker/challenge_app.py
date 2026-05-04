from ctf_runtime import get_flag


def handle_length_gate(writer, payload):
    length = len(payload.encode("utf-8"))
    if length > 40 and "magic" in payload.lower():
        value = get_flag()
        writer.write((f"accepted length={length}\n{value}\n").encode("utf-8"))
        return
    writer.write((f"rejected length={length}\nneed a longer payload with magic\n").encode("utf-8"))
