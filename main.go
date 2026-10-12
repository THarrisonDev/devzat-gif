package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	api "github.com/quackduck/devzat/devzatapi"
)

// Giphy's long-standing public beta key, used when GIF_API_KEY is unset.
// Fine for a small private server; set GIF_API_KEY for your own key.
const defaultGiphyKey = "sXpGFDGZs0Dv1mmNFvYaGUvYwKX0PWIh"

type config struct {
	provider string
	apiKey   string
	rating   string
	limit    int
}

func loadConfig() config {
	provider := strings.ToLower(getenv("GIF_PROVIDER", "giphy"))
	key := os.Getenv("GIF_API_KEY")
	if key == "" && provider == "giphy" {
		key = defaultGiphyKey
	}
	return config{
		provider: provider,
		apiKey:   key,
		rating:   getenv("GIF_RATING", "pg-13"),
		limit:    20,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

const usage = `Usage:
  devzat-gif <host:port>        Connect and serve the 'gif' command
  devzat-gif --print <args...>  Format one lookup and print it (no server)

In chat:
  gif <search terms>            Post a random matching GIF
  gif <https://image-url>       Post an image URL directly (no API key needed)

Configuration (environment):
  DEVZAT_ADDRESS   Server plugin-API address, e.g. localhost:5556 (arg overrides)
  DEVZAT_TOKEN     Plugin API token (fixed rpc.key or from 'grant')
  GIF_PROVIDER     giphy (default) or tenor
  GIF_API_KEY      Provider API key (Giphy defaults to its public beta key)
  GIF_RATING       Giphy content rating: g, pg, pg-13, r (default pg-13)
`

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage)
		return
	}
	if len(args) > 0 && args[0] == "--print" {
		runPrint(strings.Join(args[1:], " "))
		return
	}
	run()
}

func runPrint(args string) {
	reply, err := handle(loadConfig(), args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(reply)
}

func run() {
	address := os.Getenv("DEVZAT_ADDRESS")
	token := os.Getenv("DEVZAT_TOKEN")
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		address = os.Args[1]
	}
	if address == "" {
		fmt.Fprintln(os.Stderr, "no server address: pass it as the first argument or set DEVZAT_ADDRESS")
		os.Exit(1)
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "no token: set DEVZAT_TOKEN")
		os.Exit(1)
	}

	cfg := loadConfig()
	s, err := api.NewSession(address, token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating session:", err)
		os.Exit(1)
	}
	defer s.Close()

	err = s.RegisterCmd("gif", "<search terms> | <image url>", "Post a GIF (search or a direct image URL)", func(cmd api.CmdCall, e error) {
		if e != nil {
			return
		}
		reply, herr := handle(cfg, cmd.Args)
		if herr != nil {
			_ = s.SendMessage(api.Message{Room: cmd.Room, DMTo: cmd.From, Data: "gif: " + herr.Error()})
			return
		}
		if reply == "" {
			return
		}
		// Post as the invoking user, like a normal message.
		if err := s.SendMessage(api.Message{Room: cmd.Room, From: cmd.From, Data: reply}); err != nil {
			fmt.Fprintln(os.Stderr, "error sending message:", err)
		}
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error registering command:", err)
		os.Exit(1)
	}

	fmt.Printf("gif plugin connected to %s (provider: %s)\n", address, cfg.provider)
	if err := <-s.ErrorChan; err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func handle(cfg config, args string) (string, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return "", fmt.Errorf("usage: gif <search terms>  (or paste an image URL)")
	}
	if isURL(args) && !strings.ContainsAny(args, " \t\n") {
		return fmt.Sprintf("![gif](%s)", args), nil
	}

	results, err := search(cfg, args)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no GIFs found for %q", args)
	}
	r := results[rand.Intn(len(results))]
	return fmt.Sprintf("![%s](%s)\n_Powered by %s · %s_", sanitizeAlt(args), r.ImageURL, r.Provider, r.PageURL), nil
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func sanitizeAlt(s string) string {
	s = strings.NewReplacer("]", "", "[", "", "(", "", ")", "", "\n", " ", "\r", " ").Replace(s)
	if len(s) > 80 {
		s = s[:80] + "…"
	}
	return strings.TrimSpace(s)
}
