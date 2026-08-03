package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Telegram API Types
// ─────────────────────────────────────────────────────────────────────────────

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
}

type UpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Telegram HTTP Helpers
// ─────────────────────────────────────────────────────────────────────────────

func apiGet(endpoint string, params url.Values) (*http.Response, error) {

	u := BaseURL() + endpoint

	if params != nil {
		u += "?" + params.Encode()
	}

	resp, err := HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf(
			"telegram api returned %s: %s",
			resp.Status,
			strings.TrimSpace(string(body)),
		)
	}

	return resp, nil
}

func apiPost(endpoint string, body io.Reader, contentType string) (*http.Response, error) {

	resp, err := HTTPClient.Post(
		BaseURL()+endpoint,
		contentType,
		body,
	)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		b, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf(
			"telegram api returned %s: %s",
			resp.Status,
			strings.TrimSpace(string(b)),
		)
	}

	return resp, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Telegram API
// ─────────────────────────────────────────────────────────────────────────────

func getUpdates(offset int) ([]Update, int, error) {

	params := url.Values{}
	params.Set("timeout", strconv.Itoa(int(Profile.PollTimeout.Seconds())))

	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}

	resp, err := apiGet("/getUpdates", params)
	if err != nil {
		return nil, offset, err
	}
	defer resp.Body.Close()

	var result UpdatesResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, offset, err
	}

	if !result.OK {
		return nil, offset, fmt.Errorf("telegram api returned ok=false")
	}

	highest := offset

	for _, update := range result.Result {
		if update.UpdateID >= highest {
			highest = update.UpdateID + 1
		}
	}

	return result.Result, highest, nil
}

func sendMessage(text string) error {

	chunks := chunkText(text, MaxMessageSize)

	for i, chunk := range chunks {

		label := ""

		if len(chunks) > 1 {
			label = fmt.Sprintf("[%d/%d]\n", i+1, len(chunks))
		}

		params := url.Values{}
		params.Set("chat_id", CHAT_ID)
		params.Set("text", label+chunk)

		resp, err := apiGet("/sendMessage", params)
		if err != nil {
			return err
		}

		resp.Body.Close()
	}

	return nil
}

func deleteMessage(messageID int) error {

	params := url.Values{}
	params.Set("chat_id", CHAT_ID)
	params.Set("message_id", strconv.Itoa(messageID))

	resp, err := apiGet("/deleteMessage", params)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func sendFile(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	defer writer.Close()

	if err := writer.WriteField("chat_id", CHAT_ID); err != nil {
		return err
	}

	part, err := writer.CreateFormFile(
		"document",
		filepath.Base(filePath),
	)
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	resp, err := apiPost(
		"/sendDocument",
		&body,
		writer.FormDataContentType(),
	)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func chunkText(text string, size int) []string {

	runes := []rune(text)

	var chunks []string

	for len(runes) > 0 {

		if len(runes) <= size {
			chunks = append(chunks, string(runes))
			break
		}

		chunks = append(chunks, string(runes[:size]))
		runes = runes[size:]
	}

	return chunks
}
