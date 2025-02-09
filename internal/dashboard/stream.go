// SPDX-FileCopyrightText: 2025 Paulo Almeida <almeidapaulopt@gmail.com>
// SPDX-License-Identifier: MIT

package dashboard

import (
	"net/http"

	"github.com/a-h/templ"
	datastar "github.com/starfederation/datastar/sdk/go"
)

const (
	SSEQueueSize = 5

	EventAppend EventType = iota
	EventRemove
	EventMerge
	EventMergeMessage
)

// sseClient represents an SSE connection
type (
	EventType int
	sseClient struct {
		channel  chan SSEMessage
		username string
	}

	SSEMessage struct {
		Type    EventType
		Comp    templ.Component
		Message string
	}
)

func (dash *Dashboard) broadcastMessage(message SSEMessage) {
	dash.mtx.RLock()
	defer dash.mtx.RUnlock()
	for sessionID, sseClient := range dash.sseClients {
		select {
		case sseClient.channel <- message:
		default:
			dash.removeSSEClient(sessionID)
		}
	}
}

// Handler for the `/stream` endpoint
func (dash *Dashboard) streamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get("X-Session-ID")

		sse := datastar.NewSSE(w, r)

		// Create a new client
		client := &sseClient{
			channel:  make(chan SSEMessage, SSEQueueSize),
			username: "",
		}

		// Register client
		dash.mtx.Lock()
		dash.sseClients[sessionID] = client
		dash.mtx.Unlock()

		dash.Log.Info().Msg("New Client connected")
		// Ensure client is removed when disconnected
		defer dash.removeSSEClient(sessionID)

		go dash.renderList(client.channel)

		var err error
		// Send messages to the client
	LOOP:
		for {
			select {
			case <-r.Context().Done():
				break LOOP
			case message := <-client.channel:
				switch message.Type {
				case EventAppend:
					err = sse.MergeFragmentTempl(
						message.Comp,
						datastar.WithMergeMode(datastar.FragmentMergeModeAppend),
						datastar.WithSelectorID("proxy-list"),
						datastar.WithViewTransitions(),
					)
				case EventMergeMessage:
					err = sse.MergeFragments(message.Message,
						datastar.WithViewTransitions(),
					)
				}
			}

			if err != nil {
				break LOOP
			}

			dash.updateUser(r)
		}
	}
}

func (dash *Dashboard) updateUser(r *http.Request) {
}

func (dash *Dashboard) removeSSEClient(name string) {
	dash.mtx.Lock()
	client := dash.sseClients[name]
	delete(dash.sseClients, name)
	dash.mtx.Unlock()
	close(client.channel)

	dash.Log.Info().Msg("Client disconnected")
}
