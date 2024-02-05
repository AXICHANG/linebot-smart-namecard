```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/notionhq/client-go/v2/notion"
)

func main() {
	// 取得環境變數
	channelSecret := os.Getenv("CHANNEL_SECRET")
	channelAccessToken := os.Getenv("CHANNEL_ACCESS_TOKEN")
	notionIntegrationToken := os.Getenv("NOTION_INTEGRATION_TOKEN")
	notionDatabaseID := os.Getenv("NOTION_DATABASE_ID")

	// 初始化 LINE Bot 客戶端
	bot, err := linebot.New(channelSecret, channelAccessToken)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化 Notion 客戶端
	notionClient := notion.NewClient(notion.Token(notionIntegrationToken))

	// 設定 LINE Bot Callback 路由
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					// 取得使用者傳送的文字訊息
					text := message.Text

					// 將訊息儲存到 Notion Database
					err := saveToNotionDatabase(text, notionDatabaseID, notionClient)
					if err != nil {
						log.Println("Error saving to Notion Database:", err)
					} else {
						log.Println("Message saved to Notion Database:", text)
					}
				}
			}
		}
	})

	// 啟動 HTTP 伺服器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// 將訊息儲存到 Notion Database
func saveToNotionDatabase(text, databaseID string, client *notion.Client) error {
	// 建立一個新的頁面
	page := &notion.Page{
		Parent: notion.DatabaseParent{
			DatabaseID: databaseID,
		},
		Properties: notion.Properties{
			"Text": notion.RichTexts{
				{
					Text: &notion.Text{
						Content: text,
					},
				},
			},
		},
	}

	// 建立新頁面的請求
	req := &notion.CreatePageRequest{
		Parent: &notion.Parent{
			Type: notion.ParentTypeDatabaseID,
		},
		Properties: &notion.Properties{
			"Text": &notion.Property{
				Title: []notion.RichText{
					{
						Text: &notion.Text{
							Content: text,
						},
					},
				},
			},
		},
	}

	// 呼叫 Notion API 建立新頁面
	_, err := client.Pages.CreatePage(req)
	if err != nil {
		return err
	}

	return nil
}

