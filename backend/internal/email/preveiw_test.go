package email

import (
	"bytes"
	"os"
	"testing"
)

// Writes preview.html so you can see the email in a browser without sending anything.
func TestWritePreview(t *testing.T) {
	var buf bytes.Buffer
	err := codeTemplate.Execute(&buf, codeEmail{
		Eyebrow:   "VERIFY YOUR EMAIL",
		Heading:   "Confirm your email address",
		Name:      "MaryJane",
		Intro:     "Use the code below to finish creating your Noxorbit account.",
		Code:      "k7Xp3Q",
		Minutes:   5,
		Footer:    "If you did not create a Noxorbit account, you can safely ignore this email.",
		Preheader: "Use your code to finish creating your Noxorbit account.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("preview.html", buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}