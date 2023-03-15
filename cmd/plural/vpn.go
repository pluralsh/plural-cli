package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/vpn"
	"github.com/pluralsh/polly/algorithms"
	"github.com/urfave/cli"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	wireguardAppName           = "wireguard"
	wireguardNamespace         = "wireguard"
	wireguardServerName        = "wireguard"
	wireguardNotInstalledError = "wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"
)

func (p *Plural) vpnCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "list",
			Usage:       "list vpn resources",
			Subcommands: p.vpnListCommands(),
		},
		{
			Name:        "create",
			Usage:       "commands for creating vpn resources",
			Subcommands: p.vpnCreateCommands(),
		},
		{
			Name:        "delete",
			Usage:       "commands for deleting vpn resources",
			Subcommands: p.vpnDeleteCommands(),
		},
		{
			Name:      "client-config",
			ArgsUsage: "NAME",
			Usage:     "get the config for a vpn client for a server",
			Action:    latestVersion(requireArgs(highlighted(p.vpnInstalled(initKubeconfig(p.handleWireguardPeerConfig))), []string{"NAME"})),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to get the client config from",
				},
				cli.StringFlag{
					Name:  "path",
					Usage: "output path for the config wireguard client config. the filename will be NAME.conf",
				},
			},
		},
	}
}

func (p *Plural) vpnListCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "servers",
			Usage:  "lists vpn servers",
			Action: latestVersion(highlighted(p.vpnInstalled(initKubeconfig(p.handleWireguardServerList)))),
		},
		{
			Name:   "clients",
			Usage:  "lists vpn clients for a server",
			Action: latestVersion(highlighted(p.vpnInstalled(initKubeconfig(p.handleWireguardPeerList)))),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to list clients for",
				},
			},
		},
	}
}

func (p *Plural) vpnCreateCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "client",
			ArgsUsage: "NAME",
			Usage:     "create a new vpn client for a server",
			Action:    latestVersion(requireArgs(highlighted(p.vpnInstalled(initKubeconfig(p.handleWireguardPeerCreate))), []string{"NAME"})),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to create the client for",
				},
			},
		},
	}
}

func (p *Plural) vpnDeleteCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "client",
			ArgsUsage: "NAME",
			Usage:     "delete a vpn client for a server",
			Action:    latestVersion(requireArgs(highlighted(p.vpnInstalled(initKubeconfig(p.handleWireguardPeerDelete))), []string{"NAME"})),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to delete the clients from",
				},
			},
		},
	}
}

func (p *Plural) handleWireguardServerList(c *cli.Context) error {
	conf := config.Read()
	servers, err := vpn.ListServers(p.Kube, conf.Namespace(wireguardNamespace))
	if err != nil {
		return err
	}

	headers := []string{"Name", "Hostname", "Port", "Ready"}
	return utils.PrintTable(servers.Items, headers, func(s v1alpha1.WireguardServer) ([]string, error) {
		return []string{s.Name, s.Status.Hostname, s.Status.Port, strconv.FormatBool(s.Status.Ready)}, nil
	})
}

func (p *Plural) handleWireguardPeerList(c *cli.Context) error {
	var server string
	server = wireguardServerName
	if c.String("server") != "" {
		server = c.String("server")
	}

	conf := config.Read()
	peerlist, err := vpn.ListPeers(p.Kube, conf.Namespace(wireguardNamespace))
	if err != nil {
		return err
	}

	peers := algorithms.Filter(peerlist.Items, func(p v1alpha1.WireguardPeer) bool { return p.Spec.WireguardRef == server })
	headers := []string{"Name", "Address", "Config Secret", "Public Key", "Ready"}
	return utils.PrintTable(peers, headers, func(p v1alpha1.WireguardPeer) (res []string, err error) {
		res = []string{p.Name, p.Spec.Address, p.Status.ConfigRef.Name, p.Spec.PublicKey, strconv.FormatBool(p.Status.Ready)}
		return
	})
}

func (p *Plural) handleWireguardPeerCreate(c *cli.Context) error {
	var serverName string
	serverName = wireguardServerName
	if c.String("server") != "" {
		serverName = c.String("server")
	}

	name := c.Args().Get(0)
	conf := config.Read()
	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return err
	}
	peer, err := vpn.CreatePeer(p.Kube, server.Namespace,
		&v1alpha1.WireguardPeer{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.WireguardPeerSpec{
				WireguardRef: server.Name,
			},
		})
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Address", "Server", "Config Secret", "Public Key", "Ready"})
	table.Append([]string{
		peer.Name,
		peer.Spec.Address,
		peer.Spec.WireguardRef,
		peer.Status.ConfigRef.Name,
		peer.Spec.PublicKey,
		strconv.FormatBool(peer.Status.Ready),
	})
	table.Render()
	return nil
}

func (p *Plural) handleWireguardPeerConfig(c *cli.Context) error {
	var serverName string
	serverName = wireguardServerName
	if c.String("server") != "" {
		serverName = c.String("server")
	}

	name := c.Args().Get(0)
	conf := config.Read()

	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return err
	}

	peer, err := vpn.GetPeer(p.Kube, server.Namespace, name)
	if err != nil {
		return err
	}

	if !peer.Status.Ready || peer.Status.ConfigRef.Name == "" || peer.Status.ConfigRef.Key == "" {
		return fmt.Errorf("peer config not ready yet")
	}

	secret, err := vpn.GetPeerConfigSecret(p.Kube, peer.Namespace, peer.Status.ConfigRef.Name)
	if err != nil {
		return err
	}

	peerConfig, ok := secret.Data[peer.Status.ConfigRef.Key]
	if !ok {
		return fmt.Errorf("peer config not ready yet")
	}

	if c.String("path") != "" {
		path := pathing.SanitizeFilepath(filepath.Join(c.String("path"), peer.Name+".conf"))
		return utils.WriteFile(path, peerConfig)
	}
	fmt.Println(string(peerConfig))
	return nil
}

func (p *Plural) handleWireguardPeerDelete(c *cli.Context) error {
	var serverName string
	serverName = wireguardServerName
	if c.String("server") != "" {
		serverName = c.String("server")
	}

	name := c.Args().Get(0)
	conf := config.Read()
	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return err
	}

	peer, err := vpn.GetPeer(p.Kube, server.Namespace, name)
	if err != nil {
		return err
	}

	if err := vpn.DeletePeer(p.Kube, peer.Namespace, peer.Name); err != nil {
		return err
	}

	utils.Highlight(fmt.Sprintf("Deleted peer %s successfully\n", peer.Name))
	return nil
}

func (p *Plural) vpnInstalled(fn func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		p.InitPluralClient()
		if err := p.InitKube(); err != nil {
			return err
		}

		if _, err := p.GetInstallation(wireguardAppName); err != nil {
			return err
		}

		return fn(c)
	}
}
