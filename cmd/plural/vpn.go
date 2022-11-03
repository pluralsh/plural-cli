package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/vpn"
	"github.com/urfave/cli"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	wireguardAppName    = "wireguard"
	wireguardNamespace  = "wireguard"
	wireguardServerName = "wireguard"
)

func (p *Plural) vpnCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list-servers",
			Usage:  "lists vpn servers",
			Action: p.handleWireguardServerList,
		},
		{
			Name:   "list-clients",
			Usage:  "lists vpn clients for a server",
			Action: p.handleWireguardPeerList,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to list clients for",
				},
			},
		},
		{
			Name:      "create-client",
			ArgsUsage: "NAME",
			Usage:     "create a new vpn client for a server",
			Action:    requireArgs(p.handleWireguardPeerCreate, []string{"NAME"}),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to create the client for",
				},
			},
		},
		{
			Name:      "delete-client",
			ArgsUsage: "NAME",
			Usage:     "create a new vpn client for a server",
			Action:    requireArgs(p.handleWireguardPeerDelete, []string{"NAME"}),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "the vpn server to delete the clients from",
				},
			},
		},
		{
			Name:      "client-config",
			ArgsUsage: "NAME",
			Usage:     "get the config for a vpn client for a server",
			Action:    requireArgs(p.handleWireguardPeerConfig, []string{"NAME"}),
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

func (p *Plural) handleWireguardServerList(c *cli.Context) error {
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return utils.HighlightError(err)
	}

	if err := p.checkIfVPNInstalled(); err != nil {
		return utils.HighlightError(fmt.Errorf("wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"))
	}

	servers, err := vpn.ListServers(p.Kube, conf.Namespace(wireguardNamespace))
	if err != nil {
		return utils.HighlightError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Hostname", "Port", "Ready"})
	for _, s := range servers.Items {
		table.Append([]string{s.Name, s.Status.Hostname, s.Status.Port, strconv.FormatBool(s.Status.Ready)})
	}
	table.Render()
	return nil
}

func (p *Plural) handleWireguardPeerList(c *cli.Context) error {
	var server string
	server = wireguardServerName
	if c.String("server") != "" {
		server = c.String("server")
	}

	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return utils.HighlightError(err)
	}

	if err := p.checkIfVPNInstalled(); err != nil {
		return utils.HighlightError(fmt.Errorf("wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"))
	}

	peers, err := vpn.ListPeers(p.Kube, conf.Namespace(wireguardNamespace))
	if err != nil {
		return utils.HighlightError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Address", "Config Secret", "Public Key", "Ready"})
	for _, p := range peers.Items {
		if p.Spec.WireguardRef == server {
			table.Append([]string{
				p.Name,
				p.Spec.Address,
				p.Status.ConfigRef.Name,
				p.Spec.PublicKey,
				strconv.FormatBool(p.Status.Ready),
			})
		}
	}
	table.Render()
	return nil
}

func (p *Plural) handleWireguardPeerCreate(c *cli.Context) error {
	var serverName string
	serverName = wireguardServerName
	if c.String("server") != "" {
		serverName = c.String("server")
	}

	name := c.Args().Get(0)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return utils.HighlightError(err)
	}

	if err := p.checkIfVPNInstalled(); err != nil {
		return utils.HighlightError(fmt.Errorf("wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"))
	}

	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return utils.HighlightError(err)
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
		return utils.HighlightError(err)
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
	if err := p.InitKube(); err != nil {
		return utils.HighlightError(err)
	}

	if err := p.checkIfVPNInstalled(); err != nil {
		return utils.HighlightError(fmt.Errorf("wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"))
	}

	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return utils.HighlightError(err)
	}

	peer, err := vpn.GetPeer(p.Kube, server.Namespace, name)
	if err != nil {
		return utils.HighlightError(err)
	}

	if !peer.Status.Ready || peer.Status.ConfigRef.Name == "" || peer.Status.ConfigRef.Key == "" {
		return utils.HighlightError(fmt.Errorf("peer config not ready yet"))
	}

	secret, err := vpn.GetPeerConfigSecret(p.Kube, peer.Namespace, peer.Status.ConfigRef.Name)
	if err != nil {
		return utils.HighlightError(err)
	}

	peerConfig, ok := secret.Data[peer.Status.ConfigRef.Key]
	if !ok {
		return utils.HighlightError(fmt.Errorf("peer config not ready yet"))
	}

	if c.String("path") != "" {
		err := utils.WriteFile(c.String("path")+peer.Name+".conf", peerConfig)
		if err != nil {
			return utils.HighlightError(err)
		}
		return nil
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
	if err := p.InitKube(); err != nil {
		return utils.HighlightError(err)
	}

	if err := p.checkIfVPNInstalled(); err != nil {
		return utils.HighlightError(fmt.Errorf("wireguard is not installed. run `plural bundle list wireguard` to find the bundle to install"))
	}

	server, err := vpn.GetServer(p.Kube, conf.Namespace(wireguardNamespace), serverName)
	if err != nil {
		return utils.HighlightError(err)
	}

	peer, err := vpn.GetPeer(p.Kube, server.Namespace, name)
	if err != nil {
		return utils.HighlightError(err)
	}

	if err := vpn.DeletePeer(p.Kube, peer.Namespace, peer.Name); err != nil {
		return utils.HighlightError(err)
	}

	utils.Highlight(fmt.Sprintf("Deleted peer %s successfully\n", peer.Name))

	return nil
}

func (p *Plural) checkIfVPNInstalled() error {
	p.InitPluralClient()
	_, err := p.GetInstallation(wireguardAppName)
	if err != nil {
		return utils.HighlightError(err)
	}
	return nil
}
