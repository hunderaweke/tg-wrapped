package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type termAuth struct {
	reader *bufio.Reader
}

func (a termAuth) Phone(ctx context.Context) (string, error) {
	fmt.Print("Enter Phone (e.g. +1234567): ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

func (a termAuth) Password(ctx context.Context) (string, error) {
	fmt.Print("Enter 2FA Password (if any): ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

func (a termAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter Code: ")
	s, _ := a.reader.ReadString('\n')
	return strings.TrimSpace(s), nil
}

func (a termAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	fmt.Println("Telegram Terms of Service:")
	fmt.Println(tos.Text)
	fmt.Print("Do you accept the Terms of Service? (yes/no): ")
	s, _ := a.reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return nil
	default:
		return fmt.Errorf("terms of service not accepted")
	}
}

func (a termAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	fmt.Print("Enter First Name: ")
	fn, _ := a.reader.ReadString('\n')
	fmt.Print("Enter Last Name: ")
	ln, _ := a.reader.ReadString('\n')
	return auth.UserInfo{
		FirstName: strings.TrimSpace(fn),
		LastName:  strings.TrimSpace(ln),
	}, nil
}
