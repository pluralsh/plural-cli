package welcome

import "github.com/pluralsh/plural-cli/tui/navigation"

type groupID uint8

const (
	groupUp groupID = iota
	groupDeployments
	groupAccess
	groupDiagnose
	groupAI
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
		{id: groupUp, number: "1", shortcut: "u", title: "Up", blurb: "bootstrap · management cluster", route: navigation.Up},
		{id: groupDeployments, number: "2", shortcut: "d", title: "CD / Deployments", blurb: "clusters · services · repos", route: navigation.Deployments},
		{id: groupAccess, number: "3", shortcut: "a", title: "Access", blurb: "login · profiles · Console", route: navigation.Access},
		{id: groupDiagnose, number: "4", shortcut: "g", title: "Diagnose", blurb: "local context · checks", route: navigation.Diagnostics},
		{id: groupAI, number: "5", shortcut: "i", title: "AI", blurb: "agents · workbenches", route: navigation.AI},
		{id: groupHelp, number: "6", shortcut: "?", title: "Help", blurb: "shortcuts · about"},
	}
}
