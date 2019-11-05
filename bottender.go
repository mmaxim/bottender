package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/kballard/go-shellquote"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type Options struct {
	KeybaseLocation string
	Home            string
	Announcement    string
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
	var randomExtendedBody = fmt.Sprintf(`Type a cocktail ingredient (such as gin, vermouth, lime) and let the Bottender pick a random drink with a twist. Examples:
  %s
!dkrandom gin
!dkrandom elderflower%s`, mdQuotes, mdQuotes)
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
					{
						Name:        "dkrandom",
						Description: "Pick a random drink from an ingredient",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!dkrandom*
Pick a random drink`,
							DesktopBody: randomExtendedBody,
							MobileBody:  randomExtendedBody,
						},
					},
					{
						Name:        "dkaddrecipe",
						Description: "Add a new drink recipe",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!dkaddrecipe*
Add a new recipe`,
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

func (s *BotServer) handleRandom(cmd string, convID string) {
	terms := strings.Split(strings.Trim(cmd, " "), " ")
	var query *string
	if len(terms) >= 2 {
		query = new(string)
		*query = strings.Join(terms[1:], " ")
	}
	drinks, err := s.db.Random(query, 10)
	switch err {
	case nil:
		if len(drinks) == 0 {
			if _, err := s.kbc.SendMessageByConvID(convID, "no drinks found"); err != nil {
				s.debug("handleRandom: failed to send drink message: %s", err)
			}
		} else {
			if _, err := s.kbc.SendMessageByConvID(convID, "I've selected a few of my favorites, but let's let chance decide which one to make"); err != nil {
				s.debug("handleRandom: failed to send drink message: %s", err)
			}
			msg := fmt.Sprintf("/flip %s", strings.Join(DrinkNames(drinks), ", "))
			if _, err := s.kbc.SendMessageByConvID(convID, msg); err != nil {
				s.debug("handleRandom: failed to send drink message: %s", err)
			}
		}
	default:
		s.debug("handleDesc: misc error: %s", err)
	}
}

type ingredientFlags []string

func (i *ingredientFlags) String() string {
	return ""
}

func (i *ingredientFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (s *BotServer) handleAddRecipe(cmd string, sender string, convID string) {
	toks, err := shellquote.Split(cmd)
	if err != nil {
		if _, err := s.kbc.SendMessageByConvID(convID, "failed to split command"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	var ings ingredientFlags
	var name, mixing, serving, glass, notes string
	flags := flag.NewFlagSet(toks[0], flag.ContinueOnError)
	flags.Var(&ings, "ingredient", "ingredients")
	flags.StringVar(&mixing, "mixing", "", "mixing")
	flags.StringVar(&serving, "serving", "", "serving")
	flags.StringVar(&glass, "glass", "", "glass")
	flags.StringVar(&notes, "notes", "", "notes")
	s.debug("toks[1:]: %s", toks[1:])
	if err := flags.Parse(toks[1:]); err != nil {
		if _, err := s.kbc.SendMessageByConvID(convID, "failed to parse command"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	args := flags.Args()
	if len(args) != 1 {
		if _, err := s.kbc.SendMessageByConvID(convID, "must specify a drink name"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	name = args[0]
	if len(mixing) == 0 || len(serving) == 0 || len(glass) == 0 {
		if _, err := s.kbc.SendMessageByConvID(convID, "must specify all aspects of drink"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	// get all ingredients
	var ingredients []DrinkIngredient
	for _, ing := range ings {
		parts := strings.Split(ing, ",")
		if len(parts) != 2 {
			if _, err := s.kbc.SendMessageByConvID(convID, fmt.Sprintf("invalid ingredient: %s", ing)); err != nil {
				s.debug("handleAddRecipe: failed to send error message: %s", err)
			}
			return
		}
		amount, err := strconv.ParseInt(parts[1], 0, 0)
		if err != nil {
			if _, err := s.kbc.SendMessageByConvID(convID,
				fmt.Sprintf("invalid ingredient amount: %s", parts[1])); err != nil {
				s.debug("handleAddRecipe: failed to send error message: %s", err)
			}
			return
		}
		ingredient, err := s.db.DescribeIngredient(parts[0])
		if err != nil {
			if _, err := s.kbc.SendMessageByConvID(convID,
				fmt.Sprintf("failed to describe ingredient: %s", parts[0])); err != nil {
				s.debug("handleAddRecipe: failed to send error message: %s", err)
			}
			return
		}
		ingredients = append(ingredients, DrinkIngredient{
			Ingredient: ingredient,
			Amount:     int(amount),
		})
	}
	if err := s.db.AddRecipe(name, mixing, glass, serving, notes, ingredients); err != nil {
		s.debug("handleAddRecipe: failed to add recipe: %s", err)
		if _, err := s.kbc.SendMessageByConvID(convID, "failed to add recipe"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	if _, err := s.kbc.SendMessageByConvID(convID, "Success!"); err != nil {
		s.debug("handleAddRecipe: failed to send success message: %s", err)
	}
	if _, err := s.kbc.SendMessageByConvID(convID, fmt.Sprintf("!dkdesc %s", name)); err != nil {
		s.debug("handleAddRecipe: failed to send success message: %s", err)
	}
	if _, err := s.kbc.Broadcast(fmt.Sprintf("New recipe added by @%s: %s!", sender, name)); err != nil {
		s.debug("handleAddRecipe: failed to broadcast: %s", err)
	}
	if _, err := s.kbc.Broadcast(fmt.Sprintf("!dkdesc %s", name)); err != nil {
		s.debug("handleAddRecipe: failed to broadcast: %s", err)
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
	case strings.HasPrefix(msg.Content.Text.Body, "!dkrandom"):
		s.handleRandom(msg.Content.Text.Body, msg.ConvID)
	case strings.HasPrefix(msg.Content.Text.Body, "!dkaddrecipe"):
		s.handleAddRecipe(msg.Content.Text.Body, msg.Sender.Username, msg.ConvID)
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
	if s.opts.Announcement != "" {
		if _, err := s.kbc.SendMessageByTlfName(s.opts.Announcement, "I'm running."); err != nil {
			s.debug("failed to announce self: %s", err)
			return err
		}
	}
	sub, err := s.kbc.ListenForNewTextMessages()
	if err != nil {
		return err
	}
	s.debug("startup success, listening for messages...")
	for {
		msg, err := sub.Read()
		if err != nil {
			s.debug("Read() error: %s", err.Error())
			continue
		}
		s.handleCommand(msg.Message)
	}
}

func main() {
	rc := mainInner()
	os.Exit(rc)
}

func mainInner() int {
	var opts Options
	var dsn string

	flag.StringVar(&opts.KeybaseLocation, "keybase", "keybase", "keybase command")
	flag.StringVar(&opts.Home, "home", "", "Home directory")
	flag.StringVar(&opts.Announcement, "announcement", os.Getenv("BOT_ANNOUNCEMENT"),
		"Where to announce we are running")
	flag.StringVar(&dsn, "dsn", os.Getenv("BOT_DSN"), "Drink database DSN")
	flag.Parse()
	if len(dsn) == 0 {
		fmt.Printf("must specify a MySQL DSN for drink database\n")
		return 3
	}
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
