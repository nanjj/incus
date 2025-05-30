# Requirements

(requirements-go)=
## Go

Incus requires Go 1.23 or higher and is only tested with the Golang compiler.

We recommend having at least 2GiB of RAM to allow the build to complete.

## Kernel requirements

The minimum supported kernel version is 5.15.

Incus requires a kernel with support for:

* Namespaces (`pid`, `net`, `uts`, `ipc` and `mount`)
* Seccomp
* Native Linux AIO
  ([`io_setup(2)`](https://man7.org/linux/man-pages/man2/io_setup.2.html), etc.)

The following optional features also require extra kernel options:

* Namespaces (`user` and `cgroup`)
* AppArmor
* Control Groups (`blkio`, `cpuset`, `devices`, `memory` and `pids`)
* CRIU (exact details to be found with CRIU upstream)

As well as any other kernel feature required by the LXC version in use.

## LXC

Incus requires LXC 5.0.0 or higher with the following build options:

* `apparmor` (if using Incus' AppArmor support)
* `seccomp`

LXCFS is strongly recommended to properly report resource consumption inside the container.

## OCI

To run OCI containers, Incus currently rely on `skopeo`.
`skopeo` should be available in the user's `PATH`.

## QEMU

For virtual machines, QEMU 6.0 or higher is required.

When using `virtiofsd`, only the [Rust rewrite](https://gitlab.com/virtio-fs/virtiofsd) of `virtiofsd` is supported.

## Additional libraries (and development headers)

Incus uses `cowsql` for its database, to build and set it up, you can
run `make deps`.

Incus itself also uses a number of (usually packaged) C libraries:

* `libacl1`
* `libcap2`
* `libuv1` (for `cowsql`)
* `libsqlite3` >= 3.25.0 (for `cowsql`)

Make sure you have all these libraries themselves and their development
headers (`-dev` packages) installed.
