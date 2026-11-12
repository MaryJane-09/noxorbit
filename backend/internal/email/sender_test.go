package email

import (
	"bytes"
	"mime"
    "mime/multipart"
    "mime/quotedprintable"
	"io"
	"net/mail"
	"strings"
	"testing"
)

func TestEmailNotificationsHaveCleanSubjectsAndIndependentIDs(t *testing.T) {
	sender := EmailSender{GmailAddress: "noxorbit@example.com"}
	details := codeEmail{
		Heading:   "Your password reset code",
		Name:      "Sample",
		Intro:     "Use the code below to reset your password.",
		Code:      "a7bc34",
		CodeLabel: "YOUR PASSWORD RESET CODE",
		Minutes:   5,
	}
	first, err := sender.buildCodeMessage("reader@example.com", "Reset your Noxorbit password", details)
	if err != nil {
		t.Fatal(err)
	}
	second, err := sender.buildCodeMessage("reader@example.com", "Reset your Noxorbit password", details)
	if err != nil {
		t.Fatal(err)
	}
	parse := func(raw []byte) *mail.Message {
		t.Helper()
		m, err := mail.ReadMessage(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	m1, m2 := parse(first), parse(second)
	subject1, err := new(mime.WordDecoder).DecodeHeader(m1.Header.Get("Subject"))
	if err != nil {
		t.Fatal(err)
	}
	subject2, err := new(mime.WordDecoder).DecodeHeader(m2.Header.Get("Subject"))
	if err != nil {
		t.Fatal(err)
	}
	if subject1 != "Reset your Noxorbit password" || subject2 != subject1 {
		t.Fatalf("subjects should be clean and stable, got %q and %q", subject1, subject2)
	}
	if strings.Contains(subject1, "[NX-") {
		t.Fatal("visible subject must not include a random reference")
	}
	if strings.Contains(subject1, details.Code) || strings.Contains(subject2, details.Code) {
		t.Fatal("one-time code must not appear in the subject")
	}
	if m1.Header.Get("Message-ID") == "" || m1.Header.Get("Message-ID") == m2.Header.Get("Message-ID") {
		t.Fatal("each email must have its own Message-ID")
	}
	if m1.Header.Get("References") != "" || m1.Header.Get("In-Reply-To") != "" {
		t.Fatal("notifications must not reference an existing email thread")
	}
	if m1.Header.Get("Date") == "" {
		t.Fatal("missing Date header")
	}

	// Verification and reset messages keep distinct, descriptive subjects.
	verification, err := sender.buildCodeMessage("reader@example.com", "Verify your Noxorbit account", details)
	if err != nil {
		t.Fatal(err)
	}
	verificationMail := parse(verification)
	verificationSubject, err := new(mime.WordDecoder).DecodeHeader(verificationMail.Header.Get("Subject"))
	if err != nil {
		t.Fatal(err)
	}
	if verificationSubject != "Verify your Noxorbit account" || verificationSubject == subject1 {
		t.Fatalf("registration and reset subjects should differ, got %q and %q", verificationSubject, subject1)
	}
}

func TestEmailTemplateLabelsMatchPurpose(t *testing.T) {
	reset := codeEmail{
		Heading: "Your password reset code",
		Name: "Sample", Intro: "Reset your password",
		Code: "abc234", CodeLabel: "YOUR PASSWORD RESET CODE",
		Minutes: 5,
	}
	if !strings.Contains(reset.text(), reset.CodeLabel) {
		t.Fatal("plain-text email is missing the reset label")
	}
	var html bytes.Buffer
	if err := codeTemplate.Execute(&html, reset); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), reset.CodeLabel) {
		t.Fatal("HTML email is missing the reset label")
	}
	if strings.Contains(html.String(), "YOUR VERIFICATION CODE") {
		t.Fatal("reset email should not have a generic verification label")
	}
}

// Decode MIME transfer encoding before validating rendered content. Searching
// the raw multipart body is unreliable because quoted-printable may wrap or
// encode characters such as = and & across lines.
func invitationParts(t *testing.T, raw []byte) (string, string, mail.Header) {
    t.Helper()
    msg, err := mail.ReadMessage(bytes.NewReader(raw))
    if err != nil { t.Fatal(err) }
    contentType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
    if err != nil { t.Fatal(err) }
    if contentType != "multipart/alternative" {
        t.Fatalf("expected multipart/alternative, got %q", contentType)
    }
    reader := multipart.NewReader(msg.Body, params["boundary"])
    parts := make(map[string]string)
    for {
        part, err := reader.NextPart()
        if err == io.EOF { break }
        if err != nil { t.Fatal(err) }
        mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
        if err != nil { t.Fatal(err) }
        if _, exists := parts[mediaType]; exists {
            t.Fatalf("duplicate MIME part: %s", mediaType)
        }
        var decoded io.Reader = part
        if strings.EqualFold(part.Header.Get("Content-Transfer-Encoding"), "quoted-printable") {
            decoded = quotedprintable.NewReader(part)
        }
        body, err := io.ReadAll(decoded)
        if err != nil { t.Fatal(err) }
        parts[mediaType] = string(body)
    }
    return parts["text/plain"], parts["text/html"], msg.Header
}

func TestWorkspaceInvitationContainsHTMLAndPlainText(t *testing.T) {
    sender := EmailSender{GmailAddress: "noxorbit@example.com"}
    link := "https://app.example.com/#invite/sample-token"
    raw, err := sender.buildInvitationMessage("recipient@example.com", "Design & Research", link)
    if err != nil { t.Fatal(err) }
    plain, html, headers := invitationParts(t, raw)
    if plain == "" || html == "" {
        t.Fatal("invitation must contain both text/plain and text/html parts")
    }
    if !strings.Contains(html, "Accept invitation") || !strings.Contains(html, "Design &amp; Research") {
        t.Fatal("HTML part should show branded CTA and escaped workspace name")
    }
    if !strings.Contains(plain, "Design & Research") {
        t.Fatal("plain text should contain the readable workspace name")
    }
    if !strings.Contains(plain, link) || !strings.Contains(html, link) {
        t.Fatal("both alternatives must contain a working invitation link")
    }
    if strings.Contains(headers.Get("Subject"), "sample-token") {
        t.Fatal("token must not appear in subject")
    }
}

func TestWorkspaceInvitationEscapesUntrustedContent(t *testing.T) {
    sender := EmailSender{GmailAddress: "noxorbit@example.com"}
    raw, err := sender.buildInvitationMessage("recipient@example.com", "<script>alert(1)</script>", "https://app.example.com/#invite/token")
    if err != nil { t.Fatal(err) }
    plain, html, _ := invitationParts(t, raw)
    if strings.Contains(html, "<script>") || strings.Contains(html, "</script>") {
        t.Fatal("HTML alternative must escape untrusted workspace content")
    }
    if !strings.Contains(html, "&lt;script&gt;") {
        t.Fatal("HTML alternative must contain escaped untrusted content")
    }
    if !strings.Contains(plain, "<script>alert(1)</script>") {
        t.Fatal("plain-text alternative should retain readable workspace name")
    }
}
