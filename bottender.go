package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
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
  !bottender describe martini
  !bottender describe parisian daiquiri%s`, mdQuotes, mdQuotes)
	var randomExtendedBody = fmt.Sprintf(`Type a cocktail ingredient (such as gin, vermouth, lime) and let the Bottender pick a random drink with a twist. Examples:
  %s
!bottender random gin
!bottender random elderflower%s`, mdQuotes, mdQuotes)
	var addExtendedBody = fmt.Sprintf(`Submit a drink recipe for review. 
%s!bottender addrecipe <--ingredient "ingredient name",amount> [--ingredient...] <--serving "serving style"> <--mixing "mixing method"> <--glass "glass type> [--notes "notes] <drink name>%s
Example
%s!bottender addrecipe --ingredient bourbon,200 --ingredient "simple syrup",50 --ingredient 'aromatic bitters',2 --serving rocks --mixing stirred --glass rocks 'Old Fashioned'%s`, mdQuotes, mdQuotes, mdQuotes, mdQuotes)
	return kbchat.Advertisement{
		Alias: "Bartender",
		Advertisements: []chat1.AdvertiseCommandAPIParam{
			{
				Typ: "public",
				Commands: []chat1.UserBotCommandInput{
					{
						Name:        "bottender describe",
						Description: "Describe a drink",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!bottender describe*
Describe a drink`,
							DesktopBody: descExtendedBody,
							MobileBody:  descExtendedBody,
						},
					},
					{
						Name:        "bottender random",
						Description: "Pick a random drink from an ingredient",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!bottender random*
Pick a random drink`,
							DesktopBody: randomExtendedBody,
							MobileBody:  randomExtendedBody,
						},
					},
					{
						Name:        "bottender addrecipe",
						Description: "Add a new drink recipe",
						ExtendedDescription: &chat1.UserBotExtendedDescription{
							Title: `*!bottender addrecipe*
Add a new recipe`,
							DesktopBody: addExtendedBody,
							MobileBody:  addExtendedBody,
						},
					},
				},
			},
		},
	}
}

func (s *BotServer) handleDesc(cmd string, convID chat1.ConvIDStr) {
	terms := strings.Split(cmd, " ")
	if len(terms) < 3 {
		if _, err := s.kbc.SendMessageByConvID(convID, "must specify a drink query"); err != nil {
			s.debug("handleDesc: failed to send error message: %s", err)
		}
		return
	}
	query := strings.Join(terms[2:], " ")
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

func (s *BotServer) handleRandom(cmd string, convID chat1.ConvIDStr) {
	terms := strings.Split(strings.Trim(cmd, " "), " ")
	var query *string
	if len(terms) >= 3 {
		query = new(string)
		*query = strings.Join(terms[2:], " ")
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

func (s *BotServer) handleAddRecipe(cmd string, sender string, convID chat1.ConvIDStr) {
	toks, err := shellquote.Split(cmd)
	if err != nil {
		if _, err := s.kbc.SendMessageByConvID(convID, "failed to split command"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	if len(toks) < 3 {
		if _, err := s.kbc.SendMessageByConvID(convID, "must specify drink ingredients"); err != nil {
			s.debug("handleDesc: failed to send error message: %s", err)
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
	s.debug("toks[2:]: %s", toks[2:])
	if err := flags.Parse(toks[2:]); err != nil {
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
	if err := s.db.AddRecipe(name, mixing, glass, serving, notes, ingredients, sender); err != nil {
		s.debug("handleAddRecipe: failed to add recipe: %s", err)
		if _, err := s.kbc.SendMessageByConvID(convID, "failed to add recipe"); err != nil {
			s.debug("handleAddRecipe: failed to send error message: %s", err)
		}
		return
	}
	if _, err := s.kbc.SendMessageByConvID(convID, "Success!"); err != nil {
		s.debug("handleAddRecipe: failed to send success message: %s", err)
	}
	if _, err := s.kbc.SendMessageByConvID(convID, fmt.Sprintf("!bottender describe %s", name)); err != nil {
		s.debug("handleAddRecipe: failed to send success message: %s", err)
	}
	if _, err := s.kbc.Broadcast(fmt.Sprintf("New recipe added by @%s: %s!", sender, name)); err != nil {
		s.debug("handleAddRecipe: failed to broadcast: %s", err)
	}
	if _, err := s.kbc.Broadcast(fmt.Sprintf("!bottender describe %s", name)); err != nil {
		s.debug("handleAddRecipe: failed to broadcast: %s", err)
	}
}

func (s *BotServer) handleCommand(msg chat1.MsgSummary) {
	if msg.Content.Text == nil {
		s.debug("skipping non-text message")
		return
	}
	switch {
	case strings.HasPrefix(msg.Content.Text.Body, "!bottender describe"):
		s.handleDesc(msg.Content.Text.Body, msg.ConvID)
	case strings.HasPrefix(msg.Content.Text.Body, "!bottender random"):
		s.handleRandom(msg.Content.Text.Body, msg.ConvID)
	case strings.HasPrefix(msg.Content.Text.Body, "!bottender addrecipe"):
		s.handleAddRecipe(msg.Content.Text.Body, msg.Sender.Username, msg.ConvID)
	default:
		s.debug("unknown command: %s", msg.Content.Text.Body)
	}
}

func (s *BotServer) sendAnnouncement(announcement string, running string) (err error) {
	defer func() {
		if err == nil {
			s.debug("announcement success")
		}
	}()
	if _, err := s.kbc.SendMessageByConvID(chat1.ConvIDStr(announcement), running); err != nil {
		s.debug("failed to announce self as conv ID: %s", err)
	} else {
		return nil
	}
	if _, err := s.kbc.SendMessageByTlfName(announcement, running); err != nil {
		s.debug("failed to announce self as user: %s", err)
	} else {
		return nil
	}
	if _, err := s.kbc.SendMessageByTeamName(announcement, nil, running); err != nil {
		s.debug("failed to announce self as team: %s", err)
		return err
	} else {
		return nil
	}
}

func (s *BotServer) handleGet(w http.ResponseWriter, r *http.Request) {
	s.debug("handleGet: request received")
	fmt.Fprintf(w, "HELLO")
}

func (s *BotServer) Start() (err error) {
	if s.kbc, err = kbchat.Start(kbchat.RunOptions{
		KeybaseLocation: s.opts.KeybaseLocation,
		HomeDir:         s.opts.Home,
	}); err != nil {
		return err
	}
	// Start up HTTP interface
	http.HandleFunc("/", s.handleGet)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			s.debug("failed to start http server: %s", err)
			os.Exit(3)
		}
	}()

	if _, err := s.kbc.AdvertiseCommands(s.makeAdvertisement()); err != nil {
		s.debug("advertise error: %s", err)
		return err
	}
	if s.opts.Announcement != "" {
		if err := s.sendAnnouncement(s.opts.Announcement, "I'm running."); err != nil {
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
