package email

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/jackc/pgx/v5/pgxpool"
)


//go:embed templates/code.html
var codeHTML string

var codeTemplate = template.Must(template.New("code").Parse(codeHTML))

type codeEmail struct {
	Eyebrow   string
	Heading   string
	Name      string
	Intro     string
	Code      string
	CodeLabel string
	Minutes   int
	Footer    string
	Preheader string
}

func (c codeEmail) text() string {
	return fmt.Sprintf("NOXORBIT\n\n%s\n\nHello %s,\n\n%s\n\n%s\n%s\n\n"+
		"Select and copy the code above, then paste it into NoxOrbit.\n"+
		"Valid for %d minutes. The code is case-sensitive.\n"+
		"Keep this code private; NoxOrbit will never ask you to share it.\n\n%s\n",
		c.Heading, c.Name, c.Intro, c.CodeLabel, c.Code, c.Minutes, c.Footer)
}

type EmailSender struct {
	pool             *pgxpool.Pool
	GmailAddress     string
	GmailAppPassword string
	SMTPHost         string
	SMTPPort         string
}

func NewRepository(pool *pgxpool.Pool) *EmailSender {
	return &EmailSender{
		pool: pool,
	}
}

func (s EmailSender) SendVerification(name, to, code string) error {
	return s.sendCode(to, "Verify your Noxorbit account", codeEmail{
		Eyebrow:   "VERIFY YOUR EMAIL",
		Heading:   "Confirm your email address",
		Name:      name,
		Intro:     "Use the code below to finish creating your Noxorbit account.",
		Code:      code,
		CodeLabel: "YOUR EMAIL VERIFICATION CODE",
		Minutes:   otp.Time,
		Footer:    "If you did not create a Noxorbit account, you can safely ignore this email.",
		Preheader: "Use your code to finish creating your Noxorbit account.",
	})
}

func (s EmailSender) SendPasswordReset(name, to, code string) error {
	return s.sendCode(to, "Reset your Noxorbit password", codeEmail{
		Eyebrow:   "RESET YOUR PASSWORD",
		Heading:   "Your password reset code",
		Name:      name,
		Intro:     "Use the code below to choose a new password for your Noxorbit account.",
		Code:      code,
		CodeLabel: "YOUR PASSWORD RESET CODE",
		Minutes:   otp.Time,
		Footer:    "If you did not ask to reset your password, ignore this email. Your password has not been changed.",
		Preheader: "Use your code to reset your Noxorbit password.",
	})
}

func (s EmailSender) sendCode(to, subject string, c codeEmail) error {
	msg, err := s.buildCodeMessage(to, subject, c)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", s.GmailAddress, s.GmailAppPassword, s.SMTPHost)
	if err := smtp.SendMail(s.SMTPHost+":"+s.SMTPPort, auth, s.GmailAddress, []string{to}, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// buildCodeMessage constructs an independent notification, not an email reply.
// Every email receives a unique non-secret reference in its subject, so Gmail
// does not mix old verification codes and password-reset notifications.
// The code itself is never included in the subject.
func (s EmailSender) buildCodeMessage(to, subject string, c codeEmail) ([]byte, error) {
	if strings.ContainsAny(to, "\r\n") {
		return nil, errors.New("invalid recipient address")
	}
	sender, err := mail.ParseAddress(s.GmailAddress)
	if err != nil || sender.Address != s.GmailAddress {
		return nil, errors.New("invalid sender address")
	}
	at := strings.LastIndexByte(sender.Address, '@')
	if at < 0 || at == len(sender.Address)-1 {
		return nil, errors.New("invalid sender domain")
	}

	// A random reference distinguishes this notification from every other
	// registration or reset email without exposing the one-time code.
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("generate email reference: %w", err)
	}
	reference := strings.ToUpper(hex.EncodeToString(nonce[:5]))
	uniqueSubject := subject + " [NX-" + reference + "]"
	messageID := "<noxorbit." + hex.EncodeToString(nonce[:]) + "@" + sender.Address[at+1:] + ">"

	c.Name = cleanName(c.Name)
	var htmlBody bytes.Buffer
	if err := codeTemplate.Execute(&htmlBody, c); err != nil {
		return nil, fmt.Errorf("render email: %w", err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := writePart(mw, "text/plain", c.text()); err != nil {
		return nil, err
	}
	if err := writePart(mw, "text/html", htmlBody.String()); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	from := (&mail.Address{Name: "Noxorbit", Address: sender.Address}).String()
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n" +
		"Message-ID: " + messageID + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("UTF-8", uniqueSubject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=" + mw.Boundary() + "\r\n" +
		"\r\n" + body.String()
	return []byte(msg), nil
}

func writePart(mw *multipart.Writer, contentType, content string) error {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", contentType+"; charset=UTF-8")
	h.Set("Content-Transfer-Encoding", "quoted-printable")
	part, err := mw.CreatePart(h)
	if err != nil {
		return err
	}
	qp := quotedprintable.NewWriter(part)
	if _, err := qp.Write([]byte(content)); err != nil {
		return err
	}
	return qp.Close()
}

func cleanName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	if r := []rune(name); len(r) > 40 {
		name = string(r[:40])
	}
	if name == "" {
		return "there"
	}
	return name
}
