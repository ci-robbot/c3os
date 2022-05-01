package mos_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c3os-io/c3os/tests/machine"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var _ = Describe("c3os", func() {
	BeforeEach(func() {
		machine.EventuallyConnects()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			machine.SSHCommand("sudo k3s kubectl get pods -A -o json > /run/pods.json")
			machine.SSHCommand("sudo k3s kubectl get events -A -o json > /run/events.json")
			machine.SSHCommand("sudo df -h > /run/disk")
			machine.SSHCommand("sudo mount > /run/mounts")
			machine.SSHCommand("sudo blkid > /run/blkid")

			machine.GatherAllLogs(
				[]string{
					"edgevpn@c3os",
					"c3os-agent",
					"cos-setup-boot",
					"cos-setup-network",
					"c3os",
					"k3s",
				},
				[]string{
					"/var/log/edgevpn.log",
					"/var/log/c3os-agent.log",
					"/run/pods.json",
					"/run/disk",
					"/run/mounts",
					"/run/blkid",
					"/run/events.json",
				})
		}
	})

	Context("live cd", func() {
		It("has default service active", func() {
			if os.Getenv("FLAVOR") == "alpine" {
				out, _ := machine.SSHCommand("sudo rc-status")
				Expect(out).Should(ContainSubstring("c3os"))
				Expect(out).Should(ContainSubstring("c3os-agent"))
			} else {
				// Eventually(func() string {
				// 	out, _ := machine.SSHCommand("sudo systemctl status c3os-agent")
				// 	return out
				// }, 30*time.Second, 10*time.Second).Should(ContainSubstring("no network token"))

				out, _ := machine.SSHCommand("sudo systemctl status c3os")
				Expect(out).Should(ContainSubstring("loaded (/etc/systemd/system/c3os.service; enabled; vendor preset: disabled)"))
			}
		})
	})

	Context("install", func() {
		It("to disk with custom config", func() {
			err := machine.SendFile(os.Getenv("CLOUD_INIT"), "/tmp/config.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			out, _ := machine.SSHCommand("sudo elemental install --cloud-init /tmp/config.yaml /dev/sda")
			Expect(out).Should(ContainSubstring("COS_ACTIVE"))
			fmt.Println(out)
			machine.SSHCommand("sudo sync")
			machine.DetachCD()
			machine.Restart()
		})
	})

	Context("first-boot", func() {
		It("has default services on", func() {
			if os.Getenv("FLAVOR") == "alpine" {
				out, _ := machine.SSHCommand("sudo rc-status")
				Expect(out).Should(ContainSubstring("c3os"))
				Expect(out).Should(ContainSubstring("c3os-agent"))
			} else {
				// Eventually(func() string {
				// 	out, _ := machine.SSHCommand("sudo systemctl status c3os-agent")
				// 	return out
				// }, 30*time.Second, 10*time.Second).Should(ContainSubstring("no network token"))

				out, _ := machine.SSHCommand("sudo systemctl status c3os-agent")
				Expect(out).Should(ContainSubstring("loaded (/etc/systemd/system/c3os-agent.service; enabled; vendor preset: disabled)"))

				out, _ = machine.SSHCommand("sudo systemctl status systemd-timesyncd")
				Expect(out).Should(ContainSubstring("loaded (/usr/lib/systemd/system/systemd-timesyncd.service; enabled; vendor preset: disabled)"))
			}
		})

		It("has additional grub menu entry", func() {
			state, _ := machine.SSHCommand("sudo blkid -L COS_STATE")
			state = strings.TrimSpace(state)
			out, _ := machine.SSHCommand("sudo blkid")
			fmt.Println(out)
			out, _ = machine.SSHCommand("sudo mkdir -p /tmp/mnt/STATE")
			fmt.Println(out)
			out, _ = machine.SSHCommand("sudo mount " + state + " /tmp/mnt/STATE")
			fmt.Println(out)
			out, _ = machine.SSHCommand("sudo cat /tmp/mnt/STATE/grubmenu")
			Expect(out).Should(ContainSubstring("c3os remote recovery"))
			machine.SSHCommand("sudo umount /tmp/mnt/STATE")
		})

		It("configure k3s", func() {
			_, err := machine.SSHCommand("cat /run/cos/live_mode")
			Expect(err).To(HaveOccurred())
			if os.Getenv("FLAVOR") == "alpine" {
				Eventually(func() string {
					out, _ := machine.SSHCommand("sudo cat /var/log/c3os-agent.log")
					return out
				}, 90*time.Minute, 1*time.Second).Should(
					Or(
						ContainSubstring("Configuring k3s-agent"),
						ContainSubstring("Configuring k3s"),
					))
			} else {
				Eventually(func() string {
					out, _ := machine.SSHCommand("sudo systemctl status c3os-agent")
					return out
				}, 30*time.Minute, 1*time.Second).Should(
					Or(
						ContainSubstring("Configuring k3s-agent"),
						ContainSubstring("Configuring k3s"),
					))
			}
		})

		PIt("configure edgevpn", func() {
			Eventually(func() string {
				out, _ := machine.SSHCommand("sudo cat /etc/systemd/system.conf.d/edgevpn-c3os.env")
				return out
			}, 1*time.Minute, 1*time.Second).Should(
				And(
					ContainSubstring("EDGEVPNLOGLEVEL=\"debug\""),
				))
		})

		It("propagate kubeconfig", func() {
			Eventually(func() string {
				out, _ := machine.SSHCommand("c3os get-kubeconfig")
				return out
			}, 900*time.Second, 10*time.Second).Should(ContainSubstring("https:"))

			Eventually(func() string {
				machine.SSHCommand("c3os get-kubeconfig > kubeconfig")
				out, _ := machine.SSHCommand("KUBECONFIG=kubeconfig kubectl get nodes -o wide")
				fmt.Println(out)
				return out
			}, 900*time.Second, 10*time.Second).Should(ContainSubstring("Ready"))
		})

		It("has roles", func() {
			uuid, _ := machine.SSHCommand("c3os uuid")
			Expect(uuid).ToNot(Equal(""))
			Eventually(func() string {
				out, _ := machine.SSHCommand("c3os role list")
				return out
			}, 900*time.Second, 10*time.Second).Should(And(
				ContainSubstring(uuid),
				ContainSubstring("worker"),
				ContainSubstring("master"),
				HaveMinMaxRole("master", 1, 1),
				HaveMinMaxRole("worker", 1, 1),
			))
		})

		It("has machines with different IPs", func() {
			Eventually(func() string {
				out, _ := machine.SSHCommand(`curl http://localhost:8080/api/machines`)
				return out
			}, 900*time.Second, 10*time.Second).Should(And(
				ContainSubstring("10.1.0.1"),
				ContainSubstring("10.1.0.2"),
			))
		})

		It("can propagate dns and it is functional", func() {
			if os.Getenv("FLAVOR") == "alpine" {
				Skip("DNS not working on alpine yet")
			}
			Eventually(func() string {
				machine.SSHCommand(`curl -X POST http://localhost:8080/api/dns --header "Content-Type: application/json" -d '{ "Regex": "foo.bar", "Records": { "A": "2.2.2.2" } }'`)
				out, _ := machine.SSHCommand("ping -c 1 foo.bar")
				return out
			}, 900*time.Second, 10*time.Second).Should(And(
				ContainSubstring("2.2.2.2"),
			))
			Eventually(func() string {
				out, _ := machine.SSHCommand("ping -c 1 google.com")
				return out
			}, 900*time.Second, 10*time.Second).Should(And(
				ContainSubstring("64 bytes from"),
			))
		})

	})

	Context("upgrades", func() {
		It("upgrades to a specific version", func() {
			machine.Snapshot()
			defer machine.RestoreSnapshot()
			version, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")

			out, _ := machine.SSHCommand("sudo c3os upgrade v1.21.4-32")
			Expect(out).To(ContainSubstring("Upgrade completed"))

			machine.SSHCommand("sudo sync")
			machine.Restart()

			machine.EventuallyConnects(700)

			version2, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")
			Expect(version).ToNot(Equal(version2))
		})
	})

	Context("Fallback", func() {
		It("snapshot restored", func() {
			version, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")
			Expect(version).ToNot(ContainSubstring("v1.21.4-32"))
			fmt.Println(version)

		})

		//1: Delete from active.img all except /boot -> should go in fallback with upgrade_failed in /run
		It("boots in fallback when rootfs is damaged", func() {
			defer machine.RestoreSnapshot()
			currentVersion, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")

			cmdline, _ := machine.SSHCommand("sudo cat /proc/cmdline")
			Expect(cmdline).To(ContainSubstring("rd.emergency=reboot rd.shell=0 panic=5"))

			out, _ := machine.SSHCommand("sudo c3os upgrade v1.21.4-32")
			Expect(out).To(ContainSubstring("Upgrade completed"))
			fmt.Println(out)

			// Break the upgrade
			out, _ = machine.SSHCommand("sudo mount -o rw,remount /run/initramfs/cos-state")
			fmt.Println(out)

			out, _ = machine.SSHCommand("sudo cat /run/initramfs/cos-state/c3osenv")
			Expect(out).To(ContainSubstring("osupgrade=done"))

			out, _ = machine.SSHCommand("sudo mkdir -p /tmp/mnt/STATE")
			fmt.Println(out)

			machine.SSHCommand("sudo mount /run/initramfs/cos-state/cOS/active.img /tmp/mnt/STATE")

			for _, d := range []string{"usr/lib/systemd"} {
				out, _ = machine.SSHCommand("sudo rm -rfv /tmp/mnt/STATE/" + d)
			}

			out, _ = machine.SSHCommand("sudo ls -liah /tmp/mnt/STATE/")
			fmt.Println(out)

			out, _ = machine.SSHCommand("sudo umount /tmp/mnt/STATE")

			machine.Restart()
			machine.EventuallyConnects(700)

			v, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")
			Expect(v).To(Equal(currentVersion))

			cmdline, _ = machine.SSHCommand("sudo cat /proc/cmdline")
			Expect(cmdline).To(And(ContainSubstring("passive.img"), ContainSubstring("upgrade_failure")), cmdline)

			Eventually(func() string {
				out, _ := machine.SSHCommand("sudo ls -liah /run")
				return out
			}, 5*time.Minute, 10*time.Second).Should(ContainSubstring("c3os_upgrade_failure"))
		})

		//2: Delete /boot -> should go in fallback without sentinel (grub fallback)
		It("boots in fallback when boot of rootfs is damaged", func() {
			defer machine.RestoreSnapshot()
			currentVersion, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")

			out, _ := machine.SSHCommand("sudo c3os upgrade v1.21.4-32")
			Expect(out).To(ContainSubstring("Upgrade completed"))

			// Break the upgrade
			out, _ = machine.SSHCommand("sudo mount -o rw,remount /run/initramfs/cos-state")
			fmt.Println(out)

			out, _ = machine.SSHCommand("sudo mkdir -p /tmp/mnt/STATE")
			fmt.Println(out)

			machine.SSHCommand("sudo mount /run/initramfs/cos-state/cOS/active.img /tmp/mnt/STATE")

			for _, d := range []string{"bin", "usr", "etc", "sbin", "lib"} {
				machine.SSHCommand("sudo rm -rfv /tmp/mnt/STATE/" + d)
			}

			out, _ = machine.SSHCommand("sudo ls -liah /tmp/mnt/STATE/")
			fmt.Println(out)

			out, _ = machine.SSHCommand("sudo umount /tmp/mnt/STATE")

			machine.Restart()

			machine.EventuallyConnects(700)

			v, _ := machine.SSHCommand("source /etc/os-release; echo $VERSION")
			Expect(v).To(Equal(currentVersion))

			cmdline, _ := machine.SSHCommand("sudo cat /proc/cmdline")
			Expect(cmdline).To(And(ContainSubstring("passive.img")), cmdline)

			// We fallback from grub here, no sentinel from boot assessment
			out, _ = machine.SSHCommand("sudo ls -liah /run")
			Expect(out).ToNot(And(ContainSubstring("c3os_upgrade_failure")), out)
		})
	})
})

func HaveMinMaxRole(name string, min, max int) types.GomegaMatcher {
	return WithTransform(
		func(actual interface{}) (int, error) {
			switch s := actual.(type) {
			case string:
				return strings.Count(s, name), nil
			default:
				return 0, fmt.Errorf("HaveRoles expects a string, but got %T", actual)
			}
		}, SatisfyAll(
			BeNumerically(">=", min),
			BeNumerically("<=", max)))
}
