#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

__attribute__((noinline)) void unlock_relay(void) {
    const char *flag = getenv("FLAG");
    puts(flag ? flag : "flag{local_passkey_recovery_relay}");
    fflush(stdout);
}

__attribute__((noinline)) void collect_phrase(void) {
    char phrase[64];

    puts("passkey recovery relay");
    puts("paste recovery phrase:");
    read(STDIN_FILENO, phrase, 192);
    puts("request queued");
}

int main(void) {
    setbuf(stdin, NULL);
    setbuf(stdout, NULL);
    setbuf(stderr, NULL);
    alarm(60);

    collect_phrase();
    return 0;
}
