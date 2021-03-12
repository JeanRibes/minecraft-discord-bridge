package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"syscall"
)

var chatWebsocket func(s string)
var sess *discordgo.Session

//var computercraftCommand discordgo.ApplicationCommand
var chatCommand discordgo.ApplicationCommand = discordgo.ApplicationCommand{
	Name:        "mc",
	Description: "Envoie un message dans le chat Minecraft",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Message à envoyer",
			Required:    true,
		},
	},
}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

var chatChannelId string
var guildId  string

/*
Protocole: on reçoit le channelId du Computer
on répond avec OK ou FAIL
on forward les messages suivants dans Discord; et inversement
*/
func httpServ() {
	http.HandleFunc("/chat", WSChatServer)
	err := http.ListenAndServe(getenv("LISTEN_ADDR",":4042"), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
func getenv(key,default_value string) string {
    s:=os.Getenv(key)
    if s!="" {
	return s
    } else {
	return default_value
    }
}
func main() {
	var err error
	sess, err = discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	chatChannelId = getenv("CHAT_CHANNEL_ID","814779001540182016")
	guildId = getenv("GUILD_ID","740937779374194750")
	sess.AddHandler(messageCreate)
	sess.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsDirectMessages | discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessageReactions | discordgo.IntentsGuildMessageReactions)
	err = sess.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	} else {
		fmt.Println("OCbridgeBot tourne !")
	}
	go httpServ()

	sess.AddHandler(slashCommandHandler)

	cccmd, acce := sess.ApplicationCommandCreate(sess.State.User.ID, guildId, &chatCommand)
	he(acce)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	he(sess.ApplicationCommandDelete(sess.State.User.ID, "740937779374194750", cccmd.ID))

	he(sess.Close())
}
func he(err error) {
	if err != nil {
		panic(err)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content[0] != '!' && m.Content[0] != '/' && m.Content[0] != '.' && m.Content[0] != '>' && m.Content[0] != '\\' && m.Content[0] != '-' && m.Content[0] != ':' {
		if m.ChannelID == chatChannelId {
			println("got chat message, forwarding")
			if chatWebsocket != nil {
				chatWebsocket(fmt.Sprintf("Discord: <%s> %s", m.Author.Username, m.Content))
			}
		}
	}
}
func WSChatServer(w http.ResponseWriter, r *http.Request) { // un par WS
	ws, uperr := upgrader.Upgrade(w, r, nil)
	ghe(uperr)

	fmt.Printf("new client: %s\n", ws.RemoteAddr())

	sender := func(s string) {
		he(ws.WriteMessage(websocket.TextMessage,[]byte(s)))
	}
	chatWebsocket = sender
	for {
		_,msg,re := ws.ReadMessage()
		ghe(re)
		_, se := sess.ChannelMessageSend(chatChannelId, string(msg))
		ghe(se)
	}
}
func slashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Data.Name == "mc" {
		opt := i.Data.Options[0]
		fmt.Printf("mc: name=%s value=%v\n", opt.Name, opt.Value)
		/*he(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseAcknowledge,
		}))*/
		ghe(s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				TTS:             false,
				Content:         fmt.Sprintf("vous avez envoyé '%s'", opt.Value),
				Embeds:          nil,
				AllowedMentions: nil,
			},
		}))
		ghe(s.InteractionResponseDelete(sess.State.User.ID, i.Interaction)) //vire la réponse du bot dans le chat, en
		// laissant l'historique de l'utilisateur
	}

}

func ghe(err error) {
	if err != nil {
		println(err.Error())
	}
}

