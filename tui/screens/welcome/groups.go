package welcome

import "github.com/pluralsh/plural-cli/tui/navigation"

type groupID uint8

const (
	groupDeployments groupID = iota
	groupAccess
	groupDiagnose
	groupHelp
)

type group struct {
	id       groupID
	number   string
	shortcut string
	title    string
	blurb    string
	route    navigation.Route // empty for Help stub
}

func welcomeGroups() []group {
	return []group{
		{id: groupDeployments, number: "1", shortcut: "d", title: "CD / Deployments", blurb: "clusters · services · repos", route: navigation.Deployments},
		{id: groupAccess, number: "2", shortcut: "a", title: "Access", blurb: "login · profiles · Console", route: navigation.Access},
		{id: groupDiagnose, number: "3", shortcut: "g", title: "Diagnose", blurb: "local context · checks", route: navigation.Diagnostics},
		{id: groupHelp, number: "4", shortcut: "?", title: "Help", blurb: "shortcuts · about"},
	}
}
