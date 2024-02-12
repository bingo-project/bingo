package middleware

import (
	"log"

	"gopkg.in/telebot.v3"
)

// Recover from panic.
func Recover(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recover from panic: ", err)
			}
		}()

		return next(c) // continue execution chain
	}
}
