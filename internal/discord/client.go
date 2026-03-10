package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		baseURL: "https://discord.com/api/v10",
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string) ([]byte, error) {
	const maxRetries = 6

	for attempt := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", c.token)
		req.Header.Set("User-Agent", "dissync/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Rate limit: respect advisory headers
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			if resetAfter := resp.Header.Get("X-RateLimit-Reset-After"); resetAfter != "" {
				if secs, err := strconv.ParseFloat(resetAfter, 64); err == nil {
					wait := time.Duration(secs*1000)*time.Millisecond + time.Second
					if wait > 60*time.Second {
						wait = 60 * time.Second
					}
					time.Sleep(wait)
				}
			}
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return body, nil
		case http.StatusNoContent:
			return nil, nil
		case http.StatusTooManyRequests:
			wait := 5 * time.Second
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if secs, err := strconv.ParseFloat(retryAfter, 64); err == nil {
					wait = time.Duration(secs*1000)*time.Millisecond + time.Second
				}
			}
			if wait > 60*time.Second {
				wait = 60 * time.Second
			}
			time.Sleep(wait)
			continue
		case http.StatusForbidden:
			return nil, fmt.Errorf("%w: %s", ErrForbidden, path)
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("unauthorized: check your DISCORD_TOKEN")
		default:
			if resp.StatusCode >= 500 && attempt < maxRetries-1 {
				time.Sleep(backoff(attempt))
				continue
			}
			return nil, fmt.Errorf("discord API error %d: %s", resp.StatusCode, string(body))
		}
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func backoff(attempt int) time.Duration {
	secs := math.Pow(2, float64(attempt))
	if secs > 32 {
		secs = 32
	}
	return time.Duration(secs) * time.Second
}

var (
	ErrForbidden = fmt.Errorf("forbidden")
	ErrNotFound  = fmt.Errorf("not found")
)

func (c *Client) GetGuilds(ctx context.Context) ([]Guild, error) {
	var all []Guild
	after := "0"

	for {
		body, err := c.do(ctx, "GET", "/users/@me/guilds?limit=200&after="+after)
		if err != nil {
			return nil, fmt.Errorf("get guilds: %w", err)
		}

		var page []Guild
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("parse guilds: %w", err)
		}
		if len(page) == 0 {
			break
		}

		all = append(all, page...)
		after = page[len(page)-1].ID
		if len(page) < 200 {
			break
		}
	}
	return all, nil
}

func (c *Client) GetGuildChannels(ctx context.Context, guildID string) ([]Channel, error) {
	body, err := c.do(ctx, "GET", "/guilds/"+guildID+"/channels")
	if err != nil {
		return nil, fmt.Errorf("get channels: %w", err)
	}

	var channels []Channel
	if err := json.Unmarshal(body, &channels); err != nil {
		return nil, fmt.Errorf("parse channels: %w", err)
	}
	// Set guild ID on each channel (not always present in API response).
	for i := range channels {
		channels[i].GuildID = guildID
	}
	return channels, nil
}

func (c *Client) GetMessagesAfter(ctx context.Context, channelID, afterID string, limit int) ([]Message, error) {
	path := fmt.Sprintf("/channels/%s/messages?limit=%d", channelID, limit)
	if afterID != "" && afterID != "0" {
		path += "&after=" + afterID
	}

	body, err := c.do(ctx, "GET", path)
	if err != nil {
		return nil, err
	}

	var msgs []Message
	if err := json.Unmarshal(body, &msgs); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}
	// Discord returns newest first; reverse to oldest first.
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (c *Client) GetMessagesBefore(ctx context.Context, channelID, beforeID string, limit int) ([]Message, error) {
	path := fmt.Sprintf("/channels/%s/messages?limit=%d&before=%s", channelID, limit, beforeID)

	body, err := c.do(ctx, "GET", path)
	if err != nil {
		return nil, err
	}

	var msgs []Message
	if err := json.Unmarshal(body, &msgs); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}
	// Discord returns newest first; reverse to oldest first.
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (c *Client) SearchThreads(ctx context.Context, channelID string, archived bool) ([]Channel, error) {
	var all []Channel
	var beforeID string

	for {
		path := fmt.Sprintf("/channels/%s/threads/search?archived=%t&sort_by=last_message_time&sort_order=desc&limit=25",
			channelID, archived)
		if beforeID != "" {
			path += "&before=" + beforeID
		}

		body, err := c.do(ctx, "GET", path)
		if err != nil {
			// Threads endpoint may not be available on all channel types.
			if strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "not found") {
				return all, nil
			}
			return nil, fmt.Errorf("search threads: %w", err)
		}

		var resp ThreadSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse threads: %w", err)
		}

		all = append(all, resp.Threads...)

		if !resp.HasMore || len(resp.Threads) == 0 {
			break
		}
		beforeID = resp.FirstID
	}
	return all, nil
}
