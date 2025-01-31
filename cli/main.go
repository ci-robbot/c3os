package main

import (
	//"fmt"

	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	config "github.com/c3os-io/c3os/cli/config"
	"github.com/c3os-io/c3os/cli/github"
	edgeVPNClient "github.com/mudler/edgevpn/api/client"
	service "github.com/mudler/edgevpn/api/client/service"
	"github.com/mudler/edgevpn/pkg/node"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func main() {
	app := &cli.App{
		Name:        "c3os",
		Version:     "0.1",
		Author:      "Ettore Di Giacinto",
		Usage:       "c3os (register|install)",
		Description: "c3os registers and installs c3os boxes",
		UsageText:   ``,
		Copyright:   "Ettore Di Giacinto",

		Commands: []cli.Command{
			{
				Name: "upgrade",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "force",
					},
					&cli.StringFlag{
						Name: "image",
					},
				},
				Subcommands: []cli.Command{
					{
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "output",
							},
						},
						Name: "list-releases",
						Action: func(c *cli.Context) error {
							rels, err := github.FindReleases(context.Background(), "", "c3os-io/c3os")
							if err != nil {
								return err
							}

							switch strings.ToLower(c.String("output")) {
							case "yaml":
								d, _ := yaml.Marshal(rels)
								fmt.Println(string(d))
							case "json":
								d, _ := json.Marshal(rels)
								fmt.Println(string(d))
							default:
								for _, r := range rels {
									fmt.Println(r)
								}
							}

							return nil
						},
					},
				},
				Action: func(c *cli.Context) error {
					args := c.Args()
					var v string
					if len(args) == 1 {
						v = args[0]
					}

					return upgrade(v, c.String("image"), c.Bool("force"))
				},
			},
			{
				Name: "register",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "config",
					},
					&cli.StringFlag{
						Name: "device",
					},
					&cli.BoolFlag{
						Name: "reboot",
					},
					&cli.BoolFlag{
						Name: "poweroff",
					},
				},
				Action: func(c *cli.Context) error {
					args := c.Args()
					var ref string
					if len(args) == 1 {
						ref = args[0]
					}

					return register(ref, c.String("config"), c.String("device"), c.Bool("reboot"), c.Bool("poweroff"))
				},
			},
			{
				Name:      "create-config",
				Aliases:   []string{"c"},
				UsageText: "Create a config with a generated network token",
				Action: func(c *cli.Context) error {
					l := int(^uint(0) >> 1)
					args := c.Args()
					if len(args) > 0 {
						if i, err := strconv.Atoi(args[0]); err == nil {
							l = i
						}
					}
					cc := &config.Config{C3OS: &config.C3OS{NetworkToken: node.GenerateNewConnectionData(l).Base64()}}
					y, _ := yaml.Marshal(cc)
					fmt.Println(string(y))
					return nil
				},
			},
			{
				Name:      "generate-token",
				Aliases:   []string{"g"},
				UsageText: "Generate a network token",
				Action: func(c *cli.Context) error {
					l := int(^uint(0) >> 1)
					args := c.Args()
					if len(args) > 0 {
						if i, err := strconv.Atoi(args[0]); err == nil {
							l = i
						}
					}
					fmt.Println(node.GenerateNewConnectionData(l).Base64())
					return nil
				},
			},
			{
				Name:    "agent",
				Aliases: []string{"a"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "force",
					},
					&cli.StringFlag{
						Name:  "api",
						Value: "http://127.0.0.1:8080",
					},
				},
				UsageText: "Automatically setups the node",
				Action: func(c *cli.Context) error {
					dirs := []string{"/oem", "/usr/local/cloud-config"}
					args := c.Args()
					if len(args) > 0 {
						dirs = args
					}

					return agent(c.String("api"), dirs, c.Bool("force"))
				},
			},
			{
				Name:    "rotate",
				Aliases: []string{"r"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "restart",
					},
					&cli.StringFlag{
						Name:   "network-token",
						EnvVar: "NETWORK_TOKEN",
					},
					&cli.StringFlag{
						Name:  "api",
						Value: "127.0.0.1:8080",
					},
					&cli.StringFlag{
						Name: "root-dir",
					},
				},
				UsageText: "Rotate network token manually in the node",
				Action: func(c *cli.Context) error {
					dirs := []string{"/oem", "/usr/local/cloud-config"}
					args := c.Args()
					if len(args) > 0 {
						dirs = args
					}

					return rotate(dirs, c.String("network-token"), c.String("api"), c.String("root-dir"), c.Bool("restart"))
				},
			},
			{
				Name: "bridge",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "network-token",
						Required: false,
						EnvVar:   "NETWORK_TOKEN",
					},
					&cli.StringFlag{
						Name:     "log-level",
						Required: false,
						EnvVar:   "LOGLEVEL",
						Value:"info",
					},
					&cli.BoolFlag{
						Name:     "qr-code-snapshot",
						Required: false,
						EnvVar:   "QR_CODE_SNAPSHOT",
					},
					&cli.StringFlag{
						Name:     "qr-code-image",
						Required: false,
						EnvVar:   "QR_CODE_IMAGE",
					},
					&cli.StringFlag{
						Name:  "api",
						Value: "127.0.0.1:8080",
					},
					&cli.BoolFlag{
						Name:   "dhcp",
						EnvVar: "DHCP",
					},
					&cli.StringFlag{
						Value:  "10.1.0.254/24",
						Name:   "address",
						EnvVar: "ADDRESS",
					},
					&cli.StringFlag{
						Value:  "/tmp/c3os",
						Name:   "lease-dir",
						EnvVar: "lease-dir",
					},
				},
				Action: bridge,
			},
			{
				Name: "get-kubeconfig",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "api",
						Value: "http://localhost:8080",
					},
					&cli.StringFlag{
						Name:  "network-id",
						Value: "c3os",
					},
				},
				Action: func(c *cli.Context) error {
					cc := service.NewClient(
						c.String("network-id"),
						edgeVPNClient.NewClient(edgeVPNClient.WithHost(c.String("api"))))
					str, _ := cc.Get("kubeconfig", "master")
					b, _ := base64.RawURLEncoding.DecodeString(str)
					masterIP, _ := cc.Get("master", "ip")
					fmt.Println(strings.ReplaceAll(string(b), "127.0.0.1", masterIP))
					return nil
				},
			},
			{
				Name: "role",
				Subcommands: []cli.Command{
					{
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "api",
								Value: "http://localhost:8080",
							},
							&cli.StringFlag{
								Name:  "network-id",
								Value: "c3os",
							},
						},
						Name:        "set",
						Description: "Set node role. Usage: <uuid> <role>. Available roles: worker and master.",
						Action: func(c *cli.Context) error {
							cc := service.NewClient(
								c.String("network-id"),
								edgeVPNClient.NewClient(edgeVPNClient.WithHost(c.String("api"))))
							return cc.Set("role", c.Args()[0], c.Args()[1])
						},
					},
					{
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "api",
								Value: "http://localhost:8080",
							},
							&cli.StringFlag{
								Name:  "network-id",
								Value: "c3os",
							},
						},
						Name:        "list",
						Description: "List node roles",
						Action: func(c *cli.Context) error {
							cc := service.NewClient(
								c.String("network-id"),
								edgeVPNClient.NewClient(edgeVPNClient.WithHost(c.String("api"))))
							advertizing, _ := cc.AdvertizingNodes()
							fmt.Println("Node\tRole")
							for _, a := range advertizing {
								role, _ := cc.Get("role", a)
								fmt.Printf("%s\t%s\n", a, role)
							}
							return nil
						},
					},
				},
			},

			{
				Name:        "get-network-token",
				Description: "Print network token",
				Action: func(c *cli.Context) error {
					dirs := []string{"/oem", "/usr/local/cloud-config"}
					args := c.Args()
					if len(args) > 0 {
						dirs = args
					}
					cc, err := config.Scan(dirs...)
					if err != nil {
						return err
					}
					fmt.Print(cc.C3OS.NetworkToken)
					return nil
				},
			},
			{
				Name:        "uuid",
				Description: "Print node uuid",
				Aliases:     []string{"u"},
				Action: func(c *cli.Context) error {
					fmt.Print(uuid())
					return nil
				},
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Action: func(c *cli.Context) error {
					return install("/oem", "/usr/local/cloud-config")
				},
			},
			{
				Name:    "recovery",
				Aliases: []string{"r"},
				Action:  recovery,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
