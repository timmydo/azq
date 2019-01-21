package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-storage-queue-go/azqueue"
)

var (
	queue        = flag.String("queue", "", "queue name")
	downloadFile = flag.String("download-file", "", "file to download queue items to")
	uploadFile   = flag.String("upload-file", "", "file to upload queue items from")
)

func accountInfo() (string, string) {
	return os.Getenv("ACCOUNT_NAME"), os.Getenv("ACCOUNT_KEY")
}

func main() {
	flag.Parse()

	if *queue == "" {
		log.Fatalln("No queue name provided")
	}

	accountName, accountKey := accountInfo()

	if accountName == "" {
		log.Fatalln("account name (environment variable ACCOUNT_NAME) is empty")
	}

	if accountKey == "" {
		log.Fatalln("account key (environment variable ACCOUNT_KEY) is empty")
	}

	credential, err := azqueue.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal(err)
	}

	p := azqueue.NewPipeline(credential, azqueue.PipelineOptions{})

	u, _ := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net", accountName))

	serviceURL := azqueue.NewServiceURL(*u, p)
	queueURL := serviceURL.NewQueueURL(*queue)
	messagesURL := queueURL.NewMessagesURL()
	ctx := context.TODO()

	if *downloadFile != "" {
		file, err := os.Create(*downloadFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		count := int32(0)
		for {
			dequeue, err := messagesURL.Dequeue(ctx, azqueue.QueueMaxMessagesDequeue, 10*time.Second)
			if err != nil {
				log.Fatal(err)
			}
			if dequeue.NumMessages() == 0 {
				log.Printf("Downloaded %d messages...", count)
				break
			} else {
				for m := int32(0); m < dequeue.NumMessages(); m++ {
					msg := dequeue.Message(m)
					file.WriteString(msg.Text)
					file.WriteString("\n")
					file.Sync()
					msgIDURL := messagesURL.NewMessageIDURL(msg.ID)
					msgIDURL.Delete(ctx, msg.PopReceipt)
					count++
				}
			}
		}
	}

	if *uploadFile != "" {
		file, err := os.Open(*uploadFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			_, err = messagesURL.Enqueue(ctx, scanner.Text(), time.Second*0, time.Minute)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

}
