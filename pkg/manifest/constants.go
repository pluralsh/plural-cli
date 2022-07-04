package manifest

type SmtpService struct {
	Server string
	Port   int
}

var smtpConfig = map[string]*SmtpService{
	"sendgrid":  {Server: "smtp.sendgrid.net", Port: 587},
	"mailchimp": {Server: "smtp.mandrillapp.com", Port: 587},
	"mandrill":  {Server: "smtp.mandrillapp.com", Port: 587},
	"mailgun":   {Server: "smtp.mailgun.org", Port: 465},
}
