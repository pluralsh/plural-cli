package services

import (
	"fmt"
	"strings"

	servicesbridge "github.com/pluralsh/plural-cli/pkg/bridge/services"
)

type actionKind uint8

const (
	actionKick actionKind = iota
	actionEdit
	actionClone
	actionTarball
	actionWorkbench
	actionDelete
	actionCreate
)

type detailAction struct {
	kind     actionKind
	shortcut string
	title    string
	blurb    string
	danger   bool
}

func detailActions() []detailAction {
	return []detailAction{
		{kind: actionKick, shortcut: "k", title: "Kick", blurb: "force sync now"},
		{kind: actionEdit, shortcut: "e", title: "Edit", blurb: "git ref · folder · config · version"},
		{kind: actionClone, shortcut: "c", title: "Clone", blurb: "onto another cluster"},
		{kind: actionTarball, shortcut: "t", title: "Tarball", blurb: "download locally"},
		{kind: actionWorkbench, shortcut: "m", title: "Template…", blurb: "liquid / tpl / lua workbench"},
		{kind: actionDelete, shortcut: "d", title: "Delete", blurb: "remove service", danger: true},
	}
}

type pendingOp struct {
	kind     actionKind
	title    string
	cli      string
	lines    []string
	danger   bool
	create   *servicesbridge.CreateInput
	update   *servicesbridge.UpdateInput
	clone    *servicesbridge.CloneInput
	tarball  string
	deleteID string
	kickID   string
}

func (m Model) kickPlan() pendingOp {
	d := m.detail
	cluster := clusterLabel(servicesbridge.Cluster{Handle: d.ClusterHandle, Name: d.ClusterName, ID: d.ClusterID})
	cli := "plural cd services kick " + d.ID
	if d.ClusterHandle != "" {
		cli = fmt.Sprintf("plural cd services kick @%s/%s", d.ClusterHandle, d.Name)
	}
	return pendingOp{
		kind:   actionKick,
		title:  "Force sync · " + d.Name,
		cli:    cli,
		kickID: d.ID,
		lines: []string{
			"Action     Kick / force sync",
			"Cluster    " + cluster,
			"Service    " + d.Name + " · " + d.Namespace,
			"Revision   " + loCoalesce(d.RevisionSHA, "—"),
			"Status     " + loCoalesce(d.Status, "—"),
		},
	}
}

func (m Model) deletePlan() pendingOp {
	d := m.detail
	return pendingOp{
		kind:     actionDelete,
		title:    "Delete · " + d.Name,
		cli:      "plural cd services delete " + d.ID,
		danger:   true,
		deleteID: d.ID,
		lines: []string{
			"Action     Delete service",
			"Service    " + d.Name + " · " + clusterLabel(servicesbridge.Cluster{Handle: d.ClusterHandle, Name: d.ClusterName}),
			"Note       Cluster workloads are not uninstalled automatically.",
		},
	}
}

func loCoalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
