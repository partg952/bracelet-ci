package util

import (
	"slices"
	"sync"
)

type ChannelBroker struct {
	m          sync.Mutex
	ChannelMap map[string][]chan any
}

const channelBufferSize = 10

func NewHub() *ChannelBroker {
	return &ChannelBroker{
		ChannelMap: make(map[string][]chan any),
	}
}

func (h *ChannelBroker) Subscribe(streamKey string) (chan any, func()) {
	channel := make(chan any, channelBufferSize)

	h.m.Lock()
	h.ChannelMap[streamKey] = append(h.ChannelMap[streamKey], channel)
	h.m.Unlock()

	unsubscribe := func() {
		h.m.Lock()
		defer h.m.Unlock()
		for i, element := range h.ChannelMap[streamKey] {
			if element == channel {
				h.ChannelMap[streamKey] = slices.Delete(h.ChannelMap[streamKey], i, i+1)
				break
			}
		}

		if len(h.ChannelMap[streamKey]) == 0 {
			delete(h.ChannelMap, streamKey)
		}

		close(channel)

	}
	return channel, unsubscribe

}

func (h *ChannelBroker) Publish(streamKey string, message any) int {
	if streamKey == "" {
		return 0
	}

	h.m.Lock()
	defer h.m.Unlock()

	delivered := 0
	for _, channel := range h.ChannelMap[streamKey] {
		select {
		case channel <- message:
			delivered++
		default:
		}
	}

	return delivered
}
