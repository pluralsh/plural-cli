package api

import (
	"fmt"
	"strings"
)

const (
	PASSWORD     = "PASSWORD"
	PASSWORDLESS = "PASSWORDLESS"
)

const loginQuery = `
	mutation Login($email: String!, $pwd: String!) {
		login(email: $email, password: $pwd) { jwt }
	}
`

const loginMethodQuery = `
	query LoginMethod($email: String!) {
		loginMethod(email: $email) { loginMethod token }
	}
`

const pollLogin = `
	mutation Poll($token: String!) {
		loginToken(token: $token) { jwt }
	}
`

const impersonationQuery = `
	mutation Impersonate($email: String) {
		impersonateServiceAccount(email: $email) { jwt email }
	}
`

const createTokenQuery = `
	mutation {
		createToken { token }
	}
`

const listTokenQuery = `
	query {
		tokens(first: 3) {
			edges { node { token } }
		}
	}
`

const createUpgradeMut = `
	mutation Upgrade($name: String, $attributes: UpgradeAttributes!) {
		createUpgrade(name: $name, attributes: $attributes) {	id }
	} 
`

var listKeys = fmt.Sprintf(`
	query ListKeys($emails: [String]) {
		publicKeys(emails: $emails, first: 1000) {
			edges { node { ...PublicKeyFragment } }
		}
	}
	%s
`, PublicKeyFragment)

const createKey = `
	mutation Create($key: String!, $name: String!) {
		createPublicKey(attributes: {content: $key, name: $name}) { id }
	}
`

const deviceLogin = `
	mutation {
		deviceLogin { loginUrl deviceToken }
	}
`

const meQuery = `
	query {
		me { id email }
	}
`

var getEabCredential = fmt.Sprintf(`
	query Eab($cluster: String!, $provider: Provider!) {
		eabCredential(cluster: $cluster, provider: $provider) {
			...EabCredentialFragment
		}
	}
	%s
`, EabCredentialFragment)

const deleteEabCredential = `
	mutation Delete($cluster: String!, $provider: Provider!) {
		deleteEabKey(cluster: $cluster, provider: $provider) {
			id
		}
	}
`

const createEvent = `
	mutation Event($attrs: UserEventAttributes!) {
		createUserEvent(attributes: $attrs)
	}
`

type UpgradeAttributes struct {
	Message string
}

type UserEventAttributes struct {
	Event  string
	Data   string
	Status string
}

type DeviceLogin struct {
	LoginUrl    string
	DeviceToken string
}

type login struct {
	Login struct {
		Jwt string `json:"jwt"`
	}
	ImpersonateServiceAccount struct {
		Jwt   string `json:"jwt"`
		Email string `json:"email"`
	}
}

type createToken struct {
	CreateToken struct {
		Token string
	}
}

type listToken struct {
	Tokens struct {
		Edges []struct {
			Node struct {
				Token string
			}
		}
	}
}

type LoginMethod struct {
	LoginMethod string
	Token       string
}

type Me struct {
	Id    string
	Email string
}

func (client *Client) Me() (*Me, error) {
	var resp struct {
		Me *Me
	}

	req := client.Build(meQuery)
	err := client.Run(req, &resp)
	return resp.Me, err
}

func (client *Client) LoginMethod(email string) (*LoginMethod, error) {
	var resp struct {
		LoginMethod LoginMethod
	}
	req := client.Build(loginMethodQuery)
	req.Var("email", email)
	err := client.Run(req, &resp)
	return &resp.LoginMethod, err
}

func (client *Client) PollLoginToken(token string) (string, error) {
	var resp struct {
		LoginToken struct {
			Jwt string
		}
	}

	req := client.Build(pollLogin)
	req.Var("token", token)
	err := client.Run(req, &resp)
	return resp.LoginToken.Jwt, err
}

func (client *Client) DeviceLogin() (*DeviceLogin, error) {
	var resp struct {
		DeviceLogin *DeviceLogin
	}

	req := client.Build(deviceLogin)
	err := client.Run(req, &resp)
	return resp.DeviceLogin, err
}

func (client *Client) Login(email, pwd string) (string, error) {
	var resp login
	req := client.Build(loginQuery)
	req.Var("email", email)
	req.Var("pwd", pwd)
	err := client.Run(req, &resp)
	return resp.Login.Jwt, err
}

func (client *Client) ImpersonateServiceAccount(email string) (string, string, error) {
	var resp login
	req := client.Build(impersonationQuery)
	req.Var("email", email)
	err := client.Run(req, &resp)
	return resp.ImpersonateServiceAccount.Jwt, resp.ImpersonateServiceAccount.Email, err
}

func (client *Client) CreateAccessToken() (string, error) {
	var resp createToken
	req := client.Build(createTokenQuery)
	err := client.Run(req, &resp)
	return resp.CreateToken.Token, err
}

func (client *Client) GrabAccessToken() (string, error) {
	var resp listToken
	req := client.Build(listTokenQuery)
	err := client.Run(req, &resp)
	if err != nil {
		return "", err
	}
	if len(resp.Tokens.Edges) > 0 {
		return resp.Tokens.Edges[0].Node.Token, nil
	}

	return client.CreateAccessToken()
}

func (client *Client) CreateUpgrade(name string, message string) (id string, err error) {
	var resp struct {
		CreateUpgrade *Upgrade
	}

	req := client.Build(createUpgradeMut)
	req.Var("name", name)
	req.Var("attributes", UpgradeAttributes{Message: message})
	err = client.Run(req, &resp)
	if err == nil {
		id = resp.CreateUpgrade.Id
	}

	return
}

func (client *Client) ListKeys(emails []string) (keys []*PublicKey, err error) {
	var resp struct {
		PublicKeys struct {
			Edges []*PublicKeyEdge
		}
	}

	req := client.Build(listKeys)
	req.Var("emails", emails)
	err = client.Run(req, &resp)
	keys = []*PublicKey{}
	for _, edge := range resp.PublicKeys.Edges {
		keys = append(keys, edge.Node)
	}
	return
}

func (client *Client) CreateKey(name, content string) error {
	var resp struct {
		CreatePublicKey struct {
			Id string
		}
	}

	req := client.Build(createKey)
	req.Var("key", content)
	req.Var("name", name)
	return client.Run(req, &resp)
}

func (client *Client) GetEabCredential(cluster, provider string) (*EabCredential, error) {
	var resp struct {
		EabCredential *EabCredential
	}
	req := client.Build(getEabCredential)
	req.Var("cluster", cluster)
	req.Var("provider", toProvider(provider))
	err := client.Run(req, &resp)
	return resp.EabCredential, err
}

func (client *Client) DeleteEabCredential(cluster, provider string) error {
	var resp struct {
		DeleteEabKey struct {
			Id string
		}
	}

	req := client.Build(deleteEabCredential)
	req.Var("cluster", cluster)
	req.Var("provider", toProvider(provider))
	return client.Run(req, &resp)
}

func (client *Client) CreateEvent(event *UserEventAttributes) error {
	var resp struct {
		CreateUserEvent bool
	}

	req := client.Build(createEvent)
	req.Var("attrs", event)
	return client.Run(req, &resp)
}

func toProvider(prov string) string {
	if prov == "google" {
		return "GCP"
	}

	return strings.ToUpper(prov)
}
