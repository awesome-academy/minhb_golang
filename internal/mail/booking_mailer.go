package mail

import (
	"bytes"
	"context"
	"html/template"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/wneessen/go-mail"

	"cinema-booking/internal/models"
	"cinema-booking/internal/utils"
)

const (
	bookingCreatedTemplate = "booking-created.html"
	bookingPaidTemplate    = "booking-paid.html"
	mailTimeLayout         = "Mon 02 Jan 2006 15:04"
	imageTimeLayout        = "2006-01-02-15h04"
	qrSize                 = 256
	implicitTLSPort        = 465
)

type Options struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type BookingMailer struct {
	client    *mail.Client
	from      string
	templates *template.Template
}

func NewBookingMailer(opts Options, templates *template.Template) (*BookingMailer, error) {
	clientOpts := []mail.Option{mail.WithPort(opts.Port), mail.WithTLSPolicy(mail.TLSOpportunistic)}
	if opts.Port == implicitTLSPort {
		clientOpts = append(clientOpts, mail.WithSSL())
	}
	if opts.Username != "" {
		clientOpts = append(clientOpts, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), mail.WithUsername(opts.Username), mail.WithPassword(opts.Password))
	}
	client, err := mail.NewClient(opts.Host, clientOpts...)
	if err != nil {
		return nil, err
	}
	return &BookingMailer{client: client, from: opts.From, templates: templates}, nil
}

func (m *BookingMailer) SendBookingCreated(ctx context.Context, booking *models.Booking) error {
	msg, err := m.newMessage(booking, "Your booking · "+booking.Showtime.Movie.Title, bookingCreatedTemplate)
	if err != nil {
		return err
	}
	return m.client.DialAndSendWithContext(ctx, msg)
}

func (m *BookingMailer) SendBookingPaid(ctx context.Context, booking *models.Booking) error {
	msg, err := m.newMessage(booking, "Your tickets · "+booking.Showtime.Movie.Title, bookingPaidTemplate)
	if err != nil {
		return err
	}
	if err := embedTicketQRCodes(msg, booking); err != nil {
		return err
	}
	return m.client.DialAndSendWithContext(ctx, msg)
}

func (m *BookingMailer) newMessage(booking *models.Booking, subject, templateName string) (*mail.Msg, error) {
	var body bytes.Buffer
	if err := m.templates.ExecuteTemplate(&body, templateName, newBookingView(booking)); err != nil {
		return nil, err
	}
	msg := mail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return nil, err
	}
	if err := msg.AddToFormat(booking.User.FullName, booking.User.Email); err != nil {
		return nil, err
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body.String())
	return msg, nil
}

func embedTicketQRCodes(msg *mail.Msg, booking *models.Booking) error {
	for _, ticket := range booking.Tickets {
		png, err := qrcode.Encode(ticket.QRCode, qrcode.Medium, qrSize)
		if err != nil {
			return err
		}
		msg.EmbedReadSeeker(ticketImage(booking.Showtime, ticket), bytes.NewReader(png))
	}
	return nil
}

type bookingView struct {
	FullName   string
	Code       string
	MovieTitle string
	Format     string
	Theater    string
	Address    string
	Room       string
	StartsAt   string
	ExpiresAt  string
	PaidAt     string
	Total      string
	Currency   string
	Tickets    []ticketView
}

type ticketView struct {
	Seat  string
	Price string
	Image string
}

func newBookingView(booking *models.Booking) bookingView {
	showtime := booking.Showtime
	tickets := make([]ticketView, 0, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		tickets = append(tickets, ticketView{
			Seat:  seatLabel(ticket.Seat),
			Price: ticket.Price.StringFixed(2),
			Image: ticketImage(showtime, ticket),
		})
	}
	return bookingView{
		FullName:   booking.User.FullName,
		Code:       booking.Code,
		MovieTitle: showtime.Movie.Title,
		Format:     string(showtime.Format),
		Theater:    showtime.Room.Theater.Name,
		Address:    showtime.Room.Theater.Address,
		Room:       showtime.Room.Name,
		StartsAt:   utils.FormatVN(showtime.StartsAt, mailTimeLayout),
		ExpiresAt:  formatOptional(booking.ExpiresAt),
		PaidAt:     formatOptional(booking.ConfirmedAt),
		Total:      booking.Total().StringFixed(2),
		Currency:   booking.Currency,
		Tickets:    tickets,
	}
}

func formatOptional(t *time.Time) string {
	if t == nil {
		return ""
	}
	return utils.FormatVN(*t, mailTimeLayout)
}

func ticketImage(showtime models.Showtime, ticket models.Ticket) string {
	return utils.Slugify(showtime.Movie.Title) + "_" + utils.FormatVN(showtime.StartsAt, imageTimeLayout) + "_" + seatLabel(ticket.Seat) + ".png"
}

func seatLabel(seat models.Seat) string {
	return seat.RowLabel + strconv.Itoa(int(seat.SeatNumber))
}
