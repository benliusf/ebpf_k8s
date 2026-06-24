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

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(void *));
	__uint(max_entries, 1024);
} read_bufs SEC(".maps");

struct write_event {
	__s32 fd;
};

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 256 * 1024);
} write_fds SEC(".maps");

bool is_pong_server() {
	const char *pong_server = "pong_server";
	char comm[16];
	return bpf_get_current_comm(&comm, sizeof(comm)) == 0 &&
		   bpf_strncmp(comm, 12, pong_server) == 0;
}

SEC("tracepoint/syscalls/sys_enter_accept4")
int handle_sys_enter_accept4(struct trace_event_raw_sys_enter *ctx) {
	if (is_pong_server()) {
		u32 pid = bpf_get_current_pid_tgid() >> 32;
		bpf_printk("sys_enter_accept4 at pong_server: pid=%d, fd=%d\n", pid,
				   ctx->args[0]);
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_exit_accept4")
int handle_sys_exit_accept4(struct trace_event_raw_sys_exit *ctx) {
	if (is_pong_server()) {
		u32 pid = bpf_get_current_pid_tgid() >> 32;
		int fd = ctx->ret;
		bpf_printk("sys_exit_accept4 at pong_server: pid=%d, fd=%d\n", pid, fd);
		struct connection_event *e;
		e = bpf_ringbuf_reserve(&connection_fds, sizeof(*e), 0);
		if (!e)
			return 0;
		e->pid = pid;
		e->fd = fd;
		bpf_ringbuf_submit(e, 0);
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_read")
int handle_sys_enter_read(struct trace_event_raw_sys_enter *ctx) {
	if (is_pong_server()) {
		u32 pid = bpf_get_current_pid_tgid() >> 32;
		u32 tid = (u32)bpf_get_current_pid_tgid();
		int fd = ctx->args[0];
		bpf_printk("sys_enter_read at pong_server: pid=%d, tid=%d, fd=%d\n",
				   pid, tid, fd);
		void *buf = (void *)ctx->args[1];
		bpf_map_update_elem(&read_bufs, &tid, &buf, BPF_ANY);
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_exit_read")
int handle_sys_exit_read(struct trace_event_raw_sys_exit *ctx) {
	if (is_pong_server()) {
		u32 pid = bpf_get_current_pid_tgid() >> 32;
		u32 tid = (u32)bpf_get_current_pid_tgid();
		void **buf_ptr = bpf_map_lookup_elem(&read_bufs, &tid);
		if (buf_ptr) {
			void *buf = *buf_ptr;
			char data[256];
			bpf_probe_read_user_str(data, sizeof(data), buf);
			bpf_map_delete_elem(&read_bufs, &tid);
			bpf_printk("sys_exit_read at pong_server: pid=%d, tid=%d, "
					   "bytes_length=%d, data=%s\n",
					   pid, tid, ctx->ret, data);
		}
	}
	return 0;
}

SEC("tracepoint/syscalls/sys_enter_write")
int handle_sys_enter_write(struct trace_event_raw_sys_enter *ctx) {
	if (is_pong_server()) {
		u32 pid = bpf_get_current_pid_tgid() >> 32;
		int fd = ctx->args[0];
		bpf_printk("sys_enter_write at pong_server: pid=%d, fd=%d, data=%s\n",
				   pid, fd, ctx->args[1]);
		struct write_event *e;
		e = bpf_ringbuf_reserve(&write_fds, sizeof(*e), 0);
		if (!e)
			return 0;
		e->fd = fd;
		bpf_ringbuf_submit(e, 0);
	}
	return 0;
}
