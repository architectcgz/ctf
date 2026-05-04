#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

static void win(void) {
    const char *flag = getenv("FLAG");
    puts(flag ? flag : "flag{pwn_template_local}");
}

int main(void) {
    char buf[64];

    setbuf(stdin, NULL);
    setbuf(stdout, NULL);
    alarm(120);

    puts("pwn training");
    puts("name:");
    read(0, buf, 256);
    puts("bye");
    return 0;
}
