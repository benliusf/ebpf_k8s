//go:build ignore
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

char _license[] SEC("license") = "GPL";

struct connection_event {
	__u32 pid;
	__s32 fd;
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 256 * 1024);
} connection_fds SEC(".maps");

struct write_event {
	__s32 fd;
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 256 * 1024);
} write_fds SEC(".maps");

int strings_are_equal(const char *s1, const char *s2) {
	while (*s1 == *s2) {
		if (*s1 == '\0') {
			return 1;
		}
		s1++;
		s2++;
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_accept4")
int handle_sys_enter_accept4(struct trace_event_raw_sys_enter *ctx) {
	char comm[16];
	if (bpf_get_current_comm(&comm, sizeof(comm)) == 0) {
		if (strings_are_equal(comm, "pong_server")) {
			int pid = bpf_get_current_pid_tgid() >> 32;
			bpf_printk("sys_enter_accept4 at pong_server: pid=%d, fd=%d\n", pid, ctx->args[0]);
		}
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_exit_accept4")
int handle_sys_exit_accept4(struct trace_event_raw_sys_exit *ctx) {
	char comm[16];
	if (bpf_get_current_comm(&comm, sizeof(comm)) == 0) {
		int pid = bpf_get_current_pid_tgid() >> 32;
		int fd = ctx->ret;
		if (strings_are_equal(comm, "pong_server")) {
			bpf_printk("sys_exit_accept4 at pong_server: pid=%d, fd=%d\n", pid, fd);
			struct connection_event *e;
			e = bpf_ringbuf_reserve(&connection_fds, sizeof(*e), 0);
			if (!e)
				return 0;
			e->pid = pid;
			e->fd = fd;
			bpf_ringbuf_submit(e, 0);
		}
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_write")
int handle_sys_enter_write(struct trace_event_raw_sys_enter *ctx) {
	char comm[16];
	if (bpf_get_current_comm(&comm, sizeof(comm)) == 0) {
		int pid = bpf_get_current_pid_tgid() >> 32;
		int fd = ctx->args[0];
		if (strings_are_equal(comm, "pong_server")) {
			bpf_printk("sys_enter_write at pong_server: pid=%d, fd=%d, d=%s\n",	pid, fd, ctx->args[1]);
			struct write_event *e;
			e = bpf_ringbuf_reserve(&write_fds, sizeof(*e), 0);
			if (!e)
				return 0;
			e->fd = fd;
			bpf_ringbuf_submit(e, 0);
		}
	}
	return 0;
}
