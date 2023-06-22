package slack

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/slack-go/slack"
)

const (
	slackUsersLimit       = 10_000
	slackChannelsLimit    = 10_000
	maxSlackMessageLength = 3_750
	splitMessageThreshold = maxSlackMessageLength - maxSlackMessageLength/5
)

type slackProvider struct {
	client                   *slack.Client
	title                    string
	defaultChannel           string
	slackIDs                 map[string]struct{}
	slackChannelIDs          map[string]struct{}
	channelSlackID           map[string]string
	emailSlackID             map[string]string
	nameSlackID              map[string]string
	realNameSlackID          map[string]string
	normalRealNameSlackID    map[string]string
	displayNameSlackID       map[string]string
	normalDisplayNameSlackID map[string]string
	lastNameSlackID          map[string]string
}

func newSlackProvider(token, title, defaultChannel string) *slackProvider {
	return &slackProvider{
		client:                   slack.New(token),
		title:                    title,
		defaultChannel:           defaultChannel,
		slackIDs:                 make(map[string]struct{}),
		slackChannelIDs:          make(map[string]struct{}),
		channelSlackID:           make(map[string]string),
		emailSlackID:             make(map[string]string),
		nameSlackID:              make(map[string]string),
		realNameSlackID:          make(map[string]string),
		normalRealNameSlackID:    make(map[string]string),
		displayNameSlackID:       make(map[string]string),
		normalDisplayNameSlackID: make(map[string]string),
		lastNameSlackID:          make(map[string]string),
	}
}

func (s *slackProvider) readUsers(ctx context.Context) error {
	users, err := s.client.GetUsersContext(ctx, slack.GetUsersOptionLimit(slackUsersLimit))
	if err != nil {
		return err
	}
	for _, u := range users {
		if !isValidUser(u) {
			continue
		}
		s.slackIDs[u.ID] = struct{}{}
		if u.Name != "" {
			s.nameSlackID[strings.ToLower(u.Name)] = u.ID
		}
		profile := u.Profile
		if profile.Email != "" {
			s.emailSlackID[strings.ToLower(profile.Email)] = u.ID
		}
		if profile.RealNameNormalized != "" {
			name := strings.ToLower(profile.RealNameNormalized)
			s.realNameSlackID[name] = u.ID
			s.normalRealNameSlackID[normalizeName(name)] = u.ID
		}
		if profile.DisplayNameNormalized != "" {
			name := strings.ToLower(profile.DisplayNameNormalized)
			s.displayNameSlackID[name] = u.ID
			s.normalDisplayNameSlackID[normalizeName(name)] = u.ID
		}
		if profile.LastName != "" {
			s.lastNameSlackID[strings.ToLower(profile.LastName)] = u.ID
		}
	}
	return nil
}

func isValidUser(u slack.User) bool {
	return !u.Deleted && !u.IsBot && !u.IsRestricted && !u.IsUltraRestricted && !u.IsInvitedUser
}

func (s *slackProvider) readChannels(ctx context.Context) error {
	var (
		channels []slack.Channel
		cursor   string
		err      error
	)
	for {
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Limit:           slackChannelsLimit,
			Types:           []string{"public_channel"},
		}
		channels, cursor, err = s.client.GetConversationsContext(ctx, params)
		if err != nil {
			return err
		}
		for _, c := range channels {
			if !c.IsChannel {
				continue
			}
			s.slackChannelIDs[c.Conversation.ID] = struct{}{}
			s.channelSlackID[strings.ToLower(c.Conversation.NameNormalized)] = c.Conversation.ID
		}
		if cursor == "" {
			break
		}
	}
	return nil
}

func (s *slackProvider) getSlackIDByOwner(owner string) string {
	if _, ok := s.slackIDs[strings.ToUpper(owner)]; ok {
		return strings.ToUpper(owner)
	}
	if _, ok := s.slackChannelIDs[strings.ToUpper(owner)]; ok {
		return strings.ToUpper(owner)
	}
	owner = strings.ToLower(owner)
	if _, err := mail.ParseAddress(owner); err == nil {
		return s.emailSlackID[owner]
	}
	if id, ok := s.channelSlackID[owner]; ok {
		return id
	}
	if id, ok := s.displayNameSlackID[owner]; ok {
		return id
	}
	if id, ok := s.normalDisplayNameSlackID[owner]; ok {
		return id
	}
	if id, ok := s.lastNameSlackID[owner]; ok {
		return id
	}
	if id, ok := s.realNameSlackID[owner]; ok {
		return id
	}
	if id, ok := s.normalRealNameSlackID[owner]; ok {
		return id
	}
	if id, ok := s.nameSlackID[owner]; ok {
		return id
	}
	return s.defaultChannel
}

func (s *slackProvider) notify(ctx context.Context, channelMessages map[string][]string) error {
	for channel, messages := range channelMessages {
		if s.title != "" {
			_, _, _, err := s.client.SendMessageContext(ctx, channel, slack.MsgOptionText(s.title, false))
			if err != nil {
				return err
			}
		}
		for _, message := range messages {
			_, _, _, err := s.client.SendMessageContext(ctx, channel, slack.MsgOptionText(message, false))
			if err != nil {
				return err
			}
			time.Sleep(time.Millisecond * 150) // pause to avoid slack rate limits
		}
	}
	return nil
}

func splitMessage(message string) []string {
	if utf8.RuneCountInString(message) > maxSlackMessageLength {
		messages := make([]string, 0)
		builder := strings.Builder{}
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			builder.WriteString(line + "\n")
			if utf8.RuneCountInString(builder.String()) > splitMessageThreshold {
				messages = append(messages, builder.String())
				builder.Reset()
			}
		}
		if builder.Len() > 0 {
			messages = append(messages, builder.String())
		}
		return messages
	}
	return []string{message}
}

func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	return strings.ReplaceAll(name, ".", "-")
}
