package manifest

type SmtpService struct {
	Server string
	Port   int
}

var smtpConfig = map[string]*SmtpService{
	"sendgrid": &SmtpService{Server: "smtp.sendgrid.net", Port: 587},
	"mailchimp": &SmtpService{Server: "smtp.mandrillapp.com", Port: 587},
	"mandrill": &SmtpService{Server: "smtp.mandrillapp.com", Port: 587},
	"mailgun": &SmtpService{Server: "smtp.mailgun.org", Port: 465},
}