subarch: x86-64-v3
target: stage4
version_stamp: nomultilib-systemd-mergedusr-@TIMESTAMP@
rel_type: mergedusr
profile: default/linux/amd64/17.1/no-multilib/systemd/merged-usr
snapshot_treeish: 23d24f462e77786cdbba9b77a0f705287ee387c0
source_subpath: mergedusr/stage1-x86-64-v3-nomultilib-systemd-mergedusr-@TIMESTAMP@
portage_confdir: /var/tmp/catalyst/config/stages
binrepo_path: amd64/binpackages/17.1/x86-64
compression_mode: gzip

stage4/use:
    -debug
    -nls

stage4/packages:
    sys-cluster/ipvsadm

stage4/root_overlay: /var/tmp/catalyst/root_overlay

stage4/rm: /usr/share/man
stage4/unmerge:
    sys-devel/gcc
    net-misc/openssh
    sys-fs/e2fsprogs
    net-misc/rsync
