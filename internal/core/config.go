package core

type config struct {
	Hostname string
	Port     int

	ImageCacheDirectory string
	ImageCacheURL       string

	Amazoncojp struct {
		SessionToken string
		SessionID    string
	}

	Amazoncom struct {
		SessionToken string
		SessionID    string
	}

	Pixiv struct {
		PHPSessionID string
		DeviceToken  string
	}

	Twitter struct {
		AuthToken string
	}
}

var Config config
