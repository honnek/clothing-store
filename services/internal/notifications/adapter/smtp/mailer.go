// Package smtp отправляет письма через SMTP без аутентификации: в dev это Mailpit,
// в проде на его месте окажется релей внутри периметра.
package smtp

import (
	"context"
	"fmt"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

type Mailer struct {
	addr string
	from string
}

func New(addr, from string) *Mailer {
	return &Mailer{addr: addr, from: from}
}

func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	msg := m.compose(to, subject, body)

	errc := make(chan error, 1)
	go func() {
		errc <- smtp.SendMail(m.addr, nil, m.from, []string{to}, msg)
	}()

	// net/smtp не умеет context, поэтому отмену обслуживаем снаружи: горутина
	// дописывает в буферизованный канал и уходит сама.
	select {
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("smtp %s: %w", m.addr, err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Mailer) compose(to, subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", m.from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
