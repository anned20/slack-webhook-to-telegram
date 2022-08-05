package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type PayloadAttachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
	Text      string `json:"text"`
	Color     string `json:"color"`
}

type IncomingPayload struct {
	Text        string              `json:"text"`
	Attachments []PayloadAttachment `json:"attachments"`
}

type TelegramMessage struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the HTTP server that listen to the webhook request",
	Run: func(cmd *cobra.Command, args []string) {
		host := viper.GetString("host")
		port := viper.GetInt("port")

		fmt.Printf("Serving on %s:%d\n", host, port)

		http.HandleFunc("/", handleWebhook)

		err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
		if err != nil {
			log.Fatalf("Error starting the HTTP Server: %s", err)

			return
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func handleWebhook(writer http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var payload IncomingPayload

	err := decoder.Decode(&payload)

	if err != nil {
		log.Printf("Error decoding the request body: %s", err)

		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	message := formatMessage(payload)

	err = sendTelegramMessage(message)

	if err != nil {
		log.Printf("Error sending the message to Telegram: %s", err)

		writer.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func formatMessage(payload IncomingPayload) string {
	var message string

	message += fmt.Sprintf("%s\n\n", payload.Text)

	for _, attachment := range payload.Attachments {
		message += fmt.Sprintf("[*%s*](%s)\n", attachment.Title, attachment.TitleLink)
		message += fmt.Sprintf("%s\n", attachment.Text)
		message += "\n"
	}

	toReplace := map[string]string{
		"_": "\\_",
		"-": "\\-",
		"~": "\\~",
		"`": "\\`",
		".": "\\.",
		"(": "\\(",
		")": "\\)",
	}

	for key, value := range toReplace {
		message = strings.Replace(message, key, value, -1)
	}

	return message
}

func sendTelegramMessage(message string) error {
	telegramMessage := TelegramMessage{
		ChatID:    viper.GetInt64("telegram_chat_id"),
		Text:      message,
		ParseMode: "MarkdownV2",
	}

	b := new(bytes.Buffer)

	err := json.NewEncoder(b).Encode(telegramMessage)

	if err != nil {
		return err
	}

	log.Printf("Sending message to Telegram: %s", b.String())

	request, err := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", viper.GetString("telegram_bot_token")), b)

	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return err
	}

	log.Printf("Telegram response [%s]: %s", response.Status, responseBody)

	return nil
}
