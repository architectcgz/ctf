#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

__attribute__((used, naked)) void gadget_ret(void) { __asm__("ret"); }
__attribute__((used, naked)) void gadget_pop_rdi(void) { __asm__("pop %rdi; ret"); }
__attribute__((used, naked)) void gadget_pop_rsi(void) { __asm__("pop %rsi; ret"); }
__attribute__((used, naked)) void gadget_pop_rdx(void) { __asm__("pop %rdx; ret"); }

__attribute__((noinline)) void reveal(uint64_t a, uint64_t b, uint64_t c) {
    const char *flag = getenv("FLAG");
    if (a == 0x1337C0DECAFEF00DULL &&
        b == 0x4142434445464748ULL &&
        c == 0xDEADBEEF10203040ULL) {
        puts(flag ? flag : "flag{local_pwn_rop_register_vault}");
        fflush(stdout);
        return;
    }
    puts("denied");
    fflush(stdout);
}

int main(void) {
    char buf[64] = {0};

    setbuf(stdin, NULL);
    setbuf(stdout, NULL);
    setbuf(stderr, NULL);
    alarm(60);

    puts("vault:");
    read(STDIN_FILENO, buf, 256);
    puts("closed");
    return 0;
}
