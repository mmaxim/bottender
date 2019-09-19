package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type Options struct {
	KeybaseLocation string
	Home            string
	Owner           string
}

type BotServer struct {
	opts Options
	kbc  *kbchat.API
	db   *DrinkDB
}

func NewBotServer(opts Options, db *DrinkDB) *BotServer {
	return &BotServer{
		opts: opts,
		db:   db,
	}
}

func (s *BotServer) debug(msg string, args ...interface{}) {
	fmt.Printf("BotServer: "+msg+"\n", args...)
}

func (s *BotServer) getCommand() string {
	return fmt.Sprintf("hi%s", s.kbc.GetUsername())
}

func (s *BotServer) getCommandBang() string {
	return "!" + s.getCommand()
}

func (s *BotServer) makeAdvertisement() kbchat.Advertisement {
	var descExtendedBody = fmt.Sprintf(`Type a drink name and get the recipe of the closest match. Examples:
	%s
  !dkdesc martini
  !dkdesc parisian daiquiri%s`, mdQuotes, mdQuotes)
	return kbchat.Advertisement{
		Alias: "Bartender Bot",
		Advertisements: []chat1.AdvertiseCommandAPIParam{
			{
				Typ: "public",
				Commands: []chat1.UserBotCommandInput{
					{
						Name:        "dkdesc",
						Description: "Describe a drink",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!dkdesc*
Describe a drink`,
							DesktopBody: descExtendedBody,
							MobileBody:  descExtendedBody,
						},
					},
				},
			},
		},
	}
}

func (s *BotServer) handleDesc(cmd string, convID string) {
	terms := strings.Split(cmd, " ")
	if len(terms) < 2 {
		if _, err := s.kbc.SendMessageByConvID(convID, "must specify a drink query"); err != nil {
			s.debug("handleDesc: failed to send error message: %s", err)
		}
		return
	}
	query := strings.Join(terms[1:], " ")
	drink, err := s.db.Describe(query)
	switch err {
	case nil:
		if _, err := s.kbc.SendMessageByConvID(convID, DisplayDrinkFull(drink)); err != nil {
			s.debug("handleDesc: failed to send drink message: %s", err)
		}
	case errDrinkNotFound:
		if _, err := s.kbc.SendMessageByConvID(convID, "no drinks found"); err != nil {
			s.debug("handleDesc: failed to send error message: %s", err)
		}
	default:
		s.debug("handleDesc: misc error: %s", err)
	}
}

func (s *BotServer) handleCommand(msg chat1.MsgSummary) {
	if msg.Content.Text == nil {
		s.debug("skipping non-text message")
		return
	}
	switch {
	case strings.HasPrefix(msg.Content.Text.Body, "!dkdesc"):
		s.handleDesc(msg.Content.Text.Body, msg.ConvID)
	default:
		s.debug("unknown command: %s", msg.Content.Text.Body)
	}
}

func (s *BotServer) Start() (err error) {
	if s.kbc, err = kbchat.Start(kbchat.RunOptions{
		KeybaseLocation: s.opts.KeybaseLocation,
		HomeDir:         s.opts.Home,
	}); err != nil {
		return err
	}
	if _, err := s.kbc.AdvertiseCommands(s.makeAdvertisement()); err != nil {
		s.debug("advertise error: %s", err)
		return err
	}
	if s.opts.Owner != "" {
		if _, err := s.kbc.SendMessageByTlfName(s.opts.Owner, "I'm running."); err != nil {
			s.debug("failed to announce self: %s", err)
			return err
		}
	}
	sub, err := s.kbc.ListenForNewTextMessages()
	if err != nil {
		return err
	}
	s.debug("startup success, listening for messages...")
	username := s.kbc.GetUsername()
	for {
		msg, err := sub.Read()
		if err != nil {
			s.debug("Read() error: %s", err.Error())
			continue
		}
		if msg.Message.Sender.Username != username {
			s.handleCommand(msg.Message)
		}
	}
}

func main() {
	rc := mainInner()
	os.Exit(rc)
}

func mainInner() int {
	var opts Options
	flag.StringVar(&opts.KeybaseLocation, "keybase", "keybase", "keybase command")
	flag.StringVar(&opts.Home, "home", "", "Home directory")
	flag.StringVar(&opts.Owner, "owner", "", "Owner of the bot")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Printf("must specify a MySQL DSN for drink database\n")
		return 3
	}
	dsn := args[0]
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("failed to connect to MySQL: %s\n", err)
		return 3
	}

	bs := NewBotServer(opts, NewDrinkDB(db))
	if err := bs.Start(); err != nil {
		fmt.Printf("error running chat loop: %s\n", err.Error())
	}
	return 0
}
