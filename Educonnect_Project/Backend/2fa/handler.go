package twofa

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/ipt-9/EduConnect/DB"
)

func generateCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d", int(b[0])<<8|int(b[1])<<4|int(b[2])%1000000)
}

func Send2FACode(userID uint64, recipientEmail string) error {
	code := generateCode()
	expiresAt := time.Now().Add(5 * time.Minute)

	err := DB.Store2FACode(userID, code, expiresAt)
	if err != nil {
		return err
	}

	// SMTP-Konfiguration aus .env
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpName := os.Getenv("SMTP_NAME")

	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	// Betreff und Inhalt der Mail
	subject := "Subject: Dein 2FA-Code\n"
	from := fmt.Sprintf("From: %s <%s>\n", smtpName, smtpEmail)
	body := fmt.Sprintf("Hallo,\n\nDein 2FA-Code für %s lautet: %s\nGültig für 5 Minuten.\n\nViele Grüße,\n%s", smtpName, code, smtpName)

	// Vollständige Nachricht
	msg := []byte(from + subject + "MIME-Version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n" + body)

	// Mail senden
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, []string{recipientEmail}, msg)
	if err != nil {
		log.Printf("❌ Fehler beim Senden der Mail: %v", err)
		return err
	}

	log.Printf("📧 2FA-Code an %s gesendet", recipientEmail)
	return nil
}
