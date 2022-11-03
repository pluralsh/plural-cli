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

func (p *Plural) vpnCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list-servers",
			Usage:  "lists vpn servers",
			Action: p.handleWireguardServerList,
		},
		{
			Name:      "list-clients",
			ArgsUsage: "SERVER",
			Usage:     "lists vpn clients for a server",
			Action:    requireArgs(p.handleWireguardPeerList, []string{"SERVER"}),
		},
		{
			Name:      "create-client",
			ArgsUsage: "SERVER NAME",
			Usage:     "create a new vpn client for a server",
			Action:    requireArgs(p.handleWireguardPeerCreate, []string{"SERVER", "NAME"}),
		},
		{
			Name:      "client-config",
			ArgsUsage: "SERVER NAME",
			Usage:     "get the config for a vpn client for a server",
			// Flags: []cli.Flag{
			// 	cli.BoolFlag{
			// 		Name:  "qr",
			// 		Usage: "output a qr code for the config",
			// 	},
			// },
			Action: requireArgs(p.handleWireguardPeerConfig, []string{"SERVER", "NAME"}),
		},
	}
}

func (p *Plural) handleWireguardServerList(c *cli.Context) error {
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	//TODO: check if wireguard is installed here and return an error if not
	//TODO: use namespace from wireguard installation rather than hardcoding
	servers, err := vpn.ListServers(p.Kube, conf.Namespace("wireguard"))
	if err != nil {
		return err
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
	server := c.Args().Get(0)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	//TODO: check if wireguard is installed here and return an error if not
	//TODO: use namespace from wireguard installation rather than hardcoding
	peers, err := vpn.ListPeers(p.Kube, conf.Namespace("wireguard"))
	if err != nil {
		return err
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
	serverName := c.Args().Get(0)
	name := c.Args().Get(1)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	//TODO: check if wireguard is installed here and return an error if not
	//TODO: use namespace from wireguard installation rather than hardcoding
	server, err := vpn.GetServer(p.Kube, conf.Namespace("wireguard"), serverName)
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Address", "Config Secret", "Public Key", "Ready"})
	table.Append([]string{
		peer.Name,
		peer.Spec.Address,
		peer.Status.ConfigRef.Name,
		peer.Spec.PublicKey,
		strconv.FormatBool(peer.Status.Ready),
	})
	table.Render()
	return nil
}

func (p *Plural) handleWireguardPeerConfig(c *cli.Context) error {
	serverName := c.Args().Get(0)
	name := c.Args().Get(1)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	//TODO: check if wireguard is installed here and return an error if not
	//TODO: use namespace from wireguard installation rather than hardcoding
	server, err := vpn.GetServer(p.Kube, conf.Namespace("wireguard"), serverName)
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

	//TODO: handle file download. need flag to specify file name/location
	//TODO: handle qr code generation. use https://github.com/yeqown/go-qrcode
	utils.Highlight(string(peerConfig))
	return nil
}

//TODO: handle wireguard peer delete

// func (p *Plural) checkIfVPNExists() error {
// 	conf := config.Read()
// 	if err := p.InitKube(); err != nil {
// 		return err
// 	}
// }
