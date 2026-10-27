package email

import (
	"bytes"
	_ "embed"
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

	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
)

type EmailSender struct {
	GmailAddress     string
	GmailAppPassword string
	SMTPHost         string
	SMTPPort         string
}

//
//go:embed templates/code.html
var codeHTML string

var codeTemplate = template.Must(template.New("code").Parse(codeHTML))

type codeEmail struct {
	Eyebrow   string
	Heading   string
	Name      string
	Intro     string
	Code      string
	Minutes   int
	Footer    string
	Preheader string
}

func (c codeEmail) text() string {
	return fmt.Sprintf("NOXORBIT\n\n%s\n\nHello %s,\n\n%s\n\n    %s\n\n"+
		"This code expires in %d minutes. Enter it exactly as shown; it is case-sensitive.\n"+
		"Never share it with anyone. Noxorbit will never ask you for it.\n\n%s\n",
		c.Heading, c.Name, c.Intro, c.Code, c.Minutes, c.Footer)
}

func (s EmailSender) SendVerification(name, to, code string) error {
	return s.sendCode(to, "Verify your Noxorbit account", codeEmail{
		Eyebrow:   "VERIFY YOUR EMAIL",
		Heading:   "Confirm your email address",
		Name:      name,
		Intro:     "Use the code below to finish creating your Noxorbit account.",
		Code:      code,
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
		Minutes:   otp.Time,
		Footer:    "If you did not ask to reset your password, ignore this email. Your password has not been changed.",
		Preheader: "Use your code to reset your Noxorbit password.",
	})
}

func (s EmailSender) sendCode(to, subject string, c codeEmail) error {
	if strings.ContainsAny(to, "\r\n") {
		return errors.New("invalid recipient address")
	}
	c.Name = cleanName(c.Name)

	var htmlBody bytes.Buffer
	if err := codeTemplate.Execute(&htmlBody, c); err != nil {
		return fmt.Errorf("render email: %w", err)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := writePart(mw, "text/plain", c.text()); err != nil {
		return err
	}
	if err := writePart(mw, "text/html", htmlBody.String()); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	from := (&mail.Address{Name: "Noxorbit", Address: s.GmailAddress}).String()
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=" + mw.Boundary() + "\r\n" +
		"\r\n" + body.String()

	auth := smtp.PlainAuth("", s.GmailAddress, s.GmailAppPassword, s.SMTPHost)
	err := smtp.SendMail(s.SMTPHost+":"+s.SMTPPort, auth, s.GmailAddress, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
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
