from ctf_runtime import get_flag

SUPPORT_KEY = "sync-support-2026"


def handle_sync_gateway(writer, line):
    if line.upper() == "STATUS":
        writer.write(b"STATUS queued=1 replay=0\n")
        return

    if line.upper().startswith("PUSH "):
        parts = line.split(" ", 2)
        if len(parts) < 3:
            writer.write(b"ERR usage: PUSH <device> <payload>\n")
            return
        writer.write(f"QUEUED device={parts[1]}\n".encode("utf-8"))
        return

    if line.upper().startswith("EXPORT "):
        key = line.split(" ", 1)[1].strip()
        if key == SUPPORT_KEY:
            writer.write(f"SNAPSHOT {get_flag()}\n".encode("utf-8"))
            return
        writer.write(b"ERR unauthorized\n")
        return

    writer.write(b"ERR unsupported command\n")
