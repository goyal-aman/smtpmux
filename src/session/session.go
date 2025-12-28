package session

import (
	"errors"
	"io"
	"log"
	netsmtp "net/smtp"
	"os"
	"strings"

	"os/exec"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/goyal-aman/mailmux/src/shared"
	"github.com/goyal-aman/mailmux/src/types"
	"github.com/hashicorp/go-plugin"
)

type Session struct {
	CurrentUser types.UserConfig
	From        string
	To          []string

	cfg types.Config
}

func NewSession(cfg types.Config) *Session {
	return &Session{cfg: cfg}
}

// Data handles the mail transmission
func (s *Session) Data(r io.Reader) error {
	mailData, _ := io.ReadAll(r)

	// 1. Setup go-plugin Client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"selector": &shared.SelectorPlugin{},
		},
		Cmd: exec.Command(s.CurrentUser.SelectorAlgoPath),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
	})
	defer client.Kill()

	// 2. Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Printf("Error connecting to plugin: %v", err)
		return err
	}

	// 3. Request the plugin
	raw, err := rpcClient.Dispense("selector")
	if err != nil {
		log.Printf("Error dispensing plugin: %v", err)
		return err
	}

	selector := raw.(shared.Selector)

	// 4. Call the plugin
	log.Println("Calling plugin selector...")
	selectedAddr, err := selector.Select(s.CurrentUser.Downstreams)
	if err != nil {
		log.Printf("Plugin selection failed: %v", err)
		return err
	}
	log.Printf("Plugin selected: %s", selectedAddr)

	// 5. Find the selected downstream credentials
	var selectedDS types.Downstream
	found := false
	for _, ds := range s.CurrentUser.Downstreams {
		if ds.Addr == selectedAddr {
			selectedDS = ds
			found = true
			break
		}
	}
	if !found {
		return errors.New("selected downstream not found in config")
	}

	if err := send(selectedDS.User, selectedDS.Pass, selectedDS.Addr, s.From, s.To, mailData); err != nil {
		return err
	}

	return nil

}

func send(user, pass, addr, from string, to []string, mailData []byte) error {
	var auth netsmtp.Auth
	useInsecureAuth := os.Getenv("USE_INSECURE_AUTH")
	if useInsecureAuth == "true" {
		auth = &insecurePlainAuth{"", user, pass, strings.Split(addr, ":")[0]}
	} else {
		auth = netsmtp.PlainAuth("", user, pass, strings.Split(addr, ":")[0])
	}

	// Use InsecurePlainAuth to allow sending credentials to downstream containers (which are not localhost) without TLS
	return netsmtp.SendMail(addr, auth, from, to, mailData)
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}

func (s *Session) Logout() error {
	return nil
}

// 1. The Session must tell the server what it can do
func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

// 2. The Session handles the actual verification
func (s *Session) Auth(mech string) (sasl.Server, error) {
	if mech == sasl.Plain {
		return sasl.NewPlainServer(func(identity, username, password string) error {
			if identity != "" && identity != username {
				return errors.New("identities not supported")
			}
			return s.AuthPlain(username, password)
		}), nil
	}
	return nil, errors.New("unsupported mechanism")
}

// 2. The Session handles the actual verification
// AuthPlain is the specific helper method for PLAIN auth
func (s *Session) AuthPlain(username, password string) error {
	for _, u := range s.cfg.Users {
		if u.Email == username && u.Password == password {
			s.CurrentUser = u // Ensure this is stored to prevent the 'index out of range' panic
			return nil
		}
	}
	return errors.New("Invalid credentials")
}

func (s *Session) AuthLogin(username, password string) error {
	return s.AuthPlain(username, password)
}
