#include <libavutil/log.h>

extern void goavLog(int level, char *str);

#define LINE_SZ 1024

void goavLogCallback(void *class_ptr, int level, const char *fmt, va_list vl) {
    char line[LINE_SZ];
    int print_prefix = 1;
    av_log_format_line(class_ptr, level, fmt, vl, line, LINE_SZ, &print_prefix);
    goavLog(level, line);
}

void goavLogSetup() {
    av_log_set_callback(goavLogCallback);
}
