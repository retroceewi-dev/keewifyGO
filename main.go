package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lrstanley/go-ytdlp"
)

type Entry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var (
	videoCache []Entry
	cacheMu    sync.RWMutex
)
var adminIds = []string{
	"1285018696951140487",
	"1403573321316040837",
	"287929568069554209",
	"1446991754476916779",
}

func main() {
	enverr := godotenv.Load()
	if enverr != nil {
		log.Fatal("Error loading .env file")
	}
	mainToken := os.Getenv("TOKEN")
	altToken := os.Getenv("ALT_TOKEN")
	// sess, err := discordgo.New("Bot " + altToken) // Test Bot
	print(mainToken)
	print(altToken)
	sess, err := discordgo.New("Bot " + mainToken) // Main Bot
	if err != nil {
		log.Fatal(err)
	}
	updateCache()
	sess.AddHandler(messagecreated)
	sess.AddHandler(memberjoined)
	sess.AddHandler(memberbanned)
	sess.AddHandler(reactionadd)
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("Bot online.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
func messagecreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == "1468498879896096852" {
		if strings.HasPrefix(strings.ToLower(m.Content), "!keewify") {
			fmt.Println("keewify")
			s.ChannelMessageSend(m.ChannelID, keewifytext(m.Content))
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!dihify") {
			fmt.Println("dihify")
			s.ChannelMessageSend(m.ChannelID, sentenceToDih(m.Content))
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!randomframe") {
			fmt.Println("randomframe")
			randomframe(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!gamble") {
			fmt.Println("gamble")
			gamble(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!supergamble") {
			fmt.Println("supergamble")
			_allow := false
			for _, str := range m.Member.Roles {
				if slices.Contains(adminIds, str) {
					_allow = true
					break
				}
			}
			if _allow {
				for range 10 {
					gamble(s, m)
				}
			}
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!getpfp") {
			fmt.Println("getpfp")
			getpfp(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!getemoji") {
			fmt.Println("getemoji")
			getemoji(s, m)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!getsticker") {
			fmt.Println("getsticker")
			getsticker(s, m)
		}
	}
	if m.ChannelID == "1471758642003837123" {
		if strings.HasPrefix(strings.ToLower(m.Content), "!latex") {
			fmt.Println("latex")
			getLatex(s, m, m.Content)
		}
		if strings.HasPrefix(strings.ToLower(m.Content), "!l") {
			fmt.Println("l")
			getLatex(s, m, m.Content)
		}
	}
}
func channelbyname(name string, g *discordgo.Guild) *discordgo.Channel {
	for _, chanl := range g.Channels {
		if chanl.Name == name {
			return chanl
		}

	}
	return nil
}
func memberjoined(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guild, err := s.State.Guild(m.GuildID)
	welcomemsgs := []string{
		// I use an array to avoid clutter. If it didn't lead to clutter, I would absolutely
		// have this be an inline message.
		"You're now a shatling ",
		"Hey lil twin, you're looking gurtilicious today! <:emoji_53:1467954916533207091> ",
		"Hey shatling! Keewi isn't gay, nor is she ginger. <:O_O:1462370057194831873>",
		"Welcome to the Keewiverse. We have been awaiting your arrival. <:gurt:1461601994857775282>",
		"mrrrrp mrow mrrp mrrp mrow meeoowwww mrrp mrrp meow mrrp purrrr",
	}
	if err != nil {
		log.Fatal(err)
	}
	welcome := channelbyname("welcome", guild)
	if welcome != nil {
		s.ChannelMessageSend(
			welcome.ID,
			"-#"+"<@"+m.User.ID+">"+"\n"+welcomemsgs[rand.Intn(len(welcomemsgs))]+"\n\nYou are member #"+strconv.Itoa(guild.MemberCount)+"! \n Make sure you get reactions roles from <#1283449236209270815>!")
	}
}
func memberbanned(s *discordgo.Session, m *discordgo.GuildBanAdd) {
	guild, err := s.State.Guild(m.GuildID)
	print(guild, err)
	if err != nil {
		log.Fatal(err)
	}
	if guild != nil {
		welcome := channelbyname("welcome", guild)
		s.ChannelMessageSend(welcome.ID,
			"<@"+m.User.ID+"> ("+m.User.Username+")"+" was banned! Cya!")
	}
}
func reactionadd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	blunderboard, err := s.State.Channel("1495212872417017897")
	msg, merr := s.ChannelMessage(m.ChannelID, m.MessageID)
	s.State.MaxMessageCount = 100
	if err != nil {
		log.Fatal(err)
	}
	if merr != nil {
		log.Fatal(merr)
	}
	allow := false
	count := 0
	if blunderboard != nil {
		if m.Emoji.ID == "1443027776771719329" {
			// count := 0
			for _, reaction := range msg.Reactions {
				if reaction.Emoji.APIName() == m.Emoji.APIName() {
					if reaction.Count >= 4 {
						count = reaction.Count
						allow = true
					}
				}
			}
		}
	}
	if allow {
		if count == 4 {
			embed := &discordgo.MessageEmbed{
				Title: "Blunder! x" + strconv.Itoa(count),
				Author: &discordgo.MessageEmbedAuthor{
					Name:    msg.Author.DisplayName(),
					IconURL: msg.Author.AvatarURL(strconv.Itoa(int(math.Pow(2, 10))))},
				Description: msg.Content + "\n\n[Jump to Message](" + fmt.Sprintf("https://discord.com/channels/%s/%s/%s", m.GuildID, msg.ChannelID, msg.ID) + ")",
				Color:       0xe74c3c,

				Footer: &discordgo.MessageEmbedFooter{
					Text: msg.ID},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://cdn.discordapp.com/emojis/1443027776771719329.webp"},
			}
			if len(msg.Attachments) != 0 {
				embed.Description += "\n\n[Jump to Attachment](" + msg.Attachments[0].URL + ")"
				embed.Image = &discordgo.MessageEmbedImage{
					URL: msg.Attachments[0].URL,
				}
			}
			send := true
			for _, imsg := range blunderboard.Messages {
				if !strings.Contains(imsg.Embeds[0].Footer.Text, msg.ID) {
					continue
				} else {
					send = true
					break
				}
			}
			if send {
				s.ChannelMessageSendEmbed(blunderboard.ID, embed)
			}
		} else if count >= 5 {
			for _, imsg := range blunderboard.Messages {
				if !strings.Contains(imsg.Embeds[0].Footer.Text, msg.ID) {
					continue
				} else {
					embed := &discordgo.MessageEmbed{
						Title: "Blunder! x" + strconv.Itoa(count),
						Author: &discordgo.MessageEmbedAuthor{
							Name:    msg.Author.DisplayName(),
							IconURL: msg.Author.AvatarURL(strconv.Itoa(int(math.Pow(2, 10))))},
						Description: msg.Content + "\n\n[Jump to Message](" + fmt.Sprintf("https://discord.com/channels/%s/%s/%s", msg.GuildID, msg.ChannelID, msg.ID) + ")",
						Color:       0xe74c3c,

						Footer: &discordgo.MessageEmbedFooter{
							Text: msg.ID},
						Thumbnail: &discordgo.MessageEmbedThumbnail{
							URL: "https://cdn.discordapp.com/emojis/1443027776771719329.webp"},
					}
					if len(msg.Attachments) != 0 {
						embed.Description += "\n\n[Jump to Attachment](" + msg.Attachments[0].URL + ")"
						embed.Image = &discordgo.MessageEmbedImage{
							URL: msg.Attachments[0].URL,
						}
					}
					s.ChannelMessageEditEmbed(blunderboard.ID, imsg.ID, embed)
					break
				}
			}
		}
	}
}
func isalnum(c string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(c)
}
func keewifytext(s string) string {
	ignored := [...]string{
		"as",
		"is",
		"was",
		"case",
		"a",
		"of",
		"the",
		"by",
		"has",
		" ",
		"  ",
		"   ",
		"'s",
		"'nt",
		"'re",
		"than",
		"it",
		"and",
		"my",
		"also",
		"in",
		"hey",
		"i",
	}
	punc := [...]string{
		",",
		".",
		";",
		":",
		"?",
		"!",
		"`",
		"```",
		"\"",
		"'",
		"[",
		"(",
		"{",
		"]",
		")",
		"}",
		"<",
		">",
		"_",
		"-",
		"=",
		"+",
		"#",
		"$",
		"%",
		"^",
		"&",
		"*",
		"\\",
		"/",
		"|",
		"...",
	}
	fmt.Println(s)
	sentence := strings.Split(s[8:], "\n")
	fmt.Println(sentence)
	temp := make([]string, 4)
	for _, word := range sentence {
		if word == "\n" {
			temp = append(temp, word)
		} else {
			temp = append(temp, strings.Split(word, " ")...)
		}
	}
	sentence = temp
	fmt.Println(sentence)
	retsentence := make([]string, 4)

	for _, i := range sentence {
		if i != "\n" && len(i) > 0 {
			w := string(strings.Replace(strings.Replace(strings.TrimSpace(i), "k", "kwi", -1), "K", "KWI", -1))
			for z := 0; z < 3; z++ {
				if len(w) > 0 {
					if slices.Contains(punc[:], string(w[len(w)-1])) {
						w = w[:len(w)-1]
					}
				}
			}
			if len(w) < 2 {
				if w != "i" && w != "a" {
					retsentence = append(retsentence, w+"wi")
				} else {
					retsentence = append(retsentence, w+"i")
				}
			}
			if len(w) > 1 {
				if w[len(w)-1] == 'y' && len(w) > 2 {
					if w[len(w)-3:len(w)-1] == "eey" {
						retsentence = append(retsentence, w[:len(w)-1]+"wi")
					} else if w[len(w)-2:len(w)-1] == "ey" {
						retsentence = append(retsentence, w+"wi")
					} else {
						if !(strings.ToLower(w[len(w)-2:len(w)-1]) == "e") && !(strings.ToLower(w[len(w)-2:len(w)-1]) == "o") {
							retsentence = append(retsentence, w[:len(w)-1]+"eewi")
						} else {
							if !(w == "boy") {
								retsentence = append(retsentence, w+"wi")
							} else {
								retsentence = append(retsentence, w[:len(w)-1]+"eewi") // The Syno Exception
							}
						}
					}
				} else {
					if !slices.Contains(ignored[:], strings.ToLower(w)) && !(strings.ToLower(w)[len(w)-2:len(w)-1] == "ed") && !slices.Contains(punc[:], w) && !(slices.Contains(ignored[:], strings.ToLower(w)[len(w)-2:len(w)-1])) && !(strings.ToLower(w)[len(w)-2:] == "wi") {
						if strings.ToLower(w)[len(w)-1] != 'w' {
							retsentence = append(retsentence, w+"wi")
						} else {
							retsentence = append(retsentence, w)
						}
					} else {
						retsentence = append(retsentence, w)
					}
				}
			}
			if len(strings.TrimSpace(i)) > 3 && slices.Contains(punc[:], string(strings.TrimSpace(i)[len(i)-1])) {
				if !(strings.TrimSpace(i)[len(i)-2:len(i)-1] == "``" || strings.TrimSpace(i)[len(i)-3:len(i)-1] == "```" || strings.TrimSpace(i)[len(i)-1] == '`' || strings.TrimSpace(i)[len(i)-3:len(i)-1] == "...") {
					retsentence = append(retsentence, string(strings.TrimSpace(i)[len(i)-1]))
				} else {
					if strings.TrimSpace(i)[len(i)-3:len(i)-1] == "```" || strings.TrimSpace(i)[len(i)-3:len(i)-1] == "..." {
						retsentence = append(retsentence, strings.Replace(i[len(i)-3:len(i)-1], " ", "", -1))
					} else {
						if string(strings.TrimSpace(i))[len(i)-1] == '`' {
							retsentence = append(retsentence, string(strings.TrimSpace(i)[len(i)-1]))
						}
					}
				}
			}
		} else {
			retsentence = append(retsentence, "\n")
		}
	}

	return strings.Join(retsentence, " ")
}
func sentenceToDih(s string) string {
	fmt.Println(s)
	sentence := strings.Split(s[7:], "\n")
	fmt.Println(sentence)
	temp := make([]string, 0)
	for _, word := range sentence {
		if word == "\n" {
			temp = append(temp, word)
		} else {
			temp = append(temp, strings.Split(word, " ")...)
		}
	}
	sentence = temp
	fmt.Println(sentence)
	retsentence := make([]string, 0)
	vowels := [...]string{
		"a",
		"e",
		"i",
		"o",
		"u",
	}
	ignoredih := [...]string{
		"ld",
		"re",
		"as",
		"ed",
	}
	for _, i := range sentence {
		t1 := make([]string, 0)
		for _, j := range i {
			fmt.Println(isalnum(string(j)))
			if isalnum(string(j)) {
				t1 = append(t1, string(j))
			}
		}
		if len(t1) < 2 {
			retsentence = append(retsentence, strings.Join(t1, " "))
			continue
		}
		if strings.ToLower(strings.Join(t1[len(t1)-2:], "")) == "ck" || strings.ToLower(strings.Join(t1[len(t1)-2:], "")) == "sh" {
			// tt1 := []rune(t1)
			t1 = append(t1[:len(t1)-2], "h")
			// tt1.pop(-2)
			// tt1.pop(-1)

		} else {
			if slices.Contains(vowels[:], t1[len(t1)-2]) {
				if strings.ToLower(t1[len(t1)-1]) != "s" || strings.ToLower(t1[len(t1)-1]) != "t" || strings.ToLower(t1[len(t1)-2]) != "o" {
					t1[len(t1)-1] = "h"
				}
			} else if slices.Contains(vowels[:], t1[len(t1)-1]) {
				t1 = append(t1, "h")
			} else {
				if !(slices.Contains(ignoredih[:], strings.ToLower(strings.Join(t1[len(t1)-2:], "")))) {
					t1 = append(t1, "ih")
				}
			}
			fmt.Printf("t1 %v\n", t1)
			retsentence = append(retsentence, strings.Join(t1, ""))
			fmt.Println(retsentence)
		}
	}
	return strings.Join(retsentence, " ") // [/]
}
func getpfp(s *discordgo.Session, m *discordgo.MessageCreate) {
	sentence := strings.Split(m.Content[7:], " ")
	doneleast := false
	for _, i := range sentence {
		re := regexp.MustCompile(`[^0-9]+`)
		if len(re.ReplaceAllString(i, "")) > 5 {
			user, err := s.User(re.ReplaceAllString(i, ""))
			if err == nil {
				doneleast = true
				s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
					Content: user.AvatarURL(fmt.Sprintf("%f", math.Pow(2, 10))),
					Reference: &discordgo.MessageReference{
						MessageID: m.ID,
						ChannelID: m.ChannelID,
						GuildID:   m.GuildID,
					},
				})
			}
		}
	}
	if !doneleast {
		s.ChannelMessageSend(m.ChannelID, "Could not get profile picture.")
	}
}
func getemoji(s *discordgo.Session, m *discordgo.MessageCreate) {
	sentence := strings.Split(m.Content[9:], " ")
	doneleast := false
	for _, i := range sentence {
		re := regexp.MustCompile(`[^0-9]+`)
		emojiID := re.ReplaceAllString(i, "")
		if len(re.ReplaceAllString(i, "")) > 5 {
			formats := []string{"webp", "png", "gif"}

			sendingURL := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", emojiID)
			// ^ Just a fallback
			for _, ext := range formats {
				url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s", emojiID, ext)

				resp, err := http.Head(url)
				if err == nil && resp.StatusCode == http.StatusOK {
					doneleast = true
					sendingURL = url
					break
				} else if ext == "webp" || ext == "gif" {
					url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s?animated=true", emojiID, ext)
					resp1, err1 := http.Head(url)
					if err1 == nil && resp1.StatusCode == http.StatusOK {
						doneleast = true
						sendingURL = url
					}
					break
				}
			}
			s.ChannelMessageSendComplex(m.ChannelID,
				&discordgo.MessageSend{
					Content: sendingURL,
					Reference: &discordgo.MessageReference{
						MessageID: m.ID,
						ChannelID: m.ChannelID,
						GuildID:   m.GuildID,
					},
				})
		}
	}
	if !doneleast {
		s.ChannelMessageSend(m.ChannelID, "Could not get emoji.")
	}
}
func getsticker(s *discordgo.Session, m *discordgo.MessageCreate) {
	doneleast := false
	if len(m.ReferencedMessage.StickerItems) > 0 {
		doneleast = true
		id := m.ReferencedMessage.StickerItems[0].ID
		formats := []string{"webp", "png", "gif"}

		sendingURL := fmt.Sprintf("https://cdn.discordapp.com/stickers/%s.png", id)
		// ^ Just a fallback
		for _, ext := range formats {
			url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s", id, ext)

			resp, err := http.Head(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				doneleast = true
				sendingURL = url
				break
			} else if ext == "webp" || ext == "gif" {
				url := fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s?animated=true", id, ext)
				resp1, err1 := http.Head(url)
				if err1 == nil && resp1.StatusCode == http.StatusOK {
					doneleast = true
					sendingURL = url
					break
				}

			}
		}
		s.ChannelMessageSendComplex(m.ChannelID,
			&discordgo.MessageSend{
				Content: sendingURL,
				Reference: &discordgo.MessageReference{
					MessageID: m.ID,
					ChannelID: m.ChannelID,
					GuildID:   m.GuildID,
				},
			})
	}

	if !doneleast {
		s.ChannelMessageSend(m.ChannelID, "Could not get sticker.")
	}
}
func randomframe(s *discordgo.Session, m *discordgo.MessageCreate) {
	urls := []string{
		"https://www.youtube.com/@keewidraws/videos",
		"https://www.youtube.com/@KeewiExtras/videos",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ydl := ytdlp.New().FlatPlaylist().DumpSingleJSON().Quiet().Simulate()

	res1, err1 := ydl.Run(ctx, urls[0])
	if err1 != nil {
		log.Fatal(err1)
	}

	res2, err2 := ydl.Run(ctx, urls[1])
	if err2 != nil {
		log.Fatal(err2)
	}
	var tempPlaylist struct {
		Entries []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"entries"`
	}

	var allEntr []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	json.Unmarshal([]byte(res1.Stdout), &tempPlaylist)
	allEntr = append(allEntr, tempPlaylist.Entries...)

	json.Unmarshal([]byte(res2.Stdout), &tempPlaylist)
	allEntr = append(allEntr, tempPlaylist.Entries...)

	selected := allEntr[rand.Intn(len(allEntr))]

	streamCmd := ytdlp.New().
		DumpSingleJSON().
		Simulate().
		Format("best[height=360][ext=mp4]").
		NoPlaylist().
		NoCheckCertificates().
		NoWarnings()

	res, _ := streamCmd.Run(ctx, "https://youtu.be/"+selected.ID)

	var vidDetails struct {
		Title    string  `json:"title"`
		URL      string  `json:"url"`
		Duration float64 `json:"duration"`
	}
	json.Unmarshal([]byte(res.Stdout), &vidDetails)
	randomSecond := rand.Float64() * vidDetails.Duration
	timestamp := fmt.Sprintf("%f", randomSecond)

	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", timestamp,
		"-i", vidDetails.URL,
		"-vframes", "1",
		"-q:v", "2",
		"-f", "image2",
		"pipe:1")

	frameData, _ := ffmpegCmd.Output()
	fmt.Println(format_seconds(int(vidDetails.Duration)))

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("<https://youtu.be/%s> (%s), approximately %s",
			selected.ID,
			vidDetails.Title,
			format_seconds(int(vidDetails.Duration))),
		Files: []*discordgo.File{
			{
				Name:   "frame.jpg",
				Reader: bytes.NewReader(frameData),
			},
		},
		Reference: &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		},
	})
}
func gamble(s *discordgo.Session, m *discordgo.MessageCreate) {
	cacheMu.RLock()
	if len(videoCache) == 0 {
		cacheMu.RUnlock()
		s.ChannelMessageSend(m.ChannelID, "Cache is empty.")
		return
	}
	localCache := videoCache
	cacheMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	sendingfiles := make([]*discordgo.File, 0)
	mcontent := ""

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Loop until we actually get a frame
			for {
				selected := localCache[rand.Intn(len(localCache))]

				res, err := ytdlp.New().
					DumpSingleJSON().
					Format("best[height=360][ext=mp4]").
					NoPlaylist().
					Run(ctx, "https://youtu.be/"+selected.ID)

				if err != nil {
					continue // Try a different video
				}

				var info struct {
					Title    string  `json:"title"`
					URL      string  `json:"url"`
					Duration float64 `json:"duration"`
				}
				json.Unmarshal([]byte(res.Stdout), &info)

				if info.URL == "" {
					continue // Try a different video
				}

				ts := fmt.Sprintf("%f", rand.Float64()*info.Duration)
				ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", "-ss", ts, "-i", info.URL, "-vframes", "1", "-f", "image2", "pipe:1")

				frame, err := ffmpegCmd.Output()
				if err != nil {
					continue // Try a different video
				}

				mu.Lock()
				sendingfiles = append(sendingfiles, &discordgo.File{
					Name:   uuid.NewString() + ".jpg",
					Reader: bytes.NewReader(frame),
				})
				mcontent += fmt.Sprintf("<https://youtu.be/%s> (%s), ~%s\n",
					selected.ID, info.Title, format_seconds(int(info.Duration)))
				mu.Unlock()

				break // Success! Exit the 'for' loop for this goroutine
			}
		}()
	}

	wg.Wait()

	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: mcontent,
		Files:   sendingfiles,
		Reference: &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		},
	})
}
func updateCache() {
	urls := []string{
		"https://www.youtube.com/@keewidraws/videos",
		"https://www.youtube.com/@KeewiExtras/videos",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var tempEntries []Entry
	ydl := ytdlp.New().FlatPlaylist().DumpSingleJSON().Quiet()

	for _, url := range urls {
		res, err := ydl.Run(ctx, url)
		if err != nil {
			continue
		}
		var p struct {
			Entries []Entry `json:"entries"`
		}
		json.Unmarshal([]byte(res.Stdout), &p)
		tempEntries = append(tempEntries, p.Entries...)
	}

	cacheMu.Lock()
	videoCache = tempEntries
	cacheMu.Unlock()
}

func getLatex(s *discordgo.Session, m *discordgo.MessageCreate, latexString string) {
	unique := uuid.NewString()
	tempTex := fmt.Sprintf("temp_%s.tex", unique)
	tempDvi := fmt.Sprintf("temp_%s.dvi", unique)
	tempPng := fmt.Sprintf("temp_%s.png", unique)
	finalPng := fmt.Sprintf("final_%s.png", unique)
	if strings.HasPrefix(m.Content, "!latex") {
		latexString = latexString[6:]
	} else if strings.HasPrefix(m.Content, "!l") {
		latexString = latexString[2:]
	}
	clean := strings.ReplaceAll(latexString, "_{ }", "")
	clean = strings.ReplaceAll(clean, "^{ }", "")
	fmt.Print(1)
	fullDoc := fmt.Sprintf(`\documentclass[varwidth, border=20pt]{standalone}
\usepackage{amsmath,amsfonts,amssymb}
\begin{document}
\begin{align*}
%s
\end{align*}
\end{document}`, clean)
	fmt.Print(2)
	_ = os.WriteFile(tempTex, []byte(fullDoc), 0644)

	defer func() {
		os.Remove(tempTex)
		os.Remove(tempDvi)
		os.Remove(tempPng)
		os.Remove(finalPng)
		os.Remove(fmt.Sprintf("temp_%s.log", unique))
		os.Remove(fmt.Sprintf("temp_%s.aux", unique))
	}()
	fmt.Print(3)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "latex", "-interaction=nonstopmode", "-jobname="+fmt.Sprintf("temp_%s", unique), tempTex)
	if err := cmd.Run(); err != nil {
		return
	}
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("Latex failed with error: %v\n", err)
	// 	fmt.Printf("Latex Output: %s\n", string(output)) // This is the key!
	// 	return
	// }

	cmd = exec.CommandContext(ctx, "dvipng", "-D", "1000", "-bg", "Transparent", "-o", tempPng, tempDvi)
	if err := cmd.Run(); err != nil {
		return
	}
	fmt.Print(4)
	discordBg := "#313338"
	cmd = exec.CommandContext(ctx, "magick", tempPng,
		"-trim",
		"-fill", "white", "-opaque", "black",
		"-background", discordBg,
		"-alpha", "remove", "-alpha", "off",
		"-bordercolor", discordBg, "-border", "30",
		finalPng)

	if err := cmd.Run(); err != nil {
		return
	}

	data, err := os.ReadFile(finalPng)
	if err != nil {
		return
	}
	fmt.Print(5)
	s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:   "latex.png",
				Reader: bytes.NewReader(data),
			},
		},
		Reference: &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			RepliedUser: false,
		},
	})
}

func format_seconds(seconds int) string {
	if seconds <= 0 {
		return "0s"
	}
	years := seconds / 31536000
	weeks := (seconds % (31536000)) / 604800
	days := (seconds % 604800) / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	parts := make([]string, 0)
	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if weeks > 0 {
		parts = append(parts, fmt.Sprintf("%dw", weeks))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if secs > 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}

	return /* " ".join(parts) */ strings.Join(parts, " ")
}
