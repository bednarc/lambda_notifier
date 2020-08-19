package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Records []struct {
		SNS struct {
			Type       string `json:"Type"`
			Timestamp  string `json:"Timestamp"`
			SNSMessage string `json:"Message"`
		} `json:"Sns"`
	} `json:"Records"`
}

type SNSMessage struct {
	AlarmName      string `json:"AlarmName"`
	NewStateValue  string `json:"NewStateValue"`
	NewStateReason string `json:"NewStateReason"`
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Title string `json:"title"`
}

func handler(request Request) {
	var snsMessage SNSMessage
	err := json.Unmarshal([]byte(request.Records[0].SNS.SNSMessage), &snsMessage)
	if err != nil {
		fmt.Println("Unmarshal error: ", request.Records[0].SNS.SNSMessage)
	}

	slackMessage := createSlackMessage(snsMessage.NewStateValue,
		snsMessage.AlarmName,
		snsMessage.NewStateReason)

	err = postToSlack(slackMessage)

	if err != nil {
		fmt.Println("Sending failed. ", err.Error())
	}

	return
}

func createSlackMessage(state, alarmName, newStateReason string) SlackMessage {

	color := "warning"

	if state == "OK" {
		color = "good"
	} else if state == "ALARM" {
		color = "danger"
	}

	return SlackMessage{
		Text: alarmName,
		Attachments: []Attachment{
			Attachment{
				Title: "Elastic Beanstalk notification",
				Text:  "`" + newStateReason + "`",
				Color: color,
			},
		},
	}
}

func postToSlack(message SlackMessage) error {
	client := &http.Client{}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", os.Getenv("WEBHOOK_URL"), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
