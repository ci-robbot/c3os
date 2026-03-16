ARG BASE_IMAGE=registry.suse.com/suse/sle-micro/5.5:latest
FROM $BASE_IMAGE

ADD repos/SUSE* /etc/zypp/repos.d/
ADD services/* /etc/zypp/services.d/
RUN zypper --gpg-auto-import-keys ref
RUN zypper in --force-resolution -y \
    cryptsetup \
    nfs-client \
    kernel-firmware-all \
    systemd-network \
    polkit \
    rng-tools \
    nano \
    growpart \
    && zypper cc

ADD repos/opensuse* /etc/zypp/repos.d/
RUN zypper --gpg-auto-import-keys ref
RUN zypper in --force-resolution -y --no-allow-vendor-change \
    htop \
    fail2ban \
    lldpd \
    nethogs \
    rsyslog \
    logrotate \
    && zypper cc

#### REMOVE ELEMENTAL AND CLEANUP  #######
RUN zypper rm -y elemental \
    elemental-register1.5 \
    elemental-support1.5 \
    elemental-system-agent \
    elemental-toolkit \
    elemental-updater && \
    zypper clean -a

RUN systemctl mask elemental-setup-reconcile.service \
    systemctl mask elemental-register-reset.service \
    systemctl mask elemental-register.service \
    systemctl mask elemental-setup-network.service \
    systemctl mask elemental-setup-fs.service \
    systemctl mask elemental-populate-node-labels.service \
    systemctl mask elemental-system-agent.service \
    systemctl mask elemental-setup-rootfs.service \
    systemctl mask elemental-setup-boot.service \
    systemctl mask elemental-setup-initramfs.service \
    systemctl mask elemental-register-install.service \
    systemctl mask elemental-immutable-rootfs.service
RUN rm -rf /system/oem/*_elemental-* \
    /etc/dracut.conf.d/99-elemental-systemd.conf \
    /etc/dracut.conf.d/02-elemental-immutable-rootfs.conf \
    /etc/dracut.conf.d/02-elemental-setup-initramfs.conf \
    /etc/dracut.conf.d/50-elemental-initrd.conf

# NOTE: removed dhcp-client
RUN mkdir -p /run/lock
RUN mkdir -p /usr/libexec
RUN touch /usr/libexec/.keep

COPY --from=quay.io/kairos/framework:v2.11.7 / /

RUN sed -i 's/\bsource\b/./g' /system/oem/00_rootfs.yaml
RUN sed -i 's/\bsource\b/./g' /system/oem/09_openrc_services.yaml
RUN sed -i 's/\bsource\b/./g' /system/oem/50_recovery.yaml


# Remove file below to allow dracut to build initrd without dhcp-client
RUN rm -rf /usr/lib/dracut/modules.d/35network-legacy

## Generate initrd
RUN kernel=$(ls /boot/vmlinuz-* | head -n1) && \
            ln -sf "${kernel#/boot/}" /boot/vmlinuz
RUN kernel=$(ls /lib/modules | head -n1) && \
            dracut -v -N -f "/boot/initrd-${kernel}" "${kernel}" && \
            ln -sf "initrd-${kernel}" /boot/initrd && depmod -a "${kernel}"
RUN kernel=$(ls /lib/modules | head -n1) && dracut -f "/boot/initrd-${kernel}" "${kernel}" && ln -sf "initrd-${kernel}" /boot/initrd

# Cleanup
RUN rm -rf /boot/initramfs-*
RUN rm -rf /etc/zypp/repos.d/*
RUN rm -rf /etc/zypp/services.d/*
