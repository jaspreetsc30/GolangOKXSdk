package common

type Auth struct {
	ApiKey     string
	SecretKey  string
	Passphrase string
	DebugMode  bool
}

func NewAuth( apiKey, secretKey, passphrase string, debugMode bool) Auth {
	return Auth{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		DebugMode:  debugMode,
	}
}

func (a Auth) Signature(method, path, body string, isUnix bool) *Signature {
	return &Signature{
		Key:    a.SecretKey,
		Method: method,
		Path:   path,
		Body:   body,
		IsUnix: isUnix,
	}
}
