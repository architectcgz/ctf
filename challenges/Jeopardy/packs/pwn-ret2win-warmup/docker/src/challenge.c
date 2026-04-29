#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

__attribute__((noinline)) void win(void) {
    const char *flag = getenv("FLAG");
    puts(flag ? flag : "flag{local_ret2win_warmup}");
    fflush(stdout);
}

__attribute__((noinline)) void vuln(void) {
    char name[64];

    puts("ret2win warmup");
    puts("What is your name?");
    read(STDIN_FILENO, name, 160);
    puts("bye");
}

int main(void) {
    setbuf(stdin, NULL);
    setbuf(stdout, NULL);
    setbuf(stderr, NULL);
    alarm(60);

    vuln();
    return 0;
}
